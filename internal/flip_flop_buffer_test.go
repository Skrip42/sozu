package internal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
	"go.uber.org/mock/gomock"
)

type FlipFlopFabricSuite struct {
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

func (s *FlipFlopFabricSuite) SetupTest() {
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

	s.fabric = NewFlipFlopFabric(s.base, s.criteria)
}

func (s *FlipFlopFabricSuite) TearDownTest() {
	goleak.VerifyNone(s.T())
}

func TestFlipFlopFabric(t *testing.T) {
	suite.Run(t, &FlipFlopFabricSuite{})
}

func (s *FlipFlopFabricSuite) TestOk() {
	flushCounter := 0
	flush := func() {
		flushCounter++
	}

	ctx := context.Background()
	inputCh := make(chan int)
	flushCh := make(chan func())
	capacity := 10

	outputCh := make(chan []int)
	var sendF func(int, func())

	s.base.EXPECT().Create(
		ctx,
		inputCh,
		flushCh,
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		capacity,
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
		) <-chan []int {
			sendF = beforeSend
			// Check is function original
			afterSend(1, func() {})
			beforeFlush()
			afterFlush(1)

			s.Equal(1, s.afterSendCounter)
			s.Equal(1, s.beforeFlushCounter)
			s.Equal(1, s.afterFlushCounter)

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
	sendF(3, flush)
	s.Equal(0, flushCounter)
	s.Equal(2, s.beforeSendCounter)
	sendF(2, flush)
	s.Equal(1, flushCounter)
	s.Equal(3, s.beforeSendCounter)
	sendF(2, flush)
	s.Equal(1, flushCounter)
	s.Equal(4, s.beforeSendCounter)
	sendF(5, flush)
	s.Equal(2, flushCounter)
	s.Equal(5, s.beforeSendCounter)
}
