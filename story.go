package bots

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// Hot is the sign for a hot story, either because it has high score or it has
// large number of discussions.
const Hot = "🔥"

// Story is a struct represents an item stored in datastore.
// Part of the fields will be saved to datastore.
type Story struct {
	ID                  int64     `json:"id"`
	URL                 string    `json:"url"`
	Title               string    `json:"title"`
	Descendants         int64     `json:"descendants"`
	Score               int64     `json:"score"`
	MessageID           int64     `json:"-"`
	LastSave            time.Time `json:"-"`
	Type                string    `json:"type"`
	missingFieldsLoaded bool
}

// NewFromDatastore create a Story from datastore.
func NewFromDatastore(ctx context.Context, id int64) (Story, error) {
	var story Story
	if err := datastore.Get(ctx, GetKey(ctx, id), &story); err != nil {
		return story, errors.WithStack(err)
	}
	return story, nil
}

// Load implements the PropertyLoadSaver interface.
func (s *Story) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(s, ps)
}

// Save implements the PropertyLoadSaver interface.
func (s *Story) Save() ([]datastore.Property, error) {
	return []datastore.Property{
		{
			Name:  "MessageID",
			Value: s.MessageID,
		},
		{
			Name:  "ID",
			Value: s.ID,
		},
		{
			Name:  "LastSave",
			Value: time.Now(),
		},
	}, nil
}

// FillMissingFields is used to fill the missing story data from HN API.
func (s *Story) FillMissingFields(ctx context.Context) error {
	resp, err := myHTTPClient(ctx).Get(ItemURL(s.ID))
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(s)
	if err != nil {
		return errors.WithStack(err)
	}
	s.missingFieldsLoaded = true
	return nil
}

// ShouldIgnore is a filter for story.
func (s *Story) ShouldIgnore() bool {
	return s.Type != "story" ||
		s.Score < ScoreThreshold ||
		s.Descendants < NumCommentsThreshold ||
		s.URL == ""
}

// ToSendMessageRequest will return a new SendMessageRequest object
func (s *Story) ToSendMessageRequest() SendMessageRequest {
	return SendMessageRequest{
		ChatID:      DefaultChatID,
		Text:        fmt.Sprintf("<b>%s</b>  %s", s.Title, s.URL),
		ParseMode:   "HTML",
		ReplyMarkup: s.GetReplyMarkup(),
	}
}

// ToEditMessageTextRequest will return a new EditMessageTextRequest object
func (s *Story) ToEditMessageTextRequest() EditMessageTextRequest {
	return EditMessageTextRequest{
		ChatID:      DefaultChatID,
		MessageID:   s.MessageID,
		Text:        fmt.Sprintf("<b>%s</b>  %s", s.Title, s.URL),
		ParseMode:   "HTML",
		ReplyMarkup: s.GetReplyMarkup(),
	}
}

// GetReplyMarkup will return the markup for the story.
func (s *Story) GetReplyMarkup() InlineKeyboardMarkup {
	var scoreSuffix, commentSuffix string
	if s.Score > 100 {
		scoreSuffix = " " + Hot
	}
	if s.Descendants > 100 {
		commentSuffix = " " + Hot
	}
	return InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{
					Text: fmt.Sprintf("Score: %d+%s", s.Score, scoreSuffix),
					URL:  s.URL,
				},
				{
					Text: fmt.Sprintf("Comments: %d+%s", s.Descendants, commentSuffix),
					URL:  NewsURL(s.ID),
				},
			},
		},
	}
}

// ToDeleteMessageRequest returns a DeleteMessageRequest.
func (s *Story) ToDeleteMessageRequest() DeleteMessageRequest {
	return DeleteMessageRequest{
		ChatID:    DefaultChatID,
		MessageID: s.MessageID,
	}
}

// EditMessage send a request to edit a message.
func (s *Story) EditMessage(ctx context.Context) error {
	if !s.missingFieldsLoaded {
		if err := s.FillMissingFields(ctx); err != nil {
			return errors.WithStack(err)
		}
	}
	if s.ShouldIgnore() {
		return errors.WithStack(ErrIgnoredItem)
	}

	req := s.ToEditMessageTextRequest()
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return errors.WithStack(err)
	}

	resp, err := myHTTPClient(ctx).Post(TelegramAPI("editMessageText"), "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)
	return nil
}

// InDatastore checks if the story is already in datastore.
func (s *Story) InDatastore(ctx context.Context) bool {
	log.Infof(ctx, "calling InDatastore")
	key := GetKey(ctx, s.ID)
	q := datastore.NewQuery("Story").Filter("__key__ =", key).KeysOnly()
	keys, _ := q.GetAll(ctx, nil)
	return len(keys) != 0
}

// SendMessage send a request to send a new message.
func (s *Story) SendMessage(ctx context.Context) error {
	if !s.missingFieldsLoaded {
		if err := s.FillMissingFields(ctx); err != nil {
			return errors.WithStack(err)
		}
	}

	if s.ShouldIgnore() {
		return ErrIgnoredItem
	} else if s.InDatastore(ctx) {
		return errors.WithStack(fmt.Errorf("story already posted: %#v", s))
	}
	req := s.ToSendMessageRequest()
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return errors.WithStack(err)
	}

	resp, err := myHTTPClient(ctx).Post(TelegramAPI("sendMessage"), "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()

	var response SendMessageResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return errors.WithStack(err)
	}
	s.MessageID = response.Result.MessageID
	return nil
}

// DeleteMessage delete a message from telegram Channel and from channel.
func (s *Story) DeleteMessage(ctx context.Context) error {
	req := s.ToDeleteMessageRequest()
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return errors.WithStack(err)
	}

	resp, err := myHTTPClient(ctx).Post(TelegramAPI("deleteMessage"), "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()

	var response DeleteMessageResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return errors.WithStack(err)
	}

	if !response.OK {
		if !response.ShouldIgnoreError() {
			return errors.WithStack(fmt.Errorf("%#v", response))
		}
		log.Warningf(ctx, "ignoring %#v", response)
	}

	key := GetKey(ctx, s.ID)
	if err := datastore.Delete(ctx, key); err != nil {
		return errors.WithStack(err)
	}
	log.Infof(ctx, "%d (messageID: %d) deleted", s.ID, s.MessageID)
	return nil
}
