package bot

import (
	"dso_bot/pkg/data"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
)

func genInlineYesNoKeyboardWithData(yes *callbackMessage, no *callbackMessage) (*tgbotapi.InlineKeyboardMarkup, error) {
	// Generate jsons for data
	jsonYes, err := json.Marshal(yes)
	if err != nil {
		log.Println(err)
	}
	jsonNo, err := json.Marshal(no)
	if err != nil {
		log.Println(err)
	}

	// Maximum data size allowed by Telegram is 64b https://github.com/yagop/node-telegram-bot-api/issues/706
	if len(jsonNo) > 64 {
		return nil, fmt.Errorf("jsonNo size is greater then 64b: %v", len(jsonNo))
	} else if len(jsonYes) > 64 {
		return nil, fmt.Errorf("jsonYes size is greater then 64b: %v", len(jsonNo))
	}

	// Create numeric inline keyboard
	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Yes", string(jsonYes)),
			tgbotapi.NewInlineKeyboardButtonData("No", string(jsonNo)),
		),
	)
	return &numericKeyboard, nil
}

func genInlineOffDutyKeyboardWithData(offDutyList []string, cm callbackMessage) (*tgbotapi.InlineKeyboardMarkup, error) {
	// Create numeric inline keyboard
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, v := range offDutyList {
		cm.Answer = strconv.Itoa(i) // Save current index to data
		jsonData, err := json.Marshal(cm)
		if err != nil {
			log.Println(err)
			return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
		}
		// Maximum data size allowed by Telegram is 64b https://github.com/yagop/node-telegram-bot-api/issues/706
		if len(jsonData) > 64 {
			return nil, fmt.Errorf("jsonNo size is greater then 64b: %v", len(jsonData))
		}
		row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(v, string(jsonData)))
		rows = append(rows, row)
	}

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(rows...)
	return &numericKeyboard, nil
}

// Generate keyboard with available args
func genArgsKeyboard(bc *botCommands, command tCmd) [][]tgbotapi.KeyboardButton {
	var rows [][]tgbotapi.KeyboardButton
	for _, cmd := range bc.commands {
		if cmd.command.name == command {
			for _, arg := range *cmd.command.args {
				row := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(fmt.Sprintf("/%s %s",
					cmd.command.name, arg.name)))
				rows = append(rows, row)
			}
		}
	}
	return rows
}

// Generate keyboard with men on-duty indexes
func genIndexKeyboard(dm *[]data.DutyMan, cm callbackMessage) (*tgbotapi.InlineKeyboardMarkup, error) {
	// Create numeric inline keyboard
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, v := range *dm {
		cm.Answer = strconv.Itoa(i) // Save current index to data
		jsonData, err := json.Marshal(cm)
		if err != nil {
			log.Println(err)
			return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
		}
		// Maximum data size allowed by Telegram is 64b https://github.com/yagop/node-telegram-bot-api/issues/706
		if len(jsonData) > 64 {
			return nil, fmt.Errorf("jsonNo size is greater then 64b: %v", len(jsonData))
		}
		row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d. %s (%s)",
			i+1, v.FullName, v.UserName), string(jsonData)))
		rows = append(rows, row)
	}

	// Add row with ok/cancel buttons
	cmYes, cmNo := cm, cm
	cmYes.Answer = inlineKeyboardYes
	cmNo.Answer = inlineKeyboardNo
	jsonDataYes, err := json.Marshal(cmYes)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	jsonDataNo, err := json.Marshal(cmNo)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Готово", string(jsonDataYes)),
		tgbotapi.NewInlineKeyboardButtonData("Отмена", string(jsonDataNo)))
	rows = append(rows, row)

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(rows...)
	return &numericKeyboard, nil
}

// Generate keyboard with edit-duty data
func genEditDutyKeyboard(dm *[]data.DutyMan, cm callbackMessage) (*[][]tgbotapi.InlineKeyboardButton, error) {
	// Create numeric inline keyboard
	var rows [][]tgbotapi.InlineKeyboardButton
	// Generate columns names
	var keyboardButtons []tgbotapi.InlineKeyboardButton
	keyboardButtons = append(keyboardButtons,
		tgbotapi.NewInlineKeyboardButtonData("ИМЯ", inlineKeyboardVoid))
	for _, dt := range data.DutyNames {
		keyboardButtons = append(keyboardButtons,
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s", strings.ToUpper(dt)), inlineKeyboardVoid))
	}
	row := tgbotapi.NewInlineKeyboardRow(keyboardButtons...)
	rows = append(rows, row)
	// Iterate over all duty men
	for manIndex, man := range *dm {
		jsonData, err := marshalCallbackData(cm, manIndex, 0)
		if err != nil {
			return nil, err
		}
		// Add leftmost button to hold man name
		var keyboardButtons []tgbotapi.InlineKeyboardButton
		var manButtonCaption string
		if man.Enabled {
			manButtonCaption = man.FullName
		} else {
			manButtonCaption = fmt.Sprintf("❗️%s", man.FullName)
		}
		keyboardButtons = append(keyboardButtons,
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s",
				manButtonCaption),
				string(jsonData)))
		// Iterate over currently supported duty types
		for _, dt := range data.DutyTypes {
			for dutyIndex, d := range man.DutyType {
				if dt == d.Type {
					// Generate jsonData with current man's duty type state (false/true)
					jsonData, err := marshalCallbackData(cm, manIndex, dutyIndex, d.Enabled)
					if err != nil {
						return nil, err
					}
					// Generate correct buttons based on current duty type state
					if d.Enabled {
						keyboardButtons = append(keyboardButtons,
							tgbotapi.NewInlineKeyboardButtonData("✅", string(jsonData)))
					} else {
						keyboardButtons = append(keyboardButtons,
							tgbotapi.NewInlineKeyboardButtonData("❌", string(jsonData)))
					}
				}
			}
		}
		// Check if keyboard is generated correctly
		if len(keyboardButtons) == 1 {
			return nil, fmt.Errorf("unable to generate keyboard buttons for: *@%s*", man.FullName)
		}
		row := tgbotapi.NewInlineKeyboardRow(keyboardButtons...)
		rows = append(rows, row)
	}

	// Add row with ok/cancel buttons
	cmYes, cmNo := cm, cm
	cmYes.Answer = inlineKeyboardYes
	cmNo.Answer = inlineKeyboardNo
	jsonDataYes, err := json.Marshal(cmYes)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	jsonDataNo, err := json.Marshal(cmNo)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	row = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Готово", string(jsonDataYes)),
		tgbotapi.NewInlineKeyboardButtonData("Отмена", string(jsonDataNo)))
	rows = append(rows, row)

	return &rows, nil
}

// Generate final message for user after he is hit "ok" button at inline keyboard and delete keyboard with message
func (t *TgBot) delInlineKeyboardWithMessage(messageText string, chatId int64, messageId int, update *tgbotapi.Update) {
	if err := t.sendMessage(messageText,
		chatId,
		&messageId,
		nil); err != nil {
		log.Printf("unable to send message: %v", err)
	}
	// Deleting access request message in admin group
	del := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID)
	_, err := t.bot.Request(del)
	if err != nil {
		log.Printf("unable to delete message with off-duty inline keyboard: %v", err)
	}
}

// Generate keyboard with announce data
func genAnnounceKeyboard(jg []data.JoinedGroup, cm callbackMessage) ([][]tgbotapi.InlineKeyboardButton, error) {
	// Create numeric inline keyboard
	var rows [][]tgbotapi.InlineKeyboardButton
	// Generate columns names
	var keyboardButtons []tgbotapi.InlineKeyboardButton
	keyboardButtons = append(keyboardButtons,
		tgbotapi.NewInlineKeyboardButtonData("ИМЯ ГРУППЫ", inlineKeyboardVoid))
	keyboardButtons = append(keyboardButtons,
		tgbotapi.NewInlineKeyboardButtonData("ВКЛ?", inlineKeyboardVoid))

	row := tgbotapi.NewInlineKeyboardRow(keyboardButtons...)
	rows = append(rows, row)
	// Iterate over all joined groups
	for groupIndex, group := range jg {
		jsonData, err := marshalCallbackData(cm, groupIndex, 0)
		if err != nil {
			return nil, err
		}
		// Add leftmost button to hold group title
		var keyboardButtons []tgbotapi.InlineKeyboardButton
		groupButtonCaption := group.Title

		keyboardButtons = append(keyboardButtons,
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s",
				groupButtonCaption),
				string(jsonData)))

		// Generate jsonData with current group's announce type state (false/true)
		jsonData, err = marshalCallbackData(cm, groupIndex, groupIndex, group.Announce)
		if err != nil {
			return nil, err
		}
		// Generate correct buttons based on current announce type state
		if group.Announce {
			keyboardButtons = append(keyboardButtons,
				tgbotapi.NewInlineKeyboardButtonData("✅", string(jsonData)))
		} else {
			keyboardButtons = append(keyboardButtons,
				tgbotapi.NewInlineKeyboardButtonData("❌", string(jsonData)))
		}

		// Check if keyboard is generated correctly
		if len(keyboardButtons) != 2 {
			return nil, fmt.Errorf("unable to generate keyboard buttons for: *@%s*", group.Title)
		}
		row := tgbotapi.NewInlineKeyboardRow(keyboardButtons...)
		rows = append(rows, row)
	}

	// Add row with ok/cancel buttons
	cmYes, cmNo := cm, cm
	cmYes.Answer = inlineKeyboardYes
	cmNo.Answer = inlineKeyboardNo
	jsonDataYes, err := json.Marshal(cmYes)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	jsonDataNo, err := json.Marshal(cmNo)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	row = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Готово", string(jsonDataYes)),
		tgbotapi.NewInlineKeyboardButtonData("Отмена", string(jsonDataNo)))
	rows = append(rows, row)

	return rows, nil
}
