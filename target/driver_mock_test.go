// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/jamillosantos/migrations-sql/target (interfaces: Driver)

// Package target is a generated GoMock package.
package target

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	migrations "github.com/jamillosantos/migrations"
)

// MockDriver is a mock of Driver interface.
type MockDriver struct {
	ctrl     *gomock.Controller
	recorder *MockDriverMockRecorder
}

// MockDriverMockRecorder is the mock recorder for MockDriver.
type MockDriverMockRecorder struct {
	mock *MockDriver
}

// NewMockDriver creates a new mock instance.
func NewMockDriver(ctrl *gomock.Controller) *MockDriver {
	mock := &MockDriver{ctrl: ctrl}
	mock.recorder = &MockDriverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDriver) EXPECT() *MockDriverMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockDriver) Add(arg0 migrations.Migration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockDriverMockRecorder) Add(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockDriver)(nil).Add), arg0)
}

// Lock mocks base method.
func (m *MockDriver) Lock() (migrations.Unlocker, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Lock")
	ret0, _ := ret[0].(migrations.Unlocker)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Lock indicates an expected call of Lock.
func (mr *MockDriverMockRecorder) Lock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Lock", reflect.TypeOf((*MockDriver)(nil).Lock))
}

// Remove mocks base method.
func (m *MockDriver) Remove(arg0 migrations.Migration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove.
func (mr *MockDriverMockRecorder) Remove(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockDriver)(nil).Remove), arg0)
}