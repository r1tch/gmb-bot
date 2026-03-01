package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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

func (c *BotAPIClient) SendVideo(ctx context.Context, req SendRequest, video io.Reader, _ string) (SendResult, error) {
	chatID := strings.TrimSpace(req.ChatID)
	if chatID == "" {
		return SendResult{}, errs.Wrap(errs.KindConfig, "resolve-chat-target", fmt.Errorf("no chat target configured"))
	}
	if video == nil {
		return SendResult{}, errs.Wrap(errs.KindConfig, "send-video", fmt.Errorf("nil video stream"))
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	if err := mw.WriteField("chat_id", chatID); err != nil {
		return SendResult{}, errs.Wrap(errs.KindTransport, "send-video-chat-field", err)
	}
	if req.Caption != "" {
		_ = mw.WriteField("caption", req.Caption)
	}
	if req.ParseMode != "" {
		_ = mw.WriteField("parse_mode", req.ParseMode)
	}

	fw, err := mw.CreateFormFile("video", "video.mp4")
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
		return SendResult{}, errs.Wrap(errs.KindTransport, "send-video-api", errors.New(payload.Description))
	}

	return SendResult{MessageID: payload.Result.MessageID}, nil
}
