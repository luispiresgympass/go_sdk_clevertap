package clevertap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// Service ...
type Service struct {
	cO Options
}

func (c *Service) setOptions(clevertapOptions Options) BuildClevertap {
	c.cO = clevertapOptions
	return c
}

// SendEvent ...
func (c *Service) SendEvent(identity string, evtName string, evtData map[string]interface{}, responseInterface interface{}) error {
	sendEventReq := []SendEventRequest{
		{
			Identity:  identity,
			EventName: evtName,
			Type:      Event,
			Timestamp: time.Now().Unix(),
			EventData: evtData,
		},
	}

	payload := make(map[string]interface{})
	payload["d"] = sendEventReq

	req, err := c.newRequest(http.MethodPost, ClevertapSendEventURL, payload)
	if err != nil {
		return err
	}

	resp, err := c.do(req, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		body := buf.String()

		err := fmt.Errorf("non 200 response - %s", body)
		return err
	}

	return nil
}

func (c *Service) SendProfile(identity string, profileData map[string]interface{}) error {
	sendProfileReq := []SendProfileRequest{
		{
			Identity:    identity,
			Type:        Profile,
			Timestamp:   time.Now().Unix(),
			ProfileData: profileData,
		},
	}

	payload := make(map[string]interface{})
	payload["d"] = sendProfileReq

	req, err := c.newRequest(http.MethodPost, ClevertapSendEventURL, payload)
	if err != nil {
		return err
	}

	resp, err := c.do(req, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errors.New("non 200 response")
	}

	return nil
}

func (c *Service) newRequest(method, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.cO.baseURL.ResolveReference(rel)
	var buf io.ReadWriter

	if body != nil {
		buf = new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}

	if req, err := http.NewRequest(method, u.String(), buf); err != nil {
		return nil, err
	} else {
		req.Header.Set(ContentType, ApplicationJSONCharsetUtf8)
		req.Header.Set(XClevertapAccountID, c.cO.AccountID)
		req.Header.Set(XClevertapPasscode, c.cO.Passcode)
		return req, nil
	}
}

func (c *Service) do(req *http.Request, v interface{}) (*http.Response, error) {
	if resp, err := c.cO.httpClient.Do(req); err != nil {
		return nil, err
	} else {
		defer func() {
			_ = resp.Body.Close()
		}()

		if bodyByte, err := ioutil.ReadAll(resp.Body); err != nil {
			return resp, err
		} else {
			return resp, json.Unmarshal(bodyByte, &v)
		}
	}
}
