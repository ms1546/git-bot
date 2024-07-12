package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/stretchr/testify/assert"
)

// GitHubClientInterface defines the methods used from the GitHub client
type GitHubClientInterface interface {
	ListEventsPerformedByUser(ctx context.Context, user string, publicOnly bool, opt *github.ListOptions) ([]*github.Event, *github.Response, error)
}

type GitHubClientWrapper struct {
	client *github.Client
}

func (g *GitHubClientWrapper) ListEventsPerformedByUser(ctx context.Context, user string, publicOnly bool, opt *github.ListOptions) ([]*github.Event, *github.Response, error) {
	return g.client.Activity.ListEventsPerformedByUser(ctx, user, publicOnly, opt)
}

// LineBotClientInterface defines the methods used from the LINE bot client
type LineBotClientInterface interface {
	PushMessage(to string, messages ...linebot.SendingMessage) *linebot.PushMessageCall
}

// MockGitHubClient is a mock implementation of GitHubClientInterface
type MockGitHubClient struct{}

func (m *MockGitHubClient) ListEventsPerformedByUser(ctx context.Context, user string, publicOnly bool, opt *github.ListOptions) ([]*github.Event, *github.Response, error) {
	timestamp := time.Now().AddDate(0, 0, -1)
	return []*github.Event{
		{
			Type:      github.String("PushEvent"),
			Repo:      &github.Repository{Name: github.String("mock/repo")},
			CreatedAt: &github.Timestamp{Time: timestamp},
		},
	}, &github.Response{NextPage: 0}, nil
}

// MockLineBotClient is a mock implementation of LineBotClientInterface
type MockLineBotClient struct{}

func (m *MockLineBotClient) PushMessage(to string, messages ...linebot.SendingMessage) *linebot.PushMessageCall {
	return &linebot.PushMessageCall{}
}

func TestGetGithubEvents(t *testing.T) {
	client := &MockGitHubClient{}
	ctx := context.Background()

	events, err := getGithubEvents(ctx, client, "mockUser", time.Now().AddDate(0, 0, -1).Format("2006-01-02"))
	assert.NoError(t, err)
	assert.NotEmpty(t, events)
}

func TestBuildMessage(t *testing.T) {
	timestamp := time.Now().AddDate(0, 0, -1)
	createdAt := github.Timestamp{Time: timestamp}
	events := []*github.Event{
		{
			Type:      github.String("PushEvent"),
			Repo:      &github.Repository{Name: github.String("mock/repo")},
			CreatedAt: &createdAt,
		},
	}

	message := buildMessage(events, false)
	assert.Contains(t, message, "mock/repo")
}

func TestSendLineMessage(t *testing.T) {
	client := &MockLineBotClient{}
	err := sendLineMessage(client, "mockUserID", "test message")
	assert.NoError(t, err)
}
