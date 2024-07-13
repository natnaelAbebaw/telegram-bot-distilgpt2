package main

import (
	"encoding/json"
	"fmt"
	"log"

	// "os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/levigross/grequests"
)

const (
    telegramBotToken     = "7234333877:AAEBCFBMIYBotWsMCx4wQ0wunsq3nWP1F24"
    huggingFaceAPIKey    = "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"
    huggingFaceModelURL  = "https://api-inference.huggingface.co/models/distilgpt2"
)

type HFResponse struct {
    GeneratedText string `json:"generated_text"`
}

func main() {
    bot, err := tgbotapi.NewBotAPI(telegramBotToken)
    if err != nil {
        log.Panic(err)
    }

    bot.Debug = true
    log.Printf("Authorized on account %s", bot.Self.UserName)

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates := bot.GetUpdatesChan(u)

    for update := range updates {
        if update.Message == nil {
            continue
        }

        log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

        response, err := getGPTJResponse(update.Message.Text)
        if err != nil {
            log.Printf("Error getting GPT-J response: %v", err)
            response = "Sorry, I couldn't process that."
        }

        msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
        bot.Send(msg)
    }
}

func getGPTJResponse(prompt string) (string, error) {
    requestBody := map[string]string{
        "inputs": prompt,
    }

    jsonData, err := json.Marshal(requestBody)
    if err != nil {
        return "", err
    }

    requestOptions := &grequests.RequestOptions{
        Headers: map[string]string{
            "Authorization": "Bearer " + huggingFaceAPIKey,
            "Content-Type":  "application/json",
        },
        JSON: jsonData,
    }

    resp, err := grequests.Post(huggingFaceModelURL, requestOptions)
    if err != nil {
        return "", err
    }

    if !resp.Ok {
        return "", fmt.Errorf("Request failed: %s", resp.String())
    }

    var response []HFResponse
    if err := json.Unmarshal(resp.Bytes(), &response); err != nil {
        return "", err
    }

    if len(response) == 0 {
        return "", fmt.Errorf("Empty response from API")
    }

    return response[0].GeneratedText, nil
}