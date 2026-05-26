package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const ItemTypeThread = "thread"

type NotificationEventPublisher interface {
	PublishThreadReply(ctx context.Context, recipientAccountID uuid.UUID, threadID, postID uuid.UUID, actorAccountID uuid.UUID) error
	PublishThreadSolution(ctx context.Context, recipientAccountID uuid.UUID, threadID, postID uuid.UUID, actorAccountID uuid.UUID) error
	PublishThreadMention(ctx context.Context, mentionedAccountID uuid.UUID, threadID, postID uuid.UUID, actorAccountID uuid.UUID) error
}

type notificationEventPublisherImpl struct {
	db *core.Database
}

type threadInfo struct {
	Title string
	Slug  string
}

func NewNotificationEventPublisher(db *core.Database) NotificationEventPublisher {
	return &notificationEventPublisherImpl{db: db}
}

func (p *notificationEventPublisherImpl) PublishThreadReply(ctx context.Context, recipientAccountID uuid.UUID, threadID, postID uuid.UUID, actorAccountID uuid.UUID) error {
	thread, err := p.resolveThread(ctx, threadID)
	if err != nil {
		return err
	}

	actorName := p.resolveActorName(ctx, actorAccountID)
	key := p.idempotencyKey("thread.reply", postID)
	db := p.getDB(ctx)

	return db.Create(&outboxRow{
		EventType:      "thread.reply",
		SchemaVersion:  "1.0.0",
		SourceModule:   "community",
		AccountID:      recipientAccountID,
		IdempotencyKey: key,
		Payload: datatypes.JSONMap{
			"schemaVersion":    "1.0.0",
			"eventType":        "thread.reply",
			"occurredAt":       time.Now().UTC().Format(time.RFC3339Nano),
			"sourceModule":     "community",
			"accountId":        recipientAccountID.String(),
			"notificationType": "community_reply",
			"channelPolicy":    "all_enabled",
			"variables": map[string]interface{}{
				"authorName":  actorName,
				"threadTitle": thread.Title,
				"threadSlug":  thread.Slug,
			},
			"metadata": map[string]interface{}{
				"idempotencyKey": key,
				"itemType":       ItemTypeThread,
				"itemId":         threadID.String(),
				"actorAccountId": actorAccountID.String(),
				"threadId":       threadID.String(),
				"postId":         postID.String(),
			},
		},
	}).Error
}

func (p *notificationEventPublisherImpl) PublishThreadSolution(ctx context.Context, recipientAccountID uuid.UUID, threadID, postID uuid.UUID, actorAccountID uuid.UUID) error {
	thread, err := p.resolveThread(ctx, threadID)
	if err != nil {
		return err
	}

	key := p.idempotencyKey("thread.solution", postID)
	db := p.getDB(ctx)

	return db.Create(&outboxRow{
		EventType:      "thread.solution",
		SchemaVersion:  "1.0.0",
		SourceModule:   "community",
		AccountID:      recipientAccountID,
		IdempotencyKey: key,
		Payload: datatypes.JSONMap{
			"schemaVersion":    "1.0.0",
			"eventType":        "thread.solution",
			"occurredAt":       time.Now().UTC().Format(time.RFC3339Nano),
			"sourceModule":     "community",
			"accountId":        recipientAccountID.String(),
			"notificationType": "community_solution",
			"channelPolicy":    "all_enabled",
			"variables": map[string]interface{}{
				"threadTitle": thread.Title,
				"threadSlug":  thread.Slug,
			},
			"metadata": map[string]interface{}{
				"idempotencyKey": key,
				"itemType":       ItemTypeThread,
				"itemId":         threadID.String(),
				"actorAccountId": actorAccountID.String(),
				"threadId":       threadID.String(),
				"postId":         postID.String(),
			},
		},
	}).Error
}

func (p *notificationEventPublisherImpl) PublishThreadMention(ctx context.Context, mentionedAccountID uuid.UUID, threadID, postID uuid.UUID, actorAccountID uuid.UUID) error {
	thread, err := p.resolveThread(ctx, threadID)
	if err != nil {
		return err
	}

	actorName := p.resolveActorName(ctx, actorAccountID)
	key := p.idempotencyKey("thread.mention", postID)
	db := p.getDB(ctx)

	return db.Create(&outboxRow{
		EventType:      "thread.mention",
		SchemaVersion:  "1.0.0",
		SourceModule:   "community",
		AccountID:      mentionedAccountID,
		IdempotencyKey: key,
		Payload: datatypes.JSONMap{
			"schemaVersion":    "1.0.0",
			"eventType":        "thread.mention",
			"occurredAt":       time.Now().UTC().Format(time.RFC3339Nano),
			"sourceModule":     "community",
			"accountId":        mentionedAccountID.String(),
			"notificationType": "community_mention",
			"channelPolicy":    "all_enabled",
			"variables": map[string]interface{}{
				"authorName":  actorName,
				"threadTitle": thread.Title,
				"threadSlug":  thread.Slug,
			},
			"metadata": map[string]interface{}{
				"idempotencyKey": key,
				"itemType":       ItemTypeThread,
				"itemId":         threadID.String(),
				"actorAccountId": actorAccountID.String(),
				"threadId":       threadID.String(),
				"postId":         postID.String(),
			},
		},
	}).Error
}

func (p *notificationEventPublisherImpl) resolveThread(ctx context.Context, threadID uuid.UUID) (*threadInfo, error) {
	var info threadInfo
	db := p.getDB(ctx)
	if err := db.Table("discussion_threads").
		Select("title, slug").
		Where("id = ?", threadID).
		Limit(1).
		Scan(&info).Error; err != nil {
		return nil, err
	}
	return &info, nil
}

func (p *notificationEventPublisherImpl) resolveActorName(ctx context.Context, actorAccountID uuid.UUID) string {
	db := p.getDB(ctx)

	var username *string
	if err := db.Table("accounts").
		Select("username").
		Where("id = ?", actorAccountID).
		Limit(1).
		Pluck("username", &username).Error; err != nil || username == nil || *username == "" {

		var firstName, lastName string
		row := db.Table("users u").
			Joins("JOIN accounts a ON a.user_id = u.id").
			Select("u.first_name, u.last_name").
			Where("a.id = ?", actorAccountID).
			Limit(1).
			Row()
		if err := row.Scan(&firstName, &lastName); err != nil {
			return "User " + actorAccountID.String()[:6]
		}
		if firstName != "" || lastName != "" {
			return firstName + " " + lastName
		}
		return "User " + actorAccountID.String()[:6]
	}

	return *username
}

func (p *notificationEventPublisherImpl) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return p.db.WithContext(ctx)
}

func (p *notificationEventPublisherImpl) idempotencyKey(eventType string, postID uuid.UUID) string {
	raw := fmt.Sprintf("%s:%s", eventType, postID.String())
	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", hash)
}

type outboxRow struct {
	EventType      string            `gorm:"column:event_type"`
	SchemaVersion  string            `gorm:"column:schema_version"`
	SourceModule   string            `gorm:"column:source_module"`
	AccountID      uuid.UUID         `gorm:"column:account_id"`
	IdempotencyKey string            `gorm:"column:idempotency_key"`
	Payload        datatypes.JSONMap `gorm:"column:payload;type:jsonb"`
}

func (outboxRow) TableName() string {
	return "notification_outbox"
}
