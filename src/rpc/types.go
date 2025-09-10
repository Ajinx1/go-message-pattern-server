package rpc

import "encoding/json"

type Request struct {
	Pattern struct {
		Cmd string `json:"cmd"`
	} `json:"pattern"`
	Data json.RawMessage `json:"data"`
}

type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}
