package internal

import (
	"context"
	"testing"

	"github.com/Skrip42/sozu/internal/helper"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
)

type AggregatorFabricSuite struct {
	suite.Suite

	aggregateFunc func(a, b int) int

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

	fabric Fabric[int]
}

func (s *AggregatorFabricSuite) SetupTest() {

	s.aggregateFunc = func(a, b int) int {
		return a + b
	}

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

	s.fabric = NewAggregatorFabric[int](s.aggregateFunc)
}

func (s *AggregatorFabricSuite) TearDownTest() {
	goleak.VerifyNone(s.T())
}

func TestAggregatorFabric(t *testing.T) {
	suite.Run(t, &AggregatorFabricSuite{})
}

func (s *AggregatorFabricSuite) TestWithAfterSendFlush() {
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
		inputChan <- 1
		inputChan <- 2

	}()

	values, ok := <-output
	s.True(ok)
	s.Equal([]int{1}, values)
	values, ok = <-output
	s.True(ok)
	s.Equal([]int{3}, values)
	cancel()
	_, ok = <-output
	s.False(ok)

	s.Equal(2, s.beforeSendCounter)
	s.Equal(2, s.afterSendCounter)
	s.Equal(2, s.beforeFlushCounter)
	s.Equal(2, s.afterFlushCounter)

	s.Equal([]int{1, 2}, s.beforeSendValues)
	s.Equal([]int{1, 2}, s.afterSendValues)
	s.Equal([]int{1, 1}, s.afterFlushValues)

	// check is context canceled
	s.NotNil(ctx.Err())
}

func (s *AggregatorFabricSuite) TestWithBeforeSendFlush() {
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

		inputChan <- 1
		inputChan <- 2
		flushChan <- func() { done() }
		await()

		cancel()
	}()

	values, ok := <-output
	s.True(ok)
	s.Equal([]int{1}, values)
	values, ok = <-output
	s.True(ok)
	s.Equal([]int{3}, values)
	_, ok = <-output
	s.False(ok)

	s.Equal(2, s.beforeSendCounter)
	s.Equal(2, s.afterSendCounter)
	s.Equal(2, s.beforeFlushCounter)
	s.Equal(2, s.afterFlushCounter)

	s.Equal([]int{1, 2}, s.beforeSendValues)
	s.Equal([]int{1, 2}, s.afterSendValues)
	s.Equal([]int{1, 1}, s.afterFlushValues)

	// check is context canceled
	s.NotNil(ctx.Err())
}

func (s *AggregatorFabricSuite) TestWithClosedFlush() {
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

		inputChan <- 1
		inputChan <- 2
		flushChan <- func() { done() }
		await()

		close(flushChan)
	}()

	values, ok := <-output
	s.True(ok)
	s.Equal([]int{3}, values)
	_, ok = <-output
	s.False(ok)

	s.Equal(2, s.beforeSendCounter)
	s.Equal(2, s.afterSendCounter)
	s.Equal(1, s.beforeFlushCounter)
	s.Equal(1, s.afterFlushCounter)

	s.Equal([]int{1, 2}, s.beforeSendValues)
	s.Equal([]int{1, 2}, s.afterSendValues)
	s.Equal([]int{2}, s.afterFlushValues)

	// check is context canceled
	s.NotNil(ctx.Err())
}

func (s *AggregatorFabricSuite) TestWithClosedInput() {
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

		inputChan <- 1
		inputChan <- 2
		flushChan <- func() { done() }
		await()

		close(inputChan)
	}()

	values, ok := <-output
	s.True(ok)
	s.Equal([]int{3}, values)
	_, ok = <-output
	s.False(ok)

	s.Equal(2, s.beforeSendCounter)
	s.Equal(2, s.afterSendCounter)
	s.Equal(1, s.beforeFlushCounter)
	s.Equal(1, s.afterFlushCounter)

	s.Equal([]int{1, 2}, s.beforeSendValues)
	s.Equal([]int{1, 2}, s.afterSendValues)
	s.Equal([]int{2}, s.afterFlushValues)

	// check is context canceled
	s.NotNil(ctx.Err())
}

func (s *AggregatorFabricSuite) TestOk() {
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

		inputChan <- 1
		inputChan <- 2
		flushChan <- func() { done() }
		await()
		inputChan <- 3
		flushChan <- func() { done() }
		await()
		flushChan <- func() { done() }
		await()

		cancel()
	}()

	values, ok := <-output
	s.True(ok)
	s.Equal([]int{3}, values)
	values, ok = <-output
	s.True(ok)
	s.Equal([]int{6}, values)
	_, ok = <-output
	s.False(ok)

	s.Equal(3, s.beforeSendCounter)
	s.Equal(3, s.afterSendCounter)
	s.Equal(2, s.beforeFlushCounter)
	s.Equal(2, s.afterFlushCounter)

	s.Equal([]int{1, 2, 3}, s.beforeSendValues)
	s.Equal([]int{1, 2, 3}, s.afterSendValues)
	s.Equal([]int{2, 1}, s.afterFlushValues)

	// check is context canceled
	s.NotNil(ctx.Err())
}
