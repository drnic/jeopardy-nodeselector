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

type FakeNodeArchQuery struct {
	nodeArchs []string
	err       error
}

func (query FakeNodeArchQuery) NodeArchs() (archs []string, err error) {
	return query.nodeArchs, query.err
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
func TestExistingNodeSelectorNoPatch(t *testing.T) {
	pod := PodInspectImpl{
		podSpec: &core.PodSpec{
			NodeSelector: map[string]string{"some-label": "is-set"},
		},
	}
	resp := newAdmissionResponse()
	err := pod.ApplyPatchToAdmissionResponse(resp)
	assert.NoError(t, err, "no error expected when applying patch")
	assert.Equal(t, []byte("[]"), resp.Patch, "expect no patch applied")
}

// TODO: test this scenario in production
func _TestUnknownImageNoPatch(t *testing.T) {
	podSpec := core.PodSpec{
		Containers: []core.Container{
			core.Container{Image: "unknown-foobar"},
		},
	}
	pod := NewFromPodSpec(nil, &podSpec, "/spec/template")
	pod.imageQuery = FakeSingleImageQuery{nil, false, nil}
	pod.nodeArchQuery = FakeNodeArchQuery{[]string{"amd64"}, nil}
	err := pod.discoverContainerImagesArchitectures()
	assert.NoError(t, err, "unknown image should not result in error")
	assert.Equal(t, 0, len(*pod.containerImagesArchitectures), "unknown image should be ignored")

	resp := newAdmissionResponse()
	err = pod.ApplyPatchToAdmissionResponse(resp)
	assert.NoError(t, err, "no error expected when applying patch")
	assert.Equal(t, []byte("[]"), resp.Patch, "expect no patch applied")
	assert.Nil(t, pod.patchApplied, "expected no patch applied")
}

func TestSingleArchImage(t *testing.T) {
	podSpec := core.PodSpec{
		Containers: []core.Container{
			core.Container{Image: "bitnami/nginx:latest"},
		},
	}
	pod := NewFromPodSpec(nil, &podSpec, "/spec/template")
	pod.imageQuery = FakeSingleImageQuery{[]string{"amd64"}, true, nil}
	pod.nodeArchQuery = FakeNodeArchQuery{[]string{"arm", "amd64"}, nil}
	err := pod.discoverContainerImagesArchitectures()
	assert.NoError(t, err, "should not result in error")
	expected := map[string][]string{"bitnami/nginx:latest": []string{"amd64"}}
	assert.Equal(t, expected, *pod.containerImagesArchitectures, "single arch should match one of node arches")

	resp := newAdmissionResponse()
	err = pod.ApplyPatchToAdmissionResponse(resp)
	assert.NoError(t, err, "no error expected when applying patch")
	assert.NotEqual(t, []byte("[]"), resp.Patch, "expect some patch to be applied")
	assert.Equal(t, expectedPatch("amd64"), pod.patchApplied, "expected patch mismatch")
}

func TestManyArchitecturesButSingleCommonArch(t *testing.T) {
	podSpec := core.PodSpec{
		Containers: []core.Container{
			core.Container{Image: "nginx"},
			core.Container{Image: "nginx"},
			core.Container{Image: "nginx-arm"},
			core.Container{Image: "nginx-many"},
		},
		InitContainers: []core.Container{
			core.Container{Image: "nginx"},
			core.Container{Image: "something-unknown"},
		},
	}
	pod := NewFromPodSpec(nil, &podSpec, "/spec/template")
	pod.imageQuery = FakeManyImagesQuery{
		"nginx":      FakeSingleImageQuery{[]string{"amd64", "arm"}, true, nil},
		"nginx-arm":  FakeSingleImageQuery{[]string{"arm"}, true, nil},
		"nginx-many": FakeSingleImageQuery{[]string{"amd64", "arm64", "arm"}, true, nil},
		"nginx4":     FakeSingleImageQuery{[]string{"ignored"}, true, nil},
	}
	pod.nodeArchQuery = FakeNodeArchQuery{[]string{"arm", "arm64", "amd64"}, nil}
	err := pod.discoverContainerImagesArchitectures()
	assert.NoError(t, err, "should not result in error")
	expected := map[string][]string{
		"nginx":      []string{"amd64", "arm"},
		"nginx-arm":  []string{"arm"},
		"nginx-many": []string{"amd64", "arm64", "arm"},
	}
	assert.Equal(t, expected, *pod.containerImagesArchitectures, "image is published with a manifest")

	resp := newAdmissionResponse()
	err = pod.ApplyPatchToAdmissionResponse(resp)
	assert.NoError(t, err, "no error expected when applying patch")
	assert.NotEqual(t, []byte("[]"), resp.Patch, "expect some patch to be applied")
	assert.Equal(t, expectedPatch("arm"), pod.patchApplied, "expected patch mismatch")
}

func expectedPatch(arch string) *[]nodeSelectorPodPatch {
	return &[]nodeSelectorPodPatch{
		nodeSelectorPodPatch{
			Op:   "replace",
			Path: "/spec/template/spec/nodeSelector",
			Value: map[string]string{
				"kubernetes.io/arch": arch,
			},
		},
	}
}
