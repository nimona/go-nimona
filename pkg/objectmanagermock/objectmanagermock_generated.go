// Code generated by MockGen. DO NOT EDIT.
// Source: objectmanager.go

// Package objectmanagermock is a generated GoMock package.
package objectmanagermock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	context "nimona.io/pkg/context"
	did "nimona.io/pkg/did"
	object "nimona.io/pkg/object"
	objectmanager "nimona.io/pkg/objectmanager"
	tilde "nimona.io/pkg/tilde"
)

// MockObjectManager is a mock of ObjectManager interface.
type MockObjectManager struct {
	ctrl     *gomock.Controller
	recorder *MockObjectManagerMockRecorder
}

// MockObjectManagerMockRecorder is the mock recorder for MockObjectManager.
type MockObjectManagerMockRecorder struct {
	mock *MockObjectManager
}

// NewMockObjectManager creates a new mock instance.
func NewMockObjectManager(ctrl *gomock.Controller) *MockObjectManager {
	mock := &MockObjectManager{ctrl: ctrl}
	mock.recorder = &MockObjectManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockObjectManager) EXPECT() *MockObjectManagerMockRecorder {
	return m.recorder
}

// AddStreamSubscription mocks base method.
func (m *MockObjectManager) AddStreamSubscription(ctx context.Context, rootHash tilde.Digest, subscriber did.DID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddStreamSubscription", ctx, rootHash, subscriber)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddStreamSubscription indicates an expected call of AddStreamSubscription.
func (mr *MockObjectManagerMockRecorder) AddStreamSubscription(ctx, rootHash, subscriber interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddStreamSubscription", reflect.TypeOf((*MockObjectManager)(nil).AddStreamSubscription), ctx, rootHash, subscriber)
}

// Append mocks base method.
func (m *MockObjectManager) Append(ctx context.Context, o *object.Object) (*object.Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Append", ctx, o)
	ret0, _ := ret[0].(*object.Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Append indicates an expected call of Append.
func (mr *MockObjectManagerMockRecorder) Append(ctx, o interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Append", reflect.TypeOf((*MockObjectManager)(nil).Append), ctx, o)
}

// Put mocks base method.
func (m *MockObjectManager) Put(ctx context.Context, o *object.Object) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Put", ctx, o)
	ret0, _ := ret[0].(error)
	return ret0
}

// Put indicates an expected call of Put.
func (mr *MockObjectManagerMockRecorder) Put(ctx, o interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockObjectManager)(nil).Put), ctx, o)
}

// Request mocks base method.
func (m *MockObjectManager) Request(ctx context.Context, hash tilde.Digest, id did.DID) (*object.Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Request", ctx, hash, id)
	ret0, _ := ret[0].(*object.Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Request indicates an expected call of Request.
func (mr *MockObjectManagerMockRecorder) Request(ctx, hash, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Request", reflect.TypeOf((*MockObjectManager)(nil).Request), ctx, hash, id)
}

// RequestStream mocks base method.
func (m *MockObjectManager) RequestStream(ctx context.Context, rootHash tilde.Digest, id did.DID) (object.ReadCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RequestStream", ctx, rootHash, id)
	ret0, _ := ret[0].(object.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RequestStream indicates an expected call of RequestStream.
func (mr *MockObjectManagerMockRecorder) RequestStream(ctx, rootHash, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestStream", reflect.TypeOf((*MockObjectManager)(nil).RequestStream), ctx, rootHash, id)
}

// Subscribe mocks base method.
func (m *MockObjectManager) Subscribe(lookupOptions ...objectmanager.LookupOption) objectmanager.ObjectSubscription {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range lookupOptions {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Subscribe", varargs...)
	ret0, _ := ret[0].(objectmanager.ObjectSubscription)
	return ret0
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockObjectManagerMockRecorder) Subscribe(lookupOptions ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockObjectManager)(nil).Subscribe), lookupOptions...)
}
