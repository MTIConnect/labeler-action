package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	gh "github.com/google/go-github/v29/github"
	"gopkg.in/yaml.v3"

	"github.com/MTIConnect/labeler-action/github"
	"github.com/MTIConnect/labeler-action/label"
)

func main() {
	if err := run(); err != nil {
		log.Printf("Action failed to complete: %s", err)
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
	var config map[string]label.Operations
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
	state, err := prStateFromEvent(eventName, payload)
	if err != nil {
		return fmt.Errorf("failed to process state from webhook data: %w", err)
	}

	// Replace labels after configured label operations.
	labels := config[state.state].Apply(state.labels)
	err = repo.ReplaceLabelsForIssue(state.issueID, labels)
	if err != nil {
		return fmt.Errorf("failed to replace labels on pull request: %w", err)
	}

	return nil
}

// TODO: Pull this out and make it more generic?
type prState struct {
	issueID int
	state   string
	labels  []string
}

func prStateFromEvent(eventName string, payload []byte) (prState, error) {
	event, err := gh.ParseWebHook(eventName, payload)
	if err != nil {
		return prState{}, fmt.Errorf("failed to parse event data: %w", err)
	}

	var state string
	var labels []string
	var issueID int
	switch event := event.(type) {
	case *gh.PullRequestEvent:
		pr := event.GetPullRequest()
		labels = labelsToStrings(pr.Labels)
		issueID = int(pr.GetNumber())
		state = event.GetAction()
	case *gh.PullRequestReviewEvent:
		pr := event.GetPullRequest()
		labels = labelsToStrings(pr.Labels)
		issueID = int(pr.GetNumber())
		switch event.GetAction() {
		case "submitted":
			switch strings.ToLower(event.GetReview().GetState()) {
			case "approve":
				state = "approved"
			case "request_changes":
				state = "changes_requested"
			}
		case "dismissed":
			state = "ready_for_review"
		}
	}

	return prState{
		state:   state,
		labels:  labels,
		issueID: issueID,
	}, nil
}

func labelsToStrings(labels []*gh.Label) []string {
	strings := make([]string, 0, len(labels))
	for _, label := range labels {
		strings = append(strings, label.GetName())
	}
	return strings
}
