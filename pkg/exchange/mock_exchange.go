// Code generated by mockery v1.0.0. DO NOT EDIT.

package exchange

import context "nimona.io/pkg/context"
import mock "github.com/stretchr/testify/mock"
import object "nimona.io/pkg/object"

// MockExchange is an autogenerated mock type for the Exchange type
type MockExchange struct {
	mock.Mock
}

// Request provides a mock function with given fields: ctx, _a1, address, options
func (_m *MockExchange) Request(ctx context.Context, _a1 object.Hash, address string, options ...Option) error {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, _a1, address)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, object.Hash, string, ...Option) error); ok {
		r0 = rf(ctx, _a1, address, options...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Send provides a mock function with given fields: ctx, _a1, address, options
func (_m *MockExchange) Send(ctx context.Context, _a1 object.Object, address string, options ...Option) error {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, _a1, address)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, object.Object, string, ...Option) error); ok {
		r0 = rf(ctx, _a1, address, options...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Subscribe provides a mock function with given fields: filters
func (_m *MockExchange) Subscribe(filters ...EnvelopeFilter) EnvelopeSubscription {
	_va := make([]interface{}, len(filters))
	for _i := range filters {
		_va[_i] = filters[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 EnvelopeSubscription
	if rf, ok := ret.Get(0).(func(...EnvelopeFilter) EnvelopeSubscription); ok {
		r0 = rf(filters...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(EnvelopeSubscription)
		}
	}

	return r0
}
