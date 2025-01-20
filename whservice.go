package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func SendNotification(newclm []*Clientmodel) {
	webhookURL := "https://chat.googleapis.com/v1/spaces/AAAAN15k9x8/messages?key=AIzaSyDdI0hCZtE6vySjMm-WEfRq3CPzqKqqsHI&token=JFvbraw4ovOqhkVWTAwUgpSUExV2e0wVu5FoDQSGYdA"
	var pmsg string
	pmap := make(map[string][]*Clientmodel)

	for _, p := range newclm {
		key := p.GroupNames + p.Model
		pmap[key] = append(pmap[key], p)
	}
	slno := 1
	for _, plist := range pmap {
		p := plist[0]
		pmsg += fmt.Sprintf("%d *GN:* %s | *M:* %s\n",
			slno, p.GroupNames, p.Model)
		slno++
	}

	message := map[string]interface{}{
		"text": pmsg,
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		logger.Printf("Error marshalling message data: %v\n", err)
		return
	}

	if err := sendMessageToGoogleChat(webhookURL, messageData); err != nil {
		logger.Printf("Error sending message: %v\n", err)
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
