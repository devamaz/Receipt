package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// loadImage reads and encodes an image file to base64
func loadImage(filepath string) (string, error) {
	imageData, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("reading image file: %w", err)
	}
	return base64.StdEncoding.EncodeToString(imageData), nil
}

// AnalyzeReceipt sends a receipt image to Claude for analysis
func (c *Client) AnalyzeReceipt(imagePath, imageType string) (*Response, error) {
	// Load and encode image
	base64Image, err := loadImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("loading image: %w", err)
	}

	// Construct request
	request := Request{
		Model:     c.config.claude.ModelName,
		MaxTokens: 1024,
		Messages: []Message{{
			Role: "user",
			Content: []Content{
				{
					Type: "image",
					Source: &Source{
						Type:      "base64",
						MediaType: imageType,
						Data:      base64Image,
					},
				},
				{
					Type: "text",
					Text: "This image is an invoice. Extract key invoice details in JSON format. Do not include any other text.",
				},
			},
		},
		},
	}

	// Marshal request
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", c.config.claude.APIEndpoint, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	req.Header.Add("x-api-key", c.config.claude.APIKey)
	req.Header.Add("anthropic-version", c.config.claude.APIVersion)
	req.Header.Add("content-type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	// Debug: Print response body when status is not 200
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Response Headers:\n")
		for k, v := range resp.Header {
			fmt.Printf("%s: %v\n", k, v)
		}
		fmt.Printf("Response Body: %s\n", string(body))
	}

	// Check for API error
	if response.Error != nil {
		return nil, fmt.Errorf("API error: %s - %s", response.Error.Type, response.Error.Message)
	}

	return &response, nil
}
