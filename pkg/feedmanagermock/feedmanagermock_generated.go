// Code generated by MockGen. DO NOT EDIT.
// Source: feedmanager.go

// Package feedmanagermock is a generated GoMock package.
package feedmanagermock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	context "nimona.io/pkg/context"
	object "nimona.io/pkg/object"
	peer "nimona.io/pkg/peer"
)

// MockFeedManager is a mock of FeedManager interface.
type MockFeedManager struct {
	ctrl     *gomock.Controller
	recorder *MockFeedManagerMockRecorder
}

// MockFeedManagerMockRecorder is the mock recorder for MockFeedManager.
type MockFeedManagerMockRecorder struct {
	mock *MockFeedManager
}

// NewMockFeedManager creates a new mock instance.
func NewMockFeedManager(ctrl *gomock.Controller) *MockFeedManager {
	mock := &MockFeedManager{ctrl: ctrl}
	mock.recorder = &MockFeedManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFeedManager) EXPECT() *MockFeedManagerMockRecorder {
	return m.recorder
}

// RequestFeed mocks base method.
func (m *MockFeedManager) RequestFeed(ctx context.Context, rootCID object.CID, recipients ...*peer.ConnectionInfo) (object.ReadCloser, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, rootCID}
	for _, a := range recipients {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RequestFeed", varargs...)
	ret0, _ := ret[0].(object.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RequestFeed indicates an expected call of RequestFeed.
func (mr *MockFeedManagerMockRecorder) RequestFeed(ctx, rootCID interface{}, recipients ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, rootCID}, recipients...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestFeed", reflect.TypeOf((*MockFeedManager)(nil).RequestFeed), varargs...)
}
