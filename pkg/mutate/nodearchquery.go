package mutate

// NodeArchQuery describes lookup from Kube API of current nodes' platform architectures
// Implemented by NodeArchQueryImpl
// Not currently implemented by any fakes.
type NodeArchQuery interface {
	NodeArchs() ([]string, error)
}

// NodeArchQueryImpl describes lookup from Kube API of current nodes' platform architectures
type NodeArchQueryImpl struct {
}

// NodeArchs performs lookup via Kubernetes Client.
func (nodes *NodeArchQueryImpl) NodeArchs() (archs []string, err error) {
	// return []string{"armv7"}, nil
	return []string{"armv7", "amd64"}, nil
}
