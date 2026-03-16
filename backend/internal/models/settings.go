package models

import "encoding/json"

type SiteSetting struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

type SiteSettings map[string]json.RawMessage
