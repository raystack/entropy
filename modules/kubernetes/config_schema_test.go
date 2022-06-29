package kubernetes

import (
	"testing"
	
	"github.com/xeipuuv/gojsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/odpf/entropy/pkg/errors"
)

func TestModule_KubernetesJSONSchema(t *testing.T) {
	tests := []struct{
		Case string
		wantErr error
		want bool
	}{
		{
			Case : `{
				"host": "http:0.0.0.0:1234",
				"insecure": true,
				"token": "token"
			}`,
			wantErr: nil,
			want: true,
		},
		{
			Case : `{
				"host": "http:0.0.0.0:1234",
				"insecure": false,
				"token": "token"
			  }`,
			wantErr: nil,
			want: false,
		},
		{
			Case : `{
				"host": "http:0.0.0.0:1234",
				"insecure": false,
				"cluster_ca_certificate": "c_ca_cert",
				"token": "token"
			  }`,
			wantErr: nil,
			want: true,
		},
		{
			Case : `{
				"host": "http:0.0.0.0:1234",
				"cluster_ca_certificate": "c_ca_cert",
				"token": "token"
			  }`,
			wantErr: nil,
			want: true,
		},
		{
			Case : `{
				"host": "http:0.0.0.0:1234",
				"insecure": true,
				"client_key": "c_key",
				"client_certificate": "c_cert"
			  }`,
			wantErr: nil,
			want: true,
		},
		{
			Case : `  "host": "http:0.0.0.0:1234",
			"insecure": true,
			"client_key": "c_key"
		  }`,
			wantErr: nil,
			want: false,
		},
		{
			Case : `{
				"host": "http:0.0.0.0:1234",
				"insecure": true,
				"token": "token",
				"client_key": "c_key",
				"client_certificate": "c_cert"
			  }`,
			wantErr: nil,
			want: false,
		},
		{
			Case : `{
				"host": "http:0.0.0.0:1234",
				"insecure": true,
				"token": "token",
				"client_key": "c_key"
			  }`,
			wantErr: nil,
			want: true,
		},
	}
	loader := gojsonschema.NewStringLoader(configSchema)
	schema, _ := gojsonschema.NewSchema(loader)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Case, func(t *testing.T) {
			t.Parallel()

			c := gojsonschema.NewStringLoader(tt.Case)
			result, err := schema.Validate(c)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, result.Valid())
		})
	}
}
