package models

import (
	"time"

	db "github.com/TF2Stadium/Helen/database"
)

// ChatMessage Represents a chat mesasge sent by a particular player
type ChatMessage struct {
	// Message ID
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"-"`

	// Because the frontend needs the unix timestamp for the message. Not stored in the DB
	Timestamp int64 `sql:"-" json:"timestamp"`

	// ID of the player who sent the message
	PlayerID uint `json:"-"`
	// Not in the DB, used by frontend to retrieve player information
	Player PlayerSummary `json:"player" sql:"-"`

	// Room to which the message was sent
	Room int `json:"room"`
	// The actual Message, limited to 120 characters
	Message string `json:"message" sql:"type:varchar(120)"`
	// True if the message has been deleted by a moderator
	Deleted bool `json:"-"`
}

// Return a new ChatMessage sent from specficied player
func NewChatMessage(message string, room int, player *Player) *ChatMessage {
	record := &ChatMessage{
		Timestamp: time.Now().Unix(),

		PlayerID: player.ID,
		Player:   DecoratePlayerSummary(player),

		Room:    room,
		Message: message,
	}

	return record
}

// Return a list of ChatMessages spoken in room
func GetRoomMessages(room int) ([]*ChatMessage, error) {
	var messages []*ChatMessage

	err := db.DB.Table("chat_messages").Where("room = ?", room).Order("created_at").Find(&messages).Error

	return messages, err
}

// Return all messages sent by player to room
func GetPlayerMessages(player *Player) ([]*ChatMessage, error) {
	var messages []*ChatMessage

	err := db.DB.Table("chat_messages").Where("player_id = ?", player.ID).Order("room, created_at").Find(&messages).Error

	return messages, err

}

// Get a list of last 20 messages sent to room, used by frontend for displaying the chat history/scrollback
func GetScrollback(room int) ([]*ChatMessage, error) {
	var messages []*ChatMessage

	err := db.DB.Table("chat_messages").Where("room = ?", room).Order("id desc").Limit(20).Find(&messages).Error

	for _, message := range messages {
		var player Player
		db.DB.First(&player, message.PlayerID)
		message.Player = DecoratePlayerSummary(&player)
		message.Timestamp = message.CreatedAt.Unix()
	}
	return messages, err
}
