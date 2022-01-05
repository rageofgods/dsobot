package bot

import (
	"dso_bot/pkg/data"
	"encoding/json"
	"fmt"
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
			// Set default name mode to markdown
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
			// Create null message
			var msg tgbotapi.MessageConfig
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
			case callbackHandleDeleteOffDuty:
				err := t.callbackDeleteOffDuty(message.Answer, message.ChatId, message.UserId, message.MessageId)
				if err != nil {
					msg.Text = fmt.Sprintf("Возникла ошибка обработки запроса: %v", err)
					msg.ReplyToMessageID = update.CallbackQuery.Message.MessageID
					msg.ChatID = update.CallbackQuery.Message.Chat.ID
					// Send a message to user who was request access.
					if _, err := t.bot.Send(msg); err != nil {
						log.Printf("unable to send message to user: %v", err)
					}
				}
			}
		}
	}
}
