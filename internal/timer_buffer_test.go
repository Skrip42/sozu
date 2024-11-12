package internal

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
	"go.uber.org/mock/gomock"
)

type TimerFabricSuite struct {
	suite.Suite

	beforeSendCounter  int
	afterSendCounter   int
	beforeFlushCounter int
	afterFlushCounter  int

	beforeSend  func(int, func())
	afterSend   func(int, func())
	beforeFlush func()
	afterFlush  func(int)

	base     *MockFabric[int]
	criteria func(int) bool

	fabric Fabric[int]
}

func (s *TimerFabricSuite) SetupTest() {
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
	s.criteria = func(i int) bool {
		return i%2 == 0
	}

	s.fabric = NewTimerFabric(s.base, time.Millisecond)
}

func (s *TimerFabricSuite) TearDownTest() {
	goleak.VerifyNone(s.T())
}

func TestTimerFabric(t *testing.T) {
	suite.Run(t, &TimerFabricSuite{})
}

func (s *TimerFabricSuite) TestOk() {
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
		10,
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
	<-flushCh
	s.Equal(1, s.afterSendCounter)

	flushF(0)
	s.Equal(1, s.afterFlushCounter)

	cancel()
}
