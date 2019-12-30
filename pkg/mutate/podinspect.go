package mutate

import (
	"encoding/json"
	"fmt"

	"github.com/starkandwayne/jeopardy-nodeselector/pkg/mquery"
	admission "k8s.io/api/admission/v1beta1"
	core "k8s.io/api/core/v1"
)

// PodInpsect describes interface for modifying an admission controller response
// Implemented by PodInspectImpl
// TODO: PodInpsect is not currently used by any fake implementations
type PodInpsect interface {
	ApplyPatchToAdmissionResponse(resp *admission.AdmissionResponse)
}

// PodInspectImpl describes how to modify an admission controller response
// to mutate a Pod/Deployment/etc to set the nodeSelector of their PodSpec
// to ensure the Pod is allocated to a subset of Nodes that are capable of
// running the Pod's Containers' images.
type PodInspectImpl struct {
	podSpec           *core.PodSpec
	relativePatchPath string
	imageQuery        mquery.ImageQuery
}

// NewFromPodSpec consumes either a Pod or PodSpec
func NewFromPodSpec(podSpec *core.PodSpec, relativePatchPath string) *PodInspectImpl {
	podImpl := &PodInspectImpl{
		podSpec:           podSpec,
		relativePatchPath: relativePatchPath,
		imageQuery:        mquery.ImageQueryImpl{},
	}
	return podImpl
}

// ApplyPatchToAdmissionResponse applies a patch to
func (pod *PodInspectImpl) ApplyPatchToAdmissionResponse(resp *admission.AdmissionResponse) error {
	multiarchMapping, err := pod.containerImagesArchitectures()
	if err != nil {
		return err
	}
	fmt.Printf("image archs: %#v\n", multiarchMapping)
	patch := pod.patchFromSingleArchRestriction("amd64")
	if patch != nil {
		patchStr, err := json.Marshal(patch)
		if err != nil {
			return err
		}
		resp.Patch = patchStr
	}
	fmt.Printf("patch: %#v\n", patch)
	return nil
}

// nodeSelectorPodPatch describes a proposed JSONPatch to a Pod
type nodeSelectorPodPatch struct {
	Op    string            `json:"op"`
	Path  string            `json:"path"`
	Value map[string]string `json:"value"`
}

func (pod *PodInspectImpl) patchFromSingleArchRestriction(restrictArch string) *[]nodeSelectorPodPatch {
	if len(restrictArch) > 0 {
		return &[]nodeSelectorPodPatch{
			nodeSelectorPodPatch{
				Op:   "replace",
				Path: fmt.Sprintf("%s/spec/nodeSelector", pod.relativePatchPath),
				Value: map[string]string{
					"kubernetes.io/arch": restrictArch,
				},
			},
		}
	}
	return nil
}

// containerImages composes a list of the images (uri:tag)
// It is a unique list, with duplicates already removed.
func (pod *PodInspectImpl) containerImages() (images []string) {
	uniqImages := map[string]bool{}
	for _, c := range pod.podSpec.InitContainers {
		if _, ok := uniqImages[c.Image]; !ok {
			images = append(images, c.Image)
			uniqImages[c.Image] = true
		}
	}
	for _, c := range pod.podSpec.Containers {
		if _, ok := uniqImages[c.Image]; !ok {
			images = append(images, c.Image)
			uniqImages[c.Image] = true
		}
	}
	return
}

func (pod *PodInspectImpl) containerImagesArchitectures() (mapping map[string][]string, err error) {
	mapping = map[string][]string{}
	for _, image := range pod.containerImages() {
		architectures, found, err := pod.imageQuery.LookupImageArchitectures(image)
		if err != nil {
			return mapping, err
		}
		// If image not found, then assume its private and supports all required architectures
		if found {
			mapping[image] = architectures
		}
	}
	return
}
