package client

import (
	"encoding/json"
)

// GetArticles ...
func (c *Client) GetArticles(tag string) []byte {
	a := map[string]string{
		"consumer_key": c.consumerKey,
		"access_token": c.accessToken,
		"tag":          tag,
	}
	reqJSON, _ := json.Marshal(a)
	response, _ := c.postRequest("https://getpocket.com/v3/get", reqJSON)
	return response
}
