package checkly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// NewClient takes a Checkly API key, and returns a Client ready to use.
func NewClient(apiKey string) Client {
	return Client{
		apiKey:     apiKey,
		URL:        getEnv("CHECKLY_API_URL", "https://api.checklyhq.com"),
		HTTPClient: http.DefaultClient,
	}
}

// Create creates a new check with the specified details. It returns the
// newly-created check, or an error.
func (c *Client) Create(check Check) (Check, error) {
	data, err := json.Marshal(check)
	if err != nil {
		return Check{}, err
	}
	status, res, err := c.MakeAPICall(
		http.MethodPost,
		withAutoAssignAlertsFlag("checks"),
		data,
	)
	if err != nil {
		return Check{}, err
	}
	if status != http.StatusCreated {
		return Check{}, fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	var result Check
	if err = json.NewDecoder(strings.NewReader(res)).Decode(&result); err != nil {
		return Check{}, fmt.Errorf("decoding error for data %s: %v", res, err)
	}
	return result, nil
}

// Update updates an existing check with the specified details. It returns the
// updated check, or an error.
func (c *Client) Update(ID string, check Check) (Check, error) {
	data, err := json.Marshal(check)
	if err != nil {
		return Check{}, err
	}
	status, res, err := c.MakeAPICall(
		http.MethodPut,
		withAutoAssignAlertsFlag(fmt.Sprintf("checks/%s", ID)),
		data,
	)
	if err != nil {
		return Check{}, err
	}
	if status != http.StatusOK {
		return Check{}, fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	var result Check
	if err = json.NewDecoder(strings.NewReader(res)).Decode(&result); err != nil {
		return Check{}, fmt.Errorf("decoding error for data %s: %v", res, err)
	}
	return result, nil
}

// Delete deletes the check with the specified ID. It returns a non-nil
// error if the request failed.
func (c *Client) Delete(ID string) error {
	status, res, err := c.MakeAPICall(http.MethodDelete, fmt.Sprintf("checks/%s", ID), nil)
	if err != nil {
		return err
	}
	if status != http.StatusNoContent {
		return fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	return nil
}

// Get takes the ID of an existing check, and returns the check parameters, or
// an error.
func (c *Client) Get(ID string) (Check, error) {
	status, res, err := c.MakeAPICall(http.MethodGet, fmt.Sprintf("checks/%s", ID), nil)
	if err != nil {
		return Check{}, err
	}
	if status != http.StatusOK {
		return Check{}, fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	check := Check{}
	if err = json.NewDecoder(strings.NewReader(res)).Decode(&check); err != nil {
		return Check{}, fmt.Errorf("decoding error for data %s: %v", res, err)
	}
	return check, nil
}

// CreateGroup creates a new check group with the specified details. It returns
// the newly-created group, or an error.
func (c *Client) CreateGroup(group Group) (Group, error) {
	data, err := json.Marshal(group)
	if err != nil {
		return Group{}, err
	}
	status, res, err := c.MakeAPICall(
		http.MethodPost,
		withAutoAssignAlertsFlag("check-groups"),
		data,
	)
	if err != nil {
		return Group{}, err
	}
	if status != http.StatusCreated {
		return Group{}, fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	var result Group
	if err = json.NewDecoder(strings.NewReader(res)).Decode(&result); err != nil {
		return Group{}, fmt.Errorf("decoding error for data %s: %v", res, err)
	}
	return result, nil
}

// GetGroup takes the ID of an existing check group, and returns the
// corresponding group, or an error.
func (c *Client) GetGroup(ID int64) (Group, error) {
	status, res, err := c.MakeAPICall(http.MethodGet, fmt.Sprintf("check-groups/%d", ID), nil)
	if err != nil {
		return Group{}, err
	}
	if status != http.StatusOK {
		return Group{}, fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	group := Group{}
	if err = json.NewDecoder(strings.NewReader(res)).Decode(&group); err != nil {
		return Group{}, fmt.Errorf("decoding error for data %q: %v", res, err)
	}
	return group, nil
}

// UpdateGroup takes the ID of an existing check group, and updates the
// corresponding check group to match the supplied group. It returns the updated
// group, or an error.
func (c *Client) UpdateGroup(ID int64, group Group) (Group, error) {
	data, err := json.Marshal(group)
	if err != nil {
		return Group{}, err
	}
	status, res, err := c.MakeAPICall(
		http.MethodPut,
		withAutoAssignAlertsFlag(fmt.Sprintf("check-groups/%d", ID)),
		data,
	)
	if err != nil {
		return Group{}, err
	}
	if status != http.StatusOK {
		return Group{}, fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	var result Group
	if err = json.NewDecoder(strings.NewReader(res)).Decode(&result); err != nil {
		return Group{}, fmt.Errorf("decoding error for data %s: %v", res, err)
	}
	return result, nil
}

// DeleteGroup deletes the check group with the specified ID. It returns a
// non-nil error if the request failed.
func (c *Client) DeleteGroup(ID int64) error {
	status, res, err := c.MakeAPICall(http.MethodDelete, fmt.Sprintf("check-groups/%d", ID), nil)
	if err != nil {
		return err
	}
	if status != http.StatusNoContent {
		return fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	return nil
}

// GetCheckResult gets a specific Check result
func (c *Client) GetCheckResult(checkID, checkResultID string) (CheckResult, error) {
	status, res, err := c.MakeAPICall(
		http.MethodGet,
		fmt.Sprintf("check-results/%s/%s", checkID, checkResultID),
		nil,
	)
	result := CheckResult{}
	if err != nil {
		return result, err
	}
	if status != http.StatusOK {
		return result, fmt.Errorf("unexpected response status %d: %q", status, res)
	}

	err = json.NewDecoder(strings.NewReader(res)).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("decoding error for data %q: %v", res, err)
	}
	return result, nil
}

// GetCheckResults gets the results of the given Check
func (c *Client) GetCheckResults(
	checkID string,
	filters *CheckResultsFilter,
) ([]CheckResult, error) {
	uri := fmt.Sprintf("check-results/%s", checkID)
	if filters != nil {
		q := url.Values{}
		if filters.Page > 0 {
			q.Add("page", fmt.Sprintf("%d", filters.Page))
		}
		if filters.Limit > 0 {
			q.Add("limit", fmt.Sprintf("%d", filters.Limit))
		}
		if filters.From > 0 {
			q.Add("from", fmt.Sprintf("%d", filters.From))
		}
		if filters.To > 0 {
			q.Add("to", fmt.Sprintf("%d", filters.To))
		}
		if filters.CheckType == TypeBrowser || filters.CheckType == TypeAPI {
			q.Add("checkType", string(filters.CheckType))
		}
		if filters.HasFailures {
			q.Add("hasFailures", "1")
		}
		if len(filters.Location) > 0 {
			q.Add("location", filters.Location)
		}
		uri = uri + "?" + q.Encode()
	}

	status, res, err := c.MakeAPICall(
		http.MethodGet,
		uri,
		nil,
	)

	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	result := []CheckResult{}
	err = json.NewDecoder(strings.NewReader(res)).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("decoding error for data %q: %v", res, err)
	}
	return result, nil
}

// MakeAPICall calls the Checkly API with the specified URL and data, and
// returns the HTTP status code and string data of the response.
func (c *Client) MakeAPICall(method string, URL string, data []byte) (statusCode int, response string, err error) {
	requestURL := c.URL + "/v1/" + URL
	req, err := http.NewRequest(method, requestURL, bytes.NewBuffer(data))
	if err != nil {
		return 0, "", fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	req.Header.Add("content-type", "application/json")
	if c.Debug != nil {
		requestDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return 0, "", fmt.Errorf("error dumping HTTP request: %v", err)
		}
		fmt.Fprintln(c.Debug, string(requestDump))
		fmt.Fprintln(c.Debug)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()
	if c.Debug != nil {
		c.dumpResponse(resp)
	}
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}
	return resp.StatusCode, string(res), nil
}

// dumpResponse writes the raw response data to the debug output, if set, or
// standard error otherwise.
func (c *Client) dumpResponse(resp *http.Response) {
	// ignore errors dumping response - no recovery from this
	responseDump, _ := httputil.DumpResponse(resp, true)
	fmt.Fprintln(c.Debug, string(responseDump))
	fmt.Fprintln(c.Debug)
}

// CreateSnippet creates a new snippet with the specified details. It returns
// the newly-created snippet, or an error.
func (c *Client) CreateSnippet(snippet Snippet) (Snippet, error) {
	data, err := json.Marshal(snippet)
	if err != nil {
		return Snippet{}, err
	}
	status, res, err := c.MakeAPICall(http.MethodPost, "snippets", data)
	if err != nil {
		return Snippet{}, err
	}
	if status != http.StatusCreated {
		return Snippet{}, fmt.Errorf("unexpected response status: %d, res: %q", status, res)
	}
	var result Snippet
	err = json.NewDecoder(strings.NewReader(res)).Decode(&result)
	if err != nil {
		return Snippet{}, fmt.Errorf("decoding error for data %s: %v", res, err)
	}
	return result, nil
}

// GetSnippet takes the ID of an existing snippet, and returns the
// corresponding snippet, or an error.
func (c *Client) GetSnippet(ID int64) (Snippet, error) {
	status, res, err := c.MakeAPICall(http.MethodGet, fmt.Sprintf("snippets/%d", ID), nil)
	if err != nil {
		return Snippet{}, err
	}
	if status != http.StatusOK {
		return Snippet{}, fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	snippet := Snippet{}
	if err = json.NewDecoder(strings.NewReader(res)).Decode(&snippet); err != nil {
		return Snippet{}, fmt.Errorf("decoding error for data %q: %v", res, err)
	}
	return snippet, nil
}

// UpdateSnippet takes the ID of an existing snippet, and updates the
// corresponding snippet to match the supplied snippet. It returns the updated
// snippet, or an error.
func (c *Client) UpdateSnippet(ID int64, snippet Snippet) (Snippet, error) {
	data, err := json.Marshal(snippet)
	if err != nil {
		return Snippet{}, err
	}
	status, res, err := c.MakeAPICall(http.MethodPut, fmt.Sprintf("snippets/%d", ID), data)
	if err != nil {
		return Snippet{}, err
	}
	if status != http.StatusOK {
		return Snippet{}, fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	var result Snippet
	if err = json.NewDecoder(strings.NewReader(res)).Decode(&result); err != nil {
		return Snippet{}, fmt.Errorf("decoding error for data %s: %v", res, err)
	}
	return result, nil
}

// DeleteSnippet deletes the snippet with the specified ID. It returns a
// non-nil error if the request failed.
func (c *Client) DeleteSnippet(ID int64) error {
	status, res, err := c.MakeAPICall(http.MethodDelete, fmt.Sprintf("snippets/%d", ID), nil)
	if err != nil {
		return err
	}
	if status != http.StatusNoContent {
		return fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	return nil
}

// CreateEnvironmentVariable creates a new environment variable with the
// specified details.  It returns the newly-created environment variable,
// or an error.
func (c *Client) CreateEnvironmentVariable(envVar EnvironmentVariable) (EnvironmentVariable, error) {
	data, err := json.Marshal(envVar)
	if err != nil {
		return EnvironmentVariable{}, err
	}
	status, res, err := c.MakeAPICall(http.MethodPost, "variables", data)
	if err != nil {
		return EnvironmentVariable{}, err
	}
	if status != http.StatusCreated {
		return EnvironmentVariable{}, fmt.Errorf("unexpected response status: %d, res: %q", status, res)
	}
	var result EnvironmentVariable
	err = json.NewDecoder(strings.NewReader(res)).Decode(&result)
	if err != nil {
		return EnvironmentVariable{}, fmt.Errorf("decoding error for data %s: %v", res, err)
	}
	return result, nil
}

// GetEnvironmentVariable takes the ID of an existing environment variable, and returns the
// corresponding environment variable, or an error.
func (c *Client) GetEnvironmentVariable(key string) (EnvironmentVariable, error) {
	status, res, err := c.MakeAPICall(http.MethodGet, fmt.Sprintf("variables/%s", key), nil)
	if err != nil {
		return EnvironmentVariable{}, err
	}
	if status != http.StatusOK {
		return EnvironmentVariable{}, fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	envVar := EnvironmentVariable{}
	err = json.NewDecoder(strings.NewReader(res)).Decode(&envVar)
	if err != nil {
		return EnvironmentVariable{}, fmt.Errorf("decoding error for data %q: %v", res, err)
	}
	return envVar, nil
}

// UpdateEnvironmentVariable takes the ID of an existing environment variable, and updates the
// corresponding environment variable to match the supplied environment variable. It returns the updated
// environment variable, or an error.
func (c *Client) UpdateEnvironmentVariable(key string, envVar EnvironmentVariable) (EnvironmentVariable, error) {
	var result EnvironmentVariable
	data, err := json.Marshal(envVar)
	if err != nil {
		return result, err
	}

	status, res, err := c.MakeAPICall(http.MethodPut, fmt.Sprintf("variables/%s", key), data)
	if err != nil {
		return result, err
	}
	if status != http.StatusOK {
		return result, fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	err = json.NewDecoder(strings.NewReader(res)).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("decoding error for data %s: %v", res, err)
	}
	return result, nil
}

// DeleteEnvironmentVariable deletes the environment variable with the specified ID. It returns a
// non-nil error if the request failed.
func (c *Client) DeleteEnvironmentVariable(key string) error {
	status, res, err := c.MakeAPICall(http.MethodDelete, fmt.Sprintf("variables/%s", key), nil)
	if err != nil {
		return err
	}
	if status != http.StatusNoContent {
		return fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	return nil
}

// CreateAlertChannel creates a new alert channel with the specified details. It returns
// the newly-created alert channel, or an error.
func (c *Client) CreateAlertChannel(ac AlertChannel) (*AlertChannel, error) {
	payload := payloadFromAlertChannel(ac)
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	status, res, err := c.MakeAPICall(http.MethodPost, "alert-channels", data)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("unexpected response status: %d, res: %q, payload: %v", status, res, string(data))
	}
	return alertChannelFromJSON(res)
}

// GetAlertChannel takes the ID of an existing alert channel, and returns the
// corresponding alert channel, or an error.
func (c *Client) GetAlertChannel(ID int64) (*AlertChannel, error) {
	status, res, err := c.MakeAPICall(http.MethodGet, fmt.Sprintf("alert-channels/%d", ID), nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	result := map[string]interface{}{}
	if err = json.NewDecoder(strings.NewReader(res)).Decode(&result); err != nil {
		return nil, fmt.Errorf("GetAlertChannel: decoding error for data %q: %v", res, err)
	}
	return alertChannelFromJSON(res)
}

// UpdateAlertChannel takes the ID of an existing alert channel, and updates the
// corresponding alert channel to match the supplied alert channel. It returns the updated
// alert channel, or an error.
func (c *Client) UpdateAlertChannel(ID int64, ac AlertChannel) (*AlertChannel, error) {
	payload := payloadFromAlertChannel(ac)
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	status, res, err := c.MakeAPICall(http.MethodPut, fmt.Sprintf("alert-channels/%d", ID), data)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	return alertChannelFromJSON(res)
}

// DeleteAlertChannel deletes the alert channel with the specified ID. It returns a
// non-nil error if the request failed.
func (c *Client) DeleteAlertChannel(ID int64) error {
	status, res, err := c.MakeAPICall(http.MethodDelete, fmt.Sprintf("alert-channels/%d", ID), nil)
	if err != nil {
		return err
	}
	if status != http.StatusNoContent {
		return fmt.Errorf("unexpected response status %d: %q", status, res)
	}
	return nil
}

func payloadFromAlertChannel(ac AlertChannel) map[string]interface{} {
	payload := map[string]interface{}{
		"id":     ac.ID,
		"type":   ac.Type,
		"config": ac.GetConfig(),
	}
	if ac.SendRecovery != nil {
		payload["sendRecovery"] = *ac.SendRecovery
	}
	if ac.SendDegraded != nil {
		payload["sendDegraded"] = *ac.SendDegraded
	}
	if ac.SendFailure != nil {
		payload["sendFailure"] = *ac.SendFailure
	}
	if ac.SSLExpiry != nil {
		payload["sslExpiry"] = *ac.SSLExpiry
	}
	if ac.SSLExpiryThreshold != nil {
		payload["sslExpiryThreshold"] = *ac.SSLExpiryThreshold
	}
	return payload
}

func alertChannelFromJSON(response string) (*AlertChannel, error) {
	result := map[string]interface{}{}
	if err := json.NewDecoder(strings.NewReader(response)).Decode(&result); err != nil {
		return nil, fmt.Errorf("UpdateAlertChannel: decoding error for data Res(%s), Err(%w)", response, err)
	}
	resultAc := &AlertChannel{}
	if v, ok := result["id"]; ok {
		switch v.(type) {
		case int, int64:
			resultAc.ID = v.(int64)
		case float32, float64:
			resultAc.ID = int64(v.(float64))
		}
	}
	if v, ok := result["type"]; ok {
		resultAc.Type = v.(string)
	}
	if v, ok := result["sendRecovery"]; ok {
		sr := v.(bool)
		resultAc.SendRecovery = &sr
	}
	if v, ok := result["sendFailure"]; ok {
		sf := v.(bool)
		resultAc.SendFailure = &sf
	}
	if v, ok := result["sendDegraded"]; ok {
		sd := v.(bool)
		resultAc.SendDegraded = &sd
	}
	if v, ok := result["sslExpiry"]; ok {
		expiry := v.(bool)
		resultAc.SSLExpiry = &expiry
	}
	if v, ok := result["sslExpiryThreshold"]; ok {
		switch v.(type) {
		case int, int64:
			t := v.(int)
			resultAc.SSLExpiryThreshold = &t
		case float32, float64:
			t := int(v.(float64))
			resultAc.SSLExpiryThreshold = &t
		}
	}
	if cfg, ok := result["config"]; ok {
		cfgJSON, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		c, err := AlertChannelConfigFromJSON(resultAc.Type, cfgJSON)
		if err != nil {
			//TODO check this
			return nil, err
		}
		resultAc.SetConfig(c)
	}
	return resultAc, nil
}

func withAutoAssignAlertsFlag(url string) string {
	return url + "?autoAssignAlerts=true"
}
