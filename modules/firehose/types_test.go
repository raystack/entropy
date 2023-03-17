package firehose

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_generateSafeReleaseName(t *testing.T) {
	t.Parallel()

	table := []struct {
		project string
		name    string
		want    string
	}{
		{
			project: "g-pilotdata-gl",
			name:    "g-pilotdata-gl-test-firehose-des-4011-firehose",
			want:    "firehose-g-pilotdata-gl-g-pilotdata-gl-test-fi-63acaa",
		},
		{
			project: "abc",
			name:    "xyz",
			want:    "firehose-abc-xyz",
		},
		{
			project: "abcdefghijklmnopqrstuvxyz",
			name:    "ABCDEFGHIJKLMNOPQRSTUVXYZ",
			want:    "firehose-abcdefghijklmnopqrstuvxyz-ABCDEFGHIJK-0f1383",
		},
		{
			project: "abcdefghijklmnopqrstuvxyz",
			name:    "ABCDEFGHIJ-LMNOPQRSTUVXYZ",
			want:    "firehose-abcdefghijklmnopqrstuvxyz-ABCDEFGHIJ-c164fe",
		},
	}

	for i, tt := range table {
		t.Run(fmt.Sprintf("Case#%d", i), func(t *testing.T) {
			got := generateSafeReleaseName(tt.project, tt.name)
			assert.Equal(t, tt.want, got)
			assert.True(t, len(got) <= 53)
		})
	}
}
