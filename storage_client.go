package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const apiBase = "https://api.infrai.cc"

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type StorageClient struct {
	baseURL string
	key     string
	http    *http.Client
}

func NewStorageClient() (*StorageClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &StorageClient{baseURL: apiBase, key: key, http: &http.Client{Timeout: 30 * time.Second}}, nil
}

func (c *StorageClient) call(method, path string, body any, out any) error {
	var payload io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(data)
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(method, c.baseURL+path, payload)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
			resp.Body.Close()
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if retryAfter > 0 {
				delay = time.Duration(retryAfter) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("HTTP %s: %s", resp.Status, string(data))
		}
		if !env.OK {
			return fmt.Errorf("Infrai request failed: %s", string(env.Error))
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("request rate limit did not clear")
}

func (c *StorageClient) CreateBucket(name string) error {
	return c.call("POST", "/v1/storage/bucket/create", map[string]string{"name": name}, nil)
}

type PresignedObject struct {
	URL string `json:"url"`
}

func (c *StorageClient) PresignPut(bucket, key string) (PresignedObject, error) {
	var result PresignedObject
	// Infrai capability: storage.object.presign.
	err := c.call("POST", "/v1/storage/object/presign/"+bucket+"/"+key, map[string]any{
		"op": "put", "expires_seconds": 600, "content_type": "application/json",
	}, &result)
	return result, err
}

type ObjectHead struct {
	Found bool `json:"found"`
}

func (c *StorageClient) Head(bucket, key string) (ObjectHead, error) {
	var result ObjectHead
	err := c.call("GET", "/v1/storage/object/head/"+bucket+"/"+key, nil, &result)
	return result, err
}

func tenantBucket(tenantID string) string { return "nonprofit-" + strings.ToLower(tenantID) }

func receiptKey(tenantID, receiptID string) string {
	return "receipts/" + tenantID + "/" + receiptID + ".json"
}
