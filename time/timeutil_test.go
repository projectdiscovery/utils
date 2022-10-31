package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRFC3339ToTime(t *testing.T) {
	orig := time.Now()
	// converts back
	tt, err := RFC3339ToTime(orig.Format(time.RFC3339))
	require.Nil(t, err, "couldn't parse string time")
	require.Equal(t, orig.Unix(), tt.Unix(), "times don't match")
}

func TestMsToTime(t *testing.T) {
	// TBD in chaos + bbsh
}

func TestSToTime(t *testing.T) {
	// TBD in chaos + bbsh
}
