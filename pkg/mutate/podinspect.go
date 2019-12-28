package mutate

import (
	"encoding/json"
	"fmt"

	admission "k8s.io/api/admission/v1beta1"
	core "k8s.io/api/core/v1"
)

type PodInspectImpl struct {
	pod               *core.Pod
	podSpec           *core.PodSpec
	relativePatchPath string
}

type PodInpsect interface {
	ApplyPatchToAdmissionResponse(resp *admission.AdmissionResponse)
}

// NewFromPodOrPodSpec consumes either a Pod or PodSpec
func NewFromPodOrPodSpec(pod *core.Pod, podSpec *core.PodSpec) *PodInspectImpl {
	podImpl := &PodInspectImpl{
		pod:               pod,
		podSpec:           podSpec,
		relativePatchPath: "/spec/template",
	}
	if podImpl != nil {
		podImpl.relativePatchPath = ""
	}
	return podImpl
}

// ApplyPatchToAdmissionResponse applies a patch to
func (pod *PodInspectImpl) ApplyPatchToAdmissionResponse(resp *admission.AdmissionResponse) error {
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
