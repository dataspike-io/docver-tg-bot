package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/dataspike-io/docver-tg-bot/internal/models"
)

type (
	bot interface {
		SendVerificationStatus(context.Context, string, string) error
		CheckLiveness(context.Context, string, string) error
	}

	TgBotHandler struct {
		bot bot
	}
)

func NewTgBotHandler(bot bot) *TgBotHandler {
	return &TgBotHandler{
		bot: bot,
	}
}

func (t *TgBotHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	var webhook models.WebhookEvent
	b, err := io.ReadAll(r.Body)
	if err != nil {
		// TODO: logging
		return
	}
	defer r.Body.Close()
	if err = json.Unmarshal(b, &webhook); err != nil {
		// TODO: logging
		return
	}

	switch webhook.Type {
	case "DOCVER":
		var docver models.Docver
		err = json.Unmarshal(webhook.Payload, &docver)
		if err != nil {
			// TODO: logging
			return
		}

		if err = t.bot.SendVerificationStatus(r.Context(), docver.ApplicantId, docver.Status); err != nil {
			// TODO: logging
		}
	case "DOCVER_CHECKS":
		var docverCheck models.DocverCheck
		err = json.Unmarshal(webhook.Payload, &docverCheck)
		if err != nil {
			// TODO: logging
			return
		}

		if docverCheck.Step == "liveness" {
			if err = t.bot.CheckLiveness(r.Context(), docverCheck.ApplicantId, docverCheck.Result.Status); err != nil {
				// TODO: logging
			}
		}
	}

}
