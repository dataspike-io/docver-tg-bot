package telegram_bot

import (
	"context"
	"errors"
	"fmt"
	"github.com/ayush6624/go-chatgpt"
	"github.com/gofrs/uuid"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/dataspike-io/docver-sdk-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// IHTTPClient is the type needed for the bot to perform HTTP requests.
type IHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ICache is the type needed for the bot to keep data of verifications.
type ICache interface {
	GetVerification(context.Context, string) (*dataspike.Verification, error)
	SetVerification(context.Context, string, *dataspike.Verification) error
	RemoveVerification(context.Context, string) error
}

type ITelegramBot interface {
	Start(ctx context.Context, offset int, timeout int)
	ParseCallback(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) error
	ParseCommand(ctx context.Context, message *tgbotapi.Message) error
	ParseDocument(ctx context.Context, message *tgbotapi.Message) error
	ParseText(ctx context.Context, message *tgbotapi.Message) error
	SendVerificationStatus(ctx context.Context, applicantID string, status string) error
	CheckLiveness(ctx context.Context, applicantId string, status string) error
}

type Option func(bot *TelegramBot)

type TelegramBot struct {
	bot        *tgbotapi.BotAPI
	dsClient   dataspike.IDataspikeClient
	gptClient  *chatgpt.Client
	httpClient IHTTPClient
	cache      ICache
	dev        bool
	prompt     string
}

func (t *TelegramBot) Start(ctx context.Context, offset, timeout int) {
	u := tgbotapi.NewUpdate(offset)
	u.Timeout = timeout

	updates := t.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			if update.CallbackQuery != nil {
				err := t.ParseCallback(ctx, update.CallbackQuery)
				if err != nil {
					// TODO: logging
					continue
				}
			}

			if update.Message == nil { // ignore any non-Message updates
				continue
			}

			if update.Message.IsCommand() {
				err := t.ParseCommand(ctx, update.Message)
				if err != nil {
					// TODO: logging
				}
			} else if update.Message.Photo != nil || update.Message.Document != nil {
				err := t.ParseDocument(ctx, update.Message)
				if err != nil {
					// TODO: logging
					_, err = t.bot.Send(tgbotapi.NewMessage(update.Message.From.ID, "For start verification, please use command /start_verification"))
					if err != nil {
						// TODO: logging
					}
				}
			} else if update.Message.Text != "" {
				err := t.ParseText(ctx, update.Message)
				if err != nil {
					// TODO: logging
					_, err = t.bot.Send(tgbotapi.NewMessage(update.Message.From.ID, "Sorry, I didn't understand the message"))
					if err != nil {
						// TODO: logging
					}
				}
			}
		}
	}
}

func (t *TelegramBot) ParseCallback(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) error {
	switch callbackQuery.Data {
	case learnMrz:
		_, err := t.bot.Send(tgbotapi.NewMessage(callbackQuery.From.ID, mzrText))
		if err != nil {
			return err
		}
		photo := tgbotapi.NewPhoto(callbackQuery.From.ID, tgbotapi.FileURL(mrzLink))
		_, err = t.bot.Send(photo)
		return err
	case skipPoa:
		tgID := strconv.FormatInt(callbackQuery.From.ID, 10)
		verification, err := t.cache.GetVerification(ctx, tgID)
		if err != nil {
			return err
		}
		verification.Checks.Poa = nil
		err = t.cache.SetVerification(ctx, tgID, verification)
		if err != nil {
			// TODO: logging
			return err
		}

		return t.nextCheck(callbackQuery.From.ID, verification)
	default:
		return errors.New("undefined button data")
	}
}

func (t *TelegramBot) ParseCommand(ctx context.Context, message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.From.ID, "")
	switch message.Command() {
	case "start":
		if message.From.IsBot {
			_, err := t.bot.Send(tgbotapi.NewMessage(message.From.ID, verificationForBotIsDisabled))
			return err
		}

		arg := message.CommandArguments()
		if arg == "" {
			_, err := t.bot.Send(msg)
			return err
		}

		verification, err := t.dsClient.GetVerificationByShortID(arg)
		if err != nil {
			return err
		}

		switch verification.Status {
		case "expired":
			msg.Text = expiredText
			msg.ParseMode = tgbotapi.ModeHTML
			msg.ReplyMarkup = contactUsKeyboard
			_, err = t.bot.Send(msg)
			return err
		case verified:
			_, err = t.bot.Send(tgbotapi.NewMessage(message.From.ID, verificationCompleted))
			return err
		}

		applicant, err := t.dsClient.GetApplicantByID(uuid.FromStringOrNil(verification.ApplicantID))
		if err != nil {
			return err
		}

		tgID := strconv.FormatInt(message.From.ID, 10)
		err = t.dsClient.LinkTelegramProfile(applicant.ApplicantId, tgID)
		if err != nil {
			return err
		}

		err = t.cache.SetVerification(ctx, tgID, verification)
		if err != nil {
			return err
		}

		_, err = t.bot.Send(tgbotapi.NewMessage(message.From.ID, startText))
		if err != nil {
			return err
		}
		_, err = t.bot.Send(tgbotapi.NewMessage(message.From.ID, startVerificationInit))
		if err != nil {
			return err
		}

		return t.nextCheck(message.From.ID, verification)
	case "help":
		msg.Text = helpText
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = contactUsKeyboard
	case "cancel":
		tgID := strconv.FormatInt(message.From.ID, 10)
		v, err := t.cache.GetVerification(ctx, tgID)
		if err != nil {
			return err
		}

		err = t.dsClient.CancelVerification(uuid.FromStringOrNil(v.Id))
		if err != nil {
			return err
		}

		err = t.cache.RemoveVerification(ctx, tgID)
		if err != nil {
			return err
		}

		msg.Text = cancelText
		msg.ReplyMarkup = contactUsKeyboard
	case "ask_expert":
		msg.Text = askExpertText
	case "create_verification":
		return t.createVerificationCommand(ctx, message)
	case "customize_bot":
		msg.Text = customizeBotText
		msg.ReplyMarkup = contactUsKeyboard
	case "start_verification":
		return t.startVerificationCommand(ctx, message)
	default:
		msg.Text = "Oops, that command is new to me!"
	}

	_, err := t.bot.Send(msg)
	return err
}

func (t *TelegramBot) createVerificationCommand(ctx context.Context, message *tgbotapi.Message) error {
	if !t.dev {
		_, err := t.bot.Send(tgbotapi.NewMessage(message.From.ID, "Oops, that command is new to me!"))
		return err
	}
	tgID := strconv.FormatInt(message.From.ID, 10)
	var applicantID string
	applicant, err := t.dsClient.GetApplicantByExternalID(fmt.Sprintf("tg_user_%d", message.From.ID))
	if err != nil {
		if strings.Contains(err.Error(), "dataspike error: 404") {
			applicantID, err = t.dsClient.CreateApplicant(&dataspike.ApplicantCreate{ExternalId: fmt.Sprintf("tg_user_%d", message.From.ID)})
			if err != nil {
				return err
			}
			err = t.dsClient.LinkTelegramProfile(applicantID, tgID)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		applicantID = applicant.ApplicantId
	}

	verification, err := t.dsClient.CreateVerification(&dataspike.VerificationCreate{
		ApplicantId:       applicantID,
		ProfileId:         &defaultProfile,
		ExpirationMinutes: 600,
	})
	if err != nil {
		return err
	}

	err = t.cache.SetVerification(ctx, tgID, verification)
	if err != nil {
		return err
	}

	_, err = t.bot.Send(tgbotapi.NewMessage(message.From.ID, fmt.Sprintf("Verification created successfully. VerificationShortID: %s, ApplicantID: %s", verification.VerificationUrlId, applicantID)))
	if err != nil {
		return err
	}

	return t.nextCheck(message.From.ID, verification)
}

func (t *TelegramBot) startVerificationCommand(ctx context.Context, message *tgbotapi.Message) error {
	if message.From.IsBot {
		_, err := t.bot.Send(tgbotapi.NewMessage(message.From.ID, verificationForBotIsDisabled))
		return err
	}
	verification, err := t.cache.GetVerification(ctx, strconv.FormatInt(message.From.ID, 10))
	if err != nil {
		msg := tgbotapi.NewMessage(message.From.ID, verificationNotFound)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = contactUsKeyboard
		_, err2 := t.bot.Send(msg)
		if err2 != nil {
			return err2
		}
		return err
	}

	if verification.Status == verified {
		_, err = t.bot.Send(tgbotapi.NewMessage(message.From.ID, verificationCompleted))
		return err
	}

	_, err = t.bot.Send(tgbotapi.NewMessage(message.From.ID, startVerificationInit))
	if err != nil {
		return err
	}

	return t.nextCheck(message.From.ID, verification)
}

func (t *TelegramBot) ParseText(ctx context.Context, message *tgbotapi.Message) error {
	if t.gptClient == nil {
		msg := tgbotapi.NewMessage(message.From.ID, helpText)
		msg.ParseMode = tgbotapi.ModeHTML
		_, err := t.bot.Send(msg)
		return err
	}
	resp, err := t.gptClient.Send(ctx, &chatgpt.ChatCompletionRequest{
		Model: chatgpt.GPT35Turbo,
		Messages: []chatgpt.ChatMessage{
			{
				Role:    chatgpt.ChatGPTModelRoleSystem,
				Content: t.prompt,
			},
			{
				Role:    chatgpt.ChatGPTModelRoleUser,
				Content: message.Text,
			},
		},
	})
	if err != nil {
		return err
	}

	_, err = t.bot.Send(tgbotapi.NewMessage(message.From.ID, resp.Choices[0].Message.Content))
	return err
}

func (t *TelegramBot) getLink(message *tgbotapi.Message) (string, string, error) {
	switch {
	case message.Document != nil:
		doc, err := t.bot.GetFile(tgbotapi.FileConfig{FileID: message.Document.FileID})
		if err != nil {
			return "", "", err
		}

		return message.Document.FileName, doc.Link(t.bot.Token), nil
	case message.Photo != nil:
		var maxSize int
		var fileID string
		for _, photo := range message.Photo {
			if photo.FileSize > maxSize {
				maxSize = photo.FileSize
				fileID = photo.FileID
			}
		}

		f, err := t.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
		if err != nil {
			// TODO: logging
			return "", "", err
		}

		return "photo.jpg", f.Link(t.bot.Token), nil
	}

	return "", "", errors.New("file not found")
}

func (t *TelegramBot) ParseDocument(ctx context.Context, message *tgbotapi.Message) error {
	tgID := strconv.FormatInt(message.From.ID, 10)
	verification, err := t.cache.GetVerification(ctx, tgID)
	if err != nil {
		return err
	}

	filename, url, err := t.getLink(message)
	if err != nil {
		return err
	}

	defer t.bot.Send(tgbotapi.NewDeleteMessage(message.From.ID, message.MessageID))

	// Get the data
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch {
	case verification.Checks.DocumentMrz != nil:
		verification.Checks.DocumentMrz = nil
		return t.uploadDocument(ctx, message.From.ID, Poi, filename, verification, resp.Body)
	case verification.Checks.FaceComparison != nil:
		verification.Checks.FaceComparison = nil
		return t.uploadDocument(ctx, message.From.ID, Selfie, filename, verification, resp.Body)
	case verification.Checks.Poa != nil:
		verification.Checks.Poa = nil
		return t.uploadDocument(ctx, message.From.ID, Poa, filename, verification, resp.Body)
	default:
		return errors.New("status not supported for upload document")
	}
}

func (t *TelegramBot) uploadDocument(ctx context.Context, tgID int64, docType, filename string, verification *dataspike.Verification, file io.Reader) error {
	respDoc, err := t.dsClient.UploadDocument(&dataspike.DocumentUpload{
		DocType:     docType,
		FileName:    filename,
		ApplicantID: verification.ApplicantID,
		Reader:      file,
	})
	if err != nil {
		return err
	}
	if respDoc.Errors != nil {
		return errors.New(respDoc.Errors.String())
	}

	if docType == Poi && respDoc.DetectedTwoSideDocument != nil && *respDoc.DetectedTwoSideDocument &&
		respDoc.DetectedDocumentSide != nil && *respDoc.DetectedDocumentSide == Front {
		msg := tgbotapi.NewMessage(tgID, AttachBackSideOfDoc)
		msg.ParseMode = tgbotapi.ModeHTML
		_, err = t.bot.Send(msg)
		return err
	}

	err = t.cache.SetVerification(ctx, strconv.FormatInt(tgID, 10), verification)
	if err != nil {
		return err
	}

	return t.nextCheck(tgID, verification)
}

func (t *TelegramBot) nextCheck(chatID int64, verification *dataspike.Verification) error {
	msg := tgbotapi.NewMessage(chatID, "")
	switch {
	case verification.Checks.DocumentMrz != nil && verification.Checks.DocumentMrz.Status == pending:
		msg.Text = poiHelpForButton
		msg.ReplyMarkup = mzrKeyboard
		_, err := t.bot.Send(tgbotapi.NewMessage(chatID, poiAttachDocument))
		if err != nil {
			return err
		}
	case verification.Checks.Liveness != nil && verification.Checks.Liveness.Status == pending:
		msg.Text = LivenessPrompt
		msg.ReplyMarkup = generateLivenessKeyboard(fmt.Sprintf("%s?source=telegram&botName=%s", verification.VerificationUrl, t.bot.Self.UserName))
	case verification.Checks.FaceComparison != nil && verification.Checks.FaceComparison.Status == pending:
		msg.Text = AttachSelfiePrompt
	case verification.Checks.Poa != nil && verification.Checks.Poa.Status == pending:
		if verification.Settings.PoaRequired {
			msg.Text = SelectedPoaDocument
		} else {
			msg.Text = PoaSkipPrompt
			msg.ReplyMarkup = skipPoaKeyboard
			_, err := t.bot.Send(tgbotapi.NewMessage(chatID, SelectedPoaDocument))
			if err != nil {
				return err
			}
		}
	default:
		msg.Text = VerificationStartedPleaseWait
		err := t.dsClient.ProceedVerification(verification.VerificationUrlId)
		if err != nil {
			return err
		}
	}

	_, err := t.bot.Send(msg)
	return err
}

func (t *TelegramBot) SendVerificationStatus(ctx context.Context, applicantID, status string) error {
	applicant, err := t.dsClient.GetApplicantByID(uuid.FromStringOrNil(applicantID))
	if err != nil {
		// TODO: logging
		return err
	}

	v, err := t.cache.GetVerification(ctx, applicant.TgProfile)
	if err != nil {
		// TODO: logging
		return err
	}
	v.Status = status
	err = t.cache.SetVerification(ctx, applicant.TgProfile, v)
	if err != nil {
		// TODO: logging
		return err
	}

	tgID, err := strconv.ParseInt(applicant.TgProfile, 10, 64)
	if err != nil {
		// TODO: logging
		return err
	}

	if status != verified {
		err = t.cache.RemoveVerification(ctx, applicant.TgProfile)
		if err != nil {
			// TODO: logging
			return err
		}
		msg := tgbotapi.NewMessage(tgID, VerificationFailed)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = contactUsKeyboard
		_, err = t.bot.Send(msg)
		return err
	}

	_, err = t.bot.Send(tgbotapi.NewMessage(tgID, VerificationOk))
	return err
}

func (t *TelegramBot) CheckLiveness(ctx context.Context, applicantId, status string) error {
	applicant, err := t.dsClient.GetApplicantByID(uuid.FromStringOrNil(applicantId))
	if err != nil {
		// TODO: logging
		return err
	}

	tgID, err := strconv.ParseInt(applicant.TgProfile, 10, 64)
	if err != nil {
		// TODO: logging
		return err
	}

	v, err := t.cache.GetVerification(ctx, applicant.TgProfile)
	if err != nil {
		// TODO: logging
		return err
	}

	if status != verified {
		err = t.cache.RemoveVerification(ctx, applicant.TgProfile)
		if err != nil {
			// TODO: logging
			return err
		}
		msg := tgbotapi.NewMessage(tgID, LivenessFailed)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = contactUsKeyboard
		_, err = t.bot.Send(msg)
		return err
	}

	v.Checks.Liveness = nil
	v.Checks.FaceComparison = nil
	err = t.cache.SetVerification(ctx, applicant.TgProfile, v)
	if err != nil {
		// TODO: logging
		return err
	}

	return t.nextCheck(tgID, v)
}

// WithBuffer is a Option that allows you set size of bot buffer.
// Default value is 100
func WithBuffer(buffer int) Option {
	return func(t *TelegramBot) {
		t.bot.Buffer = buffer
	}
}

// WithGPT is a Option that allows you using chatgpt client for answer to text messages.
// When this option is nil, bot will answer default message.
func WithGPT(client *chatgpt.Client, prompt string) Option {
	return func(t *TelegramBot) {
		t.gptClient = client
		if prompt != "" {
			t.prompt = prompt
		}
	}
}

// WithSandbox is a Option add create_verification command.
func WithSandbox() Option {
	return func(t *TelegramBot) {
		t.dev = true
		menu.Commands = append(menu.Commands, tgbotapi.BotCommand{
			Command:     "/create_verification",
			Description: "Create new verification",
		})
	}
}

// WithHTTPClient is a Option that allows you set http client.
func WithHTTPClient(client IHTTPClient) Option {
	return func(t *TelegramBot) {
		t.httpClient = client
	}
}

func NewTelegramBot(bot *tgbotapi.BotAPI, dsClient dataspike.IDataspikeClient, cache ICache, options ...Option) (ITelegramBot, error) {
	dsTgBot := &TelegramBot{
		bot:        bot,
		dsClient:   dsClient,
		cache:      cache,
		httpClient: http.DefaultClient,
		prompt: `You are a helpful KYC assistant created by dataspike.io.
You're limited to respond only to requests that belong to KYC and AML domain 
and all other requests should be politely rejected as not fitting your work responsibilities.`,
	}

	for _, o := range options {
		o(dsTgBot)
	}

	_, err := bot.Request(menu)
	if err != nil {
		return nil, err
	}

	return dsTgBot, nil
}
