package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
	"time"
)

func genInlineYesNoKeyboardWithData(yes *callbackMessage, no *callbackMessage) (*tgbotapi.InlineKeyboardMarkup, error) {
	// Generate jsons for data
	jsonYes, err := marshalCallbackData(*yes)
	if err != nil {
		log.Println(err)
	}
	jsonNo, err := marshalCallbackData(*no)
	if err != nil {
		log.Println(err)
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
		jsonData, err := marshalCallbackData(cm)
		if err != nil {
			log.Println(err)
			return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
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
		jsonData, err := marshalCallbackData(cm)
		if err != nil {
			log.Println(err)
			return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
		}
		row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d. %s (@%s)",
			i+1, v.CustomName, v.UserName), string(jsonData)))
		rows = append(rows, row)
	}

	// Add row with ok/cancel buttons
	cmYes, cmNo := cm, cm
	cmYes.Answer = inlineKeyboardYes
	cmNo.Answer = inlineKeyboardNo
	jsonDataYes, err := marshalCallbackData(cmYes)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	jsonDataNo, err := marshalCallbackData(cmNo)
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
			tgbotapi.NewInlineKeyboardButtonData(strings.ToUpper(dt), inlineKeyboardVoid))
	}
	row := tgbotapi.NewInlineKeyboardRow(keyboardButtons...)
	rows = append(rows, row)
	// Iterate over all duty men
	for manIndex, man := range *dm {
		jsonData, err := marshalCallbackDataWithIndex(cm, manIndex, 0)
		if err != nil {
			return nil, err
		}
		// Add leftmost button to hold man name
		var keyboardButtons []tgbotapi.InlineKeyboardButton
		var manButtonCaption string
		if man.Enabled {
			manButtonCaption = man.CustomName
		} else {
			manButtonCaption = fmt.Sprintf("❗️%s", man.CustomName)
		}
		keyboardButtons = append(keyboardButtons,
			tgbotapi.NewInlineKeyboardButtonData(manButtonCaption, string(jsonData)))
		// Iterate over currently supported duty types
		for _, dt := range data.DutyTypes {
			for dutyIndex, d := range man.DutyType {
				if dt == d.Type {
					// Generate jsonData with current man's duty type state (false/true)
					jsonData, err := marshalCallbackDataWithIndex(cm, manIndex, dutyIndex, d.Enabled)
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
			return nil, fmt.Errorf("unable to generate keyboard buttons for: *@%s*", man.CustomName)
		}
		row := tgbotapi.NewInlineKeyboardRow(keyboardButtons...)
		rows = append(rows, row)
	}

	// Add row with ok/cancel buttons
	cmYes, cmNo := cm, cm
	cmYes.Answer = inlineKeyboardYes
	cmNo.Answer = inlineKeyboardNo
	jsonDataYes, err := marshalCallbackData(cmYes)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	jsonDataNo, err := marshalCallbackData(cmNo)
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
		jsonData, err := marshalCallbackDataWithIndex(cm, groupIndex, 0)
		if err != nil {
			return nil, err
		}
		// Add leftmost button to hold group title
		var keyboardButtons []tgbotapi.InlineKeyboardButton
		groupButtonCaption := group.Title

		keyboardButtons = append(keyboardButtons,
			tgbotapi.NewInlineKeyboardButtonData(groupButtonCaption, string(jsonData)))

		// Generate jsonData with current group's announce type state (false/true)
		jsonData, err = marshalCallbackDataWithIndex(cm, groupIndex, groupIndex, group.Announce)
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
	jsonDataYes, err := marshalCallbackData(cmYes)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	jsonDataNo, err := marshalCallbackData(cmNo)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	row = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Готово", string(jsonDataYes)),
		tgbotapi.NewInlineKeyboardButtonData("Отмена", string(jsonDataNo)))
	rows = append(rows, row)

	return rows, nil
}

// Generate keyboard with calendar data
func genInlineCalendarKeyboard(date time.Time,
	cm callbackMessage,
	selectedDay ...int) (*tgbotapi.InlineKeyboardMarkup, error) {
	// Create numeric inline keyboard
	var rows [][]tgbotapi.InlineKeyboardButton
	// Generate header with next/prev buttons
	var buttonsHeader []tgbotapi.InlineKeyboardButton
	// Generate answer with 'buttonType-currentDate'
	cm.Answer = fmt.Sprintf("%s-%s", inlineKeyboardPrev, date.Format(botDataShort4))
	jsonDataPrev, err := marshalCallbackData(cm)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	buttonsHeader = append(buttonsHeader, tgbotapi.NewInlineKeyboardButtonData("⬅️", string(jsonDataPrev)))
	buttonsHeader = append(buttonsHeader, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s %d",
		locMonth(date.Month()), date.Year()),
		inlineKeyboardVoid))
	// Generate answer with 'buttonType-currentDate'
	cm.Answer = fmt.Sprintf("%s-%s", inlineKeyboardNext, date.Format(botDataShort4))
	jsonDataNext, err := marshalCallbackData(cm)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	buttonsHeader = append(buttonsHeader, tgbotapi.NewInlineKeyboardButtonData("➡️", string(jsonDataNext)))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(buttonsHeader...))
	// Generate header with types of the day
	var dayTypesHeader []tgbotapi.InlineKeyboardButton
	dayTypes := [7]struct {
		weekday time.Weekday
		dayName string
	}{
		{weekday: time.Monday, dayName: "П"},
		{weekday: time.Tuesday, dayName: "В"},
		{weekday: time.Wednesday, dayName: "С"},
		{weekday: time.Thursday, dayName: "Ч"},
		{weekday: time.Friday, dayName: "П"},
		{weekday: time.Saturday, dayName: "С"},
		{weekday: time.Sunday, dayName: "В"},
	}
	for _, dt := range dayTypes {
		dayTypesHeader = append(dayTypesHeader, tgbotapi.NewInlineKeyboardButtonData(dt.dayName, inlineKeyboardVoid))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(dayTypesHeader...))
	// Generate rows with calendar data
	firstDay, lastDay, err := data.FirstLastMonthDay(1, date.Year(), int(date.Month()))
	if err != nil {
		return nil, err
	}
	// Iterate over days and weekday to fill up calendar buttons data
	for d := *firstDay; d.Before(*lastDay); {
		var calendarDays []tgbotapi.InlineKeyboardButton
		for _, dt := range dayTypes {
			// Generate calendar buttons data
			if d.Weekday() == dt.weekday {
				if d.Before(*lastDay) && d.After(time.Now().Add(time.Hour*-24)) {
					cm.Answer = fmt.Sprintf("%s-%s", inlineKeyboardDate, d.Format(botDataShort4)) // Append current data at short format as an answer
					jsonData, err := marshalCallbackData(cm)
					if err != nil {
						log.Println(err)
						return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
					}
					// If we have selected day - highlight it
					if len(selectedDay) == 1 && selectedDay[0] == d.Day() {
						calendarDays = append(calendarDays,
							tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("⸨%s⸩", strconv.Itoa(d.Day())),
								string(jsonData)))
					} else {
						calendarDays = append(calendarDays,
							tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(d.Day()),
								string(jsonData)))
					}
					d = d.AddDate(0, 0, 1)
				} else {
					calendarDays = append(calendarDays,
						tgbotapi.NewInlineKeyboardButtonData("✖️", inlineKeyboardVoid))
					d = d.AddDate(0, 0, 1)
				}
			} else {
				// Add stub button if current weekday is earlier when first day of month
				calendarDays = append(calendarDays,
					tgbotapi.NewInlineKeyboardButtonData("✖️", inlineKeyboardVoid))
			}
		}
		// Add new buttons rows (whole new week)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(calendarDays...))
	}

	// Add row with ok/cancel buttons
	cmYes, cmNo := cm, cm
	cmYes.Answer = inlineKeyboardYes
	cmNo.Answer = inlineKeyboardNo
	jsonDataYes, err := marshalCallbackData(cmYes)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	jsonDataNo, err := marshalCallbackData(cmNo)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Готово", string(jsonDataYes)),
		tgbotapi.NewInlineKeyboardButtonData("Отмена", string(jsonDataNo)))
	rows = append(rows, row)

	inlineMarkupKeyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return &inlineMarkupKeyboard, nil
}
