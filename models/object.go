package models

import "net/url"

type Object struct {
	ID      string     `json:"id,omitempty"`
	Name    string     `json:"name,omitempty"`
	Owner   string     `json:"owner,omitempty"`
	Meta    url.Values `json:"metadata,omitempty"`
	Payload []byte     `json:"payload,omitempty"`
}
