package core_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/odpf/entropy/core"
	"github.com/odpf/entropy/core/mocks"
)

func TestService_Sync(t *testing.T) {
	t.Parallel()
	s := core.New(&mocks.ResourceRepository{}, &mocks.Module{}, deadClock, nil)

	t.Run("CancelledContext", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // immediately cancel the context.

		var err error
		var exited bool
		go func() {
			err = s.RunSync(ctx)
			exited = true
		}()
		time.Sleep(500 * time.Millisecond)

		assert.NoError(t, err)
		assert.True(t, exited)
	})
}
