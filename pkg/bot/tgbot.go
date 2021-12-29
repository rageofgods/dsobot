package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	str "strings"
	"time"
)

// StartBot just star the bot
func StartBot(dc *data.CalData, botToken string) {
	//bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	//test := map[string]string{
	//	"Kolya":  "TG:Kolya",
	//	"Vasia":  "TG:Vasia",
	//	"Ardbeg": "TG:Ardbeg",
	//}

	//test2 := map[int]map[string]string{
	//	1: {"Kolya": "TG:Kolya"},
	//	2: {"Vasia": "TG:Vasia"},
	//	3: {"Ardbeg": "TG:Ardbeg"},
	//}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 30

	// Start polling Telegram for updates.
	updates := bot.GetUpdatesChan(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {
		// Telegram can send many types of updates depending on what your Bot
		// is up to. We only want to look at messages for now, so we can
		// discard any other updates.
		if update.Message == nil {
			continue
		}

		//var msg tgbotapi.MessageConfig
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		switch str.ToLower(update.Message.Command()) {
		case "start":
			msg.Text = "Тестовый телеграм бот"
			//msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Это ты!")
			//msg.ReplyToMessageID = update.Message.MessageID

		case "create":
			msg.Text = "Создаю записи, ждите..."
			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}

			err = dc.UpdateOnDutyEvents(2, 2, data.OnDutyTag)
			if err != nil {
				log.Printf("Error in event creating: %v", err)
				msg.Text = fmt.Sprintf("Что-то пошло не так: %s", err)
			} else {
				msg.Text = "Записи созданы"
			}

		case "whowasonduty":
			n, c, err := dc.WhoWasOnDuty(2021, time.November, data.OnDutyTag)
			if err != nil {
				log.Printf("Error in checking whowasonduty: %v", err)
				msg.Text = fmt.Sprintf("Что-то пошло не так: %s", err)
			} else {
				msg.Text = fmt.Sprintf("Человек: %s, Кол-во дней: %d", n, c)
			}

		case "createnwd":
			msg.Text = "Создаю записи, ждите..."
			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
			err = dc.UpdateNwdEvents(2)
			if err != nil {
				log.Printf("Error in event creating: %v", err)
				msg.Text = fmt.Sprintf("Что-то пошло не так: %s", err)
			} else {
				msg.Text = "Записи созданы"
			}

		case "delete":
			err = dc.DeleteDutyEvents(2, data.OnDutyTag)
			if err != nil {
				log.Printf("Error in event deletion: %v", err)
				msg.Text = fmt.Sprintf("Что-то пошло не так: %s", err)
			} else {
				msg.Text = "Записи удалены"
			}

		case "save":
			event, err := dc.SaveMenList()
			if err != nil {
				msg.Text = fmt.Sprintf("Сохранить не удалось: %v", err)
			} else {
				msg.Text = fmt.Sprintf("Успешно сохранено %s", *event)
			}

		case "load":
			menOfDuty, err := dc.LoadMenList()
			if err != nil {
				msg.Text = fmt.Sprintf("Загрузить не удалось: %v", err)
			} else {
				msg.Text = fmt.Sprintln(menOfDuty)
			}

		case "register":
			dc.AddManOnDuty(update.Message.CommandArguments(), update.Message.CommandArguments())
			msg.Text = fmt.Sprintf("Added: %s for %s", update.Message.CommandArguments(),
				update.Message.Chat.UserName)

		case "unregister":
			err := dc.DeleteManOnDuty(update.Message.CommandArguments())
			if err != nil {
				msg.Text = fmt.Sprintf("Не удалось удалить: %s", err)
			} else {
				msg.Text = fmt.Sprintf("Deleted: %s", update.Message.CommandArguments())
			}

		case "whoisondutynow":
			tn := time.Now().AddDate(0, 0, 4)
			man, err := dc.WhoIsOnDuty(&tn, data.OnDutyTag)
			if err != nil {
				msg.Text = fmt.Sprintf("Не могу найти дежурного: %s", err)
			} else {
				msg.Text = fmt.Sprintf("Дежурный: %s", man)
			}

		case "list":
			var list string
			menList, err := dc.ShowMenOnDutyList()
			if err != nil {
				msg.Text = fmt.Sprintf("Возникла ошибка при загрузке: %s", err)
				break
			}

			for _, i := range menList {
				list += fmt.Sprintf("%s\n", i)
			}
			msg.Text = "*TEST* bold"
			msg.ParseMode = "markdown"
			_, err = bot.Send(msg)
			if err != nil {
				msg.Text = fmt.Sprintf("Не могу отправить сообщение: %s", err)
			}
			msg.Text = fmt.Sprintf("Список дежурных: \n%s", list)

		default:
			// Now that we know we've gotten a new message, we can construct a
			// reply! We'll take the Chat ID and Text from the incoming message
			// and use it to create a new message.
			//msg = tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text+" hello from Bot!")
			// We'll also say that this message is a reply to the previous message.
			// For any other specifications than Chat ID or Text, you'll need to
			// set fields on the `MessageConfig`.
			msg.Text = "Команда не найдена"
			//msg.ReplyToMessageID = update.Message.MessageID
		}

		// Okay, we're sending our message off! We don't care about the message
		// we just sent, so we'll discard it.
		if _, err := bot.Send(msg); err != nil {
			// Note that panics are a bad way to handle errors. Telegram can
			// have service outages or network errors, you should retry sending
			// messages or more gracefully handle failures.
			panic(err)
		}
	}
}
