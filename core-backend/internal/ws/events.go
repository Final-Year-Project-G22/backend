package ws

import "encoding/json"

const (
	EventPostCreated   = "post.created"
	EventThreadUpdated = "thread.updated"

	MsgTypeSubscribe   = "subscribe"
	MsgTypeUnsubscribe = "unsubscribe"
)

type ClientMessage struct {
	Type     string `json:"type"`
	ThreadID string `json:"threadId,omitempty"`
}

type ServerEvent struct {
	Type     string      `json:"type"`
	Version  int         `json:"version"`
	ThreadID string      `json:"threadId"`
	Data     interface{} `json:"data,omitempty"`
}

type PostCreatedData struct {
	PostID string `json:"postId"`
}

func NewPostCreatedEvent(threadID, postID string) *ServerEvent {
	return &ServerEvent{
		Type:     EventPostCreated,
		Version:  1,
		ThreadID: threadID,
		Data:     &PostCreatedData{PostID: postID},
	}
}

func NewThreadUpdatedEvent(threadID string) *ServerEvent {
	return &ServerEvent{
		Type:     EventThreadUpdated,
		Version:  1,
		ThreadID: threadID,
	}
}

func (e *ServerEvent) Marshal() ([]byte, error) {
	return json.Marshal(e)
}
