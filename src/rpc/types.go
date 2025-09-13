package rpc

import (
	"encoding/json"
	"fmt"
)

type Pattern struct {
	Cmd string `json:"cmd"`
}

type Request struct {
	ID      string          `json:"id"`
	Pattern Pattern         `json:"pattern"`
	Data    json.RawMessage `json:"data"`
}

type Response struct {
	Id         string      `json:"id,omitempty"`
	Response   interface{} `json:"response,omitempty"`
	IsDisposed bool        `json:"isDisposed,omitempty"`
	Status     string      `json:"status,omitempty"`
	Err        string      `json:"err,omitempty"`
}

func parseRequest(msgBytes []byte) (*Request, error) {

	var raw struct {
		ID      string          `json:"id"`
		Pattern json.RawMessage `json:"pattern"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(msgBytes, &raw); err != nil {
		return nil, err
	}

	req := &Request{
		ID:   raw.ID,
		Data: raw.Data,
	}

	if err := json.Unmarshal(raw.Pattern, &req.Pattern); err == nil {
		return req, nil
	}

	var patternStr string
	if err := json.Unmarshal(raw.Pattern, &patternStr); err == nil {
		if err := json.Unmarshal([]byte(patternStr), &req.Pattern); err == nil {
			return req, nil
		}
	}

	return nil, fmt.Errorf("invalid pattern format")
}
