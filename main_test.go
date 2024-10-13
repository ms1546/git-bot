package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGitHubClient struct {
	mock.Mock
}

func (m *MockGitHubClient) Activity() *github.ActivityService {
	args := m.Called()
	return args.Get(0).(*github.ActivityService)
}

func (m *MockGitHubClient) ListEventsPerformedByUser(ctx context.Context, username string, publicOnly bool, opt *github.ListOptions) ([]*github.Event, *github.Response, error) {
	args := m.Called(ctx, username, publicOnly, opt)
	return args.Get(0).([]*github.Event), args.Get(1).(*github.Response), args.Error(2)
}

type MockLineBotClient struct {
	mock.Mock
}

func (m *MockLineBotClient) PushMessage(to string, messages ...linebot.SendingMessage) *linebot.BotResponse {
	args := m.Called(to, messages)
	return args.Get(0).(*linebot.BotResponse)
}

func TestGetGithubEvents(t *testing.T) {
	mockClient := new(MockGitHubClient)
	ctx := context.Background()
	date := time.Now().Format("2006-01-02")

	events := []*github.Event{
		{
			Repo:      &github.Repository{Name: github.String("test-repo")},
			Type:      github.String("PushEvent"),
			CreatedAt: &github.Timestamp{Time: time.Now()},
		},
	}
	resp := &github.Response{
		NextPage: 0,
	}

	mockClient.On("ListEventsPerformedByUser", ctx, "test-user", false, mock.Anything).Return(events, resp, nil)

	result, err := getGithubEvents(ctx, mockClient, "test-user", date)
	assert.NoError(t, err)
	assert.Equal(t, len(result), 1)
	mockClient.AssertExpectations(t)
}

func TestSendLineMessage(t *testing.T) {
	mockBot := new(MockLineBotClient)
	userID := "U1234567890"
	message := "Test message"
	stickerMessage := linebot.NewStickerMessage("11537", "52002735")

	mockBot.On("PushMessage", userID, mock.Anything).Return(&linebot.BotResponse{})

	err := sendLineMessage(mockBot, userID, message, true)
	assert.NoError(t, err)
	mockBot.AssertExpectations(t)
}

func TestSendErrorMessage(t *testing.T) {
	mockBot := new(MockLineBotClient)
	userID := "U1234567890"
	errMsg := "An error occurred"

	mockBot.On("PushMessage", userID, mock.Anything).Return(&linebot.BotResponse{})

	sendErrorMessage(mockBot, userID, errMsg)
	mockBot.AssertExpectations(t)
}
