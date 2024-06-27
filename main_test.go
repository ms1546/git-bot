package main

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v33/github"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/stretchr/testify/assert"
)

type MockGithubClient struct {
	*gomock.Controller
	mock *github.Client
}

type MockLineBotClient struct {
	*gomock.Controller
	mock *linebot.Client
}

func TestGetGithubEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGithubClient := NewMockGithubClient(ctrl)
	ctx := context.Background()
	username := "testuser"

	mockEvents := []*github.Event{
		{
			CreatedAt: &github.Timestamp{Time: time.Now()},
		},
	}

	mockGithubClient.EXPECT().Activity.ListEventsPerformedByUser(ctx, username, false, nil).Return(mockEvents, nil, nil)

	events, err := getGithubEvents(ctx, mockGithubClient.mock, username)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(events))
}

func TestSendLineMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLineBotClient := NewMockLineBotClient(ctrl)
	userID := "testuserid"
	message := "test message"

	mockLineBotClient.EXPECT().PushMessage(userID, linebot.NewTextMessage(message)).Return(&linebot.BasicResponse{}, nil)

	err := sendLineMessage(mockLineBotClient.mock, userID, message)
	assert.NoError(t, err)
}
