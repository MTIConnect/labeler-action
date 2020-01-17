package main

import (
	"testing"

	gh "github.com/google/go-github/v29/github"
)

func TestLabelsForPRState(t *testing.T) {
	falseCheck := false
	trueCheck := true
	config := labelerConfig{
		"WIP": {
			Draft: &trueCheck,
		},
		"Bug": {
			BranchName: "^(bug|issue)/",
		},
		"Feature": {
			BranchName: "^(feature|enhancement)/",
		},
		"Refactor": {
			Title: "^Refactor -",
		},
		"Awaiting Code Review": {
			Draft:            &falseCheck,
			ChangesRequested: &falseCheck,
			Approved:         &falseCheck,
		},
		"Changes Requested": {
			Draft:            &falseCheck,
			ChangesRequested: &trueCheck,
		},
		"Code Review Approved": {
			Draft:            &falseCheck,
			ChangesRequested: &falseCheck,
			Approved:         &trueCheck,
		},
	}

	tests := []struct {
		name     string
		state    prState
		expected []string
	}{
		{
			name: "Bug Draft",
			state: prState{
				labels:     []string{"Widgets Epic"},
				draft:      true,
				branchName: "bug/widget-factory-does-not-make-widgets",
			},
			expected: []string{"Widgets Epic", "Bug", "WIP"},
		},
		{
			name: "Bug Draft Approved",
			state: prState{
				labels:     []string{"Widgets Epic"},
				draft:      true,
				approved:   true,
				branchName: "bug/widget-factory-does-not-make-widgets",
			},
			expected: []string{"Widgets Epic", "Bug", "WIP"},
		},
		{
			name: "Feature Refactor AwaitingCodeReview",
			state: prState{
				labels:     []string{"Widgets Epic"},
				branchName: "enhancement/spinning-widgets-refactor",
				title:      "Refactor - Pull Out Spinning Widgets",
			},
			expected: []string{"Widgets Epic", "Feature", "Refactor", "Awaiting Code Review"},
		},
		{
			name: "Feature Approved",
			state: prState{
				labels:     []string{"Widgets Epic", "Feature", "Awaiting Code Review"},
				branchName: "feature/slicing-widget",
				approved:   true,
			},
			expected: []string{"Widgets Epic", "Feature", "Code Review Approved"},
		},
		{
			name: "Conflicting Reviews",
			state: prState{
				labels:           []string{"Widgets Epic", "Code Review Approved"},
				approved:         true,
				changesRequested: true,
			},
			expected: []string{"Widgets Epic", "Changes Requested"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := config.labelsForPRState(tc.state)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			assertStringSlicesEqualUnordered(t, tc.expected, actual)
		})
	}
}

func TestLabelNames(t *testing.T) {
	tests := []struct {
		name     string
		labels   []*gh.Label
		expected []string
	}{
		{
			name:     "Nil",
			labels:   nil,
			expected: []string{},
		},
		{
			name:     "Empty",
			labels:   []*gh.Label{},
			expected: []string{},
		},
		{
			name: "HasLabels",
			labels: []*gh.Label{
				&gh.Label{Name: stringToPtr("Label 1")},
				&gh.Label{Name: stringToPtr("Label 2")},
				&gh.Label{Name: stringToPtr("Label 3")},
			},
			expected: []string{"Label 1", "Label 2", "Label 3"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := labelNames(tc.labels)
			assertStringSlicesEqual(t, tc.expected, actual)
		})
	}
}

func TestAddLabel(t *testing.T) {
	tests := []struct {
		name     string
		labels   []string
		add      string
		expected []string
	}{
		{
			name:     "Nil",
			labels:   nil,
			add:      "Ready for Review",
			expected: []string{"Ready for Review"},
		},
		{
			name:     "Empty",
			labels:   []string{},
			add:      "Ready for Review",
			expected: []string{"Ready for Review"},
		},
		{
			name:     "Exists",
			labels:   []string{"Changes Requested"},
			add:      "Changes Requested",
			expected: []string{"Changes Requested"},
		},
		{
			name:     "Exists_2",
			labels:   []string{"Bug", "Changes Requested"},
			add:      "Changes Requested",
			expected: []string{"Bug", "Changes Requested"},
		},
		{
			name:     "Appends",
			labels:   []string{"Bug"},
			add:      "Changes Requested",
			expected: []string{"Bug", "Changes Requested"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := addLabel(tc.labels, tc.add)
			assertStringSlicesEqual(t, tc.expected, actual)
		})
	}
}

func TestRemoveLabel(t *testing.T) {
	tests := []struct {
		name     string
		labels   []string
		remove   string
		expected []string
	}{
		{
			name:     "Nil",
			labels:   nil,
			remove:   "remove this",
			expected: []string{},
		},
		{
			name:     "Empty",
			labels:   []string{},
			remove:   "remove this",
			expected: []string{},
		},
		{
			name:     "No matches",
			labels:   []string{"I", "Have", "Labels"},
			remove:   "remove_this",
			expected: []string{"I", "Have", "Labels"},
		},
		{
			name:     "Matches 1",
			labels:   []string{"Ready for Review", "Don't Touch"},
			remove:   "Ready for Review",
			expected: []string{"Don't Touch"},
		},
		{
			name:     "Matches 2",
			labels:   []string{"Ready for Review", "Don't Touch", "Ready for Review"},
			remove:   "Ready for Review",
			expected: []string{"Don't Touch"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := removeLabel(tc.labels, tc.remove)
			assertStringSlicesEqual(t, tc.expected, actual)
		})
	}
}

func assertStringSlicesEqualUnordered(t *testing.T, a, b []string) {
	if len(a) != len(b) {
		t.Errorf("expected slice lengths to equal: %d != %d\nExpected: %v\nActual: %v", len(a), len(b), a, b)
		return
	}

	for _, str := range a {
		found := false
		for _, other := range b {
			if str == other {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected slices to equal unordered\nExpected: %v\nActual: %v", a, b)
		}
	}

}

func assertStringSlicesEqual(t *testing.T, a, b []string) {
	if len(a) != len(b) {
		t.Errorf("expected slice lengths to equal: %d != %d\nExpected: %v\nActual: %v", len(a), len(b), a, b)
		return
	}
	for i := range a {
		if a[i] != b[i] {
			t.Errorf("expected slices to be equal\nExpected: %v\nActual: %v", a, b)
		}
	}
}

func stringToPtr(s string) *string {
	return &s
}
