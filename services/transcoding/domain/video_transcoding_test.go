package domain

import (
	"testing"
)

func TestQueueContentValidate(t *testing.T) {
	tests := map[string]struct {
		id      string
		content string
		expect  bool
	}{
		"valid queue content": {
			id:      "12341",
			content: "17/file_example_AVI_1920_2_3MG.avi",
			expect:  true,
		},
		"invalid id": {
			id:      "-1",
			content: "17/file_example_AVI_1920_2_3MG.avi",
			expect:  false,
		},
		"invalid id zero": {
			id:      "0",
			content: "",
			expect:  false,
		},
		"invalid content": {
			id:      "123",
			content: "",
			expect:  false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := NewQueueContent(tc.id, tc.content)
			if err != nil && tc.expect {
				t.Errorf("Test %s failed: %s. Expected valid but got error: %v", name, tc.content, err)
			}
		})
	}

}
