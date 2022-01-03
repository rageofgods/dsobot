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
	handleFunc  func(arg string)
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
			description: "Показать welcome сообщение",
			handleFunc:  t.handleStart},
		{command: &cmd{name: botCmdHelp, args: nil},
			description: "Показать помощь по командам",
			handleFunc:  t.handleHelp},
		{command: &cmd{name: botCmdRegister, args: nil},
			description: "Отправть заявку на регистрацию",
			handleFunc:  t.handleRegister},
		{command: &cmd{name: botCmdUnregister, args: nil},
			description: "Выйти из системы",
			handleFunc:  t.handleUnregister},
		{command: &cmd{name: botCmdWhoIsOnDuty, args: &[]arg{
			{name: botCmdArgDuty,
				handleFunc: t.handleWhoIsOnDuty,
				description: "Показать дежурного на сегодня. _Возможно указание конкретной даты " +
					"через пробел после аргумента_"},
			{name: botCmdArgValidation,
				handleFunc: t.handleWhoIsOnValidation,
				description: "Показать валидирующего на сегодня. _Возможно указание конкретной даты " +
					"через пробел после аргумента_"}}},
			description: "Показать дежурного для определенного типа дежурств",
			handleFunc:  t.handleWhoIsOn},
		{command: &cmd{name: botCmdAddOffDuty, args: &[]arg{
			{name: botCmdArgOffDuty,
				handleFunc:  nil,
				description: "Период _От-До_ (через дефис)"}}},
			description: "Добавить нерабочий период (отпуск/болезнь/etc)",
			handleFunc:  t.handleAddOffDuty},
		{command: &cmd{name: botCmdShowOffDuty, args: nil},
			description: "Показать список нерабочих периодов (отпуск/болезнь/etc)",
			handleFunc:  t.handleShowOffDuty},
	}}
}

// AdminBotCommands returns slice of admin botCommand struct
func (t *TgBot) AdminBotCommands() *botCommands {
	return &botCommands{commands: []botCommand{
		{command: &cmd{name: botCmdHelp, args: nil},
			description: "Показать помощь по командам",
			handleFunc:  t.adminHandleHelp},
		{command: &cmd{name: botCmdList, args: nil},
			description: "Вывести список участников",
			handleFunc:  t.adminHandleList},
		{command: &cmd{name: botCmdRollout, args: &[]arg{
			{name: botCmdArgDuty,
				handleFunc:  t.adminHandleRolloutDuty,
				description: "Дежурства"},
			{name: botCmdArgValidation,
				handleFunc:  t.adminHandleRolloutValidation,
				description: "Валидация задач"},
			{name: botCmdArgNonWorkingDay,
				handleFunc:  t.adminHandleRolloutNonWorkingDay,
				description: "Нерабочие дни (выходные/праздники)"}}},
			description: "Пересоздать события определенного типа для текущего месяца",
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
	botCmdAddOffDuty  tCmd = "addoffduty"
	botCmdShowOffDuty tCmd = "showoffduty"
	botCmdHelp        tCmd = "help"
	botCmdList        tCmd = "list"
	botCmdRollout     tCmd = "rollout"
)

// Bot available args
const (
	botCmdArgDuty          tArg = "duty"
	botCmdArgValidation    tArg = "validation"
	botCmdArgNonWorkingDay tArg = "nwd"
	botCmdArgOffDuty       tArg = "DDMMYYYY-DDMMYYYY"
)

// User provided data format for bot commands
const botDataShort1 = "02012006"
const botDataShort2 = "02.01.2006"
const botDataShort3 = "02/02/2006"

// Structure for saving callback data (json is shortened to be able to accommodate to 64b Telegram data limit)
type callbackMessage struct {
	Answer     string `json:"a"`
	ChatId     int64  `json:"c"`
	MessageId  int    `json:"m"`
	UserId     int64  `json:"u"`
	FromHandle string `json:"f"`
}
