package bot

import (
	"bytes"
	"dso_bot/pkg/data"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang/freetype/truetype"
	"github.com/rageofgods/gridder"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"image/color"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Get requested user info
func (t *TgBot) getChatMember(userId int64, chatId int64) (*tgbotapi.ChatMember, error) {
	u, err := t.bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatId,
			UserID: userId}})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Reply to user message
func (t *TgBot) sendMessage(message string, chatId int64, replyId *int, keyboard interface{}, pin ...bool) error {
	msg := tgbotapi.NewMessage(chatId, message)

	msg.ParseMode = "markdown"
	if replyId != nil {
		msg.ReplyToMessageID = *replyId
	}
	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}
	// Reply to message
	sentMessage, err := t.bot.Send(msg)
	if err != nil {
		return err
	}
	// Check if we want to sent message
	if len(pin) == 1 {
		if pin[0] {
			if err := pinMessage(t, chatId, sentMessage); err != nil {
				message = fmt.Sprintf("%v", err)
				if err := t.sendMessage(message, t.adminGroupId, nil, nil); err != nil {
					log.Printf("%v", err)
				}
				return err
			}
		}
	}
	return nil
}

func pinMessage(t *TgBot, chatId int64, sentMessage tgbotapi.Message) error {
	// Unpin previous announce message
	for _, v := range t.settings.JoinedGroups {
		if v.Id == chatId && v.LastMessageId != 0 {
			unpin := tgbotapi.UnpinChatMessageConfig{MessageID: v.LastMessageId,
				ChatID: chatId}
			_, err := t.bot.Request(unpin)
			if err != nil {
				return fmt.Errorf("не удалось открепить сообщение для chatID: %d\nОшибка: (%v)", chatId, err)
			}
		}
	}
	// Pin announce message
	pin := tgbotapi.PinChatMessageConfig{MessageID: sentMessage.MessageID,
		ChatID:              chatId,
		DisableNotification: true}
	_, err := t.bot.Request(pin)
	if err != nil {
		return fmt.Errorf("Не удалось закрепить сообщение для chatID: %d\nОшибка: (%v)", chatId, err)
	}
	// Save pinned message id
	for i, v := range t.settings.JoinedGroups {
		if v.Id == chatId {
			t.settings.JoinedGroups[i].LastMessageId = sentMessage.MessageID
		}
	}
	// Save bot settings with new data
	if err := t.dc.SaveBotSettings(&t.settings); err != nil {
		return fmt.Errorf("unable to save bot settings: %v", err)
	}
	return nil
}

func genHelpCmdText(commands []botCommand) string {
	var cmdList string
	for i, cmd := range commands {
		var argList string
		if cmd.command.args != nil {
			argList = "*Возможные значения аргумента:*\n"
			for index, arg := range *cmd.command.args {
				argList += fmt.Sprintf("  *%s*: *%s* %q\n",
					string(rune('a'-1+index+1)), // Convert number 1,2,3,etc. to char accordingly a,b,c,etc.
					arg.name,
					arg.description,
				)
			}
		}
		// Append <argument> suffix to command help if any arguments was found
		var argType string
		if argList != "" {
			argType = " *<аргумент>*"
		}
		// Generate lit of commands
		cmdList += fmt.Sprintf("*%d*: */%s*%s - `%s`\n%s",
			i+1,
			cmd.command.name,
			argType,
			cmd.description,
			argList)
	}
	return cmdList
}

// Covert weekday to localized weekday
func locWeekday(weekday time.Weekday) string {
	var locWeekday string
	switch weekday {
	case time.Sunday:
		locWeekday = "Воскресенье"
	case time.Monday:
		locWeekday = "Понедельник"
	case time.Tuesday:
		locWeekday = "Вторник"
	case time.Wednesday:
		locWeekday = "Среда"
	case time.Thursday:
		locWeekday = "Четверг"
	case time.Friday:
		locWeekday = "Пятница"
	case time.Saturday:
		locWeekday = "Суббота"
	}
	return locWeekday
}

// Covert weekday to localized weekday
func locMonth(month time.Month) string {
	var locMonth string
	switch month {
	case time.January:
		locMonth = "Январь"
	case time.February:
		locMonth = "Февраль"
	case time.March:
		locMonth = "Март"
	case time.April:
		locMonth = "Апрель"
	case time.May:
		locMonth = "Май"
	case time.June:
		locMonth = "Июнь"
	case time.July:
		locMonth = "Июль"
	case time.August:
		locMonth = "Август"
	case time.September:
		locMonth = "Сентябрь"
	case time.October:
		locMonth = "Октябрь"
	case time.November:
		locMonth = "Ноябрь"
	case time.December:
		locMonth = "Декабрь"
	}
	return locMonth
}

// Return string with duty types data for specified man
func typesOfDuties(m *data.DutyMan) string {
	var list string
	var isAnyDuty bool
	for _, dt := range m.DutyType {
		if dt.Enabled {
			list += fmt.Sprintf("%s, ", dt.Name)
			isAnyDuty = true
		}
	}
	if !isAnyDuty {
		return "❗️"
	}
	return strings.Trim(strings.Trim(list, " "), ",")
}

// Return json marshaled object for callback message data
func marshalCallbackDataWithIndex(cm callbackMessage, itemIndex int, buttonIndex int, enabled ...bool) ([]byte, error) {
	// Generate callback data
	// Format: 'itemIndex-buttonIndex-Answer'
	cm.Answer = strconv.Itoa(itemIndex)                        // Save current item index to data
	cm.Answer += fmt.Sprintf("-%s", strconv.Itoa(buttonIndex)) // Save current button index to data
	// If we have optional argument
	if len(enabled) == 1 {
		// Append current button state as suffix
		// '-1' - button is 'active' ✅
		// '-0' - button is 'passive' ❌
		if enabled[0] {
			cm.Answer += fmt.Sprintf("-%s", inlineKeyboardEditDutyYes)
		} else {
			cm.Answer += fmt.Sprintf("-%s", inlineKeyboardEditDutyNo)
		}
	}
	// Save our data
	jsonData, err := json.Marshal(cm)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	// Maximum data size allowed by Telegram is 64b https://github.com/yagop/node-telegram-bot-api/issues/706
	if len(jsonData) > 64 {
		return nil, fmt.Errorf("jsonNo size is greater then 64b: %v", len(jsonData))
	}
	return jsonData, nil
}

// Run separated goroutines to recreate duty event started from specific date
func (t *TgBot) updateOnDutyEvents(startFrom *time.Time, userName string, timeRangeText string) {
	// Recreate calendar duty event from current date if added duty in landed at this month
	if startFrom.Month() == time.Now().Month() {
		var wg = sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			if err := t.dc.UpdateOnDutyEventsFrom(startFrom, data.OnDutyContDays, data.OnDutyTag); err != nil {
				log.Printf("unable to update duty events: %v", err)
				messageText := fmt.Sprintf("Не удалось пересоздать события дежурства при "+
					"добавлении нового нерабочего периода для: %s (временной период: %s)",
					userName, timeRangeText)
				if err := t.sendMessage(messageText,
					t.adminGroupId,
					nil,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
			}
		}()
		go func() {
			defer wg.Done()
			if err := t.dc.UpdateOnDutyEventsFrom(startFrom, data.OnValidationContDays, data.OnValidationTag); err != nil {
				log.Printf("unable to update duty events: %v", err)
				messageText := fmt.Sprintf("Не удалось пересоздать события валидации при "+
					"добавлении нового нерабочего периода для: %s (временной период: %s)",
					userName, timeRangeText)
				if err := t.sendMessage(messageText,
					t.adminGroupId,
					nil,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}
			}
		}()
		wg.Wait()
	}
}

// Check if we have date in the command argument
func parseDateString(str string) (time.Time, error) {
	loc, err := time.LoadLocation(data.TimeZone)
	if err != nil {
		return time.Time{}, err
	}

	tn, err := time.ParseInLocation(botDataShort1, str, loc)
	if err != nil {
		tn, err = time.ParseInLocation(botDataShort2, str, loc)
		if err != nil {
			tn, err = time.ParseInLocation(botDataShort3, str, loc)
			if err != nil {
				return tn, fmt.Errorf("Не удалось произвести парсинг даты: %v\n\n"+
					"Доступны следующие форматы:\n"+
					"*%q*\n"+
					"*%q*\n"+
					"*%q*\n", err, botDataShort1, botDataShort2, botDataShort3)
			}
		}
	}
	return tn, nil
}

// Check if provided off-duty period is don't overlap with existing periods
func (t *TgBot) isOffDutyDatesOverlapWithCurrent(startDate time.Time, endDate time.Time,
	chatId int64, userId int64, messageId int) (bool, error) {
	// Get current MenData
	dutyMen := t.dc.DutyMenData()
	//Iterate over all men
	for _, man := range *dutyMen {
		// Find current man
		if man.TgID == userId {
			// Iterate over all available duties
			for _, offDuty := range man.OffDuty {
				// Convert string data to time object and check for parsing errors
				startOffDuty, err := parseDateString(offDuty.OffDutyStart)
				if err != nil {
					return true, err
				}
				// Convert string data to time object and check for parsing errors
				endOffDuty, err := parseDateString(offDuty.OffDutyEnd)
				if err != nil {
					messageText := fmt.Sprintf("%v", err)
					if err := t.sendMessage(messageText,
						chatId,
						&messageId,
						nil); err != nil {
						log.Printf("unable to send message: %v", err)
					}
					return true, err
				}

				// If new added start duty is in scope for old ones
				if (startDate.After(startOffDuty) || startDate.Equal(startOffDuty)) &&
					(startDate.Before(endOffDuty) || startDate.Equal(endOffDuty)) {
					messageText := fmt.Sprintf("Начало (%s) добавляемого нерабочего периода не должно "+
						"пересекаться с уже существующими нерабочими периодами.\n\nПересечение с периодом: %s-%s",
						startDate, offDuty.OffDutyStart, offDuty.OffDutyEnd)
					return true, fmt.Errorf(messageText)
				} else if (endDate.After(startOffDuty) || endDate.Equal(startOffDuty)) &&
					(endDate.Before(endOffDuty) || endDate.Equal(endOffDuty)) {
					messageText := fmt.Sprintf("Конец (%s) добавляемого нерабочего периода не должен "+
						"пересекаться с уже существующими нерабочими периодами.\n\nПересечение с периодом: %s-%s",
						endDate, offDuty.OffDutyStart, offDuty.OffDutyEnd)
					return true, fmt.Errorf(messageText)
				} else if startDate.Before(startOffDuty) && endDate.After(endOffDuty) {
					messageText := fmt.Sprintf("Период времени нового нерабочего периода не должен "+
						"пересекаться с уже существующими нерабочими периодами.\n\nПересечение с периодом: %s-%s",
						offDuty.OffDutyStart, offDuty.OffDutyEnd)
					return true, fmt.Errorf(messageText)
				}
			}
		}
	}
	return false, nil
}

func marshalCallbackData(data callbackMessage) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("unable to marshall json to persist data: %v", err)
	}
	// Maximum data size allowed by Telegram is 64b https://github.com/yagop/node-telegram-bot-api/issues/706
	if len(jsonData) > 64 {
		return nil, fmt.Errorf("callback data size is greater than 64b: %v", len(jsonData))
	}
	return jsonData, nil
}

// Spawn dedicated goroutine to handle graceful shutdown process
// TODO refactor this code to support contexts statuses
func (t *TgBot) gracefulWatcher() {
	c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		curGo := runtime.NumGoroutine()
		// Will block until catch signal
		sig := <-c

		for {
			if runtime.NumGoroutine() <= curGo {
				fmt.Printf("caught sig: %s", sig.String())
				messageText := fmt.Sprintf("⚠️ *%s (@%s)* штатно завершает свою работу.\nСигнал выхода: *%q*",
					t.bot.Self.FirstName,
					t.bot.Self.UserName,
					sig.String())
				if err := t.sendMessage(messageText,
					t.adminGroupId,
					nil,
					nil); err != nil {
					log.Printf("unable to send message: %v", err)
				}

				log.Printf("Exiting now...")
				os.Exit(0)
			} else {
				log.Printf("Current count is: %v, target number is: %v", runtime.NumGoroutine(), curGo)
				log.Println("Preparing graceful shutdown. Please be patient...")
			}
			time.Sleep(time.Millisecond * 500)
		}
	}()
}

// This function generate correct next month next to avoid wired behaviour with Go date normalization
func nextMonth(t time.Time) (time.Time, error) {
	_, lastMonthDay, err := data.FirstLastMonthDay(1, t.Year(), int(t.Month()))
	if err != nil {
		return time.Time{}, err
	}
	// Get zone offset to be able to truncate correctly
	_, offset := t.Zone()
	return lastMonthDay.AddDate(0,
		0,
		1).Truncate(time.Hour * 24).Add(time.Second * -time.Duration(offset)), nil
}

// This function generate correct previous month next to avoid wired behaviour with Go date normalization
func prevMonth(t time.Time) (time.Time, error) {
	firstMonthDay, _, err := data.FirstLastMonthDay(1, t.Year(), int(t.Month()))
	if err != nil {
		return time.Time{}, err
	}
	fmt.Println(*firstMonthDay)
	return firstMonthDay.AddDate(0,
		0,
		-1).Truncate(time.Second * 1), nil
}

// Check if provided day for month is in off-duty range
func isDayInOffDutyRange(offDuty *data.OffDutyData, day int, month time.Month, year int) (bool, error) {
	loc, err := time.LoadLocation(data.TimeZone)
	if err != nil {
		return false, err
	}

	targetDate := time.Date(year, month, day, 0, 0, 0, 0, loc)
	startOffDutyDate, err := time.ParseInLocation(botDataShort3, offDuty.OffDutyStart, loc)
	if err != nil {
		return false, err
	}
	endOffDutyDate, err := time.ParseInLocation(botDataShort3, offDuty.OffDutyEnd, loc)
	if err != nil {
		return false, err
	}

	if (targetDate.After(startOffDutyDate) || targetDate.Equal(startOffDutyDate)) &&
		(targetDate.Before(endOffDutyDate) || targetDate.Equal(endOffDutyDate)) {
		return true, nil
	}

	return false, nil
}

// Check if provided month is in off-duty range array
func isMonthInOffDutyData(offDutyData []data.OffDutyData, month time.Month, year int) (bool, error) {
	firstMonthDay, lastMonthDay, err := data.FirstLastMonthDay(1, year, int(month))
	if err != nil {
		return false, err
	}
	for _, v := range offDutyData {
		for d := *firstMonthDay; d.Before(*lastMonthDay); d = d.AddDate(0, 0, 1) {
			isDayOff, err := isDayInOffDutyRange(&v, d.Day(), month, year)
			if err != nil {
				return false, err
			}
			if isDayOff {
				return true, nil
			}
		}
	}
	return false, nil
}

func (t *TgBot) callbackAggregateUnmarshal(id string) (interface{}, callbackMessage, error) {
	var cbd interface{}
	var cbm callbackMessage
	if v, ok := t.callbackButton[id]; ok {
		cbd = v.getCallbackData()
		cbm = v.getCallbackMessage()
	} else {
		return nil, callbackMessage{}, fmt.Errorf("can't find callback data for id: %s", id)
	}
	return cbd, cbm, nil
}

func (t *TgBot) getOffDutyAnnounces(preAnnounceDays int) ([]*offDutyAnnounce, error) {
	md := t.dc.DutyMenData(true)
	tn := time.Now()
	var returnedData []*offDutyAnnounce

	for _, v := range *md {
		// Iterate over off-duty ranges for current man
		for _, vv := range v.OffDuty {
			startOffDuty, err := time.Parse(botDataShort3, vv.OffDutyStart)
			if err != nil {
				return nil, fmt.Errorf("%v", err)
			}
			ed, err := time.Parse(botDataShort3, vv.OffDutyEnd)
			if err != nil {
				return nil, fmt.Errorf("%v", err)
			}
			endOffDuty := ed.AddDate(0, 0, 1).Add(time.Nanosecond * -1)

			if vv.OffDutyAnnounced {
				if !vv.OffDutyPostAnnounced {
					// Check if off-duty end is before current time
					if tn.After(endOffDuty) {
						err := t.dc.UpdateOffDutyAnnounce(v.UserName, vv.OffDutyStart, vv.OffDutyEnd, uint(postAnnounce))
						if err != nil {
							return nil, fmt.Errorf("%v", err)
						}
						oda := &offDutyAnnounce{man: v,
							offDutyStart: vv.OffDutyStart,
							offDutyEnd:   vv.OffDutyEnd,
							announceType: postAnnounce}
						returnedData = append(returnedData, oda)
					}
				}
			} else {
				if !vv.OffDutyPreAnnounced {
					// Check if before off0duty range (Pre-announce)
					preStartOffDuty := startOffDuty.AddDate(0, 0, -preAnnounceDays)
					if (tn.After(preStartOffDuty) || tn.Equal(preStartOffDuty)) &&
						(tn.Before(startOffDuty) || tn.Equal(startOffDuty)) {
						err := t.dc.UpdateOffDutyAnnounce(v.UserName, vv.OffDutyStart, vv.OffDutyEnd, uint(preAnnounce))
						if err != nil {
							return nil, fmt.Errorf("%v", err)
						}
						oda := &offDutyAnnounce{man: v,
							offDutyStart: vv.OffDutyStart,
							offDutyEnd:   vv.OffDutyEnd,
							announceType: preAnnounce}
						returnedData = append(returnedData, oda)
						continue
					}
				}
				// Check if in off-duty range
				if (tn.After(startOffDuty) || tn.Equal(startOffDuty)) &&
					(tn.Before(endOffDuty) || tn.Equal(endOffDuty)) {
					err := t.dc.UpdateOffDutyAnnounce(v.UserName, vv.OffDutyStart, vv.OffDutyEnd, uint(announce))
					if err != nil {
						return nil, fmt.Errorf("%v", err)
					}
					oda := &offDutyAnnounce{man: v,
						offDutyStart: vv.OffDutyStart,
						offDutyEnd:   vv.OffDutyEnd,
						announceType: announce}
					returnedData = append(returnedData, oda)
				}
			}
		}
	}
	return returnedData, nil
}

func formatOffDutyAnnounces(oda []*offDutyAnnounce) string {
	var finalStr string
	const (
		preAnn  string = "⚠️Скоро начнется нерабочий период у:\n"
		ann     string = "⚠️Начался нерабочий период у:\n"
		postAnn string = "✅Закончился нерабочий период у:\n"
	)
	preAnnStr := preAnn
	annStr := ann
	postAnnStr := postAnn

	for _, v := range oda {
		switch v.announceType {
		case preAnnounce:
			preAnnStr += fmt.Sprintf("%s (*@%s*) (*%s - %s*)\n", v.man.CustomName, v.man.UserName,
				v.offDutyStart, v.offDutyEnd)
		case announce:
			annStr += fmt.Sprintf("%s (*%s - %s*)\n", v.man.CustomName, v.offDutyStart, v.offDutyEnd)
		case postAnnounce:
			postAnnStr += fmt.Sprintf("%s (*@%s*) (*%s - %s*)\n", v.man.CustomName, v.man.UserName,
				v.offDutyStart, v.offDutyEnd)
		}
	}

	if preAnnStr != preAnn {
		finalStr += preAnnStr + "\n"
	}
	if annStr != ann {
		finalStr += annStr + "\n"
	}
	if postAnnStr != postAnn {
		finalStr += postAnnStr + "\n"
	}
	return finalStr
}

func genRndTip() string {
	tips := []string{
		"Получить общий график дежурств на месяц - */duties_csv*",
		"Получить общий график валидаций на месяц - */validation_csv*",
		"Добавить нерабочий период - */addoffduty*",
		"Показать список нерабочих периодов - */showoffduty*",
		"Удалить нерабочий период - */deleteoffduty*",
		"Добавить дату своего рожденья - */birthday*",
	}
	return tips[rand.Intn(len(tips))]
}

func initGrid(width, height, rows, columns int) (*gridder.Gridder, error) {
	imageConfig := gridder.ImageConfig{
		Width:  width,
		Height: height,
	}
	gridConfig := gridder.GridConfig{
		Rows:            rows,
		Columns:         columns,
		MarginWidth:     50,
		LineStrokeWidth: 2,
		BackgroundColor: color.White,
		ColumnsWidthOffset: []*gridder.ColumnWidthOffset{
			{
				Column: 0,
				Offset: 200,
			},
		},
	}

	grid, err := gridder.New(imageConfig, gridConfig)
	if err != nil {
		return nil, fmt.Errorf("initGrid: %w", err)
	}

	return grid, nil
}

func genGridDutyDataMatrix(t *TgBot, lastMonthDay *time.Time) ([][]string, error) {
	// Generate header based on days count for current month
	header := []string{"Имя"}
	for i := 1; i <= lastMonthDay.Day(); i++ {
		header = append(header, strconv.Itoa(i))
	}

	var menData [][]string
	menData = append(menData, header)
	nwdDays, err := data.NwdEventsForCurMonth()
	if err != nil {
		log.Printf("unable to get non-working day data: %v", err)
		return nil, fmt.Errorf("genGridDutyDataMatrix: %w", err)
	}
	for _, man := range *t.dc.DutyMenData(true) {
		// Get man month duty dates
		dutyDates, err := t.dc.ManDutiesList(man.UserName, data.OnDutyTag)
		if err != nil {
			// if Current man don't have off-duties we can skip him, because he doesn't have duties at this month also
			isManHaveOffDutyForThisMonth, err := isMonthInOffDutyData(man.OffDuty,
				lastMonthDay.Month(),
				lastMonthDay.Year())
			if err != nil {
				log.Printf("unable to check man off-duties for %s: %v", lastMonthDay.Month(), err)
				continue
			}
			if !isManHaveOffDutyForThisMonth {
				continue
			}
		}

		manData := []string{man.CustomName}
		for i := 1; i <= lastMonthDay.Day(); i++ {
			var dayString string
			// Iterate over all man duty dates for current month
			if dutyDates != nil {
				for _, dd := range *dutyDates {
					// If man duty date is equal with current month day - add it to data slice
					if dd.Day() == i {
						// Mark duty day
						dayString = "\U0001F7E9"
						break
					}
				}
			}
			// If previous step didn't modify dayString, then we need to check it for another type
			if dayString == "" {
				for _, n := range nwdDays {
					if n == i {
						// Mark nwd day
						dayString = "\U0001F7EB"
						break
					}
				}
			}
			// If previous step didn't modify dayString, then we need to check it for another type
			if dayString == "" {
				for _, v := range man.OffDuty {
					isDayOffDuty, err := isDayInOffDutyRange(&v, i, lastMonthDay.Month(), lastMonthDay.Year())
					if err != nil {
						continue
					}
					if isDayOffDuty {
						// Mark off-duty day
						dayString = "\U0001F7E7"
						break
					}
				}
			}
			// If previous step didn't modify dayString, then we need to check it for another type
			if dayString == "" {
				// Mark free of duty day
				dayString = "⬜"
			}

			manData = append(manData, dayString)
		}
		menData = append(menData, manData)
	}

	return menData, nil
}

func genGridValidationDataMatrix(t *TgBot, lastMonthDay *time.Time) ([][]string, error) {
	// Generate header based on days count for current month
	header := []string{"Имя"}
	for i := 1; i <= lastMonthDay.Day(); i++ {
		header = append(header, strconv.Itoa(i))
	}

	var menData [][]string
	menData = append(menData, header)
	nwdDays, err := data.NwdEventsForCurMonth()
	if err != nil {
		log.Printf("unable to get non-working day data: %v", err)
		return nil, fmt.Errorf("genGridValidationDataMatrix: %w", err)
	}
	for _, man := range *t.dc.DutyMenData(true) {
		// Get man month duty dates
		dutyDates, err := t.dc.ManDutiesList(man.UserName, data.OnValidationTag)
		if err != nil {
			// if Current man don't have off-dates we can skip him, because he doesn't have duties at this month also
			isManHaveOffDutyForThisMonth, err := isMonthInOffDutyData(man.OffDuty,
				lastMonthDay.Month(),
				lastMonthDay.Year())
			if err != nil {
				log.Printf("unable to check man off-duties for %s: %v", lastMonthDay.Month(), err)
				continue
			}
			if !isManHaveOffDutyForThisMonth {
				continue
			}
		}

		manData := []string{man.CustomName}
		for i := 1; i <= lastMonthDay.Day(); i++ {
			var dayString string
			// Iterate over all man duty dates for current month
			if dutyDates != nil {
				for _, dd := range *dutyDates {
					// If man duty date is equal with current month day - add it to data slice
					if dd.Day() == i {
						// Mark duty day
						dayString = "\U0001F7E9"
						break
					}
				}
			}
			// If previous step didn't modify dayString, then we need to check it for another type
			if dayString == "" {
				for _, n := range nwdDays {
					if n == i {
						// Mark nwd day
						dayString = "\U0001F7EB"
						break
					}
				}
			}
			// If previous step didn't modify dayString, then we need to check it for another type
			if dayString == "" {
				for _, v := range man.OffDuty {
					isDayOffDuty, err := isDayInOffDutyRange(&v, i, lastMonthDay.Month(), lastMonthDay.Year())
					if err != nil {
						continue
					}
					if isDayOffDuty {
						// Mark off-duty day
						dayString = "\U0001F7E7"
						break
					}
				}
			}
			// If previous step didn't modify dayString, then we need to check it for another type
			if dayString == "" {
				// Mark free of duty day
				dayString = "⬜"
			}

			manData = append(manData, dayString)
		}
		menData = append(menData, manData)
	}
	return menData, nil
}

func renderGrid(grid *gridder.Gridder, menData [][]string) error {
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal(err)
	}
	fontBold, err := truetype.Parse(gobold.TTF)
	if err != nil {
		log.Fatal(err)
	}

	cellFont := truetype.NewFace(font, &truetype.Options{Size: 14})
	DayFont := truetype.NewFace(fontBold, &truetype.Options{Size: 18})

	for i, column := range menData {
		for ii, cell := range column {
			switch cell {
			case strconv.Itoa(time.Now().Day()):
				// Current day
				if err := grid.DrawCircle(i, ii, gridder.CircleConfig{
					Color:  colornames.Red,
					Radius: 15,
				}); err != nil {
					return fmt.Errorf("renderGrid: %w", err)
				}
				// Render day
				if err := grid.DrawString(i, ii, cell, DayFont, gridder.StringConfig{
					Color: color.White,
				}); err != nil {
					return fmt.Errorf("renderGrid: %w", err)
				}
			case "\U0001F7E9":
				// Duty Day
				if err := grid.PaintCell(i, ii, color.RGBA{R: 90, G: 108, B: 22, A: 255}); err != nil {
					return fmt.Errorf("renderGrid: %w", err)
				}
			case "\U0001F7EB":
				// NWD Day
				if err := grid.PaintCell(i, ii, color.RGBA{R: 109, G: 81, B: 67, A: 255}); err != nil {
					return fmt.Errorf("renderGrid: %w", err)
				}
			case "\U0001F7E7":
				// Off-Duty Day
				if err := grid.PaintCell(i, ii, color.RGBA{R: 235, G: 114, B: 49, A: 255}); err != nil {
					return fmt.Errorf("renderGrid: %w", err)
				}
			case "⬜":
				// Free of duty Day
				if err := grid.PaintCell(i, ii, colornames.White); err != nil {
					return fmt.Errorf("renderGrid: %w", err)
				}
			default:
				if err := grid.DrawString(i, ii, cell, cellFont); err != nil {
					return fmt.Errorf("renderGrid: %w", err)
				}
			}
		}
	}
	return nil
}

func (t *TgBot) genMonthDutyImage() (*tgbotapi.FileBytes, error) {
	_, lastMonthDay, err := data.FirstLastMonthDay(1)
	if err != nil {
		return nil, fmt.Errorf("dutyImageMessage: %w", err)
	}

	// Generate menData matrix
	menData, err := genGridDutyDataMatrix(t, lastMonthDay)
	if err != nil {
		return nil, fmt.Errorf("dutyImageMessage: %w", err)
	}

	// Initialise grid object
	grid, err := initGrid(1800, 500, len(menData), len(menData[0]))
	if err != nil {
		return nil, fmt.Errorf("dutyImageMessage: %w", err)
	}

	// Render grid with menData
	if err := renderGrid(grid, menData); err != nil {
		return nil, fmt.Errorf("dutyImageMessage: %w", err)
	}

	buf := new(bytes.Buffer)

	if err := grid.EncodePNG(buf); err != nil {
		return nil, fmt.Errorf("dutyImageMessage: %w", err)
	}

	photoFileBytes := &tgbotapi.FileBytes{
		Bytes: buf.Bytes(),
	}

	return photoFileBytes, nil
}
