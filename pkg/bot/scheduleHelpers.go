package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	"github.com/go-co-op/gocron"
	"log"
	"time"
)

type botScheduler struct {
	*gocron.Scheduler
	bot *TgBot
}

func newBotScheduler(bot *TgBot) *botScheduler {
	loc, err := time.LoadLocation(data.TimeZone)
	if err != nil {
		log.Fatal("newBotScheduler: unable to load location string")
	}
	return &botScheduler{Scheduler: gocron.NewScheduler(loc), bot: bot}
}

func (t *TgBot) scheduleAllHelpers() error {
	// Create new bot scheduler
	bs := newBotScheduler(t)

	// Schedule per-day (expect non-working days) announcements for group channels
	if err := bs.scheduleAnnounceDuty("09:00:00"); err != nil {
		log.Printf("%v", err)
	}
	// Schedule per-day (expect non-working days) announcements for group channels
	if err := bs.scheduleAnnounceBirthday("09:00:00"); err != nil {
		log.Printf("%v", err)
	}
	// Schedule per-month event creation for non-working days
	if err := bs.scheduleCreateNWD("00:00:01"); err != nil {
		log.Printf("%v", err)
	}
	// Schedule per-month event creation for on-duty days
	if err := bs.scheduleCreateOnDuty("00:00:03"); err != nil {
		log.Printf("%v", err)
	}
	// Schedule per-month event creation for on-validation days
	if err := bs.scheduleCreateOnValidation("00:00:03"); err != nil {
		log.Printf("%v", err)
	}
	// Schedule per-month event creation for on-validation days
	if err := bs.scheduleCreateBackupForData("00:00:10"); err != nil {
		log.Printf("%v", err)
	}
	// Schedule per-month event creation for on-validation days
	if err := bs.scheduleCreateBackupForSettings("00:00:11"); err != nil {
		log.Printf("%v", err)
	}
	// Schedule per-dey callbackButton data purge
	if err := bs.schedulePurgeForCallbackButtonData("03:00:00"); err != nil {
		log.Printf("%v", err)
	}

	// Start all scheduled jobs
	bs.StartAsync()
	return nil
}

// Send announce message based on announce group status
func (bs botScheduler) scheduleAnnounceDuty(timeString string) error {
	// Announce Duty
	j, err := bs.Every(1).Day().At(timeString).Do(bs.bot.announceDuty)
	if err != nil {
		return fmt.Errorf("can't schedule announce message. job: %v: error: %v", j, err)
	}
	return nil
}

// Send announce message based on announce group status
func (bs botScheduler) scheduleAnnounceBirthday(timeString string) error {
	// Setup announce Birthday at working day
	j, err := bs.Every(1).Day().At(timeString).Do(bs.bot.announceBirthdayAtWorkingDay)
	if err != nil {
		return fmt.Errorf("can't schedule announce message. job: %v: error: %v", j, err)
	}
	// Setup announce Birthday at non-working day (Shift to + 3 hours)
	timeLayout := "15:04:05"
	parsedTime, err := time.Parse(timeLayout, timeString)
	if err != nil {
		return fmt.Errorf("can't schedule announce message. job: %v: error: %v", j, err)
	}

	j, err = bs.Every(1).Day().At(parsedTime.Add(time.Hour * 3)).Do(bs.bot.announceBirthdayAtNonWorkingDay)
	if err != nil {
		return fmt.Errorf("can't schedule announce message. job: %v: error: %v", j, err)
	}
	return nil
}

// Create non-working events
func (bs botScheduler) scheduleCreateNWD(timeString string) error {
	j, err := bs.Every(1).Month(1).At(timeString).Do(bs.bot.updateNwd)
	if err != nil {
		return fmt.Errorf("can't schedule nwd creating message. job: %v: error: %v", j, err)
	}
	return nil
}

// Create non-working events
func (bs botScheduler) scheduleCreateOnDuty(timeString string) error {
	j, err := bs.Every(1).Month(1).At(timeString).Do(bs.bot.updateOnDuty)
	if err != nil {
		return fmt.Errorf("can't schedule on-duty creating message. job: %v: error: %v", j, err)
	}
	return nil
}

// Create non-working events
func (bs botScheduler) scheduleCreateOnValidation(timeString string) error {
	j, err := bs.Every(1).Month(1).At(timeString).Do(bs.bot.updateOnValidation)
	if err != nil {
		return fmt.Errorf("can't schedule on-validation creating message. job: %v: error: %v", j, err)
	}
	return nil
}

// Create and rotate backups
func (bs botScheduler) scheduleCreateBackupForData(timeString string) error {
	j, err := bs.Every(1).Day().At(timeString).Do(func() {
		if err := bs.bot.dc.BackupData(data.SaveNameForDutyMenData, 7); err != nil {
			messageText := fmt.Sprintf("Не удалось создать файл бэкапа для %s: %v",
				data.SaveNameForDutyMenData,
				err)
			if err := bs.bot.sendMessage(messageText,
				bs.bot.adminGroupId,
				nil,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		}
	})
	if err != nil {
		return fmt.Errorf("can't schedule on-validation creating message. job: %v: error: %v", j, err)
	}
	return nil
}

// Create and rotate backups
func (bs botScheduler) scheduleCreateBackupForSettings(timeString string) error {
	j, err := bs.Every(1).Day().At(timeString).Do(func() {
		if err := bs.bot.dc.BackupData(data.SaveNameForBotSettings, 7); err != nil {
			messageText := fmt.Sprintf("Не удалось создать файл бэкапа для %s: %v",
				data.SaveNameForBotSettings,
				err)
			if err := bs.bot.sendMessage(messageText,
				bs.bot.adminGroupId,
				nil,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		}
	})
	if err != nil {
		return fmt.Errorf("can't schedule on-validation creating message. job: %v: error: %v", j, err)
	}
	return nil
}

// Purge Callback data
func (bs botScheduler) schedulePurgeForCallbackButtonData(timeString string) error {
	j, err := bs.Every(1).Day().At(timeString).Do(func() {
		log.Printf("Purging callback data...")
		bs.bot.callbackButton = make(map[string]callbackButton)
	})
	if err != nil {
		return fmt.Errorf("can't schedule purge for callback data. job: %v: error: %v", j, err)
	}
	return nil
}
