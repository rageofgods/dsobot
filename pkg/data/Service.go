package data

import (
	"context"
	"encoding/base64"
	"fmt"
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
	dutyMen []DutyMan
}

// NewCalData CalData constructor
func NewCalData(token string, calID string) *CalData {
	c := context.Background() // Init background context
	return &CalData{
		ctx:    &c,
		token:  token,
		calID:  calID,
		bToken: new([]byte),
	}
}

// Read CalData file
func (t *CalData) readToken() error {
	data, err := base64.StdEncoding.DecodeString(t.token)
	if err != nil {
		return err
	}
	t.bToken = &data
	return nil
}

// Init HTTP client
func (t *CalData) httpClient() error {
	config, err := google.JWTConfigFromJSON(*t.bToken, calendar.CalendarScope)
	if err != nil {
		return err
	}
	t.httpC = config.Client(*t.ctx)
	return nil
}

// Init calendar service
func (t *CalData) service() error {
	srv, err := calendar.NewService(*t.ctx, option.WithHTTPClient(t.httpC))
	if err != nil {
		return err
	}
	t.cal = srv
	return nil
}

// InitService Init service
func (t *CalData) InitService() error {
	err := t.readToken()
	if err != nil {
		return fmt.Errorf("unable to read client secret file: %v", err)
	}
	err = t.httpClient()
	if err != nil {
		return fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	err = t.service()
	if err != nil {
		return fmt.Errorf("unable to retrieve Calendar client: %v", err)
	}
	return nil
}
