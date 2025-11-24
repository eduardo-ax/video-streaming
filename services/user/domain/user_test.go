package domain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserValidade(t *testing.T) {
	tests := map[string]struct {
		name     string
		email    string
		password string
		plan     uint8
		expected bool
		desc     string
	}{
		"valid user": {
			name:     "Eduardo",
			email:    "eduardo@email.com",
			password: "123%$#123",
			plan:     1,
			expected: true,
			desc:     "should pass validation with valid data",
		},
		"invalid user name": {
			name:     "",
			email:    "eduardo@email.com",
			password: "123%$#123",
			plan:     1,
			expected: false,
			desc:     "should not validate when user name is empty",
		},
		"invalid user email": {
			name:     "Eduardo",
			email:    "eduadsad2l.cm",
			password: "123%$#123",
			plan:     1,
			expected: false,
			desc:     "should not validate when email is empty or invalid",
		},
		"invalid user password": {
			name:     "Eduardo",
			email:    "eduardo@email.com",
			password: "",
			plan:     1,
			expected: false,
			desc:     "should not validate when password is empty",
		},
		"invalid user password 2": {
			name:     "Eduardo",
			email:    "eduardo@email.com",
			password: "2312",
			plan:     1,
			expected: false,
			desc:     "should not validate when password has less than 8 characters",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := NewUser(tc.name, tc.email, tc.password)
			if err != nil && tc.expected {
				t.Errorf("Test %s failed: %s. Expected valid but got error: %v", name, tc.desc, err)
			}
		})
	}
}

type MockStorage struct{ mock.Mock }

func (m *MockStorage) Persist(ctx context.Context, name string, email string, password string, plan uint8) (string, error) {
	args := m.Called(ctx, name, email, password, plan)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockStorage) CreateSession(ctx context.Context, session *Session) (*Session, error) {
	args := m.Called(ctx, session)
	return args.Get(0).(*Session), args.Error(1)
}

func (m *MockStorage) DeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStorage) UpdateUser(ctx context.Context, id string, name string, email *string, password *string) error {
	args := m.Called(ctx, id, name, email, password)
	return args.Error(0)
}

func (m *MockStorage) GetUser(ctx context.Context, email string) (*UserAuthData, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*UserAuthData), args.Error(1)
}

func (m *MockStorage) GetSession(ctx context.Context, id string) (*Session, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Session), args.Error(1)
}

func (m *MockStorage) DeleteSession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStorage) RevokeSession(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockToken struct{ mock.Mock }

func (m *MockToken) CreateToken(id string, email string, plan int8, sessionID string, duration time.Duration) (string, *UserClaims, error) {
	args := m.Called(id, email, plan, sessionID, duration)
	return args.Get(0).(string), args.Get(1).(*UserClaims), args.Error(2)
}

func (m *MockToken) VerifyToken(tokenStr string) (*UserClaims, error) {
	args := m.Called(tokenStr)
	return args.Get(0).(*UserClaims), args.Error(1)
}

func TestNewUserManager(t *testing.T) {
	ctx := context.Background()

	tests := map[string]struct {
		name       string
		email      string
		password   string
		plan       uint8
		setupMocks func(storage *MockStorage, token *MockToken)
		expected   bool
		desc       string
	}{
		"successful store": {
			name:     "Eduardo",
			email:    "eduardo@email.com",
			password: "!@#qwe1233",
			plan:     1,
			setupMocks: func(db *MockStorage, tk *MockToken) {
				db.On("Persist", mock.Anything, "Eduardo", "eduardo@email.com", mock.Anything, uint8(1)).Return("UserEUID-12312", nil)
				tk.On("CreateToken", "userEUID", "eduardo@email.com", int8(1), mock.AnythingOfType("string"), 15*time.Minute)
			},
			expected: true,
			desc:     "should pass validation with valid data",
		},
		"failed store name empty": {
			name:     "",
			email:    "eduardo@email.com",
			password: "!@#qwe1233",
			plan:     1,
			setupMocks: func(db *MockStorage, tk *MockToken) {
				db.On("Persist", mock.Anything, "", "eduardo@email.com", mock.Anything, uint8(1)).Return("UserEUID-12312", nil)
				tk.On("CreateToken", "userEUID", "eduardo@email.com", int8(1), mock.AnythingOfType("string"), 15*time.Minute)
			},
			expected: false,
			desc:     "should failed validation with invalid data",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dbMock := new(MockStorage)
			tkMock := new(MockToken)
			tc.setupMocks(dbMock, tkMock)
			manager := NewUserManager(dbMock, tkMock)
			err := manager.CreateUser(ctx, tc.name, tc.email, tc.password)
			if tc.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
			dbMock.AssertExpectations(t)
		})
	}
}
