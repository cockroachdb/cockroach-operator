package v1alpha1

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetClusterSpecDefaults(t *testing.T) {
	s := &CrdbClusterSpec{}

	expected := &CrdbClusterSpec{
		GRPCPort:     &DefaultGRPCPort,
		HTTPPort:     &DefaultHTTPPort,
		Cache:        "25%",
		MaxSQLMemory: "25%",
	}

	SetClusterSpecDefaults(s)

	diff := cmp.Diff(expected, s)
	if diff != "" {
		assert.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff))
	}
}
