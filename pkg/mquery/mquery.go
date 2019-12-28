package mquery

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// https://github.com/estesp/mquery

type mQueryResponse struct {
	Payload mQueryResponsePayload `json:"payload"`
}

type mQueryResponsePayload struct {
	ID           string   `json:"_id"`
	ArchList     []string `json:"archList"`
	Cachetime    int64    `json:"cachetime"`
	ManifestList string   `json:"manifestList"`
	RepoTags     []string `json:"repoTags"`
	Tag          string   `json:"tag"`
}

// TODO: look for {"error"} responses
type mQueryResponseError struct {
	Error string `json:"error"`
}

type mQueryRequest struct {
	Image string
}

func (mReq *mQueryRequest) do() (mRes *mQueryResponse, found bool, err error) {
	url := fmt.Sprintf("https://openwhisk.ng.bluemix.net/api/v1/web/estesp%%40us.ibm.com_dev/default/archList.json?image=%s", mReq.Image)

	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}

	req.Header.Set("User-Agent", "jeopardy-nodeselector")

	res, err := netClient.Do(req)
	if err != nil {
		return nil, false, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, false, err
	}

	mRes = &mQueryResponse{}
	err = json.Unmarshal(body, &mRes)
	if err != nil {
		return nil, false, err
	}

	if mRes.Payload.ID == "" {
		return nil, false, nil
	}

	return mRes, true, err
}
