package internal

import (
	"context"
	"testing"

	"github.com/Skrip42/sozu/internal/helper"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
)

type BufferFactorySuite struct {
	suite.Suite

	beforeSendCounter  int
	afterSendCounter   int
	beforeFlushCounter int
	afterFlushCounter  int

	beforeSendValues []int
	afterSendValues  []int
	afterFlushValues []int

	beforeSend  func(int, func())
	afterSend   func(int, func())
	beforeFlush func()
	afterFlush  func(int)

	fabric Factory[int]
}

func (s *BufferFactorySuite) SetupTest() {
	s.beforeSendCounter = 0
	s.afterSendCounter = 0
	s.beforeFlushCounter = 0
	s.afterFlushCounter = 0

	s.beforeSendValues = []int{}
	s.afterSendValues = []int{}
	s.afterFlushValues = []int{}

	s.beforeSend = func(value int, flush func()) {
		s.beforeSendCounter++
		s.beforeSendValues = append(s.beforeSendValues, value)
	}
	s.afterSend = func(value int, flush func()) {
		s.afterSendCounter++
		s.afterSendValues = append(s.afterSendValues, value)
	}
	s.beforeFlush = func() {
		s.beforeFlushCounter++
	}
	s.afterFlush = func(count int) {
		s.afterFlushCounter++
		s.afterFlushValues = append(s.afterFlushValues, count)
	}

	s.fabric = NewBufferFactory[int]()
}

func (s *BufferFactorySuite) TearDownTest() {
	goleak.VerifyNone(s.T())
}

func TestBufferFactory(t *testing.T) {
	suite.Run(t, &BufferFactorySuite{})
}

func (s *BufferFactorySuite) TestWithAfterSendFlush() {
	ctx, cancel := context.WithCancel(context.Background())

	inputChan := make(chan int)
	flushChan := make(chan func())

	afterSend := func(value int, flush func()) {
		s.afterSend(value, flush)
		flush()
	}

	output := s.fabric.Create(
		ctx,
		inputChan,
		flushChan,
		s.beforeSend,
		afterSend,
		s.beforeFlush,
		s.afterFlush,
		100,
		cancel,
	)

	go func() {
		inputChan <- 0
		inputChan <- 1
	}()

	values, ok := <-output
	s.True(ok)
	s.Equal([]int{0}, values)
	values, ok = <-output
	s.True(ok)
	s.Equal([]int{1}, values)
	cancel()
	_, ok = <-output
	s.False(ok)

	s.Equal(2, s.beforeSendCounter)
	s.Equal(2, s.afterSendCounter)
	s.Equal(2, s.beforeFlushCounter)
	s.Equal(2, s.afterFlushCounter)

	s.Equal([]int{0, 1}, s.beforeSendValues)
	s.Equal([]int{0, 1}, s.afterSendValues)
	s.Equal([]int{1, 1}, s.afterFlushValues)

	// check is context canceled
	s.NotNil(ctx.Err())
}

func (s *BufferFactorySuite) TestWithBeforeSendFlush() {
	ctx, cancel := context.WithCancel(context.Background())

	inputChan := make(chan int)
	flushChan := make(chan func())

	beforeSend := func(value int, flush func()) {
		s.beforeSend(value, flush)
		flush()
	}

	output := s.fabric.Create(
		ctx,
		inputChan,
		flushChan,
		beforeSend,
		s.afterSend,
		s.beforeFlush,
		s.afterFlush,
		100,
		cancel,
	)

	go func() {
		await, done := helper.NewSyncer(ctx)

		inputChan <- 0
		inputChan <- 1
		flushChan <- func() { done() }
		await()

		cancel()
	}()

	values, ok := <-output
	s.True(ok)
	s.Equal([]int{0}, values)
	values, ok = <-output
	s.True(ok)
	s.Equal([]int{1}, values)
	_, ok = <-output
	s.False(ok)

	s.Equal(2, s.beforeSendCounter)
	s.Equal(2, s.afterSendCounter)
	s.Equal(2, s.beforeFlushCounter)
	s.Equal(2, s.afterFlushCounter)

	s.Equal([]int{0, 1}, s.beforeSendValues)
	s.Equal([]int{0, 1}, s.afterSendValues)
	s.Equal([]int{1, 1}, s.afterFlushValues)

	// check is context canceled
	s.NotNil(ctx.Err())
}

func (s *BufferFactorySuite) TestWithClosedFlush() {
	ctx, cancel := context.WithCancel(context.Background())

	inputChan := make(chan int)
	flushChan := make(chan func())

	output := s.fabric.Create(
		ctx,
		inputChan,
		flushChan,
		s.beforeSend,
		s.afterSend,
		s.beforeFlush,
		s.afterFlush,
		100,
		cancel,
	)

	go func() {
		await, done := helper.NewSyncer(ctx)

		inputChan <- 0
		inputChan <- 1
		flushChan <- func() { done() }
		await()

		close(flushChan)
	}()

	values, ok := <-output
	s.True(ok)
	s.Equal([]int{0, 1}, values)
	_, ok = <-output
	s.False(ok)

	s.Equal(2, s.beforeSendCounter)
	s.Equal(2, s.afterSendCounter)
	s.Equal(1, s.beforeFlushCounter)
	s.Equal(1, s.afterFlushCounter)

	s.Equal([]int{0, 1}, s.beforeSendValues)
	s.Equal([]int{0, 1}, s.afterSendValues)
	s.Equal([]int{2}, s.afterFlushValues)

	// check is context canceled
	s.NotNil(ctx.Err())
}

func (s *BufferFactorySuite) TestWithClosedInput() {
	ctx, cancel := context.WithCancel(context.Background())

	inputChan := make(chan int)
	flushChan := make(chan func())

	output := s.fabric.Create(
		ctx,
		inputChan,
		flushChan,
		s.beforeSend,
		s.afterSend,
		s.beforeFlush,
		s.afterFlush,
		100,
		cancel,
	)

	go func() {
		await, done := helper.NewSyncer(ctx)

		inputChan <- 0
		inputChan <- 1
		flushChan <- func() { done() }
		await()

		close(inputChan)
	}()

	values, ok := <-output
	s.True(ok)
	s.Equal([]int{0, 1}, values)
	_, ok = <-output
	s.False(ok)

	s.Equal(2, s.beforeSendCounter)
	s.Equal(2, s.afterSendCounter)
	s.Equal(1, s.beforeFlushCounter)
	s.Equal(1, s.afterFlushCounter)

	s.Equal([]int{0, 1}, s.beforeSendValues)
	s.Equal([]int{0, 1}, s.afterSendValues)
	s.Equal([]int{2}, s.afterFlushValues)

	// check is context canceled
	s.NotNil(ctx.Err())
}

func (s *BufferFactorySuite) TestOk() {
	ctx, cancel := context.WithCancel(context.Background())

	inputChan := make(chan int)
	flushChan := make(chan func())

	output := s.fabric.Create(
		ctx,
		inputChan,
		flushChan,
		s.beforeSend,
		s.afterSend,
		s.beforeFlush,
		s.afterFlush,
		100,
		cancel,
	)

	go func() {
		await, done := helper.NewSyncer(ctx)

		inputChan <- 0
		inputChan <- 1
		flushChan <- func() { done() }
		await()
		inputChan <- 2
		flushChan <- func() { done() }
		await()
		flushChan <- func() { done() }
		await()

		cancel()
	}()

	values, ok := <-output
	s.True(ok)
	s.Equal([]int{0, 1}, values)
	values, ok = <-output
	s.True(ok)
	s.Equal([]int{2}, values)
	_, ok = <-output
	s.False(ok)

	s.Equal(3, s.beforeSendCounter)
	s.Equal(3, s.afterSendCounter)
	s.Equal(2, s.beforeFlushCounter)
	s.Equal(2, s.afterFlushCounter)

	s.Equal([]int{0, 1, 2}, s.beforeSendValues)
	s.Equal([]int{0, 1, 2}, s.afterSendValues)
	s.Equal([]int{2, 1}, s.afterFlushValues)

	// check is context canceled
	s.NotNil(ctx.Err())
}
