package ai_entity

import "net/http"

type MoonshotClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}
