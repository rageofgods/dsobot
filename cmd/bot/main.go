package main

import (
	"dso_bot/pkg/bot"
	"dso_bot/pkg/data"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

// App version
var (
	Version string
	Build   string
)

func main() {
	//
	// Show version info
	log.Printf("Version: %s, Build: %s", Version, Build)

	// Load env
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("error loading .env file: %s", err)
	}

	// Read variables
	calToken := readEnv("CAL_TOKEN")
	calURL := readEnv("CAL_URL")
	botToken := readEnv("BOT_TOKEN")
	botAdminGroupID := readEnv("BOT_ADMIN_GROUP_ID")

	id, err := strconv.ParseInt(botAdminGroupID, 10, 64) // Converting string to int64
	if err != nil {
		panic(fmt.Sprintf("Can't convert admin groupId to int64: %v", err))
	}

	// Init calendar service
	dc := data.NewCalData(calToken, calURL)
	err = dc.InitService()
	if err != nil {
		panic(err)
	}

	// Load DutyMen data
	_, err = dc.LoadMenList()
	if err != nil {
		log.Printf("Unable to load saved data: %v", err)
	}

	// Load BotSettings data
	botSettings, err := dc.LoadBotSettings()
	if err != nil {
		log.Printf("Unable to load saved data: %v", err)
	}

	// Start tgBot
	tgBot := bot.NewTgBot(dc, botSettings, botToken, id, true)
	tgBot.StartBot(Version, Build)
}

func readEnv(envName string) string {
	env := os.Getenv(envName)
	if env == "" {
		panic(fmt.Sprintf("%s is empty", envName))
	}
	return env
}
