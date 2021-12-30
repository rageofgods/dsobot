package bot

// Structure for available bot commands
type botCommand struct {
	command     string
	description string
	handleFunc  func()
}

// Structure to hold list of bot commands
type botCommands struct {
	commands []botCommand
}

// BotCommands returns slice of botCommand struct
func (t *TgBot) BotCommands() *botCommands {
	return &botCommands{commands: []botCommand{
		{command: "start", description: "Show welcome message", handleFunc: t.handleStart},
		{command: "register", description: "Register an user as DSO member team", handleFunc: t.handleRegister},
	}}
}

// Some const's for working with callbacks (use short names to workaround Telegram 64b callback data limit)
const (
	inlineKeyboardYes = "1"
	inlineKeyboardNo  = "0"

	callbackHandleRegister = "fhr"
)

// Structure for saving callback data (json is shortened to be able to accommodate to 64b Telegram data limit)
type callbackMessage struct {
	Answer     string `json:"a"`
	ChatId     int64  `json:"c"`
	UserId     int64  `json:"u"`
	FromHandle string `json:"f"`
}
