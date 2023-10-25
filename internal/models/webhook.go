package models

import "encoding/json"

type WebhookEvent struct {
	Id        string          `json:"id"`
	WebhookId string          `json:"webhook_id"`
	Type      string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

type Docver struct {
	Id          string `json:"id"`
	Status      string `json:"status"`
	ApplicantId string `json:"applicant_id"`
}

type DocverCheck struct {
	VerificationId string `json:"verification_id"`
	ApplicantId    string `json:"applicant_id"`
	Step           string `json:"step"`
	Result         struct {
		Status string `json:"status"`
		Errors []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	} `json:"result"`
}
