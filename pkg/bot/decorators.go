package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

// Define callback burst control variables
var (
	isCallbackHandleRegisterFired                bool
	isCallbackHandleRegisterHelperFired          bool
	isCallbackHandleUnregisterFired              bool
	isCallbackHandleDeleteOffDutyFired           bool
	isCallbackHandleReindexFired                 bool
	isCallbackHandleEnableFired                  bool
	isCallbackHandleDisableFired                 bool
	isCallbackHandleEditDutyFired                bool
	isCallbackHandleAnnounceFired                bool
	isCallbackHandleAddOffDutyFired              bool
	isCallbackHandleWhoIsOnDutyAtDateFired       bool
	isCallbackHandleWhoIsOnValidationAtDateFired bool
)

// Define function type for decorator
type callbackDecor func(string, int64, int64, int, *tgbotapi.Update) error

// Define callback burst decorator
func burstDecorator(waitTime int, isFired *bool, c callbackDecor) callbackDecor {
	return func(s string, i int64, i2 int64, i3 int, u *tgbotapi.Update) error {
		*isFired = true
		go func(t *bool) { time.Sleep(time.Duration(waitTime) * time.Second); *t = false }(isFired)
		return c(s, i, i2, i3, u)
	}
}
