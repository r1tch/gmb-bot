package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
	AppCron                    string
	OneShot                    bool
	InstagramProfile           string
	DescriptionRegex           string
	InstagramScanLimit         int
	InstagramFetchSleepSeconds float64
	DownloadDir                string
	DownloadMaxPerRun          int
	DownloadDelaySeconds       int
	SentLogPath                string
	TelegramBotToken           string
	TelegramChatID             string
	TelegramParseMode          string
	LogLevel                   string
}

func Load() (Config, error) {
	cfg := Config{
		AppCron:                    getenv("APP_CRON", "0 6 * * *"),
		OneShot:                    getenvBool("ONE_SHOT", false),
		InstagramProfile:           getenv("IG_PROFILE", "gmbadass"),
		DescriptionRegex:           getenv("DESCRIPTION_REGEX", "(?i)(good ?morning|wake)"),
		InstagramScanLimit:         getenvInt("IG_SCAN_LIMIT", 10),
		InstagramFetchSleepSeconds: getenvFloat("IG_FETCH_SLEEP_SECONDS", 1.5),
		DownloadDir:                getenv("DOWNLOAD_DIR", "/data/video"),
		DownloadMaxPerRun:          getenvInt("DOWNLOAD_MAX_PER_RUN", 100),
		DownloadDelaySeconds:       getenvInt("DOWNLOAD_DELAY_SECONDS", 2),
		SentLogPath:                getenv("SENT_LOG_PATH", "/data/state/sent.log"),
		TelegramBotToken:           os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:             os.Getenv("TELEGRAM_CHAT_ID"),
		TelegramParseMode:          os.Getenv("TELEGRAM_PARSE_MODE"),
		LogLevel:                   getenv("LOG_LEVEL", "info"),
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
	if strings.TrimSpace(c.TelegramChatID) == "" {
		return errors.New("TELEGRAM_CHAT_ID is required")
	}
	if strings.TrimSpace(c.InstagramProfile) == "" {
		return errors.New("IG_PROFILE cannot be empty")
	}
	if _, err := regexp.Compile(c.DescriptionRegex); err != nil {
		return fmt.Errorf("DESCRIPTION_REGEX invalid: %w", err)
	}
	if c.InstagramScanLimit <= 0 {
		return errors.New("IG_SCAN_LIMIT must be > 0")
	}
	if c.InstagramFetchSleepSeconds < 0 {
		return errors.New("IG_FETCH_SLEEP_SECONDS must be >= 0")
	}
	if c.DownloadMaxPerRun <= 0 {
		return errors.New("DOWNLOAD_MAX_PER_RUN must be > 0")
	}
	if c.DownloadDelaySeconds < 0 {
		return errors.New("DOWNLOAD_DELAY_SECONDS must be >= 0")
	}
	if strings.TrimSpace(c.DownloadDir) == "" {
		return errors.New("DOWNLOAD_DIR cannot be empty")
	}
	if strings.TrimSpace(c.SentLogPath) == "" {
		return errors.New("SENT_LOG_PATH cannot be empty")
	}
	if c.AppCron != "0 6 * * *" && !strings.HasPrefix(c.AppCron, "@every ") {
		return fmt.Errorf("unsupported APP_CRON expression %q; use '0 6 * * *' or '@every <duration>'", c.AppCron)
	}
	return nil
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

func getenvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return n
}
