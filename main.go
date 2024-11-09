package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)


func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Load configuration
	config := Config{
		claude: ClaudeConfig{
			APIKey:      os.Getenv("CLAUDE_API_KEY"),
			ModelName:   "claude-3-opus-20240229",
			APIEndpoint: "https://api.anthropic.com/v1/messages",
			APIVersion:  "2023-06-01",
		},
		twilio: TwilioConfig{
			AccountSID:           os.Getenv("TWILIO_ACCOUNT_SID"),
			AuthToken:            os.Getenv("TWILIO_AUTH_TOKEN"),
			TwilioWhatsAppNumber: os.Getenv("TWILIO_WHATSAPP_NUMBER"),
		},
	}

	// Create client
	client := &Client{
		config:     config,
		httpClient: &http.Client{},
	}

	// Set up HTTP server for webhook
	http.HandleFunc("/webhook", client.HandleIncomingMessage)

	err := client.SendWhatsAppMessage(os.Getenv("WHATSAPP_NUMBER"), "Welcome. Please Send me a receipt image to exract the data. Thank you!")
	if err != nil {
		log.Printf("Error sending message: %v", err)
	} 
	
	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
