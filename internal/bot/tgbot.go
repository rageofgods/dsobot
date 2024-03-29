package bot

import (
	data2 "dso_bot/internal/data"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	str "strings"
)

type TgBot struct {
	bot            *tgbotapi.BotAPI
	dc             *data2.CalData
	token          string
	msg            *tgbotapi.MessageConfig
	adminGroupId   int64
	debug          bool
	tmpData        tmpData
	settings       data2.BotSettings
	callbackButton map[string]callbackButton
}

func NewTgBot(dc *data2.CalData, settings data2.BotSettings, token string, adminGroupId int64, debug bool) *TgBot {
	return &TgBot{
		dc:             dc,
		token:          token,
		msg:            new(tgbotapi.MessageConfig),
		adminGroupId:   adminGroupId,
		debug:          debug,
		settings:       settings,
		callbackButton: make(map[string]callbackButton),
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

	// Check and announce current bot version
	t.botCheckVersion(version, build)

	// Schedule recurring helpers
	if err := t.scheduleAllHelpers(); err != nil {
		log.Printf("%v", err)
	}

	// Catch graceful exit signals
	t.gracefulWatcher()

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {
		// Process adding to new group
		if update.MyChatMember != nil {
			// Check if bot was added to some users group
			if update.MyChatMember.NewChatMember.Status == "member" &&
				(update.MyChatMember.Chat.Type == "group" || update.MyChatMember.Chat.Type == "supergroup") {
				if err := t.botAddedToGroup(update.MyChatMember.Chat.Title, update.MyChatMember.Chat.ID); err != nil {
					log.Printf("%v", err)
				}
				messageText := fmt.Sprintf("*Меня добавили в новую группу*:\n*ID*: `%d`\n*Title*: `%s`",
					update.MyChatMember.Chat.ID, update.MyChatMember.Chat.Title)
				if err := t.sendMessage(messageText,
					t.adminGroupId,
					nil,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
			}
			// Check if bot removed from some users group
			if (update.MyChatMember.NewChatMember.Status == "left" ||
				update.MyChatMember.NewChatMember.Status == "kicked") &&
				(update.MyChatMember.Chat.Type == "group" || update.MyChatMember.Chat.Type == "supergroup") {
				if err := t.botRemovedFromGroup(update.MyChatMember.Chat.ID); err != nil {
					log.Printf("%v", err)
				}
				messageText := fmt.Sprintf("*Меня удалили из группы*:\n*ID*: `%d`\n*Title*: `%s`",
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
					t.userHandleRegisterHelper(update.Message.ReplyToMessage.MessageID, &update)
				}
			}
		}
		// Process add birthday
		if update.Message != nil {
			if update.Message.ReplyToMessage != nil {
				if update.Message.ReplyToMessage.From.ID == t.bot.Self.ID &&
					str.Contains(update.Message.ReplyToMessage.Text, msgTextUserHandleBirthday) {
					t.userHandleBirthdayHelper(update.Message.ReplyToMessage.MessageID, &update)
				}
			}
		}
		// Process ordinary command messages
		if update.Message != nil && update.Message.IsCommand() {
			// Go through struct of allowed commands
			bc := t.UserBotCommands()
			abc := t.AdminBotCommands()

			// Handle admin commands
			if update.Message.Chat.ID == t.adminGroupId {
				var isCmdFound bool
				for _, cmd := range abc.commands {
					if str.ToLower(update.Message.Command()) == string(cmd.command.name) {
						cmd.handleFunc(str.ToLower(update.Message.CommandArguments()), &update)
						isCmdFound = true
						break
					}
				}
				// Show not found message
				if !isCmdFound {
					t.handleNotFound(&update)
				}
			} else { // Handle ordinary user commands
				var isCmdFound bool
				for _, cmd := range bc.commands {
					if str.ToLower(update.Message.Command()) == string(cmd.command.name) {
						cmd.handleFunc(str.ToLower(update.Message.CommandArguments()), &update)
						isCmdFound = true
						break
					}
				}
				// Show not found message
				if !isCmdFound {
					t.handleNotFound(&update)
				}
			}
		} else if update.CallbackQuery != nil { // Process callback messages
			// Respond to the callback query, telling Telegram to show the user
			// a message with the data received.
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := t.bot.Request(callback); err != nil {
				log.Printf("unable to request callback: %v", err)
				continue
			}

			// Get callback data and convert json to struct
			callbackData := update.CallbackQuery.Data
			var message callbackMessage
			err := json.Unmarshal([]byte(callbackData), &message)
			if err != nil {
				// If we can't unmarshal callback data, let's try to find it in callbackButton
				log.Printf("%v\n", err)
				log.Printf("Current callback data is: %s\n", callbackData)
				log.Printf("Trying to find saved callback data for id: %s\n", callbackData)
				if v, ok := t.callbackButton[callbackData]; ok {
					log.Printf("Callback data is found for id: %s\n", callbackData)
					message = v.callbackMessage
					message.Answer = callbackData
				}
			}

			// If we got callback from not an original user - ignore it. (Except user registration flow)
			if message.FromHandle != callbackHandleRegister && update.CallbackQuery.From.ID != message.UserId {
				continue
			}

			// Checking where callback come from and run specific function
			switch message.FromHandle {
			case callbackHandleRegister:
				if !isCallbackHandleRegisterFired {
					dec := burstDecorator(2, &isCallbackHandleRegisterFired, t.callbackRegister)
					if err := dec(message.Answer,
						message.ChatId,
						message.UserId,
						message.MessageId,
						&update); err != nil {
						log.Printf("%v", err)
					}
				}
			case callbackHandleRegisterHelper:
				if !isCallbackHandleRegisterHelperFired {
					dec := burstDecorator(2, &isCallbackHandleRegisterHelperFired, t.callbackRegisterHelper)
					if err := dec(message.Answer,
						message.ChatId,
						message.UserId,
						message.MessageId,
						&update); err != nil {
						log.Printf("%v", err)
					}
				}
			case callbackHandleUnregister:
				if !isCallbackHandleUnregisterFired {
					dec := burstDecorator(2, &isCallbackHandleUnregisterFired, t.callbackUnregister)
					if err := dec(message.Answer,
						message.ChatId,
						message.UserId,
						message.MessageId,
						&update); err != nil {
						log.Printf("%v", err)
					}
				}
			case callbackHandleDeleteOffDuty:
				if !isCallbackHandleDeleteOffDutyFired {
					dec := burstDecorator(2, &isCallbackHandleDeleteOffDutyFired, t.callbackDeleteOffDuty)
					if err := dec(message.Answer,
						message.ChatId,
						message.UserId,
						message.MessageId,
						&update); err != nil {
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
					if err := dec(message.Answer,
						message.ChatId,
						message.UserId,
						message.MessageId,
						&update); err != nil {
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
					if err := dec(message.Answer,
						message.ChatId,
						message.UserId,
						message.MessageId,
						&update); err != nil {
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
					if err := dec(message.Answer,
						message.ChatId,
						message.UserId,
						message.MessageId,
						&update); err != nil {
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
					if err := dec(message.Answer,
						message.ChatId,
						message.UserId,
						message.MessageId,
						&update); err != nil {
						messageText := fmt.Sprintf("Возникла ошибка обработки запроса: %v", err)
						if err := t.sendMessage(messageText,
							update.CallbackQuery.Message.Chat.ID,
							&update.CallbackQuery.Message.MessageID,
							nil); err != nil {
							log.Printf("unable to send message: %v", err)
						}
					}
				}
			case callbackHandleAnnounce:
				if !isCallbackHandleAnnounceFired {
					dec := burstDecorator(1, &isCallbackHandleAnnounceFired, t.callbackAnnounce)
					if err := dec(message.Answer,
						message.ChatId,
						message.UserId,
						message.MessageId,
						&update); err != nil {
						messageText := fmt.Sprintf("Возникла ошибка обработки запроса: %v", err)
						if err := t.sendMessage(messageText,
							update.CallbackQuery.Message.Chat.ID,
							&update.CallbackQuery.Message.MessageID,
							nil); err != nil {
							log.Printf("unable to send message: %v", err)
						}
					}
				}
			case callbackHandleAddOffDuty:
				if !isCallbackHandleAddOffDutyFired {
					dec := burstDecorator(1, &isCallbackHandleAddOffDutyFired, t.callbackAddOffDuty)
					if err := dec(message.Answer,
						message.ChatId,
						message.UserId,
						message.MessageId,
						&update); err != nil {
						messageText := fmt.Sprintf("Возникла ошибка обработки запроса: %v", err)
						if err := t.sendMessage(messageText,
							update.CallbackQuery.Message.Chat.ID,
							&update.CallbackQuery.Message.MessageID,
							nil); err != nil {
							log.Printf("unable to send message: %v", err)
						}
					}
				}
			case callbackHandleWhoIsOnDutyAtDate:
				if !isCallbackHandleWhoIsOnDutyAtDateFired {
					dec := burstDecorator(1, &isCallbackHandleWhoIsOnDutyAtDateFired,
						t.callbackWhoIsOnDutyAtDate)
					if err := dec(message.Answer,
						message.ChatId,
						message.UserId,
						message.MessageId,
						&update); err != nil {
						messageText := fmt.Sprintf("Возникла ошибка обработки запроса: %v", err)
						if err := t.sendMessage(messageText,
							update.CallbackQuery.Message.Chat.ID,
							&update.CallbackQuery.Message.MessageID,
							nil); err != nil {
							log.Printf("unable to send message: %v", err)
						}
					}
				}
			case callbackHandleWhoIsOnValidationAtDate:
				if !isCallbackHandleWhoIsOnValidationAtDateFired {
					dec := burstDecorator(1, &isCallbackHandleWhoIsOnValidationAtDateFired,
						t.callbackWhoIsOnValidationAtDate)
					if err := dec(message.Answer,
						message.ChatId,
						message.UserId,
						message.MessageId,
						&update); err != nil {
						messageText := fmt.Sprintf("Возникла ошибка обработки запроса: %v", err)
						if err := t.sendMessage(messageText,
							update.CallbackQuery.Message.Chat.ID,
							&update.CallbackQuery.Message.MessageID,
							nil); err != nil {
							log.Printf("unable to send message: %v", err)
						}
					}
				}
			case callbackHandleAdminAddOffDuty:
				if !isCallbackHandleAdminAddOffDutyFired {
					dec := burstDecorator(1, &isCallbackHandleAdminAddOffDutyFired,
						t.callbackAdminAddOffDuty)
					if err := dec(message.Answer,
						message.ChatId,
						message.UserId,
						message.MessageId,
						&update); err != nil {
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
