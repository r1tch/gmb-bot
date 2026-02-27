package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppCron            string
	OneShot            bool
	TikTokProfile      string
	TikTokLookback     int
	TikTokStorageState string
	TikTokSessionID    string
	MatchSubstring     string
	StateFilePath      string
	TelegramBotToken   string
	TelegramChatID     string
	TelegramUsername   string
	TelegramParseMode  string
	LogLevel           string
	PlaywrightHeadless bool
	SendMode           string
}

func Load() (Config, error) {
	cfg := Config{
		AppCron:            getenv("APP_CRON", "0 * * * *"),
		OneShot:            getenvBool("ONE_SHOT", false),
		TikTokProfile:      getenv("TIKTOK_PROFILE", "gmbadass"),
		TikTokLookback:     getenvInt("TIKTOK_LOOKBACK_LIMIT", 20),
		TikTokStorageState: os.Getenv("TIKTOK_STORAGE_STATE_PATH"),
		TikTokSessionID:    os.Getenv("TIKTOK_SESSIONID"),
		MatchSubstring:     strings.ToLower(getenv("MATCH_SUBSTRING", "good morning")),
		StateFilePath:      getenv("STATE_FILE_PATH", "/data/state.json"),
		TelegramBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:     os.Getenv("TELEGRAM_CHAT_ID"),
		TelegramUsername:   os.Getenv("TELEGRAM_USERNAME"),
		TelegramParseMode:  os.Getenv("TELEGRAM_PARSE_MODE"),
		LogLevel:           getenv("LOG_LEVEL", "info"),
		PlaywrightHeadless: getenvBool("PLAYWRIGHT_HEADLESS", true),
		SendMode:           getenv("SEND_MODE", "video_with_link_fallback"),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.TelegramBotToken) == "" {
		return errors.New("TELEGRAM_BOT_TOKEN is required")
	}
	if strings.TrimSpace(c.TelegramChatID) == "" && strings.TrimSpace(c.TelegramUsername) == "" {
		return errors.New("one of TELEGRAM_CHAT_ID or TELEGRAM_USERNAME is required")
	}
	if c.TikTokLookback <= 0 {
		return errors.New("TIKTOK_LOOKBACK_LIMIT must be > 0")
	}
	if strings.TrimSpace(c.MatchSubstring) == "" {
		return errors.New("MATCH_SUBSTRING cannot be empty")
	}
	if c.SendMode != "video_with_link_fallback" && c.SendMode != "link_only" && c.SendMode != "video_only" {
		return fmt.Errorf("SEND_MODE unsupported: %s", c.SendMode)
	}
	return nil
}

func (c Config) Recipient() string {
	if strings.TrimSpace(c.TelegramChatID) != "" {
		return c.TelegramChatID
	}
	return c.TelegramUsername
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getenvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
