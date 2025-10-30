package domain

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVideoValidate(t *testing.T) {
	tests := map[string]struct {
		title       string
		description string
		content     *multipart.FileHeader
		expected    bool
		desc        string
	}{
		"valid video": {
			title:       "Sample Video",
			description: "This is a sample video description.",
			content:     &multipart.FileHeader{Filename: "video.mp4", Size: 1024},
			expected:    true,
			desc:        "should pass validation with valid data",
		},
		"nil content": {
			title:       "Sample Video",
			description: "This is a sample video description.",
			content:     nil,
			expected:    false,
			desc:        "should fail validation when content is nil",
		},
		"empty title": {
			title:       "",
			description: "This is a sample video description.",
			content:     &multipart.FileHeader{Filename: "video.mp4", Size: 1024},
			expected:    false,
			desc:        "should fail validation when title is empty",
		},
		"empty description": {
			title:       "Sample Video",
			description: "",
			content:     &multipart.FileHeader{Filename: "video.mp4", Size: 1024},
			expected:    true,
			desc:        "should fail validation when description is empty",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := NewVideo(tc.title, tc.description, tc.content)
			if err != nil && tc.expected {
				t.Errorf("Test %s failed: %s. Expected valid but got error: %v", name, tc.desc, err)
			}
		})
	}
}

type MockStorage struct{ mock.Mock }

func (m *MockStorage) Persist(ctx context.Context, title string, description string) (int, error) {
	args := m.Called(ctx, title, description)
	return args.Get(0).(int), args.Error(1)
}

type MockMessagePublisher struct{ mock.Mock }

func (m *MockMessagePublisher) SendMessage(ctx context.Context, message string) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

type MockObjectStore struct{ mock.Mock }

func (m *MockObjectStore) UploadVideo(ctx context.Context, file *multipart.FileHeader, id int) error {
	args := m.Called(ctx, file, id)
	return args.Error(0)
}

func TestVideoManager_Store(t *testing.T) {
	ctx := context.Background()
	file := &multipart.FileHeader{Filename: "video.mp4", Size: 1024}

	tests := map[string]struct {
		title       string
		description string
		content     *multipart.FileHeader
		expected    bool
		setupMocks  func(storage *MockStorage, publisher *MockMessagePublisher, objectStore *MockObjectStore)
		desc        string
	}{
		"successful store": {
			title:       "Sample Video",
			description: "This is a sample video description.",
			content:     file,
			setupMocks: func(db *MockStorage, pub *MockMessagePublisher, store *MockObjectStore) {
				db.On("Persist", mock.Anything, "Sample Video", "This is a sample video description.").Return(1, nil)
				pub.On("SendMessage", mock.Anything, "1").Return(nil)
				store.On("UploadVideo", mock.Anything, file, 1).Return(nil)
			},
			expected: true,
			desc:     "should store video successfully with valid data",
		},
		"failed database": {
			title:       "Sample Video",
			description: "This is a sample video description.",
			content:     file,
			setupMocks: func(db *MockStorage, pub *MockMessagePublisher, store *MockObjectStore) {
				db.On("Persist", mock.Anything, "Sample Video", "This is a sample video description.").Return(-1, errors.New("failed to persist"))
			},
			expected: false,
			desc:     "should fail to store video when upload fails",
		},
		"failed send message": {
			title:       "Sample Video",
			description: "This is a sample video description.",
			content:     file,
			setupMocks: func(db *MockStorage, pub *MockMessagePublisher, store *MockObjectStore) {
				db.On("Persist", mock.Anything, "Sample Video", "This is a sample video description.").Return(1, nil)
				pub.On("SendMessage", mock.Anything, "1").Return(errors.New("failed to send message"))
			},
			expected: false,
			desc:     "should fail to store video when send message to queue fails",
		},
		"failed object store": {
			title:       "Sample Video",
			description: "This is a sample video description.",
			content:     file,
			setupMocks: func(db *MockStorage, pub *MockMessagePublisher, store *MockObjectStore) {
				db.On("Persist", mock.Anything, "Sample Video", "This is a sample video description.").Return(1, nil)
				pub.On("SendMessage", mock.Anything, "1").Return(nil)
				store.On("UploadVideo", mock.Anything, file, 1).Return(errors.New("failed to upload video"))
			},
			expected: false,
			desc:     "should fail to store video when object store upload fails",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dbMock := new(MockStorage)
			pubMock := new(MockMessagePublisher)
			storeMock := new(MockObjectStore)

			tc.setupMocks(dbMock, pubMock, storeMock)
			manager := NewVideoManager(dbMock, pubMock, storeMock)
			err := manager.Store(ctx, tc.title, tc.description, tc.content)

			if tc.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
			dbMock.AssertExpectations(t)
			pubMock.AssertExpectations(t)
			storeMock.AssertExpectations(t)
		})
	}

}
