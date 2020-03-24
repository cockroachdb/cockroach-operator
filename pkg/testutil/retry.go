package testutil

import (
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

var DefaultRetry = wait.Backoff{
	Duration: 10 * time.Microsecond,
	Factor:   1.0,
	Jitter:   0.1,
	Steps:    500,
}
