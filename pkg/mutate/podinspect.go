package mutate

import (
	"encoding/json"
	"fmt"

	arrayOp "github.com/adam-hanna/arrayOperations"
	"github.com/starkandwayne/jeopardy-nodeselector/pkg/mquery"
	admission "k8s.io/api/admission/v1beta1"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const defaultNodeArch = "amd64"

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
	clientset         *kubernetes.Clientset
	podSpec           *core.PodSpec
	relativePatchPath string

	patchApplied                 *[]nodeSelectorPodPatch
	containerImagesArchitectures *map[string][]string

	imageQuery    mquery.ImageQuery
	nodeArchQuery NodeArchQuery
}

// NewFromPodSpec consumes either a Pod or PodSpec
func NewFromPodSpec(clientset *kubernetes.Clientset, podSpec *core.PodSpec, relativePatchPath string) *PodInspectImpl {
	podImpl := &PodInspectImpl{
		clientset:         clientset,
		podSpec:           podSpec,
		relativePatchPath: relativePatchPath,
		imageQuery:        mquery.ImageQueryImpl{},
		nodeArchQuery:     &NodeArchQueryImpl{Clientset: clientset},
	}
	return podImpl
}

// ApplyPatchToAdmissionResponse applies a patch to
func (pod *PodInspectImpl) ApplyPatchToAdmissionResponse(resp *admission.AdmissionResponse) error {
	// Do not try to mutate podSpec if NodeSelector already set
	// In future, might want to only check if PodSpec already mutated by webhook,
	// so as to allow additional unrelated NodeSelector filters.
	if pod.podSpec.NodeSelector != nil || len(pod.podSpec.NodeSelector) != 0 {
		return nil
	}

	err := pod.discoverContainerImagesArchitectures()
	if err != nil {
		return err
	}
	fmt.Printf("image archs: %#v\n", *pod.containerImagesArchitectures)

	// If no container images are known, then assume they can run on
	// any node and do not apply a nodeSelector patch.
	if len(*pod.containerImagesArchitectures) == 0 {
		fmt.Println("no known images, so no nodeSelector patch applied")
		return nil
	}

	nodeArchs, err := pod.nodeArchQuery.NodeArchs()
	if err != nil {
		return err
	}

	someImagesHaveManifest, commonArchs := pod.commonImageArchitectures(nodeArchs)
	patch := pod.patchFromSingleArchRestriction(defaultNodeArch)
	if !someImagesHaveManifest {
		fmt.Printf("no images specify multiarch manifest, so defaulting nodeSelector to %v\n", defaultNodeArch)
	} else if len(commonArchs) == 0 {
		return fmt.Errorf("No commonly supported platform architecture between pod images: %v", pod.containerImages())
	} else {
		// For now, just pick the first item from list of common required archs
		// TODO: we need new node labels to allow more flexible allocation of pods to 2+ architectures if images support them
		// TODO: must pick an arch that is supported by actual nodes
		patch = pod.patchFromSingleArchRestriction(commonArchs[0])
	}
	if patch != nil {
		patchStr, err := json.Marshal(patch)
		if err != nil {
			return err
		}
		resp.Patch = patchStr
		pod.patchApplied = patch
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

// discoverContainerImagesArchitectures performs remote discovery of each container's image
// to determine what platform architectures it supports.
// Unknown images (to the remote backend API) are ignored from the result.
// It is assumed unknown images are published to support all node architectures available.
// Result is stored in containerImagesArchitectures to allow subsequent calls to function to be instantaneous.
func (pod *PodInspectImpl) discoverContainerImagesArchitectures() error {
	if pod.containerImagesArchitectures == nil {
		mapping := map[string][]string{}
		for _, image := range pod.containerImages() {
			architectures, found, err := pod.imageQuery.LookupImageArchitectures(image)
			if err != nil {
				return err
			}
			// If image not found, then assume its private and supports all required architectures
			if found {
				mapping[image] = architectures
			}
		}
		pod.containerImagesArchitectures = &mapping
	}
	return nil
}

// commonImageArchitectures finds the intersection of platform architectures
// provided by the list of images being used by a PodSpec.
// If all images are unknown, then commonArchs = [], and someImagesKnown = false
// If some images are known, but no common architectures then commonArchs = [], but someImagesKnown = true
// TODO: must pick an arch that is supported by actual nodes; start intersection loop with node arch list
func (pod *PodInspectImpl) commonImageArchitectures(nodeArchs []string) (someImagesKnown bool, commonArchs []string) {
	commonArchs = nodeArchs

	for _, imageArchs := range *pod.containerImagesArchitectures {
		fmt.Printf("commonArchs: %#v, imageArchs: %#v\n", commonArchs, imageArchs)
		if imageArchs != nil || len(imageArchs) > 0 {
			someImagesKnown = true
		}
	}
	if !someImagesKnown {
		fmt.Println("end: no images are known to backend API")
		return false, []string{}
	}

	for _, imageArchs := range *pod.containerImagesArchitectures {
		fmt.Printf("commonArchs: %#v, imageArchs: %#v\n", commonArchs, imageArchs)
		// Using https://github.com/adam-hanna/arrayOperations#intersect
		commonArchs = arrayOp.IntersectString(commonArchs, imageArchs)
	}
	fmt.Printf("end: commonArchs: %#v\n", commonArchs)
	return
}
