package main

import (
	"dso_bot/pkg/bot"
	"dso_bot/pkg/data"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	// Load env
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("error loading .env file: %s", err)
	}

	// Read variables
	calToken := os.Getenv("CAL_TOKEN")
	calURL := os.Getenv("CAL_URL")
	botToken := os.Getenv("BOT_TOKEN")

	// Init calendar service
	dc := data.NewCalData(calToken, calURL)
	err = dc.InitService()
	if err != nil {
		log.Println(err)
	}

	// Start bot
	bot.StartBot(dc, botToken)
}
