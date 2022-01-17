package data

// BotSettings Holds Telegram bot settings
type BotSettings struct {
	JoinedGroups []JoinedGroup `json:"joined_groups"`
}

// JoinedGroup Hold data for group which bot was joined
type JoinedGroup struct {
	Title    string `json:"title"`
	Id       int64  `json:"id"`
	Announce bool   `json:"announce"`
}

const (
	// SaveNameForBotSettings Save name for Calendar event with bot data
	SaveNameForBotSettings = "botconfig.json"
	// SaveBotSettingsDate Default save date for bot settings
	SaveBotSettingsDate = "2021-01-01"
)
