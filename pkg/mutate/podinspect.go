package mutate

import (
	"encoding/json"
	"fmt"

	admission "k8s.io/api/admission/v1beta1"
	core "k8s.io/api/core/v1"
)

type PodInspectImpl struct {
	podSpec           *core.PodSpec
	relativePatchPath string
}

type PodInpsect interface {
	ApplyPatchToAdmissionResponse(resp *admission.AdmissionResponse)
}

// NewFromPodOrPodSpec consumes either a Pod or PodSpec
func NewFromPodSpec(podSpec *core.PodSpec, relativePatchPath string) *PodInspectImpl {
	podImpl := &PodInspectImpl{
		podSpec:           podSpec,
		relativePatchPath: relativePatchPath,
	}
	return podImpl
}

// ApplyPatchToAdmissionResponse applies a patch to
func (pod *PodInspectImpl) ApplyPatchToAdmissionResponse(resp *admission.AdmissionResponse) error {
	fmt.Printf("images: %v\n", pod.containerImages())
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
