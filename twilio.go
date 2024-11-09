package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

// downloadMedia downloads media from Twilio's URL
func (c *Client) downloadMedia(url, messageID, mediaType string) (string, error) {

	// Create HTTP client with authentication
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Get the media
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Determine file extension based on media type
	ext := filepath.Ext(url)
	if ext == "" {
		switch mediaType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
	case "image/webp":
			ext = ".webp"
		default:
			ext = ""
		}
	}

	// Create media directory if it doesn't exist
	if err := os.MkdirAll("media", 0755); err != nil {
		return "", err
	}

	// Create unique filename
	filename := fmt.Sprintf("media/%s%s", messageID, ext)
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Save the media
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return filename, nil
}

// SendWhatsAppMessage sends a message using Twilio's API
func (c *Client) SendWhatsAppMessage(to, message string) error {
	 // Ensure the 'to' number is in WhatsApp format
  if !strings.HasPrefix(to, "whatsapp:") {
    to = "whatsapp:" + to
  }
  
	clientt := twilio.NewRestClient()

	params := &api.CreateMessageParams{}
	params.SetBody(message)
	params.SetFrom("whatsapp:"+c.config.twilio.TwilioWhatsAppNumber)
	params.SetTo(to)

	resp, err := clientt.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	} 

	if resp.Sid != nil {
    fmt.Printf("%s,%s\n", *resp.Sid, *resp.Status)
	} else {
    fmt.Printf("%s,%s\n", resp.Sid, resp.Status)
  }

	return nil
}
// HandleIncomingMessage processes incoming WhatsApp messages
func (c *Client) HandleIncomingMessage(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	message := IncomingMessage{
		From:       r.FormValue("From"),
		To:         r.FormValue("To"),
		Body:       r.FormValue("Body"),
		MessageSID: r.FormValue("MessageSid"),
		NumMedia:   r.FormValue("NumMedia"),
		MediaURL:   r.FormValue("MediaUrl0"),
		MediaType:  r.FormValue("MediaContentType0"),
	}

	var responseMessage string

	// Check if message contains media
	if message.NumMedia == "1" {
		// Handle image message
		if message.MediaType != "" && (message.MediaType == "image/jpeg" ||
			message.MediaType == "image/png" ||
			message.MediaType == "image/webp") {

			filename, err := c.downloadMedia(message.MediaURL, message.MessageSID, message.MediaType)
			if err != nil {
				log.Printf("Error downloading media: %v", err)
				responseMessage = "Sorry, there was an error processing your image."
			}

			// Process receipt
			response, err := c.AnalyzeReceipt(filename, message.MediaType)
			if err != nil {
				log.Fatalf("Error analyzing receipt: %v", err)
			}	

			responseMessage = response.Content[0].Text

			c.SendWhatsAppMessage(os.Getenv("WHATSAPP_NUMBER"), responseMessage)
		} else {
			log.Printf("Unsupported media type: %s", message.MediaType)
			responseMessage = "Sorry, we only support image files (JPEG, PNG, WebP)."
		}
	} else if message.Body != "" {
		log.Printf("Received text message: %s", message.Body)
		responseMessage = "Welcome. Please Send me a receipt image to exract the data. Thank you!"
	} else {
		log.Printf("Received empty or invalid message")
		responseMessage = "Please send either a text message or an image."
	}

	// Log complete message details
	messageDetails, _ := json.MarshalIndent(message, "", "  ")
	log.Printf("Complete message details: %s", string(messageDetails))

	// Send response back to Twilio
	w.Header().Set("Content-Type", "text/xml")
	response := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
		<Response>
				<Message>%s</Message>
		</Response>`, responseMessage)
	w.Write([]byte(response))
}
