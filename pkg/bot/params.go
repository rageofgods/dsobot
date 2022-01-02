package bot

// Custom struct for bot commands
type cmd struct {
	name tCmd
	args *[]arg
}

// Custom struct for bot command args
type arg struct {
	name        tArg
	description string
	handleFunc  func()
}

// Custom types for commands and arguments
type tCmd string
type tArg string

// Structure for available bot commands
type botCommand struct {
	command     *cmd
	description string
	handleFunc  func(cmdArgs string)
}

// Structure to hold list of bot commands
type botCommands struct {
	commands []botCommand
}

// BotCommands returns slice of ordinary user botCommand struct
func (t *TgBot) BotCommands() *botCommands {
	return &botCommands{commands: []botCommand{
		{command: &cmd{name: botCmdStart, args: nil},
			description: "Show welcome message", handleFunc: t.handleStart},
		{command: &cmd{name: botCmdHelp, args: nil},
			description: "Show help message", handleFunc: t.handleHelp},
		{command: &cmd{name: botCmdRegister, args: nil},
			description: "Register an user as DSO member team", handleFunc: t.handleRegister},
		{command: &cmd{name: botCmdUnregister, args: nil},
			description: "Unregister user", handleFunc: t.handleUnregister},
		{command: &cmd{name: botCmdWhoIsOnDuty, args: &[]arg{
			{name: botCmdArgDuty, handleFunc: t.handleWhoIsOnDuty, description: "Дежурство"},
			{name: botCmdArgValidation, handleFunc: t.handleWhoIsOnValidation, description: "Валидация задач"},
		}},
			description: "Shows who is on duty of specified type",
			handleFunc:  t.handleWhoIsOn},
	}}
}

// AdminBotCommands returns slice of admin botCommand struct
func (t *TgBot) AdminBotCommands() *botCommands {
	return &botCommands{commands: []botCommand{
		{command: &cmd{name: botCmdHelp, args: nil},
			description: "Show command help", handleFunc: t.adminHandleHelp},
		{command: &cmd{name: botCmdList, args: nil},
			description: "Show members list", handleFunc: t.adminHandleList},
		{command: &cmd{name: botCmdRollout, args: &[]arg{
			{name: botCmdArgDuty, handleFunc: t.adminHandleRolloutDuty, description: "Дежурство"},
			{name: botCmdArgValidation, handleFunc: t.adminHandleRolloutValidation, description: "Валидация задач"},
			{name: botCmdArgNonWorkingDay, handleFunc: t.adminHandleRolloutNonWorkingDay, description: "Нерабочий день"},
		}},
			description: "Recreate current month calendar for provided event type.",
			handleFunc:  t.adminHandleRollout},
	}}
}

// Some const's for working with callbacks (use short names to workaround Telegram 64b callback data limit)
const (
	inlineKeyboardYes = "1"
	inlineKeyboardNo  = "0"

	callbackHandleRegister   = "fhr"
	callbackHandleUnregister = "fhu"
)

// Bot available commands
const (
	botCmdStart       tCmd = "start"
	botCmdRegister    tCmd = "register"
	botCmdUnregister  tCmd = "unregister"
	botCmdWhoIsOnDuty tCmd = "whoison"
	botCmdHelp        tCmd = "help"
	botCmdList        tCmd = "list"
	botCmdRollout     tCmd = "rollout"
)

const (
	botCmdArgDuty          tArg = "duty"
	botCmdArgValidation    tArg = "validation"
	botCmdArgNonWorkingDay tArg = "nwd"
)

// Structure for saving callback data (json is shortened to be able to accommodate to 64b Telegram data limit)
type callbackMessage struct {
	Answer     string `json:"a"`
	ChatId     int64  `json:"c"`
	MessageId  int    `json:"m"`
	UserId     int64  `json:"u"`
	FromHandle string `json:"f"`
}
