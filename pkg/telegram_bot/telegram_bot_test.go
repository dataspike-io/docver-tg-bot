package telegram_bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/ayush6624/go-chatgpt"
	dataspike "github.com/dataspike-io/docver-sdk-go"
	"github.com/dataspike-io/docver-tg-bot/pkg/telegram_bot/mocks"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang/mock/gomock"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTelegramBot(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	httpMock := mock_telegram_bot.NewMockIHTTPClient(ctrl)

	type args struct {
		telegramToken string
		dsClient      dataspike.IDataspikeClient
		cache         ICache
		options       []Option
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "success",
			args: args{telegramToken: "123", options: []Option{WithBuffer(1), WithGPT(nil, "123"), WithSandbox(), WithHTTPClient(httpMock)}},
			err:  &tgbotapi.Error{Code: 404, Message: "Not Found", ResponseParameters: tgbotapi.ResponseParameters{}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTelegramBot(tt.args.telegramToken, tt.args.dsClient, tt.args.cache, tt.args.options...)
			assert.Equal(t, tt.err, err)

			bot, err := newBot(httpMock)
			if err != nil {
				t.Errorf("error creating bot: %s", err)
			}

			tBot := &TelegramBot{
				bot: bot,
			}
			for _, opt := range tt.args.options {
				opt(tBot)
			}
		})
	}
}

func newBot(httpMock *mock_telegram_bot.MockIHTTPClient) (*tgbotapi.BotAPI, error) {
	apiResp := tgbotapi.APIResponse{Ok: true, Result: json.RawMessage(`{"id":123}`)}
	b, err := json.Marshal(&apiResp)
	if err != nil {
		return nil, err
	}

	readCloser := io.NopCloser(bytes.NewReader(b))
	httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: readCloser}, nil)
	return tgbotapi.NewBotAPIWithClient("", "https://test.tt/bot%s/%s", httpMock)
}

func Test_telegramBot_CheckLiveness(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	httpMock := mock_telegram_bot.NewMockIHTTPClient(ctrl)
	dsMock := mock_telegram_bot.NewMockIDataspikeClient(ctrl)
	cacheMock := mock_telegram_bot.NewMockICache(ctrl)
	ctx := context.Background()

	bot, err := newBot(httpMock)
	if err != nil {
		t.Errorf("error creating bot: %s", err)
	}

	tBot := &TelegramBot{
		bot:      bot,
		dsClient: dsMock,
		cache:    cacheMock,
	}
	type args struct {
		applicantID string
		status      string
	}
	tests := []struct {
		name string
		args args
		f    func()
		err  error
	}{
		{
			name: "verified",
			args: args{"test", verified},
			f: func() {
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{TgProfile: "123"}, nil)
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Any()).Return(&dataspike.Verification{}, nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Any()).Return(nil)
				dsMock.EXPECT().ProceedVerification(gomock.Any()).Return(nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("next check failed"))
			},
			err: errors.New("next check failed"),
		},
		{
			name: "unverified",
			args: args{"test", "failed"},
			f: func() {
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{TgProfile: "123"}, nil)
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Any()).Return(&dataspike.Verification{}, nil)
				cacheMock.EXPECT().RemoveVerification(gomock.Any(), gomock.Eq("123")).Return(nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("liveness failed"))
			},
			err: errors.New("liveness failed"),
		},
		{
			name: "set verification error",
			args: args{"test", verified},
			f: func() {
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{TgProfile: "123"}, nil)
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Any()).Return(&dataspike.Verification{}, nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Any()).Return(errors.New("set verification error"))
			},
			err: errors.New("set verification error"),
		},
		{
			name: "remove verification error",
			args: args{"test", "failed"},
			f: func() {
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{TgProfile: "123"}, nil)
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Any()).Return(&dataspike.Verification{}, nil)
				cacheMock.EXPECT().RemoveVerification(gomock.Any(), gomock.Eq("123")).Return(errors.New("remove verification error"))
			},
			err: errors.New("remove verification error"),
		},
		{
			name: "get verification error",
			args: args{"test", "failed"},
			f: func() {
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{TgProfile: "123"}, nil)
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Any()).Return(nil, errors.New("get verification error"))
			},
			err: errors.New("get verification error"),
		},
		{
			name: "get applicant error",
			args: args{"test", "failed"},
			f: func() {
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(nil, errors.New("get applicant error"))
			},
			err: errors.New("get applicant error"),
		},
		{
			name: "parse tgBotID error",
			args: args{"test", "failed"},
			f: func() {
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{TgProfile: "abc"}, nil)
			},
			err: &strconv.NumError{Func: "ParseInt", Num: "abc", Err: errors.New("invalid syntax")},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.f()
			err = tBot.CheckLiveness(ctx, tt.args.applicantID, tt.args.status)
			assert.Equal(t, tt.err, err)
		})
	}
}

func Test_telegramBot_SendVerificationStatus(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	httpMock := mock_telegram_bot.NewMockIHTTPClient(ctrl)
	dsMock := mock_telegram_bot.NewMockIDataspikeClient(ctrl)
	cacheMock := mock_telegram_bot.NewMockICache(ctrl)
	ctx := context.Background()

	bot, err := newBot(httpMock)
	if err != nil {
		t.Errorf("error creating bot: %s", err)
	}

	tBot := &TelegramBot{
		bot:      bot,
		dsClient: dsMock,
		cache:    cacheMock,
	}
	type args struct {
		applicantID string
		status      string
	}
	tests := []struct {
		name string
		args args
		f    func()
		err  error
	}{
		{
			name: "verified",
			args: args{"test", verified},
			f: func() {
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{TgProfile: "123"}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Any()).Return(&dataspike.Verification{}, nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Any()).Return(nil)
			},
			err: nil,
		},
		{
			name: "unverified",
			args: args{"test", "failed"},
			f: func() {
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{TgProfile: "123"}, nil)
				cacheMock.EXPECT().RemoveVerification(gomock.Any(), gomock.Eq("123")).Return(nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("verification failed"))
			},
			err: errors.New("verification failed"),
		},
		{
			name: "remove verification error",
			args: args{"test", "failed"},
			f: func() {
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{TgProfile: "123"}, nil)
				cacheMock.EXPECT().RemoveVerification(gomock.Any(), gomock.Eq("123")).Return(errors.New("remove verification error"))
			},
			err: errors.New("remove verification error"),
		},
		{
			name: "set verification error",
			args: args{"test", verified},
			f: func() {
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{TgProfile: "123"}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Any()).Return(&dataspike.Verification{}, nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Any()).Return(errors.New("set verification error"))
			},
			err: errors.New("set verification error"),
		},
		{
			name: "get verification error",
			args: args{"test", verified},
			f: func() {
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{TgProfile: "123"}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Any()).Return(&dataspike.Verification{}, errors.New("get verification error"))
			},
			err: errors.New("get verification error"),
		},
		{
			name: "send message error",
			args: args{"test", verified},
			f: func() {
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{TgProfile: "123"}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("send message error"))
			},
			err: errors.New("send message error"),
		},
		{
			name: "get applicant error",
			args: args{"test", "failed"},
			f: func() {
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(nil, errors.New("get applicant error"))
			},
			err: errors.New("get applicant error"),
		},
		{
			name: "parse tgBotID error",
			args: args{"test", "failed"},
			f: func() {
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{TgProfile: "abc"}, nil)
			},
			err: &strconv.NumError{Func: "ParseInt", Num: "abc", Err: errors.New("invalid syntax")},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.f()
			err = tBot.SendVerificationStatus(ctx, tt.args.applicantID, tt.args.status)
			assert.Equal(t, tt.err, err)
		})
	}
}

func Test_telegramBot_Start(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	httpMock := mock_telegram_bot.NewMockIHTTPClient(ctrl)
	dsMock := mock_telegram_bot.NewMockIDataspikeClient(ctrl)
	cacheMock := mock_telegram_bot.NewMockICache(ctrl)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot, err := newBot(httpMock)
	if err != nil {
		t.Errorf("error creating bot: %s", err)
	}

	tBot := &TelegramBot{
		bot:        bot,
		dsClient:   dsMock,
		cache:      cacheMock,
		httpClient: httpMock,
	}
	type args struct {
		offset  int
		timeout int
	}
	tests := []struct {
		name string
		args args
		f    func()
	}{
		{
			name: "start",
			args: args{0, 10},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"update_id":123}`)))}, nil)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.f()
			go tBot.Start(ctx, tt.args.offset, tt.args.timeout)
			cancel()
		})
	}
}

func Test_telegramBot_callbackParse(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	httpMock := mock_telegram_bot.NewMockIHTTPClient(ctrl)
	dsMock := mock_telegram_bot.NewMockIDataspikeClient(ctrl)
	cacheMock := mock_telegram_bot.NewMockICache(ctrl)
	ctx := context.Background()

	bot, err := newBot(httpMock)
	if err != nil {
		t.Errorf("error creating bot: %s", err)
	}

	tBot := &TelegramBot{
		bot:      bot,
		dsClient: dsMock,
		cache:    cacheMock,
	}
	type args struct {
		callbackQuery *tgbotapi.CallbackQuery
	}
	tests := []struct {
		name string
		args args
		f    func()
		err  error
	}{
		{
			name: learnMrz,
			args: args{&tgbotapi.CallbackQuery{From: &tgbotapi.User{ID: 123}, Data: learnMrz}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "send message error",
			args: args{&tgbotapi.CallbackQuery{From: &tgbotapi.User{ID: 123}, Data: learnMrz}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("send message error"))
			},
			err: errors.New("send message error"),
		},
		{
			name: skipPoa,
			args: args{&tgbotapi.CallbackQuery{From: &tgbotapi.User{ID: 123}, Data: skipPoa}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Any()).Return(&dataspike.Verification{}, nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Any()).Return(nil)
				dsMock.EXPECT().ProceedVerification(gomock.Any()).Return(nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "set verification error",
			args: args{&tgbotapi.CallbackQuery{From: &tgbotapi.User{ID: 123}, Data: skipPoa}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Any()).Return(&dataspike.Verification{}, nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Any()).Return(errors.New("set verification error"))
			},
			err: errors.New("set verification error"),
		},
		{
			name: "get verification error",
			args: args{&tgbotapi.CallbackQuery{From: &tgbotapi.User{ID: 123}, Data: skipPoa}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Any()).Return(&dataspike.Verification{}, errors.New("get verification error"))
			},
			err: errors.New("get verification error"),
		},
		{
			name: "undefined button data",
			args: args{&tgbotapi.CallbackQuery{From: &tgbotapi.User{ID: 123}, Data: "default"}},
			f:    func() {},
			err:  errors.New("undefined button data"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.f()
			err = tBot.ParseCallback(ctx, tt.args.callbackQuery)
			assert.Equal(t, tt.err, err)
		})
	}
}

func Test_telegramBot_commandParse(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	httpMock := mock_telegram_bot.NewMockIHTTPClient(ctrl)
	dsMock := mock_telegram_bot.NewMockIDataspikeClient(ctrl)
	cacheMock := mock_telegram_bot.NewMockICache(ctrl)
	ctx := context.Background()

	bot, err := newBot(httpMock)
	if err != nil {
		t.Errorf("error creating bot: %s", err)
	}

	tBot := &TelegramBot{
		bot:      bot,
		dsClient: dsMock,
		cache:    cacheMock,
	}
	type args struct {
		message *tgbotapi.Message
	}
	tests := []struct {
		name string
		args args
		f    func()
		err  error
	}{
		{
			name: "start",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 6, Type: "bot_command"}}, Text: "/start=", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "start",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 6, Type: "bot_command"}}, Text: "/start=test", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				dsMock.EXPECT().GetVerificationByShortID(gomock.Eq("test")).Return(&dataspike.Verification{}, nil)
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{ApplicantId: "test"}, nil)
				dsMock.EXPECT().LinkTelegramProfile(gomock.Eq("test"), gomock.Any()).Return(nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Any()).Return(nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "set verification error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 6, Type: "bot_command"}}, Text: "/start=test", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				dsMock.EXPECT().GetVerificationByShortID(gomock.Eq("test")).Return(&dataspike.Verification{}, nil)
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{ApplicantId: "test"}, nil)
				dsMock.EXPECT().LinkTelegramProfile(gomock.Eq("test"), gomock.Any()).Return(nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Any()).Return(errors.New("set verification error"))
			},
			err: errors.New("set verification error"),
		},
		{
			name: "link tg error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 6, Type: "bot_command"}}, Text: "/start=test", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				dsMock.EXPECT().GetVerificationByShortID(gomock.Eq("test")).Return(&dataspike.Verification{}, nil)
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(&dataspike.Applicant{ApplicantId: "test"}, nil)
				dsMock.EXPECT().LinkTelegramProfile(gomock.Eq("test"), gomock.Any()).Return(errors.New("link tg error"))
			},
			err: errors.New("link tg error"),
		},
		{
			name: "get applicant error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 6, Type: "bot_command"}}, Text: "/start=test", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				dsMock.EXPECT().GetVerificationByShortID(gomock.Eq("test")).Return(&dataspike.Verification{}, nil)
				dsMock.EXPECT().GetApplicantByID(gomock.Any()).Return(nil, errors.New("get applicant error"))
			},
			err: errors.New("get applicant error"),
		},
		{
			name: "get verification error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 6, Type: "bot_command"}}, Text: "/start=test", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				dsMock.EXPECT().GetVerificationByShortID(gomock.Eq("test")).Return(nil, errors.New("get verification error"))
			},
			err: errors.New("get verification error"),
		},
		{
			name: "start expired",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 6, Type: "bot_command"}}, Text: "/start=test", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				dsMock.EXPECT().GetVerificationByShortID(gomock.Eq("test")).Return(&dataspike.Verification{Status: "expired"}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "help",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 5, Type: "bot_command"}}, Text: "/help", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "ask_expert",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 11, Type: "bot_command"}}, Text: "/ask_expert", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "customize_bot",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 14, Type: "bot_command"}}, Text: "/customize_bot", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "cancel",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 7, Type: "bot_command"}}, Text: "/cancel", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{}, nil)
				dsMock.EXPECT().CancelVerification(gomock.Any()).Return(nil)
				cacheMock.EXPECT().RemoveVerification(gomock.Any(), gomock.Eq("123")).Return(nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "remove verification error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 7, Type: "bot_command"}}, Text: "/cancel", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{}, nil)
				dsMock.EXPECT().CancelVerification(gomock.Any()).Return(nil)
				cacheMock.EXPECT().RemoveVerification(gomock.Any(), gomock.Eq("123")).Return(errors.New("remove verification error"))
			},
			err: errors.New("remove verification error"),
		},
		{
			name: "cancel verification error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 7, Type: "bot_command"}}, Text: "/cancel", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{}, nil)
				dsMock.EXPECT().CancelVerification(gomock.Any()).Return(errors.New("cancel verification error"))
			},
			err: errors.New("cancel verification error"),
		},
		{
			name: "get verification error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 7, Type: "bot_command"}}, Text: "/cancel", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(nil, errors.New("get verification error"))
			},
			err: errors.New("get verification error"),
		},
		{
			name: "start_verification",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 19, Type: "bot_command"}}, Text: "/start_verification", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				dsMock.EXPECT().ProceedVerification(gomock.Any()).Return(nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "verification verified",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 19, Type: "bot_command"}}, Text: "/start_verification", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Status: verified}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "send init message error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 19, Type: "bot_command"}}, Text: "/start_verification", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("send message error"))
			},
			err: errors.New("send message error"),
		},
		{
			name: "get verification error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 19, Type: "bot_command"}}, Text: "/start_verification", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(nil, errors.New("get verification error"))
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: errors.New("get verification error"),
		},
		{
			name: "send message error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 19, Type: "bot_command"}}, Text: "/start_verification", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(nil, errors.New("get verification error"))
				httpMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("send message error"))
			},
			err: errors.New("send message error"),
		},
		{
			name: "bot error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 19, Type: "bot_command"}}, Text: "/start_verification", From: &tgbotapi.User{ID: 123, IsBot: true}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("bot error"))
			},
			err: errors.New("bot error"),
		},
		{
			name: "create_verification",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 20, Type: "bot_command"}}, Text: "/create_verification", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				tBot.dev = true
				dsMock.EXPECT().GetApplicantByExternalID(gomock.Any()).Return(&dataspike.Applicant{ApplicantId: "test"}, nil)
				dsMock.EXPECT().CreateVerification(gomock.Any()).Return(&dataspike.Verification{}, nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Eq(&dataspike.Verification{})).Return(nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				dsMock.EXPECT().ProceedVerification(gomock.Any()).Return(nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "send message error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 20, Type: "bot_command"}}, Text: "/create_verification", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				tBot.dev = true
				dsMock.EXPECT().GetApplicantByExternalID(gomock.Any()).Return(&dataspike.Applicant{ApplicantId: "test"}, nil)
				dsMock.EXPECT().CreateVerification(gomock.Any()).Return(&dataspike.Verification{}, nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Eq(&dataspike.Verification{})).Return(nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("send message error"))
			},
			err: errors.New("send message error"),
		},
		{
			name: "set verification error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 20, Type: "bot_command"}}, Text: "/create_verification", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				tBot.dev = true
				dsMock.EXPECT().GetApplicantByExternalID(gomock.Any()).Return(&dataspike.Applicant{ApplicantId: "test"}, nil)
				dsMock.EXPECT().CreateVerification(gomock.Any()).Return(&dataspike.Verification{}, nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Eq(&dataspike.Verification{})).Return(errors.New("set verification error"))
			},
			err: errors.New("set verification error"),
		},
		{
			name: "create verification error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 20, Type: "bot_command"}}, Text: "/create_verification", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				tBot.dev = true
				dsMock.EXPECT().GetApplicantByExternalID(gomock.Any()).Return(&dataspike.Applicant{ApplicantId: "test"}, nil)
				dsMock.EXPECT().CreateVerification(gomock.Any()).Return(nil, errors.New("create verification error"))
			},
			err: errors.New("create verification error"),
		},
		{
			name: "link tg error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 20, Type: "bot_command"}}, Text: "/create_verification", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				tBot.dev = true
				dsMock.EXPECT().GetApplicantByExternalID(gomock.Any()).Return(nil, errors.New("dataspike error: 404"))
				dsMock.EXPECT().CreateApplicant(gomock.Any()).Return("test", nil)
				dsMock.EXPECT().LinkTelegramProfile(gomock.Eq("test"), gomock.Eq("123")).Return(errors.New("link tg error"))
			},
			err: errors.New("link tg error"),
		},
		{
			name: "get applicant error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 20, Type: "bot_command"}}, Text: "/create_verification", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				tBot.dev = true
				dsMock.EXPECT().GetApplicantByExternalID(gomock.Any()).Return(nil, errors.New("get applicant error"))
			},
			err: errors.New("get applicant error"),
		},
		{
			name: "create applicant error",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 20, Type: "bot_command"}}, Text: "/create_verification", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				tBot.dev = true
				dsMock.EXPECT().GetApplicantByExternalID(gomock.Any()).Return(nil, errors.New("dataspike error: 404"))
				dsMock.EXPECT().CreateApplicant(gomock.Any()).Return("", errors.New("create applicant error"))
			},
			err: errors.New("create applicant error"),
		},
		{
			name: "not dev",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 20, Type: "bot_command"}}, Text: "/create_verification", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				tBot.dev = false
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "unknown command",
			args: args{&tgbotapi.Message{Entities: []tgbotapi.MessageEntity{{Length: 5, Type: "bot_command"}}, Text: "/test", From: &tgbotapi.User{ID: 123}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.f()
			err = tBot.ParseCommand(ctx, tt.args.message)
			assert.Equal(t, tt.err, err)
		})
	}
}

func Test_telegramBot_getLink(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	httpMock := mock_telegram_bot.NewMockIHTTPClient(ctrl)

	bot, err := newBot(httpMock)
	if err != nil {
		t.Errorf("error creating bot: %s", err)
	}

	tBot := &TelegramBot{
		bot: bot,
	}
	type args struct {
		message *tgbotapi.Message
	}
	tests := []struct {
		name     string
		args     args
		f        func()
		filename string
		link     string
		err      error
	}{
		{
			name: "document",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{"file_path": "test"}}`)))}, nil)
			},
			filename: "test",
			link:     "https://api.telegram.org/file/bot/test",
			err:      nil,
		},
		{
			name: "get document error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("get document error"))
			},
			filename: "",
			link:     "",
			err:      errors.New("get document error"),
		},
		{
			name: "photo",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Photo: []tgbotapi.PhotoSize{{FileID: "123", FileSize: 321}}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{"file_path": "test"}}`)))}, nil)
			},
			filename: "photo.jpg",
			link:     "https://api.telegram.org/file/bot/test",
			err:      nil,
		},
		{
			name: "get photo error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Photo: []tgbotapi.PhotoSize{{FileID: "123"}}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("get photo error"))
			},
			filename: "",
			link:     "",
			err:      errors.New("get photo error"),
		},
		{
			name:     "not found error",
			args:     args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}}},
			f:        func() {},
			filename: "",
			link:     "",
			err:      errors.New("file not found"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.f()
			filename, link, err := tBot.getLink(tt.args.message)
			assert.Equal(t, tt.filename, filename)
			assert.Equal(t, tt.link, link)
			assert.Equal(t, tt.err, err)
		})
	}
}

func Test_telegramBot_nextCheck(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	httpMock := mock_telegram_bot.NewMockIHTTPClient(ctrl)
	dsMock := mock_telegram_bot.NewMockIDataspikeClient(ctrl)
	cacheMock := mock_telegram_bot.NewMockICache(ctrl)

	bot, err := newBot(httpMock)
	if err != nil {
		t.Errorf("error creating bot: %s", err)
	}

	tBot := &TelegramBot{
		bot:      bot,
		dsClient: dsMock,
		cache:    cacheMock,
	}
	type args struct {
		chatID       int64
		verification *dataspike.Verification
	}
	tests := []struct {
		name string
		args args
		f    func()
		err  error
	}{
		{
			name: "DocumentMrz",
			args: args{123, &dataspike.Verification{Checks: dataspike.Checks{DocumentMrz: &dataspike.DocumentMrz{Check: dataspike.Check{Status: "pending"}}}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "DocumentMrz send message error",
			args: args{123, &dataspike.Verification{Checks: dataspike.Checks{DocumentMrz: &dataspike.DocumentMrz{Check: dataspike.Check{Status: "pending"}}}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("send message error"))
			},
			err: errors.New("send message error"),
		},
		{
			name: "FaceComparison",
			args: args{123, &dataspike.Verification{Checks: dataspike.Checks{FaceComparison: &dataspike.Check{Status: "pending"}}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "Liveness",
			args: args{123, &dataspike.Verification{Checks: dataspike.Checks{Liveness: &dataspike.Check{Status: "pending"}}}},
			f: func() {
				tBot.bot.Self = tgbotapi.User{UserName: "test"}
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "Poa",
			args: args{123, &dataspike.Verification{Checks: dataspike.Checks{Poa: &dataspike.Check{Status: "pending"}}, Settings: &dataspike.Settings{PoaRequired: true}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "Poa skip send message error",
			args: args{123, &dataspike.Verification{Checks: dataspike.Checks{Poa: &dataspike.Check{Status: "pending"}}, Settings: &dataspike.Settings{PoaRequired: false}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("send message error"))
			},
			err: errors.New("send message error"),
		},
		{
			name: "Proceed verification error",
			args: args{123, &dataspike.Verification{}},
			f: func() {
				dsMock.EXPECT().ProceedVerification(gomock.Any()).Return(errors.New("proceed verification error"))
			},
			err: errors.New("proceed verification error"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.f()
			err = tBot.nextCheck(tt.args.chatID, tt.args.verification)
			assert.Equal(t, tt.err, err)
		})
	}
}

func Test_telegramBot_parseDocument(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	httpMock := mock_telegram_bot.NewMockIHTTPClient(ctrl)
	dsMock := mock_telegram_bot.NewMockIDataspikeClient(ctrl)
	cacheMock := mock_telegram_bot.NewMockICache(ctrl)
	ctx := context.Background()

	bot, err := newBot(httpMock)
	if err != nil {
		t.Errorf("error creating bot: %s", err)
	}

	tBot := &TelegramBot{
		bot:        bot,
		dsClient:   dsMock,
		cache:      cacheMock,
		httpClient: httpMock,
	}
	type args struct {
		message *tgbotapi.Message
	}
	tests := []struct {
		name string
		args args
		f    func()
		err  error
	}{
		{
			name: "DocumentMrz",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{DocumentMrz: &dataspike.DocumentMrz{}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))}, nil)
				dsMock.EXPECT().UploadDocument(gomock.Any()).Return(&dataspike.Document{}, nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Any()).Return(nil)
				dsMock.EXPECT().ProceedVerification(gomock.Any()).Return(nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "DocumentMrz front",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{DocumentMrz: &dataspike.DocumentMrz{}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))}, nil)
				twoSide := true
				side := Front
				dsMock.EXPECT().UploadDocument(gomock.Any()).Return(&dataspike.Document{DetectedTwoSideDocument: &twoSide, DetectedDocumentSide: &side}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "set verification error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{DocumentMrz: &dataspike.DocumentMrz{}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))}, nil)
				dsMock.EXPECT().UploadDocument(gomock.Any()).Return(&dataspike.Document{}, nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Any()).Return(errors.New("set verification error"))
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: errors.New("set verification error"),
		},
		{
			name: "Document error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{DocumentMrz: &dataspike.DocumentMrz{}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))}, nil)
				dsMock.EXPECT().UploadDocument(gomock.Any()).Return(&dataspike.Document{Errors: dataspike.Errors{{0, "document error"}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: errors.New("Code: 0; Message: document error"),
		},
		{
			name: "upload document error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{DocumentMrz: &dataspike.DocumentMrz{}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))}, nil)
				dsMock.EXPECT().UploadDocument(gomock.Any()).Return(nil, errors.New("upload document error"))
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: errors.New("upload document error"),
		},
		{
			name: "FaceComparison",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{FaceComparison: &dataspike.Check{}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))}, nil)
				dsMock.EXPECT().UploadDocument(gomock.Any()).Return(&dataspike.Document{}, nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Any()).Return(nil)
				dsMock.EXPECT().ProceedVerification(gomock.Any()).Return(nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "set verification error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{FaceComparison: &dataspike.Check{}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))}, nil)
				dsMock.EXPECT().UploadDocument(gomock.Any()).Return(&dataspike.Document{}, nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Any()).Return(errors.New("set verification error"))
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: errors.New("set verification error"),
		},
		{
			name: "Document error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{FaceComparison: &dataspike.Check{}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))}, nil)
				dsMock.EXPECT().UploadDocument(gomock.Any()).Return(&dataspike.Document{Errors: dataspike.Errors{{0, "document error"}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: errors.New("Code: 0; Message: document error"),
		},
		{
			name: "upload document error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{FaceComparison: &dataspike.Check{}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))}, nil)
				dsMock.EXPECT().UploadDocument(gomock.Any()).Return(nil, errors.New("upload document error"))
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: errors.New("upload document error"),
		},
		{
			name: "Poa",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{Poa: &dataspike.Check{}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))}, nil)
				dsMock.EXPECT().UploadDocument(gomock.Any()).Return(&dataspike.Document{}, nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Any()).Return(nil)
				dsMock.EXPECT().ProceedVerification(gomock.Any()).Return(nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "set verification error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{Poa: &dataspike.Check{}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))}, nil)
				dsMock.EXPECT().UploadDocument(gomock.Any()).Return(&dataspike.Document{}, nil)
				cacheMock.EXPECT().SetVerification(gomock.Any(), gomock.Eq("123"), gomock.Any()).Return(errors.New("set verification error"))
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: errors.New("set verification error"),
		},
		{
			name: "Document error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{Poa: &dataspike.Check{}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))}, nil)
				dsMock.EXPECT().UploadDocument(gomock.Any()).Return(&dataspike.Document{Errors: dataspike.Errors{{0, "document error"}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: errors.New("Code: 0; Message: document error"),
		},
		{
			name: "upload document error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{Poa: &dataspike.Check{}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))}, nil)
				dsMock.EXPECT().UploadDocument(gomock.Any()).Return(nil, errors.New("upload document error"))
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: errors.New("upload document error"),
		},
		{
			name: "empty checks error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: errors.New("status not supported for upload document"),
		},
		{
			name: "do request error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("do request error"))
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: errors.New("do request error"),
		},
		{
			name: "url error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{Poa: &dataspike.Check{}}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{"file_path":")(<>!@#  $%^&*"}}`)))}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: &url.Error{Op: "parse", URL: "https://api.telegram.org/file/bot/)(<>!@#  $%^&*", Err: url.EscapeError("%^&")},
		},
		{
			name: "get file error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(&dataspike.Verification{Checks: dataspike.Checks{}}, nil)
				httpMock.EXPECT().Do(gomock.Any()).Return(nil, errors.New("get file error"))
			},
			err: errors.New("get file error"),
		},
		{
			name: "get verification error",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}, Document: &tgbotapi.Document{FileID: "123", FileName: "test"}}},
			f: func() {
				cacheMock.EXPECT().GetVerification(gomock.Any(), gomock.Eq("123")).Return(nil, errors.New("get verification error"))
			},
			err: errors.New("get verification error"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.f()
			err = tBot.ParseDocument(ctx, tt.args.message)
			assert.Equal(t, tt.err, err)
		})
	}
}

func Test_telegramBot_textParse(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	httpMock := mock_telegram_bot.NewMockIHTTPClient(ctrl)
	dsMock := mock_telegram_bot.NewMockIDataspikeClient(ctrl)
	cacheMock := mock_telegram_bot.NewMockICache(ctrl)
	ctx := context.Background()

	bot, err := newBot(httpMock)
	if err != nil {
		t.Errorf("error creating bot: %s", err)
	}

	tBot := &TelegramBot{
		bot:      bot,
		dsClient: dsMock,
		cache:    cacheMock,
	}
	type args struct {
		message *tgbotapi.Message
	}
	tests := []struct {
		name string
		args args
		f    func()
		err  error
	}{
		{
			name: "non gpt",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}}},
			f: func() {
				httpMock.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"ok":true,"result":{}}`)))}, nil)
			},
			err: nil,
		},
		{
			name: "gpt",
			args: args{&tgbotapi.Message{From: &tgbotapi.User{ID: 123}}},
			f: func() {
				tBot.gptClient, _ = chatgpt.NewClient("test")
			},
			err: errors.New("api request failed: status Code: 401 401 Unauthorized https://api.openai.com/v1/chat/completions Message: map[error:map[code:invalid_api_key message:Incorrect API key provided: test. You can find your API key at https://platform.openai.com/account/api-keys. param:<nil> type:invalid_request_error]]"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.f()
			err = tBot.ParseText(ctx, tt.args.message)
			assert.Equal(t, tt.err, err)
		})
	}
}
