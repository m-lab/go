package siteinfo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	baseURLFormat = "https://siteinfo.%s.measurementlab.net/%s/"
)

// HTTPProvider is a data provider returning HTTP responses.
// http.Client satisfies this interface.
type HTTPProvider interface {
	Get(string) (*http.Response, error)
}

// Client is a Siteinfo client.
type Client struct {
	ProjectID  string
	Version    string
	httpClient HTTPProvider
}

// New returns a new Siteinfo client wrapping the provided *http.Client.
func New(projectID, version string, httpClient HTTPProvider) *Client {
	return &Client{
		ProjectID:  projectID,
		httpClient: httpClient,
		Version:    version,
	}
}

func (c Client) getContent(url string) ([]byte, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// Switches fetches the sites/switches.json output format and returns its
// content as a map[site]Switch.
func (c Client) Switches() (map[string]Switch, error) {
	url := c.makeBaseURL() + "sites/switches.json"
	body, err := c.getContent(url)
	if err != nil {
		return nil, err
	}
	res := make(map[string]Switch)
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Projects fetches the sites/projects.json output format and returns its
// content as a map[<short-node-name>]string.
func (c Client) Projects() (map[string]string, error) {
	url := c.makeBaseURL() + "sites/projects.json"
	body, err := c.getContent(url)
	if err != nil {
		return nil, err
	}
	res := make(map[string]string)
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Machines fetches the sites/machines.json output format and returns its
// content as a []Machine.
func (c Client) Machines() ([]Machine, error) {
	url := c.makeBaseURL() + "sites/machines.json"
	body, err := c.getContent(url)
	if err != nil {
		return nil, err
	}
	res := []Machine{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s Client) makeBaseURL() string {
	return fmt.Sprintf(baseURLFormat, s.ProjectID, s.Version)
}
