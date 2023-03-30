package firehose

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_safeReleaseName(t *testing.T) {
	t.Parallel()

	table := []struct {
		str  string
		want string
	}{
		{
			str:  "abcd-efgh",
			want: "abcd-efgh-firehose",
		},
		{
			str:  "abcd-efgh-firehose",
			want: "abcd-efgh-firehose",
		},
		{
			str:  "ABCDEFGHIJKLMNOPQRSTUVWXYZ-abcdefghijklmnopqrstuvwxyz",
			want: "ABCDEFGHIJKLMNOPQRSTUVWXYZ-abcdefghij-3801d0-firehose",
		},
		{
			str:  "ABCDEFGHIJKLMNOPQRSTUVWXYZ-abcdefghi---klmnopqrstuvwxyz",
			want: "ABCDEFGHIJKLMNOPQRSTUVWXYZ-abcdefghi-81c192-firehose",
		},
	}

	for i, tt := range table {
		t.Run(fmt.Sprintf("Case#%d", i), func(t *testing.T) {
			got := safeReleaseName(tt.str)
			assert.Equal(t, tt.want, got)
			assert.True(t, len(got) <= helmReleaseNameMaxLength, "release name has length %d", len(got))
		})
	}
}
