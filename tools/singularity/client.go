package singularity

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	return result, c.get(&result, "/requests")
}

func (c *Client) Request(id string) (*RequestParent, error) {
	var result *RequestParent
	return result, c.get(result, "/requests/request/%s", id)
}

func (c *Client) get(out interface{}, urlFormat string, a ...interface{}) error {
	url := fmt.Sprintf(urlFormat, a...)
	u := fmt.Sprintf("%s/api%s", c.baseURL, url)
	r, err := http.Get(u)
	if r != nil && r.Body != nil {
		defer r.Body.Close()
	}
	if err != nil {
		return err
	}
	if r.StatusCode == 404 {
		return nil
	}
	if r.StatusCode != 200 {
		var body string
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			body = fmt.Sprintf("!!! unable to read response body: %s", err)
		}
		body = string(b)
		return fmt.Errorf("GET %s responded %d:\n%s", url, r.StatusCode, body)
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
