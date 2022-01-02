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

// BotCommands returns slice of ordinary user botCommand struct
func (t *TgBot) BotCommands() *botCommands {
	return &botCommands{commands: []botCommand{
		{command: "start", description: "Show welcome message", handleFunc: t.handleStart},
		{command: "register", description: "Register an user as DSO member team", handleFunc: t.handleRegister},
		{command: "unregister", description: "Unregister user", handleFunc: t.handleUnregister},
	}}
}

// AdminBotCommands returns slice of admin botCommand struct
func (t *TgBot) AdminBotCommands() *botCommands {
	return &botCommands{commands: []botCommand{
		{command: "help", description: "Show command list", handleFunc: t.adminHandleHelp},
		{command: "list", description: "Show members list", handleFunc: t.adminHandleList},
	}}
}

// Some const's for working with callbacks (use short names to workaround Telegram 64b callback data limit)
const (
	inlineKeyboardYes = "1"
	inlineKeyboardNo  = "0"

	callbackHandleRegister   = "fhr"
	callbackHandleUnregister = "fhu"
)

// Structure for saving callback data (json is shortened to be able to accommodate to 64b Telegram data limit)
type callbackMessage struct {
	Answer     string `json:"a"`
	ChatId     int64  `json:"c"`
	UserId     int64  `json:"u"`
	FromHandle string `json:"f"`
}
