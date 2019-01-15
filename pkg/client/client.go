package client

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"net/http"
	"os"
)

// DefaultLocalCallbackPort is the default port for the callback oauth server to listen on
const DefaultLocalCallbackPort = "8181"

// Client represents a pokcli client
type Client struct {
	client          *http.Client
	consumerKey     string
	callbackPort    string
	accessToken     string
	optMap          map[string]bool
	credentialsFile string
}

type localCredentials struct {
	ConsumerKey string `toml:"consumer_key,omitempty"`
	AccessToken string `toml:"access_token,omitempty"`
}

// NewClientWithOpts returns a new pokcli client
func NewClientWithOpts(options ...func(*Client) error) (*Client, error) {
	var optMap = make(map[string]bool)
	h := &http.Client{
		Transport: new(http.Transport),
	}
	c := &Client{
		client:       h,
		callbackPort: DefaultLocalCallbackPort,
		optMap:       optMap,
		accessToken:  "",
		consumerKey:  "",
	}

	incompatibleOpts := [][]string{
		{"WithCredsFile", "WithAccessToken"},
		{"WithCredsFile", "WithConsumerKey"},
	}
	if err := validateOpts(incompatibleOpts, c.optMap); err != nil {
		return nil, err
	}

	for _, op := range options {
		if err := op(c); err != nil {
			return nil, err
		}
	}

	if _, ok := c.client.Transport.(http.RoundTripper); !ok {
		return nil, fmt.Errorf("unable to verify TLS configuration, invalid transport %v", c.client.Transport)
	}

	if c.accessToken == "" {
		if err := c.authorize(); err != nil {
			return nil, err
		}
	}

	if _, ok := c.optMap["WithCredsFile"]; ok {
		if err := c.writeUpdatedCredsFile(); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// WithCredsFile ...
func WithCredsFile(filePath string) func(*Client) error {
	return func(c *Client) error {
		if filePath == "" {
			return errors.New("no credentials file supplied")
		}

		var localCreds localCredentials
		if _, err := toml.DecodeFile(filePath, &localCreds); err != nil {
			return err
		}

		if localCreds.ConsumerKey == "" {
			return errors.New("can't find consumer_key in credentials file")
		}

		c.consumerKey = localCreds.ConsumerKey
		c.accessToken = localCreds.AccessToken
		c.optMap["WithCredsFile"] = true
		c.credentialsFile = filePath
		return nil
	}
}

// WithHTTPClient overrides the client http client with the specified one
func WithHTTPClient(client *http.Client) func(*Client) error {
	return func(c *Client) error {
		if client != nil {
			c.client = client
		}
		return nil
	}
}

// WithAccessToken overrides the client http client with the specified one
func WithAccessToken(token string) func(*Client) error {
	return func(c *Client) error {
		if token != "" {
			c.accessToken = token
		}
		c.optMap["WithAccessToken"] = true
		return nil
	}
}

// WithConsumerKey overrides the client http client with the specified one
func WithConsumerKey(consumerKey string) func(*Client) error {
	return func(c *Client) error {
		if consumerKey != "" {
			c.consumerKey = consumerKey
		}
		c.optMap["WithConsumerKey"] = true
		return nil
	}
}

// WithCustomCallbackPort overrides the default callback server port
func WithCustomCallbackPort(port string) func(*Client) error {
	return func(c *Client) error {
		if port != "" {
			c.callbackPort = port
		}
		return nil
	}
}

// HTTPClient returns a copy of the HTTP client bound to the server
func (c *Client) HTTPClient() *http.Client {
	return c.client
}

func (c *Client) writeUpdatedCredsFile() error {
	f, err := os.OpenFile(c.credentialsFile, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()

	e := toml.NewEncoder(f)
	localCreds := localCredentials{
		ConsumerKey: c.consumerKey,
		AccessToken: c.accessToken,
	}
	if err = e.Encode(localCreds); err != nil {
		return err
	}
	return nil
}

func validateOpts(incompatibleOpts [][]string, optMap map[string]bool) error {
	for _, o := range incompatibleOpts {
		if _, ok0 := optMap[o[0]]; ok0 {
			if _, ok1 := optMap[o[1]]; ok1 {
				return fmt.Errorf("incompatible options %s & %s", o[0], o[1])
			}
		}
	}
	return nil
}
