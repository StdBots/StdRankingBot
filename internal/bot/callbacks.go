package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/StdBots/StdRankingBot/internal/credit"
	"github.com/StdBots/StdRankingBot/internal/database"
	"github.com/StdBots/StdRankingBot/internal/graphics"
)

// HandleCallback handles all inline keyboard button taps
func (b *Bot) HandleCallback(cb *tgbotapi.CallbackQuery) {
	data := cb.Data

	switch {
	case strings.HasPrefix(data, "rank:"):
		period := strings.TrimPrefix(data, "rank:")
		if period == "me" {
			b.handleCallbackMe(cb)
		} else if period == "global" {
			b.handleCallbackGlobal(cb)
		} else {
			b.handleCallbackPeriod(cb, period)
		}

	case data == "sett:toggle_alerts":
		b.handleCallbackToggleAlerts(cb)

	case data == "help:main":
		b.handleCallbackHelp(cb)

	default:
		callback := tgbotapi.NewCallback(cb.ID, "")
		_, _ = b.api.Request(callback)
	}
}

// handleCallbackPeriod updates the photo banner and caption to the chosen period
func (b *Bot) handleCallbackPeriod(cb *tgbotapi.CallbackQuery, period string) {
	if cb.Message == nil {
		return
	}

	chatID := cb.Message.Chat.ID
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	members, err := b.db.GetLeaderboard(ctx, chatID, period, 10)
	if err != nil || len(members) == 0 {
		callback := tgbotapi.NewCallback(cb.ID, fmt.Sprintf("No stats recorded yet for %s.", period))
		_, _ = b.api.Request(callback)
		return
	}

	var rankedMembers []graphics.RankedMember
	var totalMsgs int64
	for i, m := range members {
		var cnt int64
		switch period {
		case "weekly":
			cnt = m.WeeklyCount
		case "monthly":
			cnt = m.MonthlyCount
		case "alltime":
			cnt = m.AllTimeCount
		default:
			cnt = m.DailyCount
		}
		totalMsgs += cnt

		name := m.FirstName
		if name == "" {
			name = m.Username
		}
		rankedMembers = append(rankedMembers, graphics.RankedMember{
			Rank:     i + 1,
			Name:     name,
			Username: m.Username,
			Level:    m.Level,
			Messages: cnt,
			XP:       m.XP,
		})
	}

	cardData := graphics.LeaderboardCardData{
		ChatTitle:   cb.Message.Chat.Title,
		Period:      period,
		Members:     rankedMembers,
		TotalActive: int64(len(members)),
		TotalMsgs:   totalMsgs,
		GeneratedAt: time.Now(),
	}

	imgBytes, err := graphics.RenderLeaderboard(cardData)
	if err != nil {
		callback := tgbotapi.NewCallback(cb.ID, "Failed to render banner.")
		_, _ = b.api.Request(callback)
		return
	}

	caption := b.buildLeaderboardCaption(cb.Message.Chat.Title, period, rankedMembers)
	keyboard := b.getLeaderboardKeyboard(period)

	// Update the message photo via EditMessageMedia
	media := tgbotapi.NewInputMediaPhoto(tgbotapi.FileBytes{
		Name:  fmt.Sprintf("leaderboard_%s.png", period),
		Bytes: imgBytes,
	})
	media.Caption = credit.WatermarkMessage(caption)
	media.ParseMode = tgbotapi.ModeHTML

	editMedia := tgbotapi.EditMessageMediaConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      chatID,
			MessageID:   cb.Message.MessageID,
			ReplyMarkup: &keyboard,
		},
		Media: media,
	}

	_, _ = b.api.Request(editMedia)
	callback := tgbotapi.NewCallback(cb.ID, fmt.Sprintf("Switched to %s leaderboard!", period))
	_, _ = b.api.Request(callback)
}

// handleCallbackMe shows caller's stats
func (b *Bot) handleCallbackMe(cb *tgbotapi.CallbackQuery) {
	if cb.Message == nil {
		return
	}

	chatID := cb.Message.Chat.ID
	userID := cb.From.ID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rank, stats, total, err := b.db.GetMemberRank(ctx, chatID, userID, "alltime")
	if err != nil || stats == nil {
		callback := tgbotapi.NewCallback(cb.ID, "You haven't sent any messages in this chat yet!")
		_, _ = b.api.Request(callback)
		return
	}

	info := fmt.Sprintf("Rank #%d of %d | Lv.%d (%s) | %d msgs | %d XP",
		rank, total, stats.Level, graphics.GetUserTitle(stats.Level), stats.AllTimeCount, stats.XP,
	)
	callback := tgbotapi.NewCallbackWithAlert(cb.ID, info)
	_, _ = b.api.Request(callback)
}

// handleCallbackGlobal sends global top users
func (b *Bot) handleCallbackGlobal(cb *tgbotapi.CallbackQuery) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users, err := b.db.GetGlobalLeaderboard(ctx, 5)
	if err != nil || len(users) == 0 {
		callback := tgbotapi.NewCallback(cb.ID, "Global rankings are initializing...")
		_, _ = b.api.Request(callback)
		return
	}

	var sb strings.Builder
	sb.WriteString("🌐 GLOBAL TOP 5:\n")
	for i, u := range users {
		sb.WriteString(fmt.Sprintf("#%d %s - %d msgs (Lv.%d)\n", i+1, u.FirstName, u.TotalMessages, u.GlobalLevel))
	}

	callback := tgbotapi.NewCallbackWithAlert(cb.ID, sb.String())
	_, _ = b.api.Request(callback)
}

// handleCallbackToggleAlerts allows admins to toggle level up alerts
func (b *Bot) handleCallbackToggleAlerts(cb *tgbotapi.CallbackQuery) {
	if cb.Message == nil {
		return
	}

	chatID := cb.Message.Chat.ID
	userID := cb.From.ID

	if !b.IsChatAdmin(chatID, userID) {
		callback := tgbotapi.NewCallback(cb.ID, "⛔ Administrator permission required.")
		_, _ = b.api.Request(callback)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	group, err := b.db.EnsureGroup(ctx, chatID, cb.Message.Chat.Title, cb.Message.Chat.UserName, cb.Message.Chat.Type)
	if err != nil {
		callback := tgbotapi.NewCallback(cb.ID, "Failed to load group settings.")
		_, _ = b.api.Request(callback)
		return
	}

	group.LevelUpAlerts = !group.LevelUpAlerts
	_ = b.db.UpdateGroupSettings(ctx, group)

	alertStatus := "ENABLED 🔔"
	btnAlert := "🔔 Alerts: ON"
	if !group.LevelUpAlerts {
		alertStatus = "DISABLED 🔕"
		btnAlert = "🔕 Alerts: OFF"
	}

	text := fmt.Sprintf(
		"⚙️ <b>RANKING CONFIGURATION • %s</b>\n\n"+
			"• <b>Level-Up Alerts:</b> %s\n"+
			"• <b>Anti-Spam Cooldown:</b> <code>%d seconds</code>\n"+
			"• <b>Minimum Message Length:</b> <code>%d chars</code>\n"+
			"• <b>Total Messages Counted:</b> <code>%d</code>\n\n"+
			"Toggle settings using the buttons below:",
		cb.Message.Chat.Title, alertStatus, group.CooldownSeconds, group.MinMsgLength, group.TotalMessages,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnAlert, "sett:toggle_alerts"),
		),
	)

	edit := tgbotapi.NewEditMessageText(chatID, cb.Message.MessageID, credit.WatermarkMessage(text))
	edit.ParseMode = tgbotapi.ModeHTML
	edit.ReplyMarkup = &keyboard
	_, _ = b.api.Request(edit)

	callback := tgbotapi.NewCallback(cb.ID, fmt.Sprintf("Level-up alerts %s!", alertStatus))
	_, _ = b.api.Request(callback)
}

// handleCallbackHelp displays help
func (b *Bot) handleCallbackHelp(cb *tgbotapi.CallbackQuery) {
	helpText := fmt.Sprintf(
		"📖 <b>HOW TO USE @%s:</b>\n\n"+
			"1. Add bot to your Telegram group.\n"+
			"2. Make bot an administrator.\n"+
			"3. The bot tracks all messages automatically.\n"+
			"4. Use /top or /ranking to see HD banners!\n"+
			"5. Use /me to see your gamer profile card!\n\n"+
			"%s",
		b.api.Self.UserName, credit.GetFooter(),
	)

	callback := tgbotapi.NewCallbackWithAlert(cb.ID, credit.StripWatermark(helpText))
	_, _ = b.api.Request(callback)
}
