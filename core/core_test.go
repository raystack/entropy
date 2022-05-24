package core_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/odpf/entropy/core"
	"github.com/odpf/entropy/core/mocks"
	"github.com/odpf/entropy/core/resource"
)

var (
	sampleResource = resource.Resource{
		URN:     "foo:bar:baz",
		Kind:    "foo",
		Name:    "baz",
		Project: "bar",
	}

	frozenTime = time.Unix(1650536955, 0)
	deadClock  = func() time.Time { return frozenTime }
)

func TestNew(t *testing.T) {
	t.Parallel()
	s := core.New(&mocks.ResourceRepository{}, &mocks.Module{}, deadClock, nil)
	assert.NotNil(t, s)
}
