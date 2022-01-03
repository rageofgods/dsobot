package data

//CalBlue      = "1"
//CalGreen     = "2"
//CalPurple    = "3"
//CalRed       = "4"
//CalYellow    = "5"
//CalOrange    = "6"
//CalTurquoise = "7"
//CalGray      = "8"
//CalBoldBlue  = "9"
//CalBoldGreen = "10"
//CalBoldRed   = "11"

// CalTag custom type for calendar TAGs
type CalTag string

// DutyMan struct for data save
type DutyMan struct {
	Index   int           `json:"index"`
	Name    string        `json:"name"`
	TgID    string        `json:"tg-id"`
	OffDuty []OffDutyData `json:"off-d,omitempty"`
}

// OffDutyData holds off-duty data save
type OffDutyData struct {
	OffDutyStart string `json:"off-d-s"`
	OffDutyEnd   string `json:"off-d-e"`
}

const (
	// OnValidationTag Validation calendar TAG
	OnValidationTag CalTag = "validation"
	// OnDutyTag Duty calendar TAG
	OnDutyTag CalTag = "on-duty"
	// OffDutyTag Off-duty day (Holiday/Illness)
	OffDutyTag CalTag = "off-duty"
	// NonWorkingDay Non-working day calendar TAG
	NonWorkingDay CalTag = "nonworking-day"
	// NonWorkingDaySum Non-working day calendar summary text
	NonWorkingDaySum = "Нерабочий день"

	// DateShort Calendar event date format
	DateShort = "2006-01-02"
	// DateShortSaveData Date format for save into json data
	DateShortSaveData = "02/02/2006"
	// DateShortIsDayOff isDayOff date format
	DateShortIsDayOff = "20060102"
	// TimeZone local timezone
	TimeZone = "Europe/Moscow"
	// CalBlue blue color for calendar event
	CalBlue = "1"
	// CalPurple purple color for calendar event
	CalPurple = "3"
	// CalOrange orange color for calendar event
	CalOrange = "6"
	// CalGray gray color for calendar event
	CalGray = "8"

	// SearchMaxResults Default value for search filter
	SearchMaxResults = 200

	// SaveListName Summary name of Calendar event with men json
	SaveListName = "menlist.json"
	// SaveListDate Default save date
	SaveListDate = "2021-01-01"
)
