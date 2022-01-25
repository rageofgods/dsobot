package bot

import (
	"dso_bot/pkg/data"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

/////////////////////////////////
// Structure to hold answer data for newly registered users
type tmpRegisterData struct {
	userId int64
	data   string
}

// Structure to hold temporary dutyMan data before saving it
type tmpDutyManData struct {
	userId int64
	data   []data.DutyMan
}

// Structure to hold temporary JoinedGroup (BotSettings) data before saving it
type tmpJoinedGroupData struct {
	userId int64
	data   []data.JoinedGroup
}

type tmpOffDutyData struct {
	userId int64
	data   []time.Time
}

// Structure (parent) for different types of tmp data
type tmpData struct {
	tmpRegisterData    []tmpRegisterData
	tmpDutyManData     []tmpDutyManData
	tmpJoinedGroupData []tmpJoinedGroupData
	tmpOffDutyData     []tmpOffDutyData
}

/////////////////////////////////
// Custom struct for bot commands
type cmd struct {
	name tCmd
	args *[]arg
}

// Custom struct for bot command args
type arg struct {
	name        tArg
	description string
	handleFunc  func(arg string, update *tgbotapi.Update)
}

// Custom types for commands and arguments
type tCmd string
type tArg string

// Structure for available bot commands
type botCommand struct {
	command     *cmd
	description string
	handleFunc  func(cmdArgs string, update *tgbotapi.Update)
}

// Structure to hold list of bot commands
type botCommands struct {
	commands []botCommand
}

/////////////////////////////////

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
		{command: &cmd{name: botCmdAddOffDuty, args: nil},
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
			{name: botCmdArgAll,
				handleFunc:  t.adminHandleRolloutAll,
				description: "Все события типов дежурств"},
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
		{command: &cmd{name: botCmdAnnounce, args: nil},
			description: "Включить или выключить анонс событий дежурства в для групповых чатов",
			handleFunc:  t.adminHandleAnnounce},
	}}
}

// Some const's for working with callbacks (use short names to workaround Telegram 64b callback data limit)
const (
	// Void answer for buttons without any helpful data
	inlineKeyboardVoid = "{}"

	inlineKeyboardYes = "99"
	inlineKeyboardNo  = "98"

	inlineKeyboardNext = "97"
	inlineKeyboardPrev = "96"
	inlineKeyboardDate = "95"

	inlineKeyboardEditDutyYes = "1"
	inlineKeyboardEditDutyNo  = "0"

	callbackHandleRegister       = "a"
	callbackHandleRegisterHelper = "b"
	callbackHandleUnregister     = "c"
	callbackHandleDeleteOffDuty  = "d"
	callbackHandleReindex        = "e"
	callbackHandleEnable         = "f"
	callbackHandleDisable        = "g"
	callbackHandleEditDuty       = "h"
	callbackHandleAnnounce       = "i"
	callbackHandleAddOffDuty     = "j"
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
	botCmdAnnounce      tCmd = "announce"
)

// Bot available args
const (
	botCmdArgAll           tArg = "all"
	botCmdArgDuty          tArg = "duty"
	botCmdArgValidation    tArg = "validation"
	botCmdArgNonWorkingDay tArg = "nwd"
)

// User provided data format for bot commands
const (
	botDataShort1 = "02012006"
	botDataShort2 = "02.01.2006"
	botDataShort3 = "02/01/2006"
	botDataShort4 = "020106"
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
	msgTextUserHandleRegister = "Для того, чтобы начать процесс регистрации, пожалуйста, отправьте " +
		"ваши реальные Имя и Фамилию в ❗ОТВЕТЕ❗ на это сообщение.\n\n" +
		"Например: 'Вася Пупкин' или 'Пупкин Василий'.\n\n"
	msgTextAdminHandleAnnounce = "Укажите для каких групповых чатов необходимо включить анонс дежурств\n\n" +
		"✅ - включает анонс в группу\n" +
		"❌ - выключает анонс в группу\n\n" +
		"⚠️Внимание, для того, чтобы бот мог закреплять сообщения в нужном чате " +
		"ему необходимо выдать права администратора на закрепление сообщений в соответствующем чате"
	msgTextUserHandleAddOffDuty1 = "📅 Для того, чтобы добавить новый нерабочий период " +
		"выберите дату его начала.\n"
	msgTextUserHandleAddOffDuty2     = "📅 Теперь выберите дату завершения нерабочего периода (включительно)\n"
	msgTextUserHandleAddOffDutyStart = "Начало нерабочего периода:"
	msgTextUserHandleAddOffDutyEnd   = "Конец нерабочего периода:"
)
