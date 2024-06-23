// Code generated by mockery v2.24.0. DO NOT EDIT.

package mocks

import (
	context "context"

	service "github.com/lks-go/url-shortener/internal/service"
	mock "github.com/stretchr/testify/mock"
)

// URLStorage is an autogenerated mock type for the URLStorage type
type URLStorage struct {
	mock.Mock
}

// CodeByURL provides a mock function with given fields: ctx, url
func (_m *URLStorage) CodeByURL(ctx context.Context, url string) (string, error) {
	ret := _m.Called(ctx, url)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, url)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, url)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, url)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteURLs provides a mock function with given fields: ctx, codes
func (_m *URLStorage) DeleteURLs(ctx context.Context, codes []string) error {
	ret := _m.Called(ctx, codes)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) error); ok {
		r0 = rf(ctx, codes)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Exists provides a mock function with given fields: ctx, code
func (_m *URLStorage) Exists(ctx context.Context, code string) (bool, error) {
	ret := _m.Called(ctx, code)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (bool, error)); ok {
		return rf(ctx, code)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) bool); ok {
		r0 = rf(ctx, code)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, code)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: ctx, code, url
func (_m *URLStorage) Save(ctx context.Context, code string, url string) error {
	ret := _m.Called(ctx, code, url)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, code, url)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveBatch provides a mock function with given fields: ctx, url
func (_m *URLStorage) SaveBatch(ctx context.Context, url []service.URL) error {
	ret := _m.Called(ctx, url)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []service.URL) error); ok {
		r0 = rf(ctx, url)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveUsersCode provides a mock function with given fields: ctx, userID, code
func (_m *URLStorage) SaveUsersCode(ctx context.Context, userID string, code string) error {
	ret := _m.Called(ctx, userID, code)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, userID, code)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// URL provides a mock function with given fields: ctx, id
func (_m *URLStorage) URL(ctx context.Context, id string) (string, error) {
	ret := _m.Called(ctx, id)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UsersURLCodes provides a mock function with given fields: ctx, userID
func (_m *URLStorage) UsersURLCodes(ctx context.Context, userID string) ([]string, error) {
	ret := _m.Called(ctx, userID)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]string, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []string); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UsersURLs provides a mock function with given fields: ctx, userID
func (_m *URLStorage) UsersURLs(ctx context.Context, userID string) ([]service.UsersURL, error) {
	ret := _m.Called(ctx, userID)

	var r0 []service.UsersURL
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]service.UsersURL, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []service.UsersURL); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]service.UsersURL)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewURLStorage interface {
	mock.TestingT
	Cleanup(func())
}

// NewURLStorage creates a new instance of URLStorage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewURLStorage(t mockConstructorTestingTNewURLStorage) *URLStorage {
	mock := &URLStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}