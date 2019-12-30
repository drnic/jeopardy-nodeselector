package mquery

// LookupImageArchitectures is the primary interface to discover how
// we will restrict nodeSelector
func LookupImageArchitectures(image string) (architectures []string, found bool, err error) {
	req := &mQueryRequest{image}
	resp, found, err := req.do()
	if err != nil || !found {
		return architectures, found, err
	}
	return resp.Payload.ArchList, found, nil
}
