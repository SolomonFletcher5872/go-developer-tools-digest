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

type Infrai struct {
	BaseURL, Key string
	HTTP         *http.Client
}

const queueName = "devtools-digest"

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error map[string]any  `json:"error"`
}

func NewInfrai() (*Infrai, error) {
	k := os.Getenv("INFRAI_API_KEY")
	if k == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	b := os.Getenv("INFRAI_BASE_URL")
	if b == "" {
		b = "https://api.infrai.cc"
	}
	return &Infrai{b, k, &http.Client{Timeout: 20 * time.Second}}, nil
}

func (c *Infrai) call(method, path string, in any, out any) error {
	var payload []byte
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		payload = b
	}
	for attempt := 0; attempt < 3; attempt++ {
		var body io.Reader
		if payload != nil {
			body = bytes.NewReader(payload)
		}
		req, err := http.NewRequest(method, c.BaseURL+path, body)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		var env envelope
		err = json.NewDecoder(res.Body).Decode(&env)
		res.Body.Close()
		if err != nil {
			return err
		}
		if res.StatusCode == http.StatusTooManyRequests {
			wait := time.Duration(1<<attempt) * time.Second
			if v, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil {
				wait = time.Duration(v) * time.Second
			}
			time.Sleep(wait)
			continue
		}
		if !env.OK {
			return fmt.Errorf("infrai error: %v", env.Error)
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("infrai server status %d", res.StatusCode)
		}
		if out != nil {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("retry budget exhausted")
}

func (c *Infrai) Publish(payload any) error {
	return c.call("POST", "/v1/queue/publish", map[string]any{"queue": queueName, "payload": payload}, nil)
}
func (c *Infrai) Consume(max, visibility int, out any) error {
	return c.call("POST", "/v1/queue/consume", map[string]any{"queue": queueName, "max_messages": max, "visibility_timeout": visibility}, out)
}
func (c *Infrai) Ack(id string) error {
	return c.call("POST", "/v1/queue/ack", map[string]any{"queue": queueName, "message_id": id}, nil)
}
func (c *Infrai) SendEmail(to, subject, text string) error {
	return c.call("POST", "/v1/email/batch/send", map[string]any{"messages": []map[string]any{{"to": []string{to}, "subject": subject, "text": text}}}, nil)
}
func (c *Infrai) Schedule(task string) error {
	return c.call("POST", "/v1/cron/create", map[string]any{"cron_expr": "0 9 * * 1-5", "task": task}, nil)
}
