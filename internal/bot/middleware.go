package bot

import (
	"log"
	"runtime/debug"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SafeHandle wraps handler executions with panic recovery
func (b *Bot) SafeHandle(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC RECOVERED] %v\nStack: %s", r, debug.Stack())
		}
	}()
	fn()
}

// IsChatAdmin checks if the user is an administrator or owner of the chat
func (b *Bot) IsChatAdmin(chatID, userID int64) bool {
	if userID == b.cfg.OwnerID {
		return true
	}

	chatMember, err := b.api.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: userID,
		},
	})
	if err != nil {
		return false
	}

	return chatMember.Status == "administrator" || chatMember.Status == "creator"
}

// IsOwner checks if user is bot owner
func (b *Bot) IsOwner(userID int64) bool {
	return userID == b.cfg.OwnerID
}
