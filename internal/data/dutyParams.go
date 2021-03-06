package data

// DutyMan struct for data save
type DutyMan struct {
	Index      int           `json:"index"`
	Enabled    bool          `json:"enabled"`
	FullName   string        `json:"full-name"`
	CustomName string        `json:"custom-name"`
	UserName   string        `json:"user-name"`
	TgID       int64         `json:"tg-id"`
	OffDuty    []OffDutyData `json:"off-d,omitempty"`
	DutyType   []Duty        `json:"duty-type"`
	Birthday   string        `json:"birthday,omitempty"`
}

// OffDutyData holds off-duty data save in '02/01/2006' format
type OffDutyData struct {
	OffDutyStart         string `json:"off-d-s"`
	OffDutyEnd           string `json:"off-d-e"`
	OffDutyPreAnnounced  bool   `json:"off-d-pre-a"`
	OffDutyAnnounced     bool   `json:"off-d-a"`
	OffDutyPostAnnounced bool   `json:"off-d-post-a"`
}

// DutyType type for men duties
type DutyType string

// Duty holds type of duty
type Duty struct {
	Type    DutyType `json:"duty-type"`
	Name    string   `json:"duty-name"`
	Enabled bool     `json:"duty-enabled"`
}

//
// When defining new types of duty don't forget to initialise them at AddManOnDuty() function
//

// Continuous days for duty periods
const (
	OnDutyContDays       = 2
	OnValidationContDays = 1
)

// DutyTypes is a Variable which holds currently supported duty types
var DutyTypes = [2]DutyType{OrdinaryDutyType, ValidationDutyType}
var DutyNames = [2]string{OrdinaryDutyName, ValidationDutyName}

// Duty types const's for saving data
const (
	OrdinaryDutyType   DutyType = "ordinary-duty"
	ValidationDutyType DutyType = "validation-duty"
)

// Duty names const's for saving data
const (
	OrdinaryDutyName   string = "дежурство"
	ValidationDutyName string = "валидация"
)
