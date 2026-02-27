package config

import (
	"testing"
)

func TestValidateRequiresToken(t *testing.T) {
	cfg := Config{TelegramChatID: "123"}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestValidateRequiresRecipient(t *testing.T) {
	cfg := Config{TelegramBotToken: "x"}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestRecipientPrefersChatID(t *testing.T) {
	cfg := Config{TelegramChatID: "-100123", TelegramUsername: "abc"}
	if got := cfg.Recipient(); got != "-100123" {
		t.Fatalf("expected chat ID, got %q", got)
	}
}
