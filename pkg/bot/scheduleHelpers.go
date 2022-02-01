package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	"github.com/go-co-op/gocron"
	"log"
	"time"
)

func (t *TgBot) scheduleAllHelpers() error {
	// Schedule per-day (expect non-working days) announcements for group channels
	if err := t.scheduleAnnounce("09:00:00"); err != nil {
		log.Printf("%v", err)
	}
	// Schedule per-month event creation for non-working days
	if err := t.scheduleCreateNWD("00:00:01"); err != nil {
		log.Printf("%v", err)
	}
	// Schedule per-month event creation for on-duty days
	if err := t.scheduleCreateOnDuty("00:00:03"); err != nil {
		log.Printf("%v", err)
	}
	// Schedule per-month event creation for on-validation days
	if err := t.scheduleCreateOnValidation("00:00:03"); err != nil {
		log.Printf("%v", err)
	}
	// Schedule per-month event creation for on-validation days
	if err := t.scheduleCreateBackupForData("00:00:10"); err != nil {
		log.Printf("%v", err)
	}
	// Schedule per-month event creation for on-validation days
	if err := t.scheduleCreateBackupForSettings("00:00:11"); err != nil {
		log.Printf("%v", err)
	}
	return nil
}

// Send announce message based on announce group status
func (t *TgBot) scheduleAnnounce(timeString string) error {
	loc, err := time.LoadLocation(data.TimeZone)
	if err != nil {
		return err
	}

	s := gocron.NewScheduler(loc)
	j, err := s.Every(1).Day().At(timeString).Do(t.announceDuty)
	if err != nil {
		return fmt.Errorf("can't schedule announce message. job: %v: error: %v", j, err)
	}
	s.StartAsync()
	return nil
}

// Create non-working events
func (t *TgBot) scheduleCreateNWD(timeString string) error {
	loc, err := time.LoadLocation(data.TimeZone)
	if err != nil {
		return err
	}

	s := gocron.NewScheduler(loc)
	j, err := s.Every(1).Month(1).At(timeString).Do(t.updateNwd)
	if err != nil {
		return fmt.Errorf("can't schedule nwd creating message. job: %v: error: %v", j, err)
	}
	s.StartAsync()
	return nil
}

// Create non-working events
func (t *TgBot) scheduleCreateOnDuty(timeString string) error {
	loc, err := time.LoadLocation(data.TimeZone)
	if err != nil {
		return err
	}

	s := gocron.NewScheduler(loc)
	j, err := s.Every(1).Month(1).At(timeString).Do(t.updateOnDuty)
	if err != nil {
		return fmt.Errorf("can't schedule on-duty creating message. job: %v: error: %v", j, err)
	}
	s.StartAsync()
	return nil
}

// Create non-working events
func (t *TgBot) scheduleCreateOnValidation(timeString string) error {
	loc, err := time.LoadLocation(data.TimeZone)
	if err != nil {
		return err
	}

	s := gocron.NewScheduler(loc)
	j, err := s.Every(1).Month(1).At(timeString).Do(t.updateOnValidation)
	if err != nil {
		return fmt.Errorf("can't schedule on-validation creating message. job: %v: error: %v", j, err)
	}
	s.StartAsync()
	return nil
}

// Create and rotate backups
func (t *TgBot) scheduleCreateBackupForData(timeString string) error {
	loc, err := time.LoadLocation(data.TimeZone)
	if err != nil {
		return err
	}

	s := gocron.NewScheduler(loc)
	j, err := s.Every(1).Day().At(timeString).Do(func() {
		if err := t.dc.BackupData(data.SaveNameForDutyMenData, 7); err != nil {
			messageText := fmt.Sprintf("Не удалось создать файл бэкапа для %s: %v",
				data.SaveNameForDutyMenData,
				err)
			if err := t.sendMessage(messageText,
				t.adminGroupId,
				nil,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		}
	})
	if err != nil {
		return fmt.Errorf("can't schedule on-validation creating message. job: %v: error: %v", j, err)
	}
	s.StartAsync()
	return nil
}

// Create and rotate backups
func (t *TgBot) scheduleCreateBackupForSettings(timeString string) error {
	loc, err := time.LoadLocation(data.TimeZone)
	if err != nil {
		return err
	}

	s := gocron.NewScheduler(loc)
	j, err := s.Every(1).Day().At(timeString).Do(func() {
		if err := t.dc.BackupData(data.SaveNameForBotSettings, 7); err != nil {
			messageText := fmt.Sprintf("Не удалось создать файл бэкапа для %s: %v",
				data.SaveNameForBotSettings,
				err)
			if err := t.sendMessage(messageText,
				t.adminGroupId,
				nil,
				nil); err != nil {
				log.Printf("unable to send message: %v", err)
			}
		}
	})
	if err != nil {
		return fmt.Errorf("can't schedule on-validation creating message. job: %v: error: %v", j, err)
	}
	s.StartAsync()
	return nil
}
