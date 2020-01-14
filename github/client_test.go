package github

import (
	"errors"
	"testing"
)

func TestNewRepositoryClient(t *testing.T) {
	tests := []struct {
		name          string
		repo          string
		expectedErr   error
		expectedOwner string
		expectedName  string
	}{
		{
			name:        "Empty",
			repo:        "",
			expectedErr: ErrInvalidRepository,
		},
		{
			name:        "Only Owner",
			repo:        "owner",
			expectedErr: ErrInvalidRepository,
		},
		{
			name:          "Valid",
			repo:          "owner/name",
			expectedOwner: "owner",
			expectedName:  "name",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewRepositoryClient("fake-token", tc.repo)
			if tc.expectedErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected err: %v, got: %v", tc.expectedErr, err)
				}
				return
			}

			if client == nil {
				t.Fatalf("expected populated client, got: nil")
			}

			if client.client == nil {
				t.Fatalf("expected configured internal client, got: nil")
			}

			if tc.expectedOwner != client.owner {
				t.Fatalf("expected owner: %q, got: %q", tc.expectedOwner, client.owner)
			}

			if tc.expectedName != client.name {
				t.Fatalf("expected name: %q, got: %q", tc.expectedName, client.name)
			}
		})
	}
}

// TODO: is there a good way of testing these wrappers?
// func TestDownloadFileFromDefaultBranch(t *testing.T) {
// }

// func TestReplaceLabelsForIssue(t *testing.T) {
// }
