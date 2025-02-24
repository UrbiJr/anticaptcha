package anticaptcha

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"
)

var (
	baseURL       = &url.URL{Host: "api.anti-captcha.com", Scheme: "https", Path: "/"}
	checkInterval = 2 * time.Second
)

// Client : A client used to communicate with the AntiCaptcha API
type Client struct {
	APIKey string
}

// GeeTestSolution : Holds the solution variables of a GeeTest solve
type GeeTestSolution struct {
	Challenge string `json:"challenge"`
	Validate  string `json:"validate"`
	Seccode   string `json:"seccode"`
}

// NewClient : Returns an AntiCaptcha client
func NewClient(APIKey string) *Client {
	return &Client{APIKey: APIKey}
}

// Method to create the task to process the recaptcha v2, returns the task_id
func (c *Client) createTaskRecaptchaV2(websiteURL string, recaptchaKey string) (float64, error) {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"task": map[string]interface{}{
			"type":       "NoCaptchaTaskProxyless",
			"websiteURL": websiteURL,
			"websiteKey": recaptchaKey,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/createTask"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	// TODO treat api errors and handle them properly
	if _, ok := responseBody["taskId"]; ok {
		if taskID, ok := responseBody["taskId"].(float64); ok {
			return taskID, nil
		}

		return 0, errors.New("task number of irregular format")
	}

	return 0, errors.New("task number not found in server response")
}

// Method to create the task to process the recaptcha v3, returns the task_id
func (c *Client) createTaskRecaptchaV3(websiteURL string, recaptchaKey string, minimumScore float32, pageAction string) (float64, error) {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"task": map[string]interface{}{
			"type":       "RecaptchaV3TaskProxyless",
			"websiteURL": websiteURL,
			"websiteKey": recaptchaKey,
			"minScore":   minimumScore,
			"pageAction": pageAction,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/createTask"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	// TODO treat api errors and handle them properly
	if _, ok := responseBody["taskId"]; ok {
		if taskID, ok := responseBody["taskId"].(float64); ok {
			return taskID, nil
		}

		return 0, errors.New("task number of irregular format")
	}

	return 0, errors.New("task number not found in server response")
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

// Method to create the task to process a GeeTest captcha, returns the task_id
func (c *Client) createTaskGeeTest(websiteURL string, geeTestGT string, geeTestChallenge string, geeTestAPIServerSubdomain string) (float64, error) {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"task": map[string]interface{}{
			"type":                      "GeeTestTaskProxyless",
			"websiteURL":                websiteURL,
			"gt":                        geeTestGT,
			"challenge":                 geeTestChallenge,
			"geetestApiServerSubdomain": geeTestAPIServerSubdomain,
		},
	}

	b, err := JSONMarshal(body)
	if err != nil {
		return 0, err
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/createTask"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	// TODO treat api errors and handle them properly
	if _, ok := responseBody["taskId"]; ok {
		if taskID, ok := responseBody["taskId"].(float64); ok {
			return taskID, nil
		}

		return 0, errors.New("task number of irregular format")
	}

	return 0, errors.New("task number not found in server response")
}

// Method to check the result of a given task, returns the json returned from the api
func (c *Client) getTaskResult(taskID float64) (map[string]interface{}, error) {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"taskId":    taskID,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/getTaskResult"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	return responseBody, nil
}

// SendRecaptchaV2 Method to encapsulate the processing of the recaptcha
// Given a url and a key, it sends to the api and waits until
// the processing is complete to return the evaluated key
func (c *Client) SendRecaptchaV2(websiteURL string, recaptchaKey string, timeoutInterval time.Duration) (string, error) {
	taskID, err := c.createTaskRecaptchaV2(websiteURL, recaptchaKey)
	if err != nil {
		return "", err
	}

	check := time.NewTicker(10 * time.Second)
	timeout := time.NewTimer(timeoutInterval)

	for {
		select {
		case <-check.C:
			response, err := c.getTaskResult(taskID)
			if err != nil {
				return "", err
			}
			if response["status"] == "ready" {
				return response["solution"].(map[string]interface{})["gRecaptchaResponse"].(string), nil
			}
			check = time.NewTicker(checkInterval)
		case <-timeout.C:
			return "", errors.New("antiCaptcha check result timeout")
		}
	}
}

// SendRecaptchaV3 Method to encapsulate the processing of the recaptcha
// Given a url and a key, it sends to the api and waits until
// the processing is complete to return the evaluated key
func (c *Client) SendRecaptchaV3(websiteURL string, recaptchaKey string, minimumScore float32, pageAction string, timeoutInterval time.Duration) (string, error) {
	taskID, err := c.createTaskRecaptchaV3(websiteURL, recaptchaKey, minimumScore, pageAction)
	if err != nil {
		return "", err
	}

	check := time.NewTicker(10 * time.Second)
	timeout := time.NewTimer(timeoutInterval)

	for {
		select {
		case <-check.C:
			response, err := c.getTaskResult(taskID)
			if err != nil {
				return "", err
			}
			if response["status"] == "ready" {
				return response["solution"].(map[string]interface{})["gRecaptchaResponse"].(string), nil
			}
			check = time.NewTicker(checkInterval)
		case <-timeout.C:
			return "", errors.New("antiCaptcha check result timeout")
		}
	}
}

// SendGeeTest Method to encapsulate the processing of the recaptcha
// Given a url and a key, it sends to the api and waits until
// the processing is complete to return the evaluated key
func (c *Client) SendGeeTest(websiteURL string, geeTestGT string, geeTestChallenge string, geeTestAPIServerSubdomain string, timeoutInterval time.Duration) (GeeTestSolution, error) {
	taskID, err := c.createTaskGeeTest(websiteURL, geeTestGT, geeTestChallenge, geeTestAPIServerSubdomain)
	if err != nil {
		return GeeTestSolution{}, err
	}

	check := time.NewTicker(10 * time.Second)
	timeout := time.NewTimer(timeoutInterval)

	for {
		select {
		case <-check.C:
			response, err := c.getTaskResult(taskID)
			if err != nil {
				return GeeTestSolution{}, err
			}
			if response["status"] == "ready" {
				solution := response["solution"].(map[string]interface{})
				geeTestSolution := GeeTestSolution{
					Challenge: solution["challenge"].(string),
					Validate:  solution["validate"].(string),
					Seccode:   solution["seccode"].(string),
				}
				return geeTestSolution, nil
			}
			check = time.NewTicker(checkInterval)
		case <-timeout.C:
			return GeeTestSolution{}, errors.New("antiCaptcha check result timeout")
		}
	}
}

// Method to create the task to process the image captcha, returns the task_id
func (c *Client) createTaskImage(imgString string) (float64, error) {
	// Mount the data to be sent
	body := map[string]interface{}{
		"clientKey": c.APIKey,
		"task": map[string]interface{}{
			"type": "ImageToTextTask",
			"body": imgString,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	// Make the request
	u := baseURL.ResolveReference(&url.URL{Path: "/createTask"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Decode response
	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	// TODO treat api errors and handle them properly
	return responseBody["taskId"].(float64), nil
}

// SendImage Method to encapsulate the processing of the image captcha
// Given a base64 string from the image, it sends to the api and waits until
// the processing is complete to return the evaluated key
func (c *Client) SendImage(imgString string) (string, error) {
	// Create the task on anti-captcha api and get the task_id
	taskID, err := c.createTaskImage(imgString)
	if err != nil {
		return "", err
	}

	// Check if the result is ready, if not loop until it is
	response, err := c.getTaskResult(taskID)
	if err != nil {
		return "", err
	}
	for {
		if response["status"] == "processing" {
			time.Sleep(checkInterval)
			response, err = c.getTaskResult(taskID)
			if err != nil {
				return "", err
			}
		} else {
			break
		}
	}
	return response["solution"].(map[string]interface{})["text"].(string), nil
}
