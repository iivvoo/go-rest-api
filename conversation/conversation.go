package conversation

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	messagebird "github.com/messagebird/go-rest-api/v8"
)

const (
	// ConversationStatusActive is returned when the Conversation is active.
	// Only one active conversation can ever exist for a given contact.
	ConversationStatusActive ConversationStatus = "active"

	// ConversationStatusArchived is returned when the Conversation is
	// archived. When this is the case, a new Conversation is created when a
	// message is received from a contact.
	ConversationStatusArchived ConversationStatus = "archived"
)

type Conversation struct {
	ID                   string
	ContactID            string
	Contact              *Contact
	Channels             []*Channel
	Status               ConversationStatus
	CreatedDatetime      time.Time
	UpdatedDatetime      *time.Time
	LastReceivedDatetime *time.Time
	LastUsedChannelID    string
	lastUsedPlatformId   Platform
	Messages             *MessagesCount
}

type Channel struct {
	ID              string
	Name            string
	PlatformID      string
	Status          string
	CreatedDatetime *time.Time
	UpdatedDatetime *time.Time
}

type MessagesCount struct {
	HRef          string
	TotalCount    int
	LastMessageId string
}

// ConversationStatus indicates what state a Conversation is in.
type ConversationStatus string

type Platform string

type ConversationList struct {
	Offset     int
	Limit      int
	Count      int
	TotalCount int
	Items      []*Conversation
}

type ConversationByContactList struct {
	Offset     int
	Limit      int
	Count      int
	TotalCount int
	Items      []*string // array of conversation IDs
}

// StartRequest contains the request data for the Start endpoint.
type StartRequest struct {
	ChannelID string                 `json:"channelId"`
	Content   *MessageContent        `json:"content"`
	To        MessageRecipient       `json:"to"`
	Type      MessageType            `json:"type"`
	Source    map[string]interface{} `json:"source,omitempty"`
	ReportUrl string                 `json:"reportUrl,omitempty"`
	Tag       MessageTag             `json:"tag,omitempty"`
	TrackId   string                 `json:"trackId,omitempty"`
	EventType string                 `json:"eventType,omitempty"`
	TTL       string                 `json:"ttl,omitempty"`
}

// ReplyRequest contains the request data for the Reply endpoint.
type ReplyRequest struct {
	Type      MessageType            `json:"type"`
	Content   *MessageContent        `json:"content"`
	ChannelID string                 `json:"channelId,omitempty"`
	Fallback  *Fallback              `json:"fallback,omitempty"`
	Source    map[string]interface{} `json:"source,omitempty"`
	EventType string                 `json:"eventType,omitempty"`
	ReportUrl string                 `json:"reportUrl,omitempty"`
	Tag       MessageTag             `json:"tag,omitempty"`
	TrackId   string                 `json:"trackId,omitempty"`
	TTL       string                 `json:"ttl,omitempty"`
}

// UpdateRequest contains the request data for the Update endpoint.
type UpdateRequest struct {
	Status ConversationStatus `json:"status"`
}

// ListRequest retrieves all conversations sorted by the lastReceivedDatetime field
// so that all conversations with new messages appear first.
type ListRequest struct {
	PaginationRequest
	Ids    string
	Status *ConversationStatus
}

func (lr *ListRequest) GetParams() string {
	if lr == nil {
		return ""
	}

	query := url.Values{}

	query.Set("limit", strconv.Itoa(lr.Limit))
	query.Set("offset", strconv.Itoa(lr.Offset))

	if len(lr.Ids) > 0 {
		query.Set("ids", lr.Ids)
	}
	if lr.Status != nil {
		query.Set("status", string(*lr.Status))
	}

	return query.Encode()
}

type ListByContactRequest struct {
	PaginationRequest
	Id     string
	Status *ConversationStatus
}

func (lr *ListByContactRequest) GetParams() string {
	if lr == nil {
		return ""
	}

	query := url.Values{}

	query.Set("limit", strconv.Itoa(lr.Limit))
	query.Set("offset", strconv.Itoa(lr.Offset))

	if len(lr.Id) > 0 {
		query.Set("id", lr.Id)
	}
	if lr.Status != nil {
		query.Set("status", string(*lr.Status))
	}

	return query.Encode()
}

// List gets a collection of Conversations. Pagination can be set in options.
func List(c messagebird.ClientInterface, options *ListRequest) (*ConversationList, error) {
	convList := &ConversationList{}
	if err := request(c, convList, http.MethodGet, fmt.Sprintf("%s?%s", path, options.GetParams()), nil); err != nil {
		return nil, err
	}

	return convList, nil
}

// ListByContact fetches a collection of Conversations of a specific MessageBird contact ID.
func ListByContact(c messagebird.ClientInterface, contactId string, options *PaginationRequest) (*ConversationByContactList, error) {
	reqPath := fmt.Sprintf("%s/%s/%s?%s", path, contactPath, contactId, options.GetParams())

	conv := &ConversationByContactList{}
	if err := request(c, conv, http.MethodGet, reqPath, nil); err != nil {
		return nil, err
	}

	return conv, nil
}

// Read fetches a single Conversation based on its ID.
func Read(c messagebird.ClientInterface, id string) (*Conversation, error) {
	conv := &Conversation{}
	if err := request(c, conv, http.MethodGet, path+"/"+id, nil); err != nil {
		return nil, err
	}

	return conv, nil
}

// Start creates a conversation by sending an initial message. If an active
// conversation exists for the recipient, it is resumed.
func Start(c messagebird.ClientInterface, req *StartRequest) (*Conversation, error) {
	conv := &Conversation{}
	if err := request(c, conv, http.MethodPost, path+"/"+startConversationPath, req); err != nil {
		return nil, err
	}

	return conv, nil
}

// Reply Send a new message to an existing conversation. In case the conversation is archived, a new conversation is created.
func Reply(c messagebird.ClientInterface, conversationId string, req *ReplyRequest) (*Message, error) {
	uri := fmt.Sprintf("%s/%s/%s", path, conversationId, messagesPath)

	message := &Message{}
	if err := request(c, message, http.MethodPost, uri, req); err != nil {
		return nil, err
	}

	return message, nil
}

// Update changes the conversation's status, so this can be used to (un)archive
// conversations.
func Update(c messagebird.ClientInterface, id string, req *UpdateRequest) (*Conversation, error) {
	conv := &Conversation{}
	if err := request(c, conv, http.MethodPatch, path+"/"+id, req); err != nil {
		return nil, err
	}

	return conv, nil
}
