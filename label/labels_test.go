package label_test

import (
	"testing"

	"github.com/MTIConnect/labeler-action/labels"
)

func TestOperations(t *testing.T) {
	tests := []struct {
		name     string
		action   label.Operations
		labels   []string
		expected []string
	}{
		{
			name:     "Empty",
			action:   label.Operations{},
			labels:   []string{},
			expected: []string{},
		},
		{
			name:     "Untouched",
			action:   label.Operations{},
			labels:   []string{"I", "Have", "Labels"},
			expected: []string{"I", "Have", "Labels"},
		},
		{
			name: "Removes",
			action: label.Operations{
				Remove: []string{"Ready for Review"},
			},
			labels:   []string{"Ready for Review", "Don't Touch"},
			expected: []string{"Don't Touch"},
		},
		{
			name: "Sets",
			action: label.Operations{
				Set: []string{"Ready for Review", "Bug"},
			},
			labels:   []string{"Ready for Review", "Don't Touch"},
			expected: []string{"Ready for Review", "Don't Touch", "Bug"},
		},
		{
			name: "SetsAndRemoves",
			action: label.Operations{
				Set:    []string{"Ready for Review", "Bug"},
				Remove: []string{"Remove Me"},
			},
			labels:   []string{"Ready for Review", "Remove Me"},
			expected: []string{"Ready for Review", "Bug"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.action.Apply(tc.labels)
			assertStringSlicesEqual(t, tc.expected, actual)
		})
	}
}

func assertStringSlicesEqual(t *testing.T, a, b []string) {
	if len(a) != len(b) {
		t.Errorf("expected slice lengths to equal: %d != %d\nExpected: %v\nActual: %v", len(a), len(b), a, b)
		return
	}
	for i := range a {
		if a[i] != b[i] {
			t.Errorf("Expected slices to be equal\nExpected: %v\nActual: %v", a, b)
		}
	}
}
