package singularity

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL}
}

func (c *Client) Requests() ([]RequestParent, error) {
	var result []RequestParent
	return result, c.get("/requests", &result)
}

func (c *Client) get(url string, out interface{}) error {
	u := fmt.Sprintf("%s/api%s", c.baseURL, url)
	r, err := http.Get(u)
	if r != nil && r.Body != nil {
		defer r.Body.Close()
	}
	if err != nil {
		return err
	}
	return json.NewDecoder(r.Body).Decode(&out)
}

type RequestParent struct {
	Request            Request
	RequestDeployState RequestDeployState
}

type Request struct {
	ID          string
	RequestType string
	Daemon      bool
	Instances   int
	Owners      []string
}

type RequestDeployState struct {
	RequestID                   string
	PendingDeploy, ActiveDeploy DeployMarker
}

type DeployMarker struct {
	User, RequestID, DeployID string
	Timestamp                 int64
}
