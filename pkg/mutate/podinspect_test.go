package mutate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
)

type FakeImageQuery struct {
	// fake responses
	architectures []string
	found         bool
	err           error
}

func (query *FakeImageQuery) LookupImageArchitectures(image string) (architectures []string, found bool, err error) {
	return query.architectures, query.found, query.err
}

func TestUnknownImage(t *testing.T) {
	podSpec := core.PodSpec{
		Containers: []core.Container{
			core.Container{Image: "unknown-foobar"},
		},
	}
	pod := NewFromPodSpec(&podSpec, "/spec/template")
	pod.imageQuery = &FakeImageQuery{nil, false, nil}
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
	pod.imageQuery = &FakeImageQuery{nil, true, nil}
	architectures, err := pod.containerImagesArchitectures()
	assert.NoError(t, err, "unknown image should not result in error")
	expected := map[string][]string{"bitnami/nginx:latest": nil}
	assert.Equal(t, expected, architectures, "image is published without a manifest")
}
