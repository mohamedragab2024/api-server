package models

type Response struct {
	Type         string        `json:"type"`
	ResourceType string        `json:"resourceType"`
	Data         []interface{} `json:"data"`
	Status       int           `json:"status"`
	Message      string        `json:"message"`
	Prefix       string        `json:"prefix"`
}
