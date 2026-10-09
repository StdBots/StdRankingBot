package graphics

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"strings"
	"time"
)

// RankedMember represents a single user's ranking display data
type RankedMember struct {
	Rank      int
	Name      string
	Username  string
	Level     int
	Messages  int64
	XP        int64
	IsUserReq bool
}

// LeaderboardCardData defines all inputs to render the leaderboard banner
type LeaderboardCardData struct {
	ChatTitle   string
	Period      string // "daily", "weekly", "monthly", "alltime"
	Members     []RankedMember
	TotalActive int64
	TotalMsgs   int64
	GeneratedAt time.Time
}

// Color schemes per period
type Theme struct {
	BgTop      color.RGBA
	BgBottom   color.RGBA
	Accent     color.RGBA
	CardBg     color.RGBA
	CardBorder color.RGBA
	TextMain   color.RGBA
	TextDim    color.RGBA
	HeaderBg   color.RGBA
}

func getTheme(period string) Theme {
	switch strings.ToLower(period) {
	case "weekly":
		return Theme{
			BgTop:      color.RGBA{R: 28, G: 14, B: 54, A: 255},  // #1c0e36
			BgBottom:   color.RGBA{R: 10, G: 5, B: 20, A: 255},   // #0a0514
			Accent:     color.RGBA{R: 192, G: 132, B: 252, A: 255}, // #c084fc (Purple)
			CardBg:     color.RGBA{R: 35, G: 20, B: 65, A: 200},
			CardBorder: color.RGBA{R: 147, G: 51, B: 234, A: 120},
			TextMain:   color.RGBA{R: 245, G: 243, B: 255, A: 255},
			TextDim:    color.RGBA{R: 167, G: 139, B: 250, A: 255},
			HeaderBg:   color.RGBA{R: 88, G: 28, B: 135, A: 255},
		}
	case "monthly":
		return Theme{
			BgTop:      color.RGBA{R: 6, G: 42, B: 32, A: 255},   // #062a20
			BgBottom:   color.RGBA{R: 3, G: 18, B: 14, A: 255},   // #03120e
			Accent:     color.RGBA{R: 52, G: 211, B: 153, A: 255}, // #34d399 (Emerald)
			CardBg:     color.RGBA{R: 12, G: 55, B: 42, A: 200},
			CardBorder: color.RGBA{R: 16, G: 185, B: 129, A: 120},
			TextMain:   color.RGBA{R: 236, G: 253, B: 245, A: 255},
			TextDim:    color.RGBA{R: 110, G: 231, B: 183, A: 255},
			HeaderBg:   color.RGBA{R: 6, G: 95, B: 70, A: 255},
		}
	case "alltime":
		return Theme{
			BgTop:      color.RGBA{R: 38, G: 25, B: 6, A: 255},   // #261906
			BgBottom:   color.RGBA{R: 14, G: 9, B: 2, A: 255},    // #0e0902
			Accent:     color.RGBA{R: 251, G: 191, B: 36, A: 255}, // #fbbf24 (Gold)
			CardBg:     color.RGBA{R: 50, G: 34, B: 10, A: 200},
			CardBorder: color.RGBA{R: 217, G: 119, B: 6, A: 120},
			TextMain:   color.RGBA{R: 255, G: 251, B: 235, A: 255},
			TextDim:    color.RGBA{R: 252, G: 211, B: 77, A: 255},
			HeaderBg:   color.RGBA{R: 146, G: 64, B: 14, A: 255},
		}
	default: // daily
		return Theme{
			BgTop:      color.RGBA{R: 11, G: 24, B: 48, A: 255},  // #0b1830
			BgBottom:   color.RGBA{R: 4, G: 10, B: 24, A: 255},   // #040a18
			Accent:     color.RGBA{R: 56, G: 189, B: 248, A: 255}, // #38bdf8 (Cyan)
			CardBg:     color.RGBA{R: 16, G: 35, B: 68, A: 200},
			CardBorder: color.RGBA{R: 14, G: 116, B: 144, A: 120},
			TextMain:   color.RGBA{R: 240, G: 249, B: 255, A: 255},
			TextDim:    color.RGBA{R: 125, G: 211, B: 252, A: 255},
			HeaderBg:   color.RGBA{R: 12, G: 74, B: 110, A: 255},
		}
	}
}

// RenderLeaderboard renders an HD 1000x660 PNG leaderboard banner
func RenderLeaderboard(data LeaderboardCardData) ([]byte, error) {
	const width = 1000
	const height = 660

	c := NewCanvas(width, height)
	theme := getTheme(data.Period)

	// 1. Background gradient & subtle tech grid
	c.FillVerticalGradient(theme.BgTop, theme.BgBottom)
	c.DrawGridPattern(40, color.RGBA{R: 255, G: 255, B: 255, A: 8})

	// 2. Header Area
	periodLabel := strings.ToUpper(data.Period) + " LEADERBOARD"
	if strings.ToLower(data.Period) == "alltime" {
		periodLabel = "ALL-TIME HALL OF FAME"
	}

	// Period Badge
	badgeW := len(periodLabel)*7 + 24
	c.FillRoundedRect(35, 22, badgeW, 24, 6, theme.HeaderBg)
	c.DrawText(periodLabel, 47, 27, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 1)

	// Chat Title
	chatTitle := TruncateString(data.ChatTitle, 35)
	c.DrawText(chatTitle, 35, 55, theme.TextMain, 2)

	// Subtitle
	subtitle := fmt.Sprintf("TOP 10 FIGHTERS • %s", data.GeneratedAt.Format("02 Jan 2006, 15:04 MST"))
	c.DrawText(subtitle, 35, 87, theme.TextDim, 1)

	// Stats Summary in top right
	statText := fmt.Sprintf("TOTAL MSGS: %d", data.TotalMsgs)
	statW := len(statText) * 7
	c.DrawText(statText, width-statW-40, 58, theme.Accent, 1)

	// Divider line under header
	c.FillRect(35, 106, width-70, 2, theme.CardBorder)

	// 3. Leaderboard Cards: 2 Columns of 5 items each
	// Left column: 1-5, Right column: 6-10
	colWidth := 450
	cardHeight := 88
	colSpacing := 30
	startY := 122
	rowSpacing := 14

	for i := 0; i < 10; i++ {
		col := i / 5 // 0 for left, 1 for right
		row := i % 5

		x := 35 + col*(colWidth+colSpacing)
		y := startY + row*(cardHeight+rowSpacing)

		// Draw card background
		c.FillRoundedRect(x, y, colWidth, cardHeight, 10, theme.CardBg)

		if i >= len(data.Members) {
			// Empty slot placeholder
			c.DrawText(fmt.Sprintf("#%d [ EMPTY SLOT ]", i+1), x+20, y+35, color.RGBA{R: 120, G: 120, B: 140, A: 160}, 1)
			continue
		}

		m := data.Members[i]

		// Rank Badge styling
		rankColor := theme.TextDim
		rankBadgeBg := color.RGBA{R: 30, G: 40, B: 60, A: 255}
		switch m.Rank {
		case 1:
			rankColor = color.RGBA{R: 255, G: 215, B: 0, A: 255}    // Gold
			rankBadgeBg = color.RGBA{R: 120, G: 90, B: 0, A: 255}
		case 2:
			rankColor = color.RGBA{R: 226, G: 232, B: 240, A: 255}  // Silver
			rankBadgeBg = color.RGBA{R: 80, G: 90, B: 105, A: 255}
		case 3:
			rankColor = color.RGBA{R: 245, G: 158, B: 11, A: 255}   // Bronze
			rankBadgeBg = color.RGBA{R: 110, G: 65, B: 5, A: 255}
		default:
			rankColor = theme.Accent
		}

		// Rank pill on the left
		c.FillRoundedRect(x+12, y+14, 46, 60, 8, rankBadgeBg)
		rankStr := fmt.Sprintf("#%d", m.Rank)
		rankX := x + 12 + (46-len(rankStr)*14)/2
		c.DrawText(rankStr, rankX, y+30, rankColor, 2)

		// Member Name
		dispName := m.Name
		if dispName == "" {
			dispName = m.Username
		}
		if dispName == "" {
			dispName = "Fighter"
		}
		dispName = TruncateString(dispName, 18)
		c.DrawText(dispName, x+70, y+18, theme.TextMain, 2)

		// Level Badge
		lvlStr := fmt.Sprintf("LVL %d", m.Level)
		c.FillRoundedRect(x+70, y+48, len(lvlStr)*7+16, 22, 6, theme.HeaderBg)
		c.DrawText(lvlStr, x+78, y+53, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 1)

		// Messages and XP on right side of the card
		msgStr := fmt.Sprintf("%d msgs", m.Messages)
		xpStr := fmt.Sprintf("%d XP", m.XP)

		msgW := len(msgStr) * 7
		xpW := len(xpStr) * 7

		c.DrawText(msgStr, x+colWidth-msgW-18, y+24, theme.TextDim, 1)
		c.DrawText(xpStr, x+colWidth-xpW-18, y+51, theme.Accent, 1)
	}

	// 4. Footer Banner
	footerY := height - 32
	c.FillRect(0, footerY, width, 32, color.RGBA{R: 0, G: 0, B: 0, A: 220})
	creditText := "⚡ POWERED BY @STDBOTS • STD DEEPANSHU • https://deepanshu.in"
	c.DrawText(creditText, 35, footerY+10, color.RGBA{R: 200, G: 200, B: 200, A: 255}, 1)

	botTag := "@StdRankingBot"
	botTagW := len(botTag) * 7
	c.DrawText(botTag, width-botTagW-35, footerY+10, theme.Accent, 1)

	// Encode to PNG buffer
	var buf bytes.Buffer
	if err := png.Encode(&buf, c.Img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
