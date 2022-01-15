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

// UserBotCommands returns slice of ordinary user botCommand struct
func (t *TgBot) UserBotCommands() *botCommands {
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
		{command: &cmd{name: botCmdWhoIsOn, args: &[]arg{
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
		{command: &cmd{name: botCmdShowMy, args: &[]arg{
			{name: botCmdArgDuty,
				handleFunc:  t.handleShowMyDuty,
				description: "Показать дежурства в этом месяце"},
			{name: botCmdArgValidation,
				handleFunc:  t.handleShowMyValidation,
				description: "Показать валидации в этом месяце"}}},
			description: "Показать список дежурств в текущем месяце для определенного типа дежурств",
			handleFunc:  t.handleShowMy},
		{command: &cmd{name: botCmdAddOffDuty, args: &[]arg{
			{name: botCmdArgOffDuty,
				handleFunc:  nil,
				description: "Период _От-До_ (через дефис)"}}},
			description: "Добавить нерабочий период (отпуск/болезнь/etc)",
			handleFunc:  t.handleAddOffDuty},
		{command: &cmd{name: botCmdShowOffDuty, args: nil},
			description: "Показать список нерабочих периодов (отпуск/болезнь/etc)",
			handleFunc:  t.handleShowOffDuty},
		{command: &cmd{name: botCmdDeleteOffDuty, args: nil},
			description: "Удалить нерабочий период",
			handleFunc:  t.handleDeleteOffDuty},
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
		{command: &cmd{name: botCmdShowOffDuty, args: nil},
			description: "Показать список нерабочих периодов (отпуск/болезнь/etc) для всех участников",
			handleFunc:  t.adminHandleShowOffDuty},
		{command: &cmd{name: botCmdReindex, args: nil},
			description: "Изменить порядок дежурных (повлияет на очередность дежурств)",
			handleFunc:  t.adminHandleReindex},
		{command: &cmd{name: botCmdEnable, args: nil},
			description: "Добавить активных дежурных (повлияет на очередность дежурств)",
			handleFunc:  t.adminHandleEnable},
		{command: &cmd{name: botCmdDisable, args: nil},
			description: "Добавить неактивных дежурных (повлияет на очередность дежурств)",
			handleFunc:  t.adminHandleDisable},
		{command: &cmd{name: botCmdEditDutyType, args: nil},
			description: "Отредактировать типы дежурств для всех дежурных",
			handleFunc:  t.adminHandleEditDutyType},
	}}
}

// Some const's for working with callbacks (use short names to workaround Telegram 64b callback data limit)
const (
	inlineKeyboardYes = "99"
	inlineKeyboardNo  = "98"

	inlineKeyboardEditDutyYes = "1"
	inlineKeyboardEditDutyNo  = "0"

	callbackHandleRegister       = "fhr"
	callbackHandleRegisterHelper = "fhrh"
	callbackHandleUnregister     = "fhu"
	callbackHandleDeleteOffDuty  = "fhdod"
	callbackHandleReindex        = "fhre"
	callbackHandleEnable         = "fhe"
	callbackHandleDisable        = "fhd"
	callbackHandleEditDuty       = "fhed"
)

// Bot available commands
const (
	botCmdStart         tCmd = "start"
	botCmdRegister      tCmd = "register"
	botCmdUnregister    tCmd = "unregister"
	botCmdWhoIsOn       tCmd = "whoison"
	botCmdShowMy        tCmd = "showmy"
	botCmdAddOffDuty    tCmd = "addoffduty"
	botCmdShowOffDuty   tCmd = "showoffduty"
	botCmdDeleteOffDuty tCmd = "deleteoffduty"
	botCmdHelp          tCmd = "help"
	botCmdList          tCmd = "list"
	botCmdRollout       tCmd = "rollout"
	botCmdReindex       tCmd = "reindex"
	botCmdEnable        tCmd = "enable"
	botCmdDisable       tCmd = "disable"
	botCmdEditDutyType  tCmd = "editduty"
)

// Bot available args
const (
	botCmdArgDuty          tArg = "duty"
	botCmdArgValidation    tArg = "validation"
	botCmdArgNonWorkingDay tArg = "nwd"
	botCmdArgOffDuty       tArg = "DDMMYYYY-DDMMYYYY"
)

// User provided data format for bot commands
const (
	botDataShort1 = "02012006"
	botDataShort2 = "02.01.2006"
	botDataShort3 = "02/01/2006"
)

// Continuous days for duty periods
const (
	onDutyContDays       = 2
	onValidationContDays = 1
)

// Structure for saving callback data (json is shortened to be able to accommodate to 64b Telegram data limit)
type callbackMessage struct {
	Answer     string `json:"a"`
	ChatId     int64  `json:"c"`
	MessageId  int    `json:"m"`
	UserId     int64  `json:"u"`
	FromHandle string `json:"f"`
}

// Text strings for messages
// Don't use markdown here because returned message will be always in plain text
const (
	msgTextAdminHandleReindex = "Укажите новую очередность дежурств (поочередно нажимая на кнопки участников " +
		"в нужной последовательности):"
	msgTextAdminHandleEnable = "Укажите активных дежурных из текущего списка неактивных" +
		" (поочередно нажимая на кнопки участников в нужной последовательности):"
	msgTextAdminHandleDisable = "Укажите неактивных дежурных из текущего списка активных" +
		" (поочередно нажимая на кнопки участников в нужной последовательности):"
	msgTextAdminHandleEditDuty = "Укажите нужные типы дежурства для текущего списка дежурных\n\n" +
		"✅ - включает тип дежурства\n" +
		"❌ - выключает тип дежуртсва\n\n" +
		"❗ - неактивный дежурный\n\n"
	msgTextUserHandleRegister = "Для того, чтобы начать процесс регистрации, пожалуйста пришлите мне " +
		"ваши реальные Имя и Фамилию в ОТВЕТЕ (Reply) на это сообщение.\n\n" +
		"Например: 'Вася Пупкин' или 'Пупкин Василий'.\n\n"
)
