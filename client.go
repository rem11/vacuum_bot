package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	BaseURLV2 = "http://127.0.0.1/api/v2"
)

type Client struct {
	BaseURL    string
	apiKey     string
	HTTPClient *http.Client
}

type PutBasicControlCapabilityRequest struct {
	Action string `json:"action"`
}

type GetZoneCleaningCapabilityPresetsResponse map[string]ValetudoZonePreset

type ZoneCleaningCapabilityPresetsResponse struct {
	Action string `json:"action"`
}

type ValetudoZonePreset struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type StateAttribute struct {
	Class string      `json:"__class"`
	Value interface{} `json:"value"`
	Level int         `json:"level"`
}

type GetStateAttributesResponse []StateAttribute

func NewClient() *Client {
	return &Client{
		BaseURL: BaseURLV2,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

func (c *Client) PutBasicControlCapability(ctx context.Context, action string) error {
	reqBodyJson := &PutBasicControlCapabilityRequest{
		Action: action,
	}
	reqBody, err := json.Marshal(reqBodyJson)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/robot/capabilities/BasicControlCapability", c.BaseURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	if err := c.sendRequest(req, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetZoneCleaningCapabilityPresets(ctx context.Context) (*GetZoneCleaningCapabilityPresetsResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/robot/capabilities/ZoneCleaningCapability/presets", c.BaseURL), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	resp := make(GetZoneCleaningCapabilityPresetsResponse)
	if err := c.sendRequest(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) PutZoneCleaningCapabilityPresets(ctx context.Context, id string) error {
	reqBodyJson := &PutBasicControlCapabilityRequest{
		Action: "clean",
	}
	reqBody, err := json.Marshal(reqBodyJson)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/robot/capabilities/ZoneCleaningCapability/presets/%s", c.BaseURL, id), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	if err := c.sendRequest(req, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) sendRequest(req *http.Request, resp interface{}) error {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("Recieved error with status code: %d", res.StatusCode)
	}
	if resp != nil {
		if err = json.NewDecoder(res.Body).Decode(resp); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) GetStateAttributes(ctx context.Context) (*GetStateAttributesResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/robot/state/attributes", c.BaseURL), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	resp := make(GetStateAttributesResponse, 0)
	if err := c.sendRequest(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
