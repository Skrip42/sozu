package helper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
)

type SyncerSuite struct {
	suite.Suite
}

func TestSyncer(t *testing.T) {
	suite.Run(t, &SyncerSuite{})
}

func (s *SyncerSuite) TearDownTest() {
	goleak.VerifyNone(s.T())
}

func (s *SyncerSuite) TestDoneWithContextCancel() {
	ctx, cancel := context.WithCancel(context.Background())
	_, done := NewSyncer(ctx)

	go func() {
		done()
	}()

	cancel()
}

func (s *SyncerSuite) TestAwaitWithContextCancel() {
	ctx, cancel := context.WithCancel(context.Background())
	await, _ := NewSyncer(ctx)

	go func() {
		await()
	}()

	cancel()
}

func (s *SyncerSuite) TestOk() {
	ctx, _ := context.WithCancel(context.Background())
	await, done := NewSyncer(ctx)

	go func() {
		await()
	}()

	done()
}
