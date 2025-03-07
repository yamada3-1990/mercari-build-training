// Code generated by MockGen. DO NOT EDIT.
// Source: C:\Users\maple\Desktop\mercari-build-training\go\app\infra.go
//
// Generated by this command:
//
//	mockgen -source=C:\Users\maple\Desktop\mercari-build-training\go\app\infra.go -package=app -destination=C:\Users\maple\Desktop\mercari-build-training\go\app\mock_infra.go
//

// Package app is a generated GoMock package.
package app

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockItemRepository is a mock of ItemRepository interface.
type MockItemRepository struct {
	ctrl     *gomock.Controller
	recorder *MockItemRepositoryMockRecorder
	isgomock struct{}
}

// MockItemRepositoryMockRecorder is the mock recorder for MockItemRepository.
type MockItemRepositoryMockRecorder struct {
	mock *MockItemRepository
}

// NewMockItemRepository creates a new mock instance.
func NewMockItemRepository(ctrl *gomock.Controller) *MockItemRepository {
	mock := &MockItemRepository{ctrl: ctrl}
	mock.recorder = &MockItemRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockItemRepository) EXPECT() *MockItemRepositoryMockRecorder {
	return m.recorder
}

// GetAll mocks base method.
func (m *MockItemRepository) GetAll(ctx context.Context) ([]Item, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll", ctx)
	ret0, _ := ret[0].([]Item)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockItemRepositoryMockRecorder) GetAll(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockItemRepository)(nil).GetAll), ctx)
}

// GetItemById mocks base method.
func (m *MockItemRepository) GetItemById(ctx context.Context, item_id string) (Item, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetItemById", ctx, item_id)
	ret0, _ := ret[0].(Item)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetItemById indicates an expected call of GetItemById.
func (mr *MockItemRepositoryMockRecorder) GetItemById(ctx, item_id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetItemById", reflect.TypeOf((*MockItemRepository)(nil).GetItemById), ctx, item_id)
}

// Insert mocks base method.
func (m *MockItemRepository) Insert(ctx context.Context, item *Item) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Insert", ctx, item)
	ret0, _ := ret[0].(error)
	return ret0
}

// Insert indicates an expected call of Insert.
func (mr *MockItemRepositoryMockRecorder) Insert(ctx, item any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*MockItemRepository)(nil).Insert), ctx, item)
}

// SearchItemsByKeyword mocks base method.
func (m *MockItemRepository) SearchItemsByKeyword(ctx context.Context, keyword string) ([]Item, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchItemsByKeyword", ctx, keyword)
	ret0, _ := ret[0].([]Item)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchItemsByKeyword indicates an expected call of SearchItemsByKeyword.
func (mr *MockItemRepositoryMockRecorder) SearchItemsByKeyword(ctx, keyword any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchItemsByKeyword", reflect.TypeOf((*MockItemRepository)(nil).SearchItemsByKeyword), ctx, keyword)
}
