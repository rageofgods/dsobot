package bot

import (
	"dso_bot/pkg/data"
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	str "strings"
)

type TgBot struct {
	bot          *tgbotapi.BotAPI
	dc           *data.CalData
	token        string
	msg          *tgbotapi.MessageConfig
	adminGroupId int64
	debug        bool
	update       *tgbotapi.Update
}

func NewTgBot(dc *data.CalData, token string, adminGroupId int64, debug bool) *TgBot {
	return &TgBot{
		dc:           dc,
		token:        token,
		msg:          new(tgbotapi.MessageConfig),
		adminGroupId: adminGroupId,
		debug:        debug,
		update:       new(tgbotapi.Update),
	}
}

func (t *TgBot) StartBot() {
	var err error
	t.bot, err = tgbotapi.NewBotAPI(t.token)
	if err != nil {
		panic(err)
	}

	//t.bot = bot
	t.bot.Debug = t.debug

	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 30

	// Start polling Telegram for updates.
	updates := t.bot.GetUpdatesChan(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {
		// Process ordinary command messages
		if update.Message != nil && update.Message.IsCommand() {
			// Hold pointer to the current update for access inside handlers
			t.update = &update

			// Init empty message to fill up it later
			*t.msg = tgbotapi.NewMessage(update.Message.Chat.ID, "")
			// Set default text mode to markdown
			t.msg.ParseMode = "markdown"

			// Go through struct of allowed commands
			bc := t.BotCommands()
			abc := t.AdminBotCommands()

			// Handle admin commands
			if update.Message.Chat.ID == t.adminGroupId {
				var isCmdFound bool
				for _, cmd := range abc.commands {
					if str.ToLower(update.Message.Command()) == string(cmd.command.name) {
						cmd.handleFunc(str.ToLower(update.Message.CommandArguments()))
						isCmdFound = true
						break
					}
				}
				// Show not found message
				if !isCmdFound {
					t.handleNotFound()
				}
			} else { // Handle ordinary user commands
				var isCmdFound bool
				for _, cmd := range bc.commands {
					if str.ToLower(update.Message.Command()) == string(cmd.command.name) {
						cmd.handleFunc(str.ToLower(update.Message.CommandArguments()))
						isCmdFound = true
						break
					}
				}
				// Show not found message
				if !isCmdFound {
					t.handleNotFound()
				}
			}

			// Okay, we're sending our message off! We don't care about the message
			// we just sent, so we'll discard it.
			if _, err := t.bot.Send(t.msg); err != nil {
				// Note that panics are a bad way to handle errors. Telegram can
				// have service outages or network errors, you should retry sending
				// messages or more gracefully handle failures.
				panic(err)
			}
			// Process callback messages
		} else if update.CallbackQuery != nil {
			// Respond to the callback query, telling Telegram to show the user
			// a message with the data received.
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := t.bot.Request(callback); err != nil {
				panic(err)
			}

			// Get callback data and convert json to struct
			callbackData := t.update.CallbackQuery.Data
			var message callbackMessage
			err := json.Unmarshal([]byte(callbackData), &message)
			if err != nil {
				log.Printf("Can't unmarshal data json: %v", err)
			}

			// Checking where callback come from and run specific function
			switch message.FromHandle {
			case callbackHandleRegister:
				t.callbackRegister(message.Answer, message.ChatId, message.UserId, message.MessageId)
			case callbackHandleUnregister:
				t.callbackUnregister(message.Answer, message.ChatId, message.UserId, message.MessageId)
			}
		}
	}
}

//// StartBot just star the bot
//func OldStartBot(dc *data.CalData, botToken string) {
//	bot, err := tgbotapi.NewBotAPI(botToken)
//	if err != nil {
//		panic(err)
//	}
//
//	bot.Debug = true
//
//	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
//	// to make sure Telegram knows we've handled previous values and we don't
//	// need them repeated.
//	updateConfig := tgbotapi.NewUpdate(0)
//
//	// Tell Telegram we should wait up to 30 seconds on each request for an
//	// update. This way we can get information just as quickly as making many
//	// frequent requests without having to send nearly as many.
//	updateConfig.Timeout = 30
//
//	// Start polling Telegram for updates.
//	updates := bot.GetUpdatesChan(updateConfig)
//
//	// Let's go through each update that we're getting from Telegram.
//	for update := range updates {
//		// We only want to look at command messages, so we can
//		// discard any other updates.
//		if update.Message == nil || !update.Message.IsCommand(){
//			continue
//		}
//
//		// Init empty message to fill up it later
//		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
//
//		// Go through set of allowed commands
//		switch str.ToLower(update.Message.Command()) {
//		case "start":
//			msg.Text = "Тестовый телеграм бот"
//			//msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Это ты!")
//			//msg.ReplyToMessageID = update.Message.MessageID
//
//		case "create":
//			msg.Text = "Создаю записи, ждите..."
//			if _, err := bot.Send(msg); err != nil {
//				panic(err)
//			}
//
//			err = dc.UpdateOnDutyEvents(2, 2, data.OnDutyTag)
//			if err != nil {
//				log.Printf("Error in event creating: %v", err)
//				msg.Text = fmt.Sprintf("Что-то пошло не так: %s", err)
//			} else {
//				msg.Text = "Записи созданы"
//			}
//
//		case "whowasonduty":
//			n, c, err := dc.WhoWasOnDuty(2021, time.November, data.OnDutyTag)
//			if err != nil {
//				log.Printf("Error in checking whowasonduty: %v", err)
//				msg.Text = fmt.Sprintf("Что-то пошло не так: %s", err)
//			} else {
//				msg.Text = fmt.Sprintf("Человек: %s, Кол-во дней: %d", n, c)
//			}
//
//		case "createnwd":
//			msg.Text = "Создаю записи, ждите..."
//			if _, err := bot.Send(msg); err != nil {
//				panic(err)
//			}
//			err = dc.UpdateNwdEvents(2)
//			if err != nil {
//				log.Printf("Error in event creating: %v", err)
//				msg.Text = fmt.Sprintf("Что-то пошло не так: %s", err)
//			} else {
//				msg.Text = "Записи созданы"
//			}
//
//		case "delete":
//			err = dc.DeleteDutyEvents(2, data.OnDutyTag)
//			if err != nil {
//				log.Printf("Error in event deletion: %v", err)
//				msg.Text = fmt.Sprintf("Что-то пошло не так: %s", err)
//			} else {
//				msg.Text = "Записи удалены"
//			}
//
//		case "save":
//			event, err := dc.SaveMenList()
//			if err != nil {
//				msg.Text = fmt.Sprintf("Сохранить не удалось: %v", err)
//			} else {
//				msg.Text = fmt.Sprintf("Успешно сохранено %s", *event)
//			}
//
//		case "load":
//			menOfDuty, err := dc.LoadMenList()
//			if err != nil {
//				msg.Text = fmt.Sprintf("Загрузить не удалось: %v", err)
//			} else {
//				msg.Text = fmt.Sprintln(menOfDuty)
//			}
//
//		case "register":
//			dc.AddManOnDuty(update.Message.CommandArguments(), update.Message.CommandArguments())
//			msg.Text = fmt.Sprintf("Added: %s for %s", update.Message.CommandArguments(),
//				update.Message.Chat.UserName)
//
//		case "unregister":
//			err := dc.DeleteManOnDuty(update.Message.CommandArguments())
//			if err != nil {
//				msg.Text = fmt.Sprintf("Не удалось удалить: %s", err)
//			} else {
//				msg.Text = fmt.Sprintf("Deleted: %s", update.Message.CommandArguments())
//			}
//
//		case "whoisondutynow":
//			tn := time.Now().AddDate(0, 0, 4)
//			man, err := dc.WhoIsOnDuty(&tn, data.OnDutyTag)
//			if err != nil {
//				msg.Text = fmt.Sprintf("Не могу найти дежурного: %s", err)
//			} else {
//				msg.Text = fmt.Sprintf("Дежурный: %s", man)
//			}
//
//		case "list":
//			var list string
//			menList, err := dc.ShowMenOnDutyList()
//			if err != nil {
//				msg.Text = fmt.Sprintf("Возникла ошибка при загрузке: %s", err)
//				break
//			}
//
//			for _, i := range menList {
//				list += fmt.Sprintf("%s\n", i)
//			}
//			msg.Text = "*TEST* bold"
//			msg.ParseMode = "markdown"
//			_, err = bot.Send(msg)
//			if err != nil {
//				msg.Text = fmt.Sprintf("Не могу отправить сообщение: %s", err)
//			}
//			msg.Text = fmt.Sprintf("Список дежурных: \n%s", list)
//
//		default:
//			// Now that we know we've gotten a new message, we can construct a
//			// reply! We'll take the Chat ID and Text from the incoming message
//			// and use it to create a new message.
//			//msg = tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text+" hello from Bot!")
//			// We'll also say that this message is a reply to the previous message.
//			// For any other specifications than Chat ID or Text, you'll need to
//			// set fields on the `MessageConfig`.
//			msg.Text = "Команда не найдена"
//			//msg.ReplyToMessageID = update.Message.MessageID
//		}
//
//		// Okay, we're sending our message off! We don't care about the message
//		// we just sent, so we'll discard it.
//		if _, err := bot.Send(msg); err != nil {
//			// Note that panics are a bad way to handle errors. Telegram can
//			// have service outages or network errors, you should retry sending
//			// messages or more gracefully handle failures.
//			panic(err)
//		}
//	}
//}
