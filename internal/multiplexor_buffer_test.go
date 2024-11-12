package internal

import (
	"context"
	"testing"

	"github.com/Skrip42/sozu/internal/helper"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
	gomock "go.uber.org/mock/gomock"
)

type MultiplexorFactorySuite struct {
	suite.Suite

	beforeSendCounter  int
	afterSendCounter   int
	beforeFlushCounter int
	afterFlushCounter  int

	afterFlushValues []int

	beforeSend  func(int, func())
	afterSend   func(int, func())
	beforeFlush func()
	afterFlush  func(int)

	base      *MockFactory[int]
	separator func(int) int

	fabric Factory[int]
}

func (s *MultiplexorFactorySuite) SetupTest() {
	ctrl := gomock.NewController(s.T())

	s.beforeSendCounter = 0
	s.afterSendCounter = 0
	s.beforeFlushCounter = 0
	s.afterFlushCounter = 0

	s.afterFlushValues = []int{}

	s.beforeSend = func(value int, flush func()) {
		s.beforeSendCounter++
	}
	s.afterSend = func(value int, flush func()) {
		s.afterSendCounter++
	}
	s.beforeFlush = func() {
		s.beforeFlushCounter++
	}
	s.afterFlush = func(count int) {
		s.afterFlushCounter++
		s.afterFlushValues = append(s.afterFlushValues, count)
	}

	s.base = NewMockFactory[int](ctrl)
	s.separator = func(i int) int {
		return i % 3
	}

	s.fabric = NewMultiplexorFactory(s.base, s.separator, 3)
}

func (s *MultiplexorFactorySuite) TearDownTest() {
	goleak.VerifyNone(s.T())
}

func TestMultiplexorFactory(t *testing.T) {
	suite.Run(t, &MultiplexorFactorySuite{})
}

func (s *MultiplexorFactorySuite) TestOk() {
	ctx, cancel := context.WithCancel(context.Background())

	outputCh := []chan []int{
		make(chan []int),
		make(chan []int),
		make(chan []int),
	}

	inputValues := [][]int{
		[]int{},
		[]int{},
		[]int{},
	}

	flushCounter := 0

	i := 0

	s.base.EXPECT().Create(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		10,
		gomock.Any(),
	).DoAndReturn(
		func(
			_ context.Context,
			iCh chan int,
			fCh chan func(),
			_ func(int, func()),
			_ func(int, func()),
			_ func(),
			afterFlush func(int),
			_ int,
			_ context.CancelFunc,
		) <-chan []int {
			it := i
			i++
			go func() {
				for {
					select {
					case v := <-iCh:
						inputValues[it] = append(inputValues[it], v)
					case <-ctx.Done():
						return
					}
				}
			}()

			go func() {
				for {
					select {
					case done := <-fCh:
						flushCounter++
						done()
					case <-ctx.Done():
						return
					}
				}
			}()

			afterFlush(it)
			return outputCh[it]
		},
	).Times(3)

	inputCh := make(chan int)
	flushCh := make(chan func())

	output := s.fabric.Create(
		ctx,
		inputCh,
		flushCh,
		s.beforeSend,
		s.afterSend,
		s.beforeFlush,
		s.afterFlush,
		10,
		cancel,
	)

	inputCh <- 0
	inputCh <- 1
	inputCh <- 2

	await, done := helper.NewSyncer(ctx)
	flushCh <- done
	await()

	s.Equal(3, flushCounter)
	s.Equal(3, s.beforeSendCounter)
	s.Equal(3, s.afterSendCounter)
	s.Equal(
		[][]int{
			[]int{0},
			[]int{1},
			[]int{2},
		},
		inputValues,
	)

	outputCh[0] <- []int{0}
	i1 := <-output
	s.Equal([]int{0}, i1)
	outputCh[1] <- []int{1}
	i2 := <-output
	s.Equal([]int{1}, i2)
	outputCh[2] <- []int{2}
	i3 := <-output
	s.Equal([]int{2}, i3)

	s.Equal(3, s.afterFlushCounter)
	s.Equal([]int{0, 1, 2}, s.afterFlushValues)

	close(inputCh)

	cancel()
}
