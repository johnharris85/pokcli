package client

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

// POST data to the server and return the response as bytes
func (c *Client) postRequest(url string, d []byte) ([]byte, error) {

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(d))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json; charset=UTF8")
	req.Header.Add("X-Accept", "application/json")

	bytes, err := c.doRequest(req)
	if err != nil {
		return bytes, err
	}
	return bytes, nil
}

// GET data from the server and return the response as bytes
func (c *Client) getRequest(url string, d []byte) ([]byte, error) {

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(d))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	bytes, err := c.doRequest(req)
	if err != nil {
		return bytes, err
	}
	return bytes, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 2xx Success / 3xx Redirection
	if resp.StatusCode < 400 {
		log.Debugf("[success] HTTP Status code %d", resp.StatusCode)
		return body, nil
	}
	// The error code is > 400
	log.Debugf("HTTP Error code: %d for URL: %s", resp.StatusCode, req.URL.String())

	// Catches the "Majority" of expected responses
	switch resp.StatusCode {
	case 400:
		return body, fmt.Errorf("Code %d, Bad Request", resp.StatusCode)
	case 401:
		return body, fmt.Errorf("Code %d, Unauthorised", resp.StatusCode)
	case 402:
		return body, fmt.Errorf("Code %d, Payment Required", resp.StatusCode) //unused
	case 403:
		return body, fmt.Errorf("Code %d, Forbidden", resp.StatusCode)
	case 404:
		return body, fmt.Errorf("Code %d, Not Found", resp.StatusCode)
	case 405:
		return body, fmt.Errorf("Code %d, Method Not Allowed", resp.StatusCode)
	case 500:
		return body, fmt.Errorf("Code %d, Internal Server Error", resp.StatusCode)
	case 501:
		return body, fmt.Errorf("Code %d, Not Implemented", resp.StatusCode)
	case 502:
		return body, fmt.Errorf("Code %d, Bad Gateway", resp.StatusCode)
	case 503:
		return body, fmt.Errorf("Code %d, Service Unavailable", resp.StatusCode)
	case 504:
		return body, fmt.Errorf("Code %d, Gateway Timeout", resp.StatusCode)
	default:
		log.Debugf("[Untrapped return code] %d", resp.StatusCode)
		return body, fmt.Errorf("Code %s", resp.Status)

	}
	/*
		Status Code, X-Error-Code, X-Error
		400	138	Missing consumer key.
		400	140	Missing redirect url.
		400	181	Invalid redirect uri.
		400	182	Missing code.
		400	185	Code not found.
		403	152	Invalid consumer key.
		403	158	User rejected code.
		403	159	Already used code.
		50X	199	Pocket server issue.
	*/

}
