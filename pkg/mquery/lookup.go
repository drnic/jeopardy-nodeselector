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
		return []string{}, found, err
	}
	return query.normalizeArchList(resp.Payload.ArchList), found, nil
}

// normalizeArchList converts results from backend API into nodeSelector values
// The list from backend API:
// []string{"linux/amd64", "linux/arm/v7", "linux/arm64", "linux/386", "linux/ppc64le", "linux/s390x"}
// Will become:
// []string{"amd64", "arm", "arm64", "386", "ppc64le", "s390x"}
func (query ImageQueryImpl) normalizeArchList(received []string) (internal []string) {
	for _, recvImage := range received {
		var image string
		switch recvImage {
		case "linux/amd64":
			image = "amd64"
		case "linux/arm/v7":
			image = "arm"
		case "linux/arm64":
			image = "arm64"
		// TODO: I'm unsure what the correct values are for the following:
		case "linux/ppc64le":
			image = "ppc64le"
		case "linux/386":
			image = "386"
		case "linux/s390x":
			image = "s390x"
		// If unknown, then ignore for now. Submit PR to add new items.
		default:
			break
		}
		internal = append(internal, image)
	}
	return
}
