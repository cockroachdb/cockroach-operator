package testutil

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func AssertDiff(t *testing.T, expected interface{}, actual interface{}) {
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("unexpected result (-want +got):\n%v", diff)
	}
}
