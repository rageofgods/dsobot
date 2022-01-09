package data

import (
	"context"
	"encoding/base64"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"net/http"
)

// CalData struct for creating calendar events
type CalData struct {
	ctx     *context.Context
	token   string
	calID   string
	cal     *calendar.Service
	bToken  *[]byte
	httpC   *http.Client
	dutyMen *[]DutyMan
}

// NewCalData CalData constructor
func NewCalData(token string, calID string) *CalData {
	c := context.Background() // Init background context
	return &CalData{
		ctx:     &c,
		token:   token,
		calID:   calID,
		cal:     new(calendar.Service),
		bToken:  new([]byte),
		httpC:   new(http.Client),
		dutyMen: new([]DutyMan),
	}
}

// Read CalData file
func (t *CalData) readToken() error {
	data, err := base64.StdEncoding.DecodeString(t.token)
	if err != nil {
		return CtxError("data.readToken()", err)
	}
	t.bToken = &data
	return nil
}

// Init HTTP client
func (t *CalData) httpClient() error {
	config, err := google.JWTConfigFromJSON(*t.bToken, calendar.CalendarScope)
	if err != nil {
		return CtxError("data.httpClient()", err)
	}
	t.httpC = config.Client(*t.ctx)
	return nil
}

// Init calendar service
func (t *CalData) service() error {
	srv, err := calendar.NewService(*t.ctx, option.WithHTTPClient(t.httpC))
	if err != nil {
		return CtxError("data.service()", err)
	}
	t.cal = srv
	return nil
}

// InitService Init service
func (t *CalData) InitService() error {
	if err := t.readToken(); err != nil {
		return CtxError("data.InitService()", err)
	}
	if err := t.httpClient(); err != nil {
		return CtxError("data.InitService()", err)
	}
	if err := t.service(); err != nil {
		return CtxError("data.InitService()", err)
	}
	return nil
}
