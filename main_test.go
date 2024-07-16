package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v33/github"
)

func TestCreateGithubClient(t *testing.T) {
	ctx := context.Background()
	token := "dummy_token"
	client := createGithubClient(ctx, token)
	if client == nil {
		t.Fatalf("Expected github.Client, got nil")
	}
}

func TestCreateLineBotClient(t *testing.T) {
	secret := "dummy_secret"
	token := "dummy_token"
	bot, err := createLineBotClient(secret, token)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if bot == nil {
		t.Fatalf("Expected linebot.Client, got nil")
	}
}

func TestBuildMessage(t *testing.T) {
	event := &github.Event{
		Repo: &github.Repository{Name: github.String("test/repo")},
		Type: github.String("PushEvent"),
	}
	events := []*github.Event{event}

	t.Run("With events", func(t *testing.T) {
		isFinalCheck := false
		msg := buildMessage(events, isFinalCheck)
		if msg == "" {
			t.Fatalf("Expected non-empty message, got empty string")
		}
	})

	t.Run("Without events, not final check", func(t *testing.T) {
		isFinalCheck := false
		msg := buildMessage([]*github.Event{}, isFinalCheck)
		expected := "まだ本日はGitHubに草が生えていません。"
		if msg != expected {
			t.Fatalf("Expected %v, got %v", expected, msg)
		}
	})

	t.Run("Without events, final check", func(t *testing.T) {
		isFinalCheck := true
		msg := buildMessage([]*github.Event{}, isFinalCheck)
		expected := "本日はGitHubに草が生えませんでした。"
		if msg != expected {
			t.Fatalf("Expected %v, got %v", expected, msg)
		}
	})
}

func TestGetGithubEvents(t *testing.T) {
	ctx := context.Background()
	client := createGithubClient(ctx, "dummy_token")
	username := "dummy_user"
	date := time.Now().Format("2006-01-02")

	events, err := getGithubEvents(ctx, client, username, date)
	if err == nil && len(events) > 0 {
		t.Fatalf("Expected no events for dummy user, got %d", len(events))
	}
}

func TestSendLineMessage(t *testing.T) {
	secret := "dummy_secret"
	token := "dummy_token"
	bot, err := createLineBotClient(secret, token)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	userID := "dummy_user_id"
	message := "test message"

	err = sendLineMessage(bot, userID, message)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
