// Code generated by MockGen. DO NOT EDIT.
// Source: /home/lenovo/dataspike/docver-tg-bot/internal/telegram_bot/telegram_bot.go

// Package mock_telegram_bot is a generated GoMock package.
package mock_telegram_bot

import (
	context "context"
	http "net/http"
	reflect "reflect"

	dataspike "github.com/dataspike-io/docver-sdk-go"
	uuid "github.com/gofrs/uuid"
	gomock "github.com/golang/mock/gomock"
)

// MockIHTTPClient is a mock of IHTTPClient interface.
type MockIHTTPClient struct {
	ctrl     *gomock.Controller
	recorder *MockIHTTPClientMockRecorder
}

// MockIHTTPClientMockRecorder is the mock recorder for MockIHTTPClient.
type MockIHTTPClientMockRecorder struct {
	mock *MockIHTTPClient
}

// NewMockIHTTPClient creates a new mock instance.
func NewMockIHTTPClient(ctrl *gomock.Controller) *MockIHTTPClient {
	mock := &MockIHTTPClient{ctrl: ctrl}
	mock.recorder = &MockIHTTPClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIHTTPClient) EXPECT() *MockIHTTPClientMockRecorder {
	return m.recorder
}

// Do mocks base method.
func (m *MockIHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Do", req)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Do indicates an expected call of Do.
func (mr *MockIHTTPClientMockRecorder) Do(req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Do", reflect.TypeOf((*MockIHTTPClient)(nil).Do), req)
}

// MockICache is a mock of ICache interface.
type MockICache struct {
	ctrl     *gomock.Controller
	recorder *MockICacheMockRecorder
}

// MockICacheMockRecorder is the mock recorder for MockICache.
type MockICacheMockRecorder struct {
	mock *MockICache
}

// NewMockICache creates a new mock instance.
func NewMockICache(ctrl *gomock.Controller) *MockICache {
	mock := &MockICache{ctrl: ctrl}
	mock.recorder = &MockICacheMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockICache) EXPECT() *MockICacheMockRecorder {
	return m.recorder
}

// GetVerification mocks base method.
func (m *MockICache) GetVerification(arg0 context.Context, arg1 string) (*dataspike.Verification, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVerification", arg0, arg1)
	ret0, _ := ret[0].(*dataspike.Verification)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVerification indicates an expected call of GetVerification.
func (mr *MockICacheMockRecorder) GetVerification(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVerification", reflect.TypeOf((*MockICache)(nil).GetVerification), arg0, arg1)
}

// RemoveVerification mocks base method.
func (m *MockICache) RemoveVerification(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveVerification", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveVerification indicates an expected call of RemoveVerification.
func (mr *MockICacheMockRecorder) RemoveVerification(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveVerification", reflect.TypeOf((*MockICache)(nil).RemoveVerification), arg0, arg1)
}

// SetVerification mocks base method.
func (m *MockICache) SetVerification(arg0 context.Context, arg1 string, arg2 *dataspike.Verification) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetVerification", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetVerification indicates an expected call of SetVerification.
func (mr *MockICacheMockRecorder) SetVerification(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetVerification", reflect.TypeOf((*MockICache)(nil).SetVerification), arg0, arg1, arg2)
}

// MockIDataspikeClient is a mock of IDataspikeClient interface.
type MockIDataspikeClient struct {
	ctrl     *gomock.Controller
	recorder *MockIDataspikeClientMockRecorder
}

// MockIDataspikeClientMockRecorder is the mock recorder for MockIDataspikeClient.
type MockIDataspikeClientMockRecorder struct {
	mock *MockIDataspikeClient
}

// NewMockIDataspikeClient creates a new mock instance.
func NewMockIDataspikeClient(ctrl *gomock.Controller) *MockIDataspikeClient {
	mock := &MockIDataspikeClient{ctrl: ctrl}
	mock.recorder = &MockIDataspikeClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIDataspikeClient) EXPECT() *MockIDataspikeClientMockRecorder {
	return m.recorder
}

// CancelVerification mocks base method.
func (m *MockIDataspikeClient) CancelVerification(arg0 uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CancelVerification", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CancelVerification indicates an expected call of CancelVerification.
func (mr *MockIDataspikeClientMockRecorder) CancelVerification(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CancelVerification", reflect.TypeOf((*MockIDataspikeClient)(nil).CancelVerification), arg0)
}

// CreateApplicant mocks base method.
func (m *MockIDataspikeClient) CreateApplicant(arg0 *dataspike.ApplicantCreate) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateApplicant", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateApplicant indicates an expected call of CreateApplicant.
func (mr *MockIDataspikeClientMockRecorder) CreateApplicant(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateApplicant", reflect.TypeOf((*MockIDataspikeClient)(nil).CreateApplicant), arg0)
}

// CreateVerification mocks base method.
func (m *MockIDataspikeClient) CreateVerification(arg0 *dataspike.VerificationCreate) (*dataspike.Verification, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateVerification", arg0)
	ret0, _ := ret[0].(*dataspike.Verification)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateVerification indicates an expected call of CreateVerification.
func (mr *MockIDataspikeClientMockRecorder) CreateVerification(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateVerification", reflect.TypeOf((*MockIDataspikeClient)(nil).CreateVerification), arg0)
}

// CreateWebhook mocks base method.
func (m *MockIDataspikeClient) CreateWebhook(arg0 *dataspike.WebhookCreate) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateWebhook", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateWebhook indicates an expected call of CreateWebhook.
func (mr *MockIDataspikeClientMockRecorder) CreateWebhook(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateWebhook", reflect.TypeOf((*MockIDataspikeClient)(nil).CreateWebhook), arg0)
}

// DeleteWebhook mocks base method.
func (m *MockIDataspikeClient) DeleteWebhook(arg0 uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteWebhook", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteWebhook indicates an expected call of DeleteWebhook.
func (mr *MockIDataspikeClientMockRecorder) DeleteWebhook(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteWebhook", reflect.TypeOf((*MockIDataspikeClient)(nil).DeleteWebhook), arg0)
}

// GetApplicantByExternalID mocks base method.
func (m *MockIDataspikeClient) GetApplicantByExternalID(arg0 string) (*dataspike.Applicant, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicantByExternalID", arg0)
	ret0, _ := ret[0].(*dataspike.Applicant)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicantByExternalID indicates an expected call of GetApplicantByExternalID.
func (mr *MockIDataspikeClientMockRecorder) GetApplicantByExternalID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicantByExternalID", reflect.TypeOf((*MockIDataspikeClient)(nil).GetApplicantByExternalID), arg0)
}

// GetApplicantByID mocks base method.
func (m *MockIDataspikeClient) GetApplicantByID(arg0 uuid.UUID) (*dataspike.Applicant, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicantByID", arg0)
	ret0, _ := ret[0].(*dataspike.Applicant)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicantByID indicates an expected call of GetApplicantByID.
func (mr *MockIDataspikeClientMockRecorder) GetApplicantByID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicantByID", reflect.TypeOf((*MockIDataspikeClient)(nil).GetApplicantByID), arg0)
}

// GetVerificationByID mocks base method.
func (m *MockIDataspikeClient) GetVerificationByID(arg0 uuid.UUID) (*dataspike.Verification, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVerificationByID", arg0)
	ret0, _ := ret[0].(*dataspike.Verification)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVerificationByID indicates an expected call of GetVerificationByID.
func (mr *MockIDataspikeClientMockRecorder) GetVerificationByID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVerificationByID", reflect.TypeOf((*MockIDataspikeClient)(nil).GetVerificationByID), arg0)
}

// GetVerificationByShortID mocks base method.
func (m *MockIDataspikeClient) GetVerificationByShortID(arg0 string) (*dataspike.Verification, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVerificationByShortID", arg0)
	ret0, _ := ret[0].(*dataspike.Verification)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVerificationByShortID indicates an expected call of GetVerificationByShortID.
func (mr *MockIDataspikeClientMockRecorder) GetVerificationByShortID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVerificationByShortID", reflect.TypeOf((*MockIDataspikeClient)(nil).GetVerificationByShortID), arg0)
}

// LinkTelegramProfile mocks base method.
func (m *MockIDataspikeClient) LinkTelegramProfile(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LinkTelegramProfile", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// LinkTelegramProfile indicates an expected call of LinkTelegramProfile.
func (mr *MockIDataspikeClientMockRecorder) LinkTelegramProfile(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LinkTelegramProfile", reflect.TypeOf((*MockIDataspikeClient)(nil).LinkTelegramProfile), arg0, arg1)
}

// ListWebhooks mocks base method.
func (m *MockIDataspikeClient) ListWebhooks() (*dataspike.WebhookResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListWebhooks")
	ret0, _ := ret[0].(*dataspike.WebhookResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListWebhooks indicates an expected call of ListWebhooks.
func (mr *MockIDataspikeClientMockRecorder) ListWebhooks() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListWebhooks", reflect.TypeOf((*MockIDataspikeClient)(nil).ListWebhooks))
}

// ProceedVerification mocks base method.
func (m *MockIDataspikeClient) ProceedVerification(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProceedVerification", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ProceedVerification indicates an expected call of ProceedVerification.
func (mr *MockIDataspikeClientMockRecorder) ProceedVerification(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProceedVerification", reflect.TypeOf((*MockIDataspikeClient)(nil).ProceedVerification), arg0)
}

// UploadDocument mocks base method.
func (m *MockIDataspikeClient) UploadDocument(arg0 *dataspike.DocumentUpload) (*dataspike.Document, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UploadDocument", arg0)
	ret0, _ := ret[0].(*dataspike.Document)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UploadDocument indicates an expected call of UploadDocument.
func (mr *MockIDataspikeClientMockRecorder) UploadDocument(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadDocument", reflect.TypeOf((*MockIDataspikeClient)(nil).UploadDocument), arg0)
}

// MockITelegramBot is a mock of ITelegramBot interface.
type MockITelegramBot struct {
	ctrl     *gomock.Controller
	recorder *MockITelegramBotMockRecorder
}

// MockITelegramBotMockRecorder is the mock recorder for MockITelegramBot.
type MockITelegramBotMockRecorder struct {
	mock *MockITelegramBot
}

// NewMockITelegramBot creates a new mock instance.
func NewMockITelegramBot(ctrl *gomock.Controller) *MockITelegramBot {
	mock := &MockITelegramBot{ctrl: ctrl}
	mock.recorder = &MockITelegramBotMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockITelegramBot) EXPECT() *MockITelegramBotMockRecorder {
	return m.recorder
}

// CheckLiveness mocks base method.
func (m *MockITelegramBot) CheckLiveness(ctx context.Context, applicantId, status string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckLiveness", ctx, applicantId, status)
	ret0, _ := ret[0].(error)
	return ret0
}

// CheckLiveness indicates an expected call of CheckLiveness.
func (mr *MockITelegramBotMockRecorder) CheckLiveness(ctx, applicantId, status interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckLiveness", reflect.TypeOf((*MockITelegramBot)(nil).CheckLiveness), ctx, applicantId, status)
}

// SendVerificationStatus mocks base method.
func (m *MockITelegramBot) SendVerificationStatus(ctx context.Context, applicantID, status string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendVerificationStatus", ctx, applicantID, status)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendVerificationStatus indicates an expected call of SendVerificationStatus.
func (mr *MockITelegramBotMockRecorder) SendVerificationStatus(ctx, applicantID, status interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendVerificationStatus", reflect.TypeOf((*MockITelegramBot)(nil).SendVerificationStatus), ctx, applicantID, status)
}

// Start mocks base method.
func (m *MockITelegramBot) Start(ctx context.Context, offset, timeout int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start", ctx, offset, timeout)
}

// Start indicates an expected call of Start.
func (mr *MockITelegramBotMockRecorder) Start(ctx, offset, timeout interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockITelegramBot)(nil).Start), ctx, offset, timeout)
}
