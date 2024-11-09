package main

import (
	"net/http"
)

// Configuration holds all the app settings
type ClaudeConfig struct {
	APIKey      string
	ModelName   string
	APIEndpoint string
	APIVersion  string
}

type TwilioConfig struct {
	AccountSID           string
	AuthToken            string
	TwilioWhatsAppNumber string
}

type Config struct {
	claude ClaudeConfig
	twilio TwilioConfig
}

// Message represents a chat message
type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}


// IncomingMessage represents the structure of incoming WhatsApp messages
type IncomingMessage struct {
	From       string 
	To         string 
	Body       string 
	MessageSID string 
	NumMedia   string 
	MediaURL   string 
	MediaType  string 
}

// Content represents different types of content in a message
type Content struct {
	Type   string  `json:"type"`
	Text   string  `json:"text,omitempty"`
	Source *Source `json:"source,omitempty"`
}

// Source represents the image source data
type Source struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// Request represents the API request structure
type Request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

// Response represents the API response structure
type Response struct {
	ID         string    `json:"id"`
	Content    []Content `json:"content"`
	Role       string    `json:"role"`
	Model      string    `json:"model"`
	CreatedAt  int64     `json:"created_at"`
	StopReason string    `json:"stop_reason"`
	Error      *APIError `json:"error,omitempty"`
}

// APIError represents error information from the API
type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Client handles communication with the Claude API
type Client struct {
	config     Config
	httpClient *http.Client
}
