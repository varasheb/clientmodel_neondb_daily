package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func SendHiNotification() {
	webhookURL := "https://chat.googleapis.com/v1/spaces/AAAAN15k9x8/messages?key=AIzaSyDdI0hCZtE6vySjMm-WEfRq3CPzqKqqsHI&token=JFvbraw4ovOqhkVWTAwUgpSUExV2e0wVu5FoDQSGYdA"

	// Create the message data
	message := map[string]interface{}{
		"text": "Hi",
	}

	// Marshal the message data to JSON format
	messageData, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Error marshalling message data: %v\n", err)
		return
	}

	// Send the message
	if err := sendMessageToGoogleChat(webhookURL, messageData); err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}
}

func sendMessageToGoogleChat(webhookURL string, messageData []byte) error {
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(messageData))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed with status code: %d", resp.StatusCode)
	}

	fmt.Println("Message sent successfully!")
	return nil
}

func main() {
	SendHiNotification()
}
