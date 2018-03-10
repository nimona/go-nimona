package net

import mock "github.com/stretchr/testify/mock"

// MockProtocol is an autogenerated mock type for the Protocol type
type MockProtocol struct {
	mock.Mock
}

// Handle provides a mock function with given fields: _a0
func (_m *MockProtocol) Handle(_a0 HandlerFunc) HandlerFunc {
	ret := _m.Called(_a0)

	var r0 HandlerFunc
	if rf, ok := ret.Get(0).(func(HandlerFunc) HandlerFunc); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(HandlerFunc)
		}
	}

	return r0
}

// Name provides a mock function with given fields:
func (_m *MockProtocol) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Negotiate provides a mock function with given fields: _a0
func (_m *MockProtocol) Negotiate(_a0 NegotiatorFunc) NegotiatorFunc {
	ret := _m.Called(_a0)

	var r0 NegotiatorFunc
	if rf, ok := ret.Get(0).(func(NegotiatorFunc) NegotiatorFunc); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(NegotiatorFunc)
		}
	}

	return r0
}
