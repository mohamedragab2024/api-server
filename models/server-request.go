package models

type ServerRequest struct {
	Path         string      `json:"path"`
	Verb         string      `json:"verb"`
	PayLoad      interface{} `json:"payload"`
	Prefix       string      `json:"prefix"`
	ResourceType string      `json:"resourceType"`
}
