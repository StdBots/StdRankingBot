package graphics

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"math"
)

// ProfileData holds information needed to render the personal profile card
type ProfileData struct {
	Name           string
	Username       string
	ChatTitle      string
	Rank           int
	TotalMembers   int
	Level          int
	DailyMsgs      int64
	WeeklyMsgs     int64
	MonthlyMsgs    int64
	AllTimeMsgs    int64
	CurrentXP      int64
	LevelBaseXP    int64
	NextLevelXP    int64
	ProgressRatio  float64
}

// GetUserTitle calculates a dynamic gamer title based on level
func GetUserTitle(level int) string {
	switch {
	case level >= 100:
		return "👑 OMNIPOTENT GOD"
	case level >= 75:
		return "🔱 CELESTIAL WARLORD"
	case level >= 50:
		return "⚡ GRANDMASTER FIGHTER"
	case level >= 35:
		return "⚔️ MYTHIC GLADIATOR"
	case level >= 25:
		return "🛡️ ELITE VETERAN"
	case level >= 15:
		return "🔥 BATTLE KNIGHT"
	case level >= 10:
		return "🎯 WARRIOR"
	case level >= 5:
		return "🗡️ APPRENTICE"
	default:
		return "🌱 ROOKIE FIGHTER"
	}
}

// RenderProfileCard renders an 840x500 gamer profile card PNG
func RenderProfileCard(data ProfileData) ([]byte, error) {
	const width = 840
	const height = 500

	c := NewCanvas(width, height)

	// Cyber tech gradient background
	c.FillVerticalGradient(
		color.RGBA{R: 15, G: 23, B: 42, A: 255}, // #0f172a
		color.RGBA{R: 2, G: 6, B: 23, A: 255},   // #020617
	)
	c.DrawGridPattern(35, color.RGBA{R: 56, G: 189, B: 248, A: 10})

	// Outer card glowing border
	c.FillRoundedRect(20, 20, width-40, height-40, 16, color.RGBA{R: 30, G: 41, B: 59, A: 180})

	// Header
	headerBg := color.RGBA{R: 14, G: 116, B: 144, A: 220} // Cyan/Teal
	c.FillRoundedRect(40, 36, 170, 26, 6, headerBg)
	c.DrawText("FIGHTER PROFILE", 52, 42, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 1)

	if data.ChatTitle != "" {
		chatTitle := TruncateString(data.ChatTitle, 40)
		c.DrawText(chatTitle, 225, 42, color.RGBA{R: 148, G: 163, B: 184, A: 255}, 1)
	}

	// User Name & Tag
	name := TruncateString(data.Name, 22)
	c.DrawText(name, 40, 78, color.RGBA{R: 248, G: 250, B: 252, A: 255}, 2)

	userHandle := ""
	if data.Username != "" {
		userHandle = "@" + data.Username + " • "
	}
	titleStr := GetUserTitle(data.Level)
	c.DrawText(userHandle+titleStr, 40, 110, color.RGBA{R: 56, G: 189, B: 248, A: 255}, 1)

	// Level and Rank Display on top right
	lvlBadgeW := 120
	lvlBadgeH := 70
	c.FillRoundedRect(width-lvlBadgeW-40, 68, lvlBadgeW, lvlBadgeH, 12, color.RGBA{R: 12, G: 74, B: 110, A: 255})
	c.DrawText("LEVEL", width-lvlBadgeW-15, 78, color.RGBA{R: 125, G: 211, B: 252, A: 255}, 1)
	lvlNumStr := fmt.Sprintf("%d", data.Level)
	lvlNumW := len(lvlNumStr) * 14
	c.DrawText(lvlNumStr, width-40-(lvlBadgeW+lvlNumW)/2, 98, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 2)

	// Rank Info
	rankStr := fmt.Sprintf("RANK #%d", data.Rank)
	if data.TotalMembers > 0 {
		rankStr = fmt.Sprintf("RANK #%d OF %d", data.Rank, data.TotalMembers)
	}
	rankStrW := len(rankStr) * 7
	c.DrawText(rankStr, width-rankStrW-40, 42, color.RGBA{R: 251, G: 191, B: 36, A: 255}, 1)

	// Divider
	c.FillRect(40, 150, width-80, 2, color.RGBA{R: 51, G: 65, B: 85, A: 255})

	// Stats Grid: 4 boxes
	// Box dimensions:
	boxW := 172
	boxH := 105
	boxY := 170
	boxSpacing := 24

	stats := []struct {
		Label string
		Value string
		Color color.RGBA
	}{
		{"DAILY MSGS", fmt.Sprintf("%d", data.DailyMsgs), color.RGBA{R: 56, G: 189, B: 248, A: 255}},
		{"WEEKLY MSGS", fmt.Sprintf("%d", data.WeeklyMsgs), color.RGBA{R: 192, G: 132, B: 252, A: 255}},
		{"ALL-TIME MSGS", fmt.Sprintf("%d", data.AllTimeMsgs), color.RGBA{R: 52, G: 211, B: 153, A: 255}},
		{"TOTAL XP", fmt.Sprintf("%d XP", data.CurrentXP), color.RGBA{R: 251, G: 191, B: 36, A: 255}},
	}

	for i, st := range stats {
		bx := 40 + i*(boxW+boxSpacing)
		c.FillRoundedRect(bx, boxY, boxW, boxH, 10, color.RGBA{R: 15, G: 23, B: 42, A: 220})
		c.FillRect(bx+14, boxY+12, boxW-28, 1, color.RGBA{R: 51, G: 65, B: 85, A: 160})

		// Label
		c.DrawText(st.Label, bx+16, boxY+24, color.RGBA{R: 148, G: 163, B: 184, A: 255}, 1)

		// Value
		valStr := TruncateString(st.Value, 11)
		c.DrawText(valStr, bx+16, boxY+55, st.Color, 2)
	}

	// XP Progression Area
	progY := 300
	c.FillRoundedRect(40, progY, width-80, 115, 12, color.RGBA{R: 15, G: 23, B: 42, A: 220})

	pct := int(math.Round(data.ProgressRatio * 100))
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	progLabel := fmt.Sprintf("XP PROGRESS TO LEVEL %d", data.Level+1)
	c.DrawText(progLabel, 60, progY+18, color.RGBA{R: 226, G: 232, B: 240, A: 255}, 1)

	pctStr := fmt.Sprintf("%d%%", pct)
	pctW := len(pctStr) * 7
	c.DrawText(pctStr, width-pctW-60, progY+18, color.RGBA{R: 56, G: 189, B: 248, A: 255}, 1)

	// Progress bar
	c.DrawProgressBar(
		60, progY+42, width-120, 20,
		data.ProgressRatio,
		color.RGBA{R: 30, G: 41, B: 59, A: 255},
		color.RGBA{R: 56, G: 189, B: 248, A: 255},
	)

	// XP Subtext
	xpSubtext := fmt.Sprintf("%d / %d XP  (%d XP remaining)", data.CurrentXP, data.NextLevelXP, data.NextLevelXP-data.CurrentXP)
	c.DrawText(xpSubtext, 60, progY+74, color.RGBA{R: 148, G: 163, B: 184, A: 255}, 1)

	// Footer
	footerY := height - 45
	c.DrawText("⚡ POWERED BY @STDBOTS • STD DEEPANSHU • https://deepanshu.in", 40, footerY, color.RGBA{R: 148, G: 163, B: 184, A: 255}, 1)
	botTag := "@StdRankingBot"
	botW := len(botTag) * 7
	c.DrawText(botTag, width-botW-40, footerY, color.RGBA{R: 56, G: 189, B: 248, A: 255}, 1)

	// Encode to PNG buffer
	var buf bytes.Buffer
	if err := png.Encode(&buf, c.Img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
