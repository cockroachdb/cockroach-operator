package testutil

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	"strings"
	"testing"
)

// AssertDiff compares an expected interface with an actual interface
// and fails if the two interfaces do not match.
// This func removes all metadata calico annotations.
func AssertDiff(t *testing.T, expected interface{}, actual interface{}) {

	if actual, ok := actual.(string); ok {
		decode, encode := Yamlizers(t, InitScheme(t))
		var newSlice []string
		actualSlice := strings.Split(actual, "---")

		for _, str := range actualSlice {
			// we are getting zero length string because the string starts with ---
			if len(str) == 0 {
				continue
			}
			// if we do not need to parse the string if it does not have an annotation in it
			if !strings.Contains(str, "cni.projectcalico.org") {
				newSlice = append(newSlice, str)
				continue
			}
			// TODO test if the string is actually an api definition???
			// otherwise the decode is going to throw an error

			obj := decode([]byte(str))

			// we are only removing annotations out of pods
			switch obj.(type) {
			case *v1.Pod:
				pod := obj.(*v1.Pod)
				delete(pod.ObjectMeta.Annotations, "cni.projectcalico.org/podIP")
				var b bytes.Buffer
				err := encode(pod, &b)
				require.NoError(t, err)
				newSlice = append(newSlice, "\n"+b.String()+"\n")

			default:
				// todo warn and log
				newSlice = append(newSlice, str)
			}
		}

		newYaml := strings.Join(newSlice, "---")
		newYaml = "---" + newYaml
		if diff := cmp.Diff(expected, newYaml); diff != "" {
			t.Fatalf("unexpected result (-want +got):\n%v", diff)
		}
	} else {
		// Doing the else because of scoping weirdness
		// I set actual to the strings.Join and it did not carry through
		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Fatalf("unexpected result (-want +got):\n%v", diff)
		}
	}
}
