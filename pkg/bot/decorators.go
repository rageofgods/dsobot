package bot

import "time"

// Define callback burst control variables
var (
	isCallbackHandleRegisterFired      bool
	isCallbackHandleUnregisterFired    bool
	isCallbackHandleDeleteOffDutyFired bool
	isCallbackHandleReindexFired       bool
	isCallbackHandleEnableFired        bool
	isCallbackHandleDisableFired       bool
)

// Define function type for decorator
type callbackDecor func(string, int64, int64, int) error

// Define callback burst decorator
func burstDecorator(waitTime int, isFired *bool, c callbackDecor) callbackDecor {
	return func(s string, i int64, i2 int64, i3 int) error {
		*isFired = true
		go func(t *bool) { time.Sleep(time.Duration(waitTime) * time.Second); *t = false }(isFired)
		return c(s, i, i2, i3)
	}
}
