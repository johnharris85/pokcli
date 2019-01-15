package client

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/skratchdot/open-golang/open"
	"net/http"
	"time"
)

var (
	localRedirectURLTmpl  = "http://localhost:%s/auth"
	pocketRedirectURLTmpl = "https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s"
)

const (
	pocketAuthURL        = "https://getpocket.com/v3/oauth/request"
	pocketAccessTokenURL = "https://getpocket.com/v3/oauth/authorize"
)

type requestTokenRequest struct {
	ConsumerKey string `json:"consumer_key"`
	RedirectURL string `json:"redirect_uri"`
}

type requestTokenResponse struct {
	Code string `json:"code"`
}

type accessTokenRequest struct {
	ConsumerKey string `json:"consumer_key"`
	Code        string `json:"code"`
}

type accessTokenResponse struct {
	Username    string `json:"username"`
	AccessToken string `json:"access_token"`
}

func (c *Client) authorize() error {
	localRedirectURL := fmt.Sprintf(localRedirectURLTmpl, c.callbackPort)
	requestToken, err := c.getRequestToken(localRedirectURL)
	if err != nil {
		return err
	}
	successChan, stopHTTPServerChan, cancelAuthentication := startHTTPServer(context.Background(), c.callbackPort)
	log.Println("You will now be taken to your browser for authentication")
	pocketRedirectURL := fmt.Sprintf(pocketRedirectURLTmpl, requestToken, localRedirectURL)
	log.Printf("Authentication URL: %s\n", pocketRedirectURL)
	time.Sleep(1 * time.Second)
	if err = open.Run(pocketRedirectURL); err != nil {
		return err
	}

	go func() {
		log.Println("authentication will be cancelled in 40 seconds")
		time.Sleep(40 * time.Second)
		stopHTTPServerChan <- struct{}{}
	}()

	select {
	case <-successChan:
		// After the callbackHandler returns a client, it's time to shutdown the server gracefully
		stopHTTPServerChan <- struct{}{}

		// if authentication process is cancelled first return an error
	case <-cancelAuthentication:
		log.Fatalf("authentication timed out and was cancelled")
	}

	accessToken, err := c.getAccessToken(requestToken)
	if err != nil {
		return err
	}
	c.accessToken = accessToken
	return nil
}

func (c *Client) getRequestToken(redirectURI string) (string, error) {
	requestData := requestTokenRequest{ConsumerKey: c.consumerKey, RedirectURL: redirectURI}
	reqJSON, _ := json.Marshal(requestData)
	response, err := c.postRequest(pocketAuthURL, reqJSON)
	if err != nil {
		return "", err
	}
	var tokenResponse requestTokenResponse
	if err := json.Unmarshal(response, &tokenResponse); err != nil {
		return "", err
	}
	return tokenResponse.Code, nil
}

func (c *Client) getAccessToken(requestToken string) (string, error) {
	requestData := accessTokenRequest{ConsumerKey: c.consumerKey, Code: requestToken}
	reqJSON, _ := json.Marshal(requestData)
	response, err := c.postRequest(pocketAccessTokenURL, reqJSON)
	if err != nil {
		return "", err
	}

	var aTokenResponse accessTokenResponse
	if err := json.Unmarshal(response, &aTokenResponse); err != nil {
		return "", err
	}
	return aTokenResponse.AccessToken, nil
}

func startHTTPServer(ctx context.Context, port string) (chan struct{}, chan struct{}, chan struct{}) {
	// init returns
	successChan := make(chan struct{})
	stopHTTPServerChan := make(chan struct{})
	cancelAuthentication := make(chan struct{})

	http.HandleFunc("/auth", callbackHandler(ctx, successChan))
	srv := &http.Server{Addr: ":" + port}

	// handle server shutdown signal
	go func() {
		// wait for signal on stopHTTPServerChan
		<-stopHTTPServerChan
		log.Println("Shutting down server...")

		// give it 5 sec to shutdown gracefully, else quit program
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("could not shutdown gracefully: %v", err)
		}

		// after server is shutdown, quit program
		cancelAuthentication <- struct{}{}
	}()

	// handle callback request
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
		fmt.Println("Server gracefully stopped")
	}()

	return successChan, stopHTTPServerChan, cancelAuthentication
}

func callbackHandler(ctx context.Context, success chan struct{}) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		successPage := `
		<div style="height:100px; width:100%!; display:flex; flex-direction: column; justify-content: center; align-items:center; background-color:#2ecc71; color:white; font-size:22"><div>Success!</div></div>
		<p style="margin-top:20px; font-size:18; text-align:center">You are authenticated, you can now close this window and return to the CLI.</p>
		`
		fmt.Fprintf(w, successPage)
		success <- struct{}{}
	}
}
