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

// Review enum declaration.
const (
	Approved Review = iota
	ChangesRequested
	Dismissed
	Commented
)

// PullRequestReviews returns a slice of review states on the pull request.
func (r RepositoryClient) PullRequestReviews(number int) ([]Review, error) {
	// TODO: paging might be an issue if there are a lot of reviews.
	reviews, _, err := r.client.PullRequests.ListReviews(context.TODO(), r.owner, r.name, number, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list reviews for pull request: %w", err)
	}

	states := make([]Review, 0, len(reviews))
	for _, review := range reviews {
		switch strings.ToLower(review.GetState()) {
		case "approved":
			states = append(states, Approved)
		case "changes_requested":
			states = append(states, ChangesRequested)
		case "dismissed":
			states = append(states, Dismissed)
		case "commented":
			states = append(states, Commented)
		}
	}
	return states, nil
}
