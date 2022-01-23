package data

import (
	"context"
	"encoding/base64"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"net/http"
)

// NewCalData CalData constructor
func NewCalData(token string, calID string) *CalData {
	return &CalData{
		token:   token,
		calID:   calID,
		cal:     new(calendar.Service),
		dutyMen: new([]DutyMan),
	}
}

// Read CalData file
func (t *CalData) readToken() (*[]byte, error) {
	data, err := base64.StdEncoding.DecodeString(t.token)
	if err != nil {
		return nil, CtxError("data.readToken()", err)
	}
	return &data, nil
}

// Init HTTP client
func (t *CalData) httpClient(b *[]byte) (*http.Client, error) {
	config, err := google.JWTConfigFromJSON(*b, calendar.CalendarScope)
	if err != nil {
		return nil, CtxError("data.httpClient()", err)
	}
	httpC := config.Client(context.Background())
	return httpC, nil
}

// Init calendar service
func (t *CalData) service(httpC *http.Client) error {
	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(httpC))
	if err != nil {
		return CtxError("data.service()", err)
	}
	t.cal = srv
	return nil
}

// InitService Init service
func (t *CalData) InitService() error {
	b, err := t.readToken()
	if err != nil {
		return CtxError("data.InitService()", err)
	}
	httpC, err := t.httpClient(b)
	if err != nil {
		return CtxError("data.InitService()", err)
	}
	if err := t.service(httpC); err != nil {
		return CtxError("data.InitService()", err)
	}
	return nil
}
