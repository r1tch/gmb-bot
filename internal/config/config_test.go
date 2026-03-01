package config

import "testing"

func TestValidateRequiresToken(t *testing.T) {
	cfg := Config{TelegramChatID: "123", InstagramProfile: "gmbadass", DescriptionRegex: "(?i)good", AppCron: "0 6 * * *", InstagramScanLimit: 10, DownloadMaxPerRun: 1, DownloadDir: "/data/video", SentLogPath: "/data/state/sent.log"}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestValidateRequiresChatID(t *testing.T) {
	cfg := Config{TelegramBotToken: "x", InstagramProfile: "gmbadass", DescriptionRegex: "(?i)good", AppCron: "0 6 * * *", InstagramScanLimit: 10, DownloadMaxPerRun: 1, DownloadDir: "/data/video", SentLogPath: "/data/state/sent.log"}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestValidateRegex(t *testing.T) {
	cfg := Config{TelegramBotToken: "x", TelegramChatID: "1", InstagramProfile: "gmbadass", DescriptionRegex: "(", AppCron: "0 6 * * *", InstagramScanLimit: 10, DownloadMaxPerRun: 1, DownloadDir: "/data/video", SentLogPath: "/data/state/sent.log"}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected regex validation error")
	}
}

func TestValidateFetchSleepNonNegative(t *testing.T) {
	cfg := Config{
		TelegramBotToken:           "x",
		TelegramChatID:             "1",
		InstagramProfile:           "gmbadass",
		DescriptionRegex:           "(?i)good",
		AppCron:                    "0 6 * * *",
		InstagramScanLimit:         10,
		InstagramFetchSleepSeconds: -1,
		DownloadMaxPerRun:          1,
		DownloadDir:                "/data/video",
		SentLogPath:                "/data/state/sent.log",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected fetch sleep validation error")
	}
}
