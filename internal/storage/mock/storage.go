// Code generated by MockGen. DO NOT EDIT.
// Source: internal/storage/repository.go

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	storage "github.com/PrahaTurbo/url-shortener/internal/storage"
	gomock "go.uber.org/mock/gomock"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// GetURL mocks base method.
func (m *MockRepository) GetURL(id string) (*storage.URLRecord, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetURL", id)
	ret0, _ := ret[0].(*storage.URLRecord)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetURL indicates an expected call of GetURL.
func (mr *MockRepositoryMockRecorder) GetURL(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetURL", reflect.TypeOf((*MockRepository)(nil).GetURL), id)
}

// Ping mocks base method.
func (m *MockRepository) Ping() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockRepositoryMockRecorder) Ping() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockRepository)(nil).Ping))
}

// PutBatchURLs mocks base method.
func (m *MockRepository) PutBatchURLs(urls []storage.URLRecord) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutBatchURLs", urls)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutBatchURLs indicates an expected call of PutBatchURLs.
func (mr *MockRepositoryMockRecorder) PutBatchURLs(urls interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutBatchURLs", reflect.TypeOf((*MockRepository)(nil).PutBatchURLs), urls)
}

// PutURL mocks base method.
func (m *MockRepository) PutURL(url storage.URLRecord) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutURL", url)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutURL indicates an expected call of PutURL.
func (mr *MockRepositoryMockRecorder) PutURL(url interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutURL", reflect.TypeOf((*MockRepository)(nil).PutURL), url)
}
