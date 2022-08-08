package resource_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func TestResource_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		res  resource.Resource
		want error
	}{
		{
			name: "InvalidName",
			res: resource.Resource{
				Kind:    "fake",
				Name:    "",
				Project: "bar",
			},
			want: errors.ErrInvalid,
		},
		{
			name: "InvalidKind",
			res: resource.Resource{
				Kind:    "",
				Name:    "foo",
				Project: "bar",
			},
			want: errors.ErrInvalid,
		},
		{
			name: "InvalidProject",
			res: resource.Resource{
				Kind:    "fake",
				Name:    "foo",
				Project: "978997",
			},
			want: errors.ErrInvalid,
		},

		{
			name: "ValidResource",
			res: resource.Resource{
				Kind:    "fake",
				Name:    "foo",
				Project: "odpf",
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.res.Validate(true)
			assert.Truef(t, errors.Is(got, tt.want), "want=%v, got=%v", tt.want, got)
		})
	}
}
