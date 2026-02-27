package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	errs "gmb/internal/errors"
)

type BotAPIClient struct {
	Token      string
	HTTPClient *http.Client
}

type apiResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      struct {
		MessageID int `json:"message_id"`
	} `json:"result"`
}

func NewBotAPIClient(token string) *BotAPIClient {
	return &BotAPIClient{
		Token:      token,
		HTTPClient: &http.Client{Timeout: 45 * time.Second},
	}
}

func (c *BotAPIClient) SendVideoOrFallback(ctx context.Context, req SendRequest, video io.Reader, contentType string) (SendResult, error) {
	chatID := resolveTarget(req.ChatID, req.Username)
	if chatID == "" {
		return SendResult{}, errs.Wrap(errs.KindConfig, "resolve-chat-target", fmt.Errorf("no chat target configured"))
	}

	if video != nil {
		if result, err := c.sendVideo(ctx, chatID, req.Caption, req.ParseMode, video, contentType); err == nil {
			result.Mode = "video"
			return result, nil
		}
	}

	result, err := c.sendMessage(ctx, chatID, req.VideoURL, req.ParseMode)
	if err != nil {
		return SendResult{}, err
	}
	result.Mode = "link"
	return result, nil
}

func resolveTarget(chatID, username string) string {
	if strings.TrimSpace(chatID) != "" {
		return strings.TrimSpace(chatID)
	}
	u := strings.TrimSpace(username)
	if u == "" {
		return ""
	}
	if strings.HasPrefix(u, "@") {
		return u
	}
	return "@" + u
}

func (c *BotAPIClient) sendVideo(ctx context.Context, chatID, caption, parseMode string, video io.Reader, _ string) (SendResult, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	if err := mw.WriteField("chat_id", chatID); err != nil {
		return SendResult{}, errs.Wrap(errs.KindTransport, "send-video-chat-field", err)
	}
	if caption != "" {
		_ = mw.WriteField("caption", caption)
	}
	if parseMode != "" {
		_ = mw.WriteField("parse_mode", parseMode)
	}

	fw, err := mw.CreateFormFile("video", "good-morning.mp4")
	if err != nil {
		return SendResult{}, errs.Wrap(errs.KindTransport, "send-video-formfile", err)
	}
	if _, err := io.Copy(fw, video); err != nil {
		return SendResult{}, errs.Wrap(errs.KindTransport, "copy-video", err)
	}
	if err := mw.Close(); err != nil {
		return SendResult{}, errs.Wrap(errs.KindTransport, "close-multipart", err)
	}

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendVideo", c.Token)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return SendResult{}, errs.Wrap(errs.KindTransport, "build-send-video-request", err)
	}
	httpReq.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return SendResult{}, errs.Wrap(errs.KindTransient, "send-video-request", err)
	}
	defer resp.Body.Close()

	var payload apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return SendResult{}, errs.Wrap(errs.KindTransport, "decode-send-video-response", err)
	}
	if !payload.OK {
		return SendResult{}, errs.Wrap(errs.KindTransport, "send-video-api", fmt.Errorf(payload.Description))
	}

	return SendResult{MessageID: payload.Result.MessageID}, nil
}

func (c *BotAPIClient) sendMessage(ctx context.Context, chatID, text, parseMode string) (SendResult, error) {
	v := url.Values{}
	v.Set("chat_id", chatID)
	v.Set("text", text)
	if parseMode != "" {
		v.Set("parse_mode", parseMode)
	}

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.Token)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(v.Encode()))
	if err != nil {
		return SendResult{}, errs.Wrap(errs.KindTransport, "build-send-message-request", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return SendResult{}, errs.Wrap(errs.KindTransient, "send-message-request", err)
	}
	defer resp.Body.Close()

	var payload apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return SendResult{}, errs.Wrap(errs.KindTransport, "decode-send-message-response", err)
	}
	if !payload.OK {
		return SendResult{}, errs.Wrap(errs.KindTransport, "send-message-api", fmt.Errorf(payload.Description))
	}

	return SendResult{MessageID: payload.Result.MessageID}, nil
}
