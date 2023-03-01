package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

func TestModule_KubernetesJSONSchema(t *testing.T) {
	tests := []struct {
		title         string
		Case          string
		shouldBeValid bool
	}{
		{
			title: "TokenAuthPresent_InsecureTrue",
			Case: `{
				"host": "http://0.0.0.0:1234",
				"insecure": true,
				"token": "token"
			}`,
			shouldBeValid: true,
		},
		{
			title: "TokenAuthPresent_InsecureFalse",
			Case: `{
			  "host": "http://0.0.0.0:1234",
			  "insecure": false,
			  "token": "foo"
			}`,
			shouldBeValid: true,
		},
		{
			title: "TokenAuthPresent_CertIsPresentToo",
			Case: `{
				"host": "http://0.0.0.0:1234",
				"insecure": false,
				"token": "token",
				"cluster_certificate": "c_ca_cert"
			  }`,
			shouldBeValid: true,
		},
		{
			title: "CertAuthPresent_InsecureTrue",
			Case: `{
				"host": "http://0.0.0.0:1234",
				"insecure": true,
				"client_key": "c_key",
				"client_certificate": "c_cert"
			  }`,
			shouldBeValid: true,
		},
		{
			title: "CertAuthPresent_InsecureFalse",
			Case: `{
				"host": "http://0.0.0.0:1234",
				"insecure": false,
				"client_key": "c_key",
				"client_certificate": "c_cert"
			  }`,
			shouldBeValid: false,
		},

		{
			title: "CertAuthPresent_InsecureFalse_WithCACert",
			Case: `{
				"host": "http://0.0.0.0:1234",
				"insecure": false,
				"client_key": "c_key",
				"client_certificate": "c_cert",
				"client_ca_certificate": "ca_cert"
			  }`,
			shouldBeValid: true,
		},
		{
			title: "Missing_ClientCert",
			Case: `{
				"host": "http://0.0.0.0:1234",
				"client_key": "c_key"
		  	}`,
			shouldBeValid: false,
		},
		{
			title: "Missing_ClientKey",
			Case: `{
				"host": "http://0.0.0.0:1234",
				"insecure": true,
				"client_certificate": "c_cert"
			  }`,
			shouldBeValid: false,
		},
		{
			title: "Missing_CACert",
			Case: `{
				"host": "http://0.0.0.0:1234",
				"insecure": false,
				"client_key": "foo",
				"client_certificate": "c_cert"
			  }`,
			shouldBeValid: false,
		},
	}

	schema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(configSchema))
	require.NoError(t, err)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()

			c := gojsonschema.NewStringLoader(tt.Case)
			result, err := schema.Validate(c)
			require.NoError(t, err)
			assert.Equal(t, tt.shouldBeValid, result.Valid())
		})
	}
}
