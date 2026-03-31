package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"chatly/internal/domain"

	"github.com/redis/go-redis/v9"
)

type PubSubManager struct {
	rdb *redis.Client
	hub *Hub
}

func NewPubSubManager(rdb *redis.Client, hub *Hub) *PubSubManager {
	return &PubSubManager{
		rdb: rdb,
		hub: hub,
	}
}

func (ps *PubSubManager) Publish(ctx context.Context, msg domain.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("chat:%s", msg.Room)
	return ps.rdb.Publish(ctx, channel, data).Err()
}

func (ps *PubSubManager) Subscribe(ctx context.Context) {
	pubsub := ps.rdb.PSubscribe(ctx, "chat:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var chatMsg domain.Message
		if err := json.Unmarshal([]byte(msg.Payload), &chatMsg); err != nil {
			log.Printf("pubsub unmarshal error: %v", err)
			continue
		}

		// Broadcast to local hub (only if we have clients for this room)
		ps.hub.BroadcastToRoom(chatMsg.Room, chatMsg)
	}
}
