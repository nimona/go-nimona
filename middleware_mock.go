package fabric

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockProtocol is an autogenerated mock type for the Protocol type
type MockProtocol struct {
	mock.Mock
}

// Handle provides a mock function with given fields: ctx, conn
func (_m *MockProtocol) Handle(ctx context.Context, conn Conn) (context.Context, Conn, error) {
	ret := _m.Called(ctx, conn)

	var r0 context.Context
	if rf, ok := ret.Get(0).(func(context.Context, Conn) context.Context); ok {
		r0 = rf(ctx, conn)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}

	var r1 Conn
	if rf, ok := ret.Get(1).(func(context.Context, Conn) Conn); ok {
		r1 = rf(ctx, conn)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(Conn)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, Conn) error); ok {
		r2 = rf(ctx, conn)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
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

// Negotiate provides a mock function with given fields: ctx, conn
func (_m *MockProtocol) Negotiate(ctx context.Context, conn Conn) (context.Context, Conn, error) {
	ret := _m.Called(ctx, conn)

	var r0 context.Context
	if rf, ok := ret.Get(0).(func(context.Context, Conn) context.Context); ok {
		r0 = rf(ctx, conn)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}

	var r1 Conn
	if rf, ok := ret.Get(1).(func(context.Context, Conn) Conn); ok {
		r1 = rf(ctx, conn)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(Conn)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, Conn) error); ok {
		r2 = rf(ctx, conn)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
