package mquery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnknownImage(t *testing.T) {
	mReq := mQueryRequest{Image: "unknown"}
	_, found, err := mReq.do()
	assert.False(t, found, "expected image to not be found")
	assert.NoError(t, err, "no error expected when looking up unknown image")
}

func TestSingleArchImage(t *testing.T) {
	mReq := mQueryRequest{Image: "bitnami/nginx"}
	mRes, found, err := mReq.do()
	// fmt.Printf("%#v\n", mRes.Payload.ArchList)
	var nilList []string
	assert.Equal(t, nilList, mRes.Payload.ArchList, "expected single architecture image to return empty list")
	assert.True(t, found, "expected image to be found")
	assert.NoError(t, err, "no error expected when looking up known image")
}

func TestMultiArchImage(t *testing.T) {
	mReq := mQueryRequest{Image: "nginx"}
	mRes, found, err := mReq.do()
	// fmt.Printf("%#v\n", mRes.Payload.ArchList)
	assert.Equal(t,
		[]string{"linux/amd64", "linux/arm/v7", "linux/arm64", "linux/386", "linux/ppc64le", "linux/s390x"},
		mRes.Payload.ArchList, "expected multiarch image to return list of results")
	assert.True(t, found, "expected image to be found")
	assert.NoError(t, err, "no error expected when looking up known image")
}
