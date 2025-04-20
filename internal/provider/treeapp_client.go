package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const API_BASE string = "http://127.0.0.1:5000"

// const API_BASE string = "https://api.thetreeapp.org"

type TreeappClient struct {
	APIKey string
	Client *http.Client
}

func NewTreeappClient(apiKey string) *TreeappClient {
	return &TreeappClient{
		APIKey: apiKey,
		Client: &http.Client{},
	}
}

type UsageRecordRequest struct {
	Quantity int `json:"quantity"`
}

type UsageRecordResponse struct {
	ID             string `json:"id"`
	Quantity       int    `json:"quantity"`
	PaymentProfile string `json:"payment_profile_id"`
	CreatedAt      int64  `json:"created_at"`
}

func (c *TreeappClient) CreateUsageRecord(quantity int, idempotencyKey string) (*UsageRecordResponse, error) {
	reqBody, err := json.Marshal(UsageRecordRequest{Quantity: quantity})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", API_BASE+"/v1/usage-records", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Treeapp-Api-Key", c.APIKey)
	req.Header.Set("Idempotency-Key", idempotencyKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var result UsageRecordResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

type ImpactSummary struct {
	Trees    int64 `json:"trees"`
	Unbilled struct {
		Trees int64 `json:"trees"`
	} `json:"unbilled"`
}

func (c *TreeappClient) GetImpactSummary() (*ImpactSummary, error) {
	req, err := http.NewRequest("GET", API_BASE+"/v1.1/impacts/summary", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Treeapp-Api-Key", c.APIKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var summary ImpactSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, err
	}

	return &summary, nil
}

func (c *TreeappClient) GetTotalNumberOfTrees() (*int64, error) {
	req, err := http.NewRequest("GET", API_BASE+"/v1.1/impacts/summary", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Treeapp-Api-Key", c.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch data from Treeapp API")
	}

	var data ImpactSummary
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	total := data.Trees + data.Unbilled.Trees
	return &total, nil
}
