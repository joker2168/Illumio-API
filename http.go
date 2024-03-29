package illumioapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// APIResponse contains the information from the response of the API
type APIResponse struct {
	RespBody   string
	StatusCode int
	Header     http.Header
	Request    *http.Request
}

// PCE represents an Illumio PCE and the necessary info to authenticate
type PCE struct {
	FQDN               string
	Port               int
	Org                int
	User               string
	Key                string
	DisableTLSChecking bool
}

// Unexported struct for handling the asyncResults
type asyncResults struct {
	Href        string `json:"href"`
	JobType     string `json:"job_type"`
	Description string `json:"description"`
	Result      struct {
		Href string `json:"href"`
	} `json:"result"`
	Status       string `json:"status"`
	RequestedAt  string `json:"requested_at"`
	TerminatedAt string `json:"terminated_at"`
	RequestedBy  struct {
		Href string `json:"href"`
	} `json:"requested_by"`
}

// API does the HTTP set up (including polling for async calls) and returns an APIResponse type.
// httpAction must be GET, POST, PUT, or DELETE.
// apiURL is the full endpoint being called.
// PUT and POST methods should have a body that is JSON run through the json.marshal function so it's a []byte.
// async parameter should be set to true for any GET requests returning > 500 items.
func apicall(httpAction, apiURL string, pce PCE, body []byte, async bool) (APIResponse, error) {

	var response APIResponse
	var httpBody *bytes.Buffer
	var asyncResults asyncResults

	// Validate the provided action
	httpAction = strings.ToUpper(httpAction)
	if httpAction != "GET" && httpAction != "POST" && httpAction != "PUT" && httpAction != "DELETE" {
		return response, errors.New("invalid http action string. action must be GET, POST, PUT, or DELETE")
	}

	// Get the base URL
	u, err := url.Parse(apiURL)
	baseURL := "https://" + u.Host + "/api/v1"

	// Create body
	httpBody = bytes.NewBuffer(body)

	// Create HTTP client and request
	client := &http.Client{}
	if pce.DisableTLSChecking == true {
		client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}

	req, err := http.NewRequest(httpAction, apiURL, httpBody)
	if err != nil {
		return response, err
	}

	// Set basic authentication and headers
	req.SetBasicAuth(pce.User, pce.Key)
	req.Header.Set("Content-Type", "application/json")
	if async == true {
		req.Header.Set("Prefer", "respond-async")
	}

	// Make HTTP Request
	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}

	// Process Async requests
	if async == true {
		for asyncResults.Status != "done" {
			asyncResults, err = polling(baseURL, pce, resp)
			if err != nil {
				return response, err
			}
		}

		finalReq, err := http.NewRequest("GET", baseURL+asyncResults.Result.Href, httpBody)
		if err != nil {
			return response, err
		}

		// Set basic authentication and headers
		finalReq.SetBasicAuth(pce.User, pce.Key)
		finalReq.Header.Set("Content-Type", "application/json")

		// Make HTTP Request
		resp, err = client.Do(finalReq)
		if err != nil {
			return response, err
		}
	}

	// Process response
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	// Put relevant response info into struct
	response.RespBody = string(data[:])
	response.StatusCode = resp.StatusCode
	response.Header = resp.Header
	response.Request = resp.Request

	// Check for a 200 response code
	if strconv.Itoa(resp.StatusCode)[0:1] != "2" {
		return response, errors.New("http status code of " + strconv.Itoa(response.StatusCode))
	}

	// Return data and nil error
	return response, nil
}

func polling(baseURL string, pce PCE, origResp *http.Response) (asyncResults, error) {

	var asyncResults asyncResults

	// Create HTTP client and request
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	pollReq, err := http.NewRequest("GET", baseURL+origResp.Header.Get("Location"), nil)
	if err != nil {
		return asyncResults, err
	}

	// Set basic authentication and headers
	pollReq.SetBasicAuth(pce.User, pce.Key)
	pollReq.Header.Set("Content-Type", "application/json")

	// Wait for recommended time from Retry-After
	wait, err := strconv.Atoi(origResp.Header.Get("Retry-After"))
	if err != nil {
		return asyncResults, err
	}
	duration := time.Duration(wait) * time.Second
	time.Sleep(duration)

	// Check if the data is ready
	pollResp, err := client.Do(pollReq)
	if err != nil {
		return asyncResults, err
	}

	// Process Response
	data, err := ioutil.ReadAll(pollResp.Body)
	if err != nil {
		return asyncResults, err
	}

	// Put relevant response info into struct
	json.Unmarshal(data[:], &asyncResults)

	return asyncResults, err
}

// pceSanitization cleans up the provided PCE FQDN in case of common errors
func pceSanitization(pceFQDN string) string {

	// Remove trailing slash if included
	if strings.HasSuffix(pceFQDN, "/") {
		pceFQDN = pceFQDN[:len(pceFQDN)-1]
	}

	// Remove HTTPS if included
	if strings.HasPrefix(pceFQDN, "https://") {
		pceFQDN = strings.Replace(pceFQDN, "https://", "", -1)
	}

	return pceFQDN

}
