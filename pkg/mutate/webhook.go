package mutate

import (
	"fmt"

	"golang.org/x/xerrors"

	admissioncontrol "github.com/elithrar/admission-control"

	admission "k8s.io/api/admission/v1beta1"
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// NodeSelectorMultiArch constructs an AdmitFunc to perform this project's primary functions
func NodeSelectorMultiArch(clientset *kubernetes.Clientset, ignoredNamespaces []string) admissioncontrol.AdmitFunc {
	return func(admissionReview *admission.AdmissionReview) (*admission.AdmissionResponse, error) {
		kind := admissionReview.Request.Kind.Kind
		pT := admission.PatchTypeJSONPatch
		resp := &admission.AdmissionResponse{
			Allowed:   true,
			Result:    &metav1.Status{},
			PatchType: &pT,
			Patch:     []byte("[]"),
		}

		deserializer := serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()

		// We handle all built-in Kinds that include a PodTemplateSpec, as described here:
		// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#pod-v1-core
		var podSpec *core.PodSpec
		// patch path is common for built-in resource definitions
		// for everything except Pod
		relativePatchPath := "/spec/template"

		// Extract the necessary metadata from our known Kinds
		switch kind {
		case "Pod":
			pod := &core.Pod{}
			if _, _, err := deserializer.Decode(admissionReview.Request.Object.Raw, nil, pod); err != nil {
				return nil, err
			}
			podSpec = &pod.Spec
			relativePatchPath = ""
		case "Deployment":
			deployment := apps.Deployment{}
			if _, _, err := deserializer.Decode(admissionReview.Request.Object.Raw, nil, &deployment); err != nil {
				return nil, err
			}
			podSpec = &deployment.Spec.Template.Spec
		case "StatefulSet":
			statefulset := apps.StatefulSet{}
			if _, _, err := deserializer.Decode(admissionReview.Request.Object.Raw, nil, &statefulset); err != nil {
				return nil, err
			}
			podSpec = &statefulset.Spec.Template.Spec
		case "DaemonSet":
			daemonset := apps.DaemonSet{}
			if _, _, err := deserializer.Decode(admissionReview.Request.Object.Raw, nil, &daemonset); err != nil {
				return nil, err
			}
			podSpec = &daemonset.Spec.Template.Spec
		case "Job":
			job := batch.Job{}
			if _, _, err := deserializer.Decode(admissionReview.Request.Object.Raw, nil, &job); err != nil {
				return nil, err
			}
			podSpec = &job.Spec.Template.Spec
		default:
			// TODO(drnic): except for whitelisted namespaces
			return nil, xerrors.Errorf("the submitted Kind is not supported by this admission handler: %s", kind)
		}

		fmt.Printf("kind: %s\n", kind)

		podInspect := NewFromPodSpec(clientset, podSpec, relativePatchPath)
		podInspect.ApplyPatchToAdmissionResponse(resp)

		return resp, nil
	}
}
