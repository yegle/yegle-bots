package bots

import (
	"errors"
)

// ErrIgnoredItem is returned when the story should be ignored.
var ErrIgnoredItem = errors.New("item ignored")

// SendMessageRequest is a struct that maps to a sendMessage request.
type SendMessageRequest struct {
	ChatID      string               `json:"chat_id"`
	Text        string               `json:"text"`
	ParseMode   string               `json:"parse_mode,omitempty"`
	ReplyMarkup InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

// InlineKeyboardMarkup type.
type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard,omitempty"`
}

// InlineKeyboardButton type.
type InlineKeyboardButton struct {
	Text string `json:"text,omitempty"`
	URL  string `json:"url,omitempty"`
}

// SendMessageResponse is the response from sendMessage request.
type SendMessageResponse struct {
	OK     bool   `json:"ok"`
	Result Result `json:"result"`
}

// Result is a submessage in SendMessageResponse. We only care the MessageID for now.
type Result struct {
	MessageID int64 `json:"message_id"`
}

// EditMessageTextRequest is the request to editMessageText method.
type EditMessageTextRequest struct {
	ChatID      string               `json:"chat_id"`
	MessageID   int64                `json:"message_id"`
	Text        string               `json:"text"`
	ParseMode   string               `json:"parse_mode,omitempty"`
	ReplyMarkup InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

// DeleteMessageRequest is the request to deleteMessage method.
type DeleteMessageRequest struct {
	ChatID    string `json:"chat_id"`
	MessageID int64  `json:"message_id"`
}

// DeleteMessageResponse is the response to deleteMessage method.
type DeleteMessageResponse struct {
	OK          bool   `json:"ok"`
	ErrorCode   int64  `json:"error_code"`
	Description string `json:"description"`
}
