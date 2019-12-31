package mutate

import (
	arrayOp "github.com/adam-hanna/arrayOperations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NodeArchQuery describes lookup from Kube API of current nodes' platform architectures
// Implemented by NodeArchQueryImpl
// Faked by FakeNodeArchQuery
type NodeArchQuery interface {
	NodeArchs() ([]string, error)
}

// NodeArchQueryImpl describes lookup from Kube API of current nodes' platform architectures
type NodeArchQueryImpl struct {
	Clientset *kubernetes.Clientset
}

// NodeArchs performs lookup via Kubernetes Client.
func (nodes *NodeArchQueryImpl) NodeArchs() (archs []string, err error) {
	nodeList, err := nodes.Clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, node := range nodeList.Items {
		if arch, ok := node.Labels["kubernetes.io/arch"]; ok {
			archs = append(archs, arch)
		}
	}

	// Just default to amd64 if quirky cluster has no nodes or no labels
	if len(archs) == 0 {
		archs = []string{"amd64"}
	}

	// Using https://github.com/adam-hanna/arrayOperations
	archs = arrayOp.DistinctString(archs)

	return
}
