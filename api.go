package limesurvey

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func NewClient(url string) *LimesurveyClient {
	return &LimesurveyClient{
		c:   &http.Client{},
		url: url,
	}
}

type LimesurveyClient struct {
	c   *http.Client
	url string
	key string
}

func (l *LimesurveyClient) Authenticate(ctx context.Context, username, password string) error {
	type clientRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var session string
	err := l.call(ctx, "get_session_key", []any{
		username,
		password,
	}, &session)
	if err != nil {
		return err
	}

	l.key = session
	return nil
}

type ListSurveyResponse []struct {
	Active      string `json:"active"`
	SurveyID    int    `json:"sid"`
	SurveyTitle string `json:"surveyls_title"`
	StartDate   string `json:"startdate"`
	Expires     string `json:"expires"`
}

func (l *LimesurveyClient) ListSurveys(ctx context.Context) (ListSurveyResponse, error) {
	var surveys ListSurveyResponse
	err := l.call(ctx, "list_surveys", []any{
		l.key,
		nil,
	}, &surveys)

	return surveys, err
}

func (l *LimesurveyClient) call(ctx context.Context, method string, data []any, reply any) error {
	type clientRequest struct {
		ID     uint64 `json:"id"`
		Method string `json:"method"`
		Params []any  `json:"params"`
	}

	b, err := json.Marshal(clientRequest{
		ID:     1,
		Method: method,
		Params: data,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", l.url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header = http.Header{
		"Content-Type": {"application/json"},
		"Accept":       {"application/json"},
	}
	resp, err := l.c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	type clientResponse struct {
		ID     uint64 `json:"id"`
		Result *any   `json:"result"`
		Error  any    `json:"error"`
	}

	cr := clientResponse{
		Result: &reply,
	}

	dec := json.NewDecoder(resp.Body)
	dec.DisallowUnknownFields()

	err = dec.Decode(&cr)
	if err != nil {
		return err
	}

	if cr.Error != nil {
		return fmt.Errorf("error: %v", cr.Error)
	}

	return nil
}
