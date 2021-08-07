// Code generated by MockGen. DO NOT EDIT.
// Source: objectstore.go

// Package objectstoremock is a generated GoMock package.
package objectstoremock

import (
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	object "nimona.io/pkg/object"
	tilde "nimona.io/pkg/tilde"
)

// MockGetter is a mock of Getter interface.
type MockGetter struct {
	ctrl     *gomock.Controller
	recorder *MockGetterMockRecorder
}

// MockGetterMockRecorder is the mock recorder for MockGetter.
type MockGetterMockRecorder struct {
	mock *MockGetter
}

// NewMockGetter creates a new mock instance.
func NewMockGetter(ctrl *gomock.Controller) *MockGetter {
	mock := &MockGetter{ctrl: ctrl}
	mock.recorder = &MockGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGetter) EXPECT() *MockGetterMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockGetter) Get(hash tilde.Digest) (*object.Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", hash)
	ret0, _ := ret[0].(*object.Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockGetterMockRecorder) Get(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockGetter)(nil).Get), hash)
}

// MockStore is a mock of Store interface.
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
}

// MockStoreMockRecorder is the mock recorder for MockStore.
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance.
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockStore) Get(hash tilde.Digest) (*object.Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", hash)
	ret0, _ := ret[0].(*object.Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockStoreMockRecorder) Get(hash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockStore)(nil).Get), hash)
}

// GetByStream mocks base method.
func (m *MockStore) GetByStream(arg0 tilde.Digest) (object.ReadCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByStream", arg0)
	ret0, _ := ret[0].(object.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByStream indicates an expected call of GetByStream.
func (mr *MockStoreMockRecorder) GetByStream(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByStream", reflect.TypeOf((*MockStore)(nil).GetByStream), arg0)
}

// GetByType mocks base method.
func (m *MockStore) GetByType(arg0 string) (object.ReadCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByType", arg0)
	ret0, _ := ret[0].(object.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByType indicates an expected call of GetByType.
func (mr *MockStoreMockRecorder) GetByType(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByType", reflect.TypeOf((*MockStore)(nil).GetByType), arg0)
}

// GetPinned mocks base method.
func (m *MockStore) GetPinned() ([]tilde.Digest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPinned")
	ret0, _ := ret[0].([]tilde.Digest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPinned indicates an expected call of GetPinned.
func (mr *MockStoreMockRecorder) GetPinned() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPinned", reflect.TypeOf((*MockStore)(nil).GetPinned))
}

// GetStreamLeaves mocks base method.
func (m *MockStore) GetStreamLeaves(streamRootHash tilde.Digest) ([]tilde.Digest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStreamLeaves", streamRootHash)
	ret0, _ := ret[0].([]tilde.Digest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStreamLeaves indicates an expected call of GetStreamLeaves.
func (mr *MockStoreMockRecorder) GetStreamLeaves(streamRootHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStreamLeaves", reflect.TypeOf((*MockStore)(nil).GetStreamLeaves), streamRootHash)
}

// IsPinned mocks base method.
func (m *MockStore) IsPinned(arg0 tilde.Digest) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsPinned", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsPinned indicates an expected call of IsPinned.
func (mr *MockStoreMockRecorder) IsPinned(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsPinned", reflect.TypeOf((*MockStore)(nil).IsPinned), arg0)
}

// Pin mocks base method.
func (m *MockStore) Pin(arg0 tilde.Digest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Pin", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Pin indicates an expected call of Pin.
func (mr *MockStoreMockRecorder) Pin(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pin", reflect.TypeOf((*MockStore)(nil).Pin), arg0)
}

// Put mocks base method.
func (m *MockStore) Put(arg0 *object.Object) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Put", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Put indicates an expected call of Put.
func (mr *MockStoreMockRecorder) Put(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockStore)(nil).Put), arg0)
}

// PutWithTTL mocks base method.
func (m *MockStore) PutWithTTL(arg0 *object.Object, arg1 time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutWithTTL", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutWithTTL indicates an expected call of PutWithTTL.
func (mr *MockStoreMockRecorder) PutWithTTL(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutWithTTL", reflect.TypeOf((*MockStore)(nil).PutWithTTL), arg0, arg1)
}

// RemovePin mocks base method.
func (m *MockStore) RemovePin(arg0 tilde.Digest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemovePin", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemovePin indicates an expected call of RemovePin.
func (mr *MockStoreMockRecorder) RemovePin(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemovePin", reflect.TypeOf((*MockStore)(nil).RemovePin), arg0)
}
