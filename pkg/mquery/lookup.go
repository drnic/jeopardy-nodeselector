package mquery

// ImageQuery represents the request for a query of an image for its supported architectures
// Primary implementation by ImageQueryImpl
// Interface allows for fakes FakeSingleImageQuery, FakeManyImagesQuery
type ImageQuery interface {
	LookupImageArchitectures(image string) (architectures []string, found bool, err error)
}

// ImageQueryImpl performs the remote query to backend of https://github.com/estesp/mquery
type ImageQueryImpl struct {
}

// LookupImageArchitectures is the primary interface to discover how
// we will restrict nodeSelector
func (query ImageQueryImpl) LookupImageArchitectures(image string) (architectures []string, found bool, err error) {
	req := &mQueryRequest{image}
	resp, found, err := req.do()
	if err != nil || !found {
		return architectures, found, err
	}
	return resp.Payload.ArchList, found, nil
}
