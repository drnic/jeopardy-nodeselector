package mutate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	admission "k8s.io/api/admission/v1beta1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FakeSingleImageQuery supports fake response for simple image query
type FakeSingleImageQuery struct {
	architectures []string
	found         bool
	err           error
}

func (query FakeSingleImageQuery) LookupImageArchitectures(image string) (architectures []string, found bool, err error) {
	return query.architectures, query.found, query.err
}

// FakeManyImagesQuery allows a test with many container images
type FakeManyImagesQuery map[string]FakeSingleImageQuery

func (queries FakeManyImagesQuery) LookupImageArchitectures(image string) (architectures []string, found bool, err error) {
	if query, ok := queries[image]; ok {
		return query.LookupImageArchitectures(image)
	}
	return nil, false, nil
}

func newAdmissionResponse() *admission.AdmissionResponse {
	pT := admission.PatchTypeJSONPatch
	return &admission.AdmissionResponse{
		Allowed:   true,
		Result:    &metav1.Status{},
		PatchType: &pT,
		Patch:     []byte("[]"),
	}
}
func TestDoNotMutateIfNodeSelectorSet(t *testing.T) {
	pod := PodInspectImpl{
		podSpec: &core.PodSpec{
			NodeSelector: map[string]string{"some-label": "is-set"},
		},
	}
	resp := newAdmissionResponse()
	err := pod.ApplyPatchToAdmissionResponse(resp)
	assert.NoError(t, err, "no error expected")
	assert.Equal(t, []byte("[]"), resp.Patch, "expect no patch applied")
}

func TestUnknownImage(t *testing.T) {
	podSpec := core.PodSpec{
		Containers: []core.Container{
			core.Container{Image: "unknown-foobar"},
		},
	}
	pod := NewFromPodSpec(&podSpec, "/spec/template")
	pod.imageQuery = FakeSingleImageQuery{nil, false, nil}
	architectures, err := pod.containerImagesArchitectures()
	assert.NoError(t, err, "unknown image should not result in error")
	assert.Equal(t, 0, len(architectures), "unknown image should be ignored")
}

func TestNoManifestImageDefault(t *testing.T) {
	podSpec := core.PodSpec{
		Containers: []core.Container{
			core.Container{Image: "bitnami/nginx:latest"},
		},
	}
	pod := NewFromPodSpec(&podSpec, "/spec/template")
	pod.imageQuery = FakeSingleImageQuery{nil, true, nil}
	architectures, err := pod.containerImagesArchitectures()
	assert.NoError(t, err, "should not result in error")
	expected := map[string][]string{"bitnami/nginx:latest": nil}
	assert.Equal(t, expected, architectures, "image is published without a manifest")
}

func TestManyContainersImageManifestSomeArchitectures(t *testing.T) {
	podSpec := core.PodSpec{
		Containers: []core.Container{
			core.Container{Image: "nginx"},
			core.Container{Image: "nginx"},
			core.Container{Image: "nginx2"},
			core.Container{Image: "nginx3"},
		},
		InitContainers: []core.Container{
			core.Container{Image: "nginx"},
			core.Container{Image: "something-unknown"},
		},
	}
	pod := NewFromPodSpec(&podSpec, "/spec/template")
	pod.imageQuery = FakeManyImagesQuery{
		"nginx":  FakeSingleImageQuery{[]string{"amd64", "arm64"}, true, nil},
		"nginx2": FakeSingleImageQuery{[]string{"amd64"}, true, nil},
		"nginx3": FakeSingleImageQuery{[]string{"amd64", "arm64"}, true, nil},
		"nginx4": FakeSingleImageQuery{[]string{"ignored"}, true, nil},
	}
	architectures, err := pod.containerImagesArchitectures()
	assert.NoError(t, err, "should not result in error")
	expected := map[string][]string{
		"nginx":  []string{"amd64", "arm64"},
		"nginx2": []string{"amd64"},
		"nginx3": []string{"amd64", "arm64"},
	}
	assert.Equal(t, expected, architectures, "image is published with a manifest")
}
