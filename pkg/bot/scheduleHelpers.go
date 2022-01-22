package bot

import (
	"dso_bot/pkg/data"
	"fmt"
	"github.com/go-co-op/gocron"
	"time"
)

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
