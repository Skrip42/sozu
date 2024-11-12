package internal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
	"go.uber.org/mock/gomock"
)

type LimitedFabricSuite struct {
	suite.Suite

	beforeSendCounter  int
	afterSendCounter   int
	beforeFlushCounter int
	afterFlushCounter  int

	beforeSend  func(int, func())
	afterSend   func(int, func())
	beforeFlush func()
	afterFlush  func(int)

	base *MockFabric[int]

	fabric Fabric[int]
}

func (s *LimitedFabricSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())

	s.beforeSendCounter = 0
	s.afterSendCounter = 0
	s.beforeFlushCounter = 0
	s.afterFlushCounter = 0

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
	}

	s.base = NewMockFabric[int](ctrl)

	s.fabric = NewLimitedFabric(s.base, 3)
}

func (s *LimitedFabricSuite) TearDownTest() {
	goleak.VerifyNone(s.T())
}

func TestLimitedFabric(t *testing.T) {
	suite.Run(t, &LimitedFabricSuite{})
}

func (s *LimitedFabricSuite) TestOk() {
	flushCounter := 0
	flush := func() {
		flushCounter++
	}

	ctx, cancel := context.WithCancel(context.Background())
	inputCh := make(chan int)
	flushCh := make(chan func())
	capacity := 10

	outputCh := make(chan []int)
	var sendF func(int, func())
	var flushF func(int)

	s.base.EXPECT().Create(
		ctx,
		inputCh,
		flushCh,
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		3,
		gomock.Any(),
	).DoAndReturn(
		func(
			_ context.Context,
			_ chan int,
			_ chan func(),
			beforeSend func(int, func()),
			afterSend func(int, func()),
			beforeFlush func(),
			afterFlush func(int),
			_ int,
			_ context.CancelFunc,
		) <-chan []int {
			sendF = afterSend
			flushF = afterFlush
			// Check is function original
			beforeSend(1, func() {})
			beforeFlush()

			s.Equal(1, s.beforeSendCounter)
			s.Equal(1, s.beforeFlushCounter)

			return outputCh
		},
	)

	output := s.fabric.Create(
		ctx,
		inputCh,
		flushCh,
		s.beforeSend,
		s.afterSend,
		s.beforeFlush,
		s.afterFlush,
		capacity,
		cancel,
	)

	// check correct output
	go func() {
		outputCh <- []int{1, 2, 3}
		outputCh <- []int{0}
	}()
	s.Equal([]int{1, 2, 3}, <-output)
	s.Equal([]int{0}, <-output)

	// check logick
	sendF(1, flush)
	s.Equal(1, s.afterSendCounter)
	s.Equal(0, flushCounter)
	sendF(2, flush)
	s.Equal(2, s.afterSendCounter)
	s.Equal(0, flushCounter)
	sendF(3, flush)
	s.Equal(3, s.afterSendCounter)
	s.Equal(1, flushCounter)
	flushF(3)
	s.Equal(1, s.afterFlushCounter)

	sendF(1, flush)
	s.Equal(4, s.afterSendCounter)
	s.Equal(1, flushCounter)
	sendF(1, flush)
	s.Equal(5, s.afterSendCounter)
	s.Equal(1, flushCounter)
	flushF(1)
	s.Equal(2, s.afterFlushCounter)
	s.Equal(1, flushCounter)
	sendF(1, flush)
	s.Equal(6, s.afterSendCounter)
	s.Equal(1, flushCounter)
	sendF(2, flush)
	s.Equal(7, s.afterSendCounter)
	s.Equal(2, flushCounter)
}
