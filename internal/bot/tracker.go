package bot

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/StdBots/StdRankingBot/internal/graphics"
)

// TrackGroupMessage intercepts group messages, verifies anti-farming rules, and distributes XP
func (b *Bot) TrackGroupMessage(msg *tgbotapi.Message) {
	if msg.From == nil || msg.From.IsBot {
		return
	}

	text := msg.Text
	if text == "" && msg.Caption != "" {
		text = msg.Caption
	}

	// Media without caption counts as 1 character placeholder
	if text == "" {
		if msg.Photo != nil || msg.Document != nil || msg.Video != nil || msg.Voice != nil || msg.Sticker != nil {
			text = "[media]"
		} else {
			return
		}
	}

	chatID := msg.Chat.ID
	userID := msg.From.ID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get group settings
	group, err := b.db.EnsureGroup(ctx, chatID, msg.Chat.Title, msg.Chat.UserName, msg.Chat.Type)
	if err != nil || !group.Active {
		return
	}

	cooldown := group.CooldownSeconds
	if cooldown <= 0 {
		cooldown = b.cfg.AntiSpamCooldown
	}
	minLen := group.MinMsgLength
	if minLen <= 0 {
		minLen = b.cfg.MinMsgLength
	}

	// 1. Anti-farming check
	eligible, _ := b.cache.CheckAndRecord(chatID, userID, text, cooldown, minLen)
	if !eligible {
		return
	}

	// 2. XP Distribution
	// Base XP: 15
	// Bonus: 1 XP per 10 characters (capped at +15)
	xpGain := b.cfg.BaseXpPerMsg
	bonus := len(text) / 10
	if bonus > 15 {
		bonus = 15
	}
	xpGain += bonus

	username := msg.From.UserName
	firstName := msg.From.FirstName
	if firstName == "" {
		firstName = "Fighter"
	}

	stats, leveledUp, err := b.db.RecordMessage(ctx, chatID, userID, username, firstName, text, xpGain)
	if err != nil {
		return
	}

	// 3. Level-up Celebration Alert
	if leveledUp && group.LevelUpAlerts {
		go func(lvl int, xp int64) {
			title := graphics.GetUserTitle(lvl)
			userMention := fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>", userID, firstName)

			alertText := fmt.Sprintf(
				"🎉 <b>LEVEL UP!</b>\n\n"+
					"Fighter %s just ascended to <b>LEVEL %d</b>! 🏆\n"+
					"⚔️ Title: <b>%s</b>\n"+
					"⚡ Total XP: <code>%d XP</code>\n\n"+
					"<i>Keep chatting to dominate the leaderboard!</i>",
				userMention, lvl, title, xp,
			)

			replyMsg := tgbotapi.NewMessage(chatID, alertText)
			replyMsg.ParseMode = tgbotapi.ModeHTML
			replyMsg.ReplyToMessageID = msg.MessageID
			_, _ = b.api.Send(replyMsg)
		}(stats.Level, stats.XP)
	}
}
