package github

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
)

var (
	// ErrInvalidRepository occurs when an invalid repo string is passed
	// into NewRepositoryClient.
	ErrInvalidRepository = errors.New("invalid repository format")
)

// RepositoryClient is a Github client with operations targetted
// at a specific repository.
type RepositoryClient struct {
	client *github.Client
	owner  string
	name   string
}

// NewRepositoryClient initializes a oauth client and formats data for
// further repository actions. The token is assumed valid and repo is expected
// in the "owner/name" format.
func NewRepositoryClient(token, repo string) (*RepositoryClient, error) {
	split := strings.Split(repo, "/")
	if len(split) != 2 {
		return nil, fmt.Errorf("%w: %q", ErrInvalidRepository, repo)
	}
	owner, name := split[0], split[1]

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	oauthClient := oauth2.NewClient(context.TODO(), tokenSource)
	client := github.NewClient(oauthClient)

	return &RepositoryClient{
		client: client,
		owner:  owner,
		name:   name,
	}, nil
}

// DownloadFileFromDefaultBranch synchronously downloads a file at the specified path
// on the Github configured default branch.
func (r RepositoryClient) DownloadFileFromDefaultBranch(path string) ([]byte, error) {
	file, _, _, err := r.client.Repositories.GetContents(
		context.TODO(),
		r.owner,
		r.name,
		path,
		nil, // Optional SHA, defaults to repos default branch.
	)
	if err != nil {
		return nil, fmt.Errorf("github client error: %w", err)
	}

	// If we get a nil config without error, then we were supplied a directory.
	if file == nil {
		return nil, fmt.Errorf("directory path supplied: %q", path)
	}
	fileContent, err := file.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to load file contents: %w", err)
	}

	return []byte(fileContent), nil
}

// ReplaceLabelsForIssue replaces the set of labels on an issue within the repository.
func (r RepositoryClient) ReplaceLabelsForIssue(number int, labels []string) error {
	_, _, err := r.client.Issues.ReplaceLabelsForIssue(context.TODO(), r.owner, r.name, number, labels)
	if err != nil {
		return fmt.Errorf("failed to replace labels: %w", err)
	}
	return nil
}

// Review is the current state of a pull request review.
type Review int

func (r Review) String() string {
	switch r {
	case Approved:
		return "approved"
	case ChangesRequested:
		return "changes_requested"
	case Dismissed:
		return "dismissed"
	case Commented:
		return "commented"
	default:
		return "UNKNOWN"
	}
}

// Review enum declaration.
const (
	Unknown Review = iota
	Approved
	ChangesRequested
	Dismissed
	Commented
)

// Map lookup table from string to Review enum.
var reviewLookupTable = map[string]Review{
	"approved":          Approved,
	"changes_requested": ChangesRequested,
	"dismissed":         Dismissed,
	"commented":         Commented,
}

// PullRequestReviews returns a slice of review states on the pull request.
func (r RepositoryClient) PullRequestReviews(number int) ([]Review, error) {
	opt := &github.ListOptions{PerPage: 100}
	var allReviews []*github.PullRequestReview
	for {
		reviews, resp, err := r.client.PullRequests.ListReviews(context.TODO(), r.owner, r.name, number, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list reviews for pull request: %w", err)
		}
		allReviews = append(allReviews, reviews...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return normalizedReviews(allReviews), nil
}

// normalizedReviews takes a slices of reviews and returns a list of each users latest review state.
func normalizedReviews(reviews []*github.PullRequestReview) []Review {
	// Reviews are in chronological order, overwrite previous reviews of the same user.
	statePerUser := make(map[int64]Review)
	for _, review := range reviews {
		state := reviewLookupTable[strings.ToLower(review.GetState())]

		// Ignore the commented state, as they aren't actual reviews.
		if state != Commented {
			statePerUser[review.GetUser().GetID()] = state
		}
	}

	states := make([]Review, 0, len(statePerUser))
	for _, review := range statePerUser {
		states = append(states, review)
	}
	return states
}
