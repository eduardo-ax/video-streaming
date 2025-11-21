package domain

import "testing"

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
			desc:     "should pass validation with vlaid data",
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
