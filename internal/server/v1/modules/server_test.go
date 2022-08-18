package modules

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/odpf/entropy/pkg/errors"
)

func TestAPIServer_ListModules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) *APIServer
		request *entropyv1beta1.ListModulesRequest
		want    *entropyv1beta1.ListModulesResponse
		wantErr error
	}{}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := tt.setup(t)

			got, err := srv.ListModules(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Truef(t, errors.Is(err, tt.wantErr), "'%s' != '%s'", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
