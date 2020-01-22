package github

import (
	"errors"
	"testing"

	"github.com/google/go-github/v29/github"
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

func TestNormalizedReviews(t *testing.T) {
	approved := Approved.String()
	changesRequested := ChangesRequested.String()
	dismissed := Dismissed.String()
	commented := Commented.String()

	tests := []struct {
		name     string
		reviews  []*github.PullRequestReview
		expected []Review
	}{
		{
			name:     "Nil",
			reviews:  nil,
			expected: nil,
		},
		{
			name:     "Empty",
			reviews:  []*github.PullRequestReview{},
			expected: []Review{},
		},
		{
			name: "One Of Each",
			reviews: []*github.PullRequestReview{
				&github.PullRequestReview{State: &approved, User: &github.User{ID: int64ToPtr(1)}},
				&github.PullRequestReview{State: &changesRequested, User: &github.User{ID: int64ToPtr(2)}},
				&github.PullRequestReview{State: &dismissed, User: &github.User{ID: int64ToPtr(3)}},
				&github.PullRequestReview{State: &commented, User: &github.User{ID: int64ToPtr(4)}},
			},
			expected: []Review{
				Approved,
				ChangesRequested,
				Dismissed,
				Commented,
			},
		},
		{
			name: "Multiple Reviews per User",
			reviews: []*github.PullRequestReview{
				&github.PullRequestReview{State: &changesRequested, User: &github.User{ID: int64ToPtr(1)}},
				&github.PullRequestReview{State: &changesRequested, User: &github.User{ID: int64ToPtr(2)}},
				&github.PullRequestReview{State: &dismissed, User: &github.User{ID: int64ToPtr(1)}},
				&github.PullRequestReview{State: &approved, User: &github.User{ID: int64ToPtr(2)}},
				&github.PullRequestReview{State: &approved, User: &github.User{ID: int64ToPtr(1)}},
			},
			expected: []Review{
				Approved,
				Approved,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := normalizedReviews(tc.reviews)
			assertReviewSlicesEqualUnordered(t, tc.expected, actual)
		})
	}
}

func int64ToPtr(i int64) *int64 {
	return &i
}

func assertReviewSlicesEqualUnordered(t *testing.T, a, b []Review) {
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
