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
	tmpData      tmpData
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

func (t *TgBot) StartBot(version string, build string) {
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

	// Send message to admin group about current running bot build version
	messageText := fmt.Sprintf("*%s (@%s)* был запущен.\n_версия_: %q\n_билд_: %q",
		t.bot.Self.FirstName,
		t.bot.Self.UserName,
		version,
		build)
	if err := t.sendMessage(messageText,
		t.adminGroupId,
		nil,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {
		// Process adding to new group
		if update.MyChatMember != nil {
			if update.MyChatMember.NewChatMember.Status == "member" &&
				update.MyChatMember.Chat.Type == "group" {
				messageText := fmt.Sprintf("*Меня добавили в новую группу*:\n*ID*: `%d`\n*Title*: `%s`",
					update.MyChatMember.Chat.ID, update.MyChatMember.Chat.Title)
				if err := t.sendMessage(messageText,
					t.adminGroupId,
					nil,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
			}
		}
		// Process user registration
		if update.Message != nil {
			if update.Message.ReplyToMessage != nil {
				if update.Message.ReplyToMessage.From.ID == t.bot.Self.ID &&
					str.Contains(update.Message.ReplyToMessage.Text, msgTextUserHandleRegister) {
					// Show to user Yes/No message to allow him to check his Name and Surname
					t.userHandleRegisterHelper(update.Message.ReplyToMessage.MessageID)
				}
			}
		}
		// Process ordinary command messages
		if update.Message != nil && update.Message.IsCommand() {
			// Hold pointer to the current update for access inside handlers
			t.update = &update

			// Go through struct of allowed commands
			bc := t.UserBotCommands()
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
		} else if update.CallbackQuery != nil { // Process callback messages
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

			// If we got callback from not an original user - ignore it.
			//if update.CallbackQuery.From.ID != message.UserId {
			//	continue
			//}

			// Checking where callback come from and run specific function
			switch message.FromHandle {
			case callbackHandleRegister:
				if !isCallbackHandleRegisterFired {
					dec := burstDecorator(2, &isCallbackHandleRegisterFired, t.callbackRegister)
					if err := dec(message.Answer, message.ChatId, message.UserId, message.MessageId); err != nil {
						log.Printf("%v", err)
					}
				}
			case callbackHandleRegisterHelper:
				if !isCallbackHandleRegisterHelperFired {
					dec := burstDecorator(2, &isCallbackHandleRegisterHelperFired, t.callbackRegisterHelper)
					if err := dec(message.Answer, message.ChatId, message.UserId, message.MessageId); err != nil {
						log.Printf("%v", err)
					}
				}
			case callbackHandleUnregister:
				if !isCallbackHandleUnregisterFired {
					dec := burstDecorator(2, &isCallbackHandleUnregisterFired, t.callbackUnregister)
					if err := dec(message.Answer, message.ChatId, message.UserId, message.MessageId); err != nil {
						log.Printf("%v", err)
					}
				}
			case callbackHandleDeleteOffDuty:
				if !isCallbackHandleDeleteOffDutyFired {
					dec := burstDecorator(2, &isCallbackHandleDeleteOffDutyFired, t.callbackDeleteOffDuty)
					if err := dec(message.Answer, message.ChatId, message.UserId, message.MessageId); err != nil {
						messageText := fmt.Sprintf("Возникла ошибка обработки запроса: %v", err)
						if err := t.sendMessage(messageText,
							update.CallbackQuery.Message.Chat.ID,
							&update.CallbackQuery.Message.MessageID,
							nil); err != nil {
							log.Printf("unable to send message: %v", err)
						}
					}
				}
			case callbackHandleReindex:
				if !isCallbackHandleReindexFired {
					dec := burstDecorator(1, &isCallbackHandleReindexFired, t.callbackReindex)
					if err := dec(message.Answer, message.ChatId, message.UserId, message.MessageId); err != nil {
						messageText := fmt.Sprintf("Возникла ошибка обработки запроса: %v", err)
						if err := t.sendMessage(messageText,
							update.CallbackQuery.Message.Chat.ID,
							&update.CallbackQuery.Message.MessageID,
							nil); err != nil {
							log.Printf("unable to send message: %v", err)
						}
					}
				}
			case callbackHandleEnable:
				if !isCallbackHandleEnableFired {
					dec := burstDecorator(1, &isCallbackHandleEnableFired, t.callbackEnable)
					if err := dec(message.Answer, message.ChatId, message.UserId, message.MessageId); err != nil {
						messageText := fmt.Sprintf("Возникла ошибка обработки запроса: %v", err)
						if err := t.sendMessage(messageText,
							update.CallbackQuery.Message.Chat.ID,
							&update.CallbackQuery.Message.MessageID,
							nil); err != nil {
							log.Printf("unable to send message: %v", err)
						}
					}
				}
			case callbackHandleDisable:
				if !isCallbackHandleDisableFired {
					dec := burstDecorator(1, &isCallbackHandleDisableFired, t.callbackDisable)
					if err := dec(message.Answer, message.ChatId, message.UserId, message.MessageId); err != nil {
						messageText := fmt.Sprintf("Возникла ошибка обработки запроса: %v", err)
						if err := t.sendMessage(messageText,
							update.CallbackQuery.Message.Chat.ID,
							&update.CallbackQuery.Message.MessageID,
							nil); err != nil {
							log.Printf("unable to send message: %v", err)
						}
					}
				}
			case callbackHandleEditDuty:
				if !isCallbackHandleEditDutyFired {
					dec := burstDecorator(1, &isCallbackHandleEditDutyFired, t.callbackEditDuty)
					if err := dec(message.Answer, message.ChatId, message.UserId, message.MessageId); err != nil {
						messageText := fmt.Sprintf("Возникла ошибка обработки запроса: %v", err)
						if err := t.sendMessage(messageText,
							update.CallbackQuery.Message.Chat.ID,
							&update.CallbackQuery.Message.MessageID,
							nil); err != nil {
							log.Printf("unable to send message: %v", err)
						}
					}
				}
			}
		}
	}
}
