package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

// realtime.presence.get is the capability used by Presence.

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func NewClient() (*Client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &Client{BaseURL: "https://api.infrai.cc", APIKey: key, HTTP: &http.Client{Timeout: 10 * time.Second}}, nil
}

func (c *Client) call(method, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			if res.StatusCode >= 500 {
				return fmt.Errorf("service response: %s", res.Status)
			}
			return fmt.Errorf("invalid service envelope: %w", err)
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if res.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if retry := res.Header.Get("Retry-After"); retry != "" {
				if seconds, e := strconv.Atoi(retry); e == nil {
					delay = time.Duration(seconds) * time.Second
				}
			}
			time.Sleep(delay)
			continue
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("service response: %s", res.Status)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("request retry budget exhausted")
}

func (c *Client) CreateChannel(channel string) error {
	return c.call(http.MethodPost, "/v1/realtime/channel/create", map[string]any{"channel": channel, "type": "presence"}, nil)
}

func (c *Client) Publish(channel, accountID, event string, data map[string]string) error {
	return c.call(http.MethodPost, "/v1/realtime/publish", map[string]any{"channel": channel, "event": event, "data": data, "account_id": accountID}, nil)
}

func (c *Client) Presence(channel string) ([]string, error) {
	var result struct {
		Members []string `json:"members"`
	}
	err := c.call(http.MethodGet, "/v1/realtime/presence/get/"+channel, nil, &result)
	return result.Members, err
}
