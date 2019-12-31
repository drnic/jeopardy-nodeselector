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
		var namespace string
		annotations := make(map[string]string)
		var podSpec *core.PodSpec
		// patch path is common for everything except raw Pod
		relativePatchPath := "/spec/template"

		// Extract the necessary metadata from our known Kinds
		switch kind {
		case "Pod":
			pod := &core.Pod{}
			if _, _, err := deserializer.Decode(admissionReview.Request.Object.Raw, nil, pod); err != nil {
				return nil, err
			}

			namespace = pod.GetNamespace()
			podSpec = &pod.Spec
			relativePatchPath = ""
		case "Deployment":
			deployment := apps.Deployment{}
			if _, _, err := deserializer.Decode(admissionReview.Request.Object.Raw, nil, &deployment); err != nil {
				return nil, err
			}

			deployment.GetNamespace()
			annotations = deployment.Spec.Template.GetAnnotations()
			podSpec = &deployment.Spec.Template.Spec
		case "StatefulSet":
			statefulset := apps.StatefulSet{}
			if _, _, err := deserializer.Decode(admissionReview.Request.Object.Raw, nil, &statefulset); err != nil {
				return nil, err
			}

			namespace = statefulset.GetNamespace()
			annotations = statefulset.Spec.Template.GetAnnotations()
			podSpec = &statefulset.Spec.Template.Spec
		case "DaemonSet":
			daemonset := apps.DaemonSet{}
			if _, _, err := deserializer.Decode(admissionReview.Request.Object.Raw, nil, &daemonset); err != nil {
				return nil, err
			}

			namespace = daemonset.GetNamespace()
			annotations = daemonset.Spec.Template.GetAnnotations()
			podSpec = &daemonset.Spec.Template.Spec
		case "Job":
			job := batch.Job{}
			if _, _, err := deserializer.Decode(admissionReview.Request.Object.Raw, nil, &job); err != nil {
				return nil, err
			}

			namespace = job.Spec.Template.GetNamespace()
			annotations = job.Spec.Template.GetAnnotations()
			podSpec = &job.Spec.Template.Spec
		default:
			// TODO(drnic): except for whitelisted namespaces
			return nil, xerrors.Errorf("the submitted Kind is not supported by this admission handler: %s", kind)
		}

		fmt.Printf("namespace: %s, annotations: %v\n", namespace, annotations)
		fmt.Printf("kind: %s, podSpec: %#v\n", kind, podSpec)

		podInspect := NewFromPodSpec(clientset, podSpec, relativePatchPath)
		fmt.Printf("podInspect: %#v\n", podInspect)

		podInspect.ApplyPatchToAdmissionResponse(resp)

		return resp, nil
	}
}
