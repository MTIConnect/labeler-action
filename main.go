package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	gh "github.com/google/go-github/v29/github"
	"gopkg.in/yaml.v3"

	"github.com/MTIConnect/labeler-action/github"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Action failed to complete: %s", err)
	}
}

func run() error {
	repo, err := github.NewRepositoryClient(
		os.Getenv("INPUT_GITHUB_TOKEN"),
		os.Getenv("GITHUB_REPOSITORY"),
	)

	// Get the config from default branch.
	configFile, err := repo.DownloadFileFromDefaultBranch(os.Getenv("INPUT_CONFIG_PATH"))
	if err != nil {
		return fmt.Errorf("failed to retrieve config file: %w", err)
	}

	// Parse the labels.
	var config labelerConfig
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal labeler config: %w", err)
	}
	log.Printf("Loaded action config: %s", os.Getenv("INPUT_CONFIG_PATH"))

	// Get event details for processing.
	eventName := os.Getenv("GITHUB_EVENT_NAME")
	payload, err := ioutil.ReadFile(os.Getenv("GITHUB_EVENT_PATH"))
	if err != nil {
		return fmt.Errorf("failed to read event payload: %w", payload)
	}

	// Get PR State using event details.
	state, err := prStateFromEvent(repo, eventName, payload)
	if err != nil {
		return fmt.Errorf("failed to process state from webhook data: %w", err)
	}

	// Evaluate the config rules.
	labels, err := config.labelsForPRState(state)
	if err != nil {
		return fmt.Errorf("failed to evaluate labels: %w", err)
	}

	// Replace labels after configured label operations.
	if !stringSlicesEqual(state.labels, labels) {
		err = repo.ReplaceLabelsForIssue(state.issueNumber, labels)
		if err != nil {
			return fmt.Errorf("failed to replace labels on pull request: %w", err)
		}
	}

	return nil
}

type labelerConfig map[string]struct {
	Draft            *bool
	BranchName       string `yaml:"branch_name"`
	Title            string
	ChangesRequested *bool `yaml:"changes_requested"`
	Approved         *bool
}

func (c labelerConfig) labelsForPRState(state prState) ([]string, error) {
	labels := state.labels
	for label, conditions := range c {
		if conditions.Approved != nil &&
			*conditions.Approved != state.approved {
			labels = removeLabel(labels, label)
			continue
		}

		if conditions.ChangesRequested != nil &&
			*conditions.ChangesRequested != state.changesRequested {
			labels = removeLabel(labels, label)
			continue
		}

		if conditions.Draft != nil &&
			*conditions.Draft != state.draft {
			labels = removeLabel(labels, label)
			continue
		}

		if conditions.Title != "" {
			matched, err := regexp.MatchString(conditions.Title, state.title)
			if err != nil {
				return nil, fmt.Errorf("failed to compile title regexp: %w", err)
			}
			if !matched {
				labels = removeLabel(labels, label)
				continue
			}
		}

		if conditions.BranchName != "" {
			matched, err := regexp.MatchString(conditions.BranchName, state.branchName)
			if err != nil {
				return nil, fmt.Errorf("failed to compile branch name regexp: %w", err)
			}
			if !matched {
				labels = removeLabel(labels, label)
				continue
			}
		}

		labels = addLabel(labels, label)
	}

	return labels, nil
}

type prState struct {
	issueNumber int
	labels      []string

	draft            bool
	branchName       string
	title            string
	changesRequested bool
	approved         bool
}

type reviewsLister interface {
	PullRequestReviews(int) ([]github.Review, error)
}

func prStateFromEvent(client reviewsLister, eventName string, payload []byte) (prState, error) {
	event, err := gh.ParseWebHook(eventName, payload)
	if err != nil {
		return prState{}, fmt.Errorf("failed to parse event data: %w", err)
	}

	prGetter, ok := event.(interface {
		GetPullRequest() *gh.PullRequest
	})
	if !ok {
		return prState{}, fmt.Errorf("event didn't relate to pull request")
	}
	pr := prGetter.GetPullRequest()

	reviews, err := client.PullRequestReviews(int(pr.GetNumber()))
	if err != nil {
		return prState{}, fmt.Errorf("couldn't list pull request reviews: %w", err)
	}

	var approved, changesRequested bool
	for _, review := range reviews {
		if review == github.Approved {
			approved = true
		}
		if review == github.ChangesRequested {
			changesRequested = true
		}
	}

	return prState{
		issueNumber: int(pr.GetNumber()),
		labels:      labelNames(pr.Labels),

		draft:            pr.GetDraft(),
		title:            pr.GetTitle(),
		branchName:       pr.GetHead().GetRef(),
		approved:         approved,
		changesRequested: changesRequested,
	}, nil
}

func labelNames(labels []*gh.Label) []string {
	strings := make([]string, 0, len(labels))
	for _, label := range labels {
		strings = append(strings, label.GetName())
	}
	return strings
}

func removeLabel(labels []string, removal string) []string {
	n := 0
	for _, label := range labels {
		if label != removal {
			labels[n] = label
			n++
		}
	}
	return labels[:n]
}

func addLabel(labels []string, addition string) []string {
	for _, label := range labels {
		if label == addition {
			return labels
		}
	}
	return append(labels, addition)
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
