# 🏆 StdRankingBot — Enterprise Gamification & Chat Ranking Bot

<p align="center">
  <img src="https://graph.org/file/00ea4effe5d2dfbb8d5be.jpg" alt="StdRankingBot Banner" width="450"/>
</p>

<p align="center">
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Language-Go%201.22+-00ADD8?style=for-the-badge&logo=go" alt="Go"/></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-AGPL%20v3-blue?style=for-the-badge" alt="License"/></a>
  <a href="https://t.me/STDBOTS"><img src="https://img.shields.io/badge/Channel-%40STDBOTS-2CA5E0?style=for-the-badge&logo=telegram" alt="Telegram Channel"/></a>
  <a href="https://deepanshu.in"><img src="https://img.shields.io/badge/Author-STD%20DEEPANSHU-FF4500?style=for-the-badge" alt="Author"/></a>
</p>

<p align="center">
  <a href="https://dashboard.heroku.com/new?template=https://github.com/StdBots/StdRankingBot">
    <img src="https://img.shields.io/badge/Deploy%20To%20Heroku-7056bf?style=for-the-badge&logo=heroku" alt="Deploy to Heroku"/>
  </a>
  <a href="https://railway.app/template/new?template=https://github.com/StdBots/StdRankingBot">
    <img src="https://img.shields.io/badge/Deploy%20On%20Railway-0B0D0E?style=for-the-badge&logo=railway" alt="Deploy on Railway"/>
  </a>
</p>

---

## ⚡ Overview

**StdRankingBot** is an enterprise-grade gamification, XP progression, and chat ranking bot engineered in **Go (Golang)** with **MongoDB** persistence.

Modeled after Telegram's popular `@chatfightbot` (ChatFight), it revolutionizes Telegram group engagement by transforming passive conversations into competitive esports tournaments with **pure Go lightning-fast HD banner generation** (<15ms per card without any external headless browsers or ImageMagick dependencies).

Developed by **[STD DEEPANSHU](https://deepanshu.in)** as the 5th bot of Phase 1 in the **[STD BOTS Ecosystem](https://t.me/STDBOTS)**.

---

## ✨ Features

- 🖼 **Dynamic HD Leaderboard Banners:**
  - Pure Go 1000x660 PNG generation in <15ms.
  - Interactive inline period switcher: **[📅 Daily]**, **[📆 Weekly]**, **[🗓 Monthly]**, and **[👑 All-Time]**.
  - Distinct cyberpunk color palettes: Cyan (Daily), Royal Purple (Weekly), Cyber Emerald (Monthly), and Imperial Obsidian Gold (All-Time).
  - 2-Column tournament layout with Top 1-10 rankings, custom medals (🥇, 🥈, 🥉), levels, message counters, and XP.
- 🎮 **Personal Gamer Cards (`/me`):**
  - High-resolution 840x500 PNG profile card.
  - Displays user rank in the chat, 4-stat metric box (Daily, Weekly, All-Time, Total XP), active tier title, and a visual gradient XP progress bar.
- 🛡 **Anti-Farming Shield:**
  - Built-in thread-safe memory debounce (3s per user).
  - Cryptographic SHA-256 message hashing to detect and discard duplicate text repeats within 60 seconds.
  - Minimum message length filter (configurable).
- ⚡ **Mathematical XP Curve & RPG Tiers:**
  - Quadratic leveling curve: `Level = floor(sqrt(XP / 50)) + 1`.
  - 9 dynamic prestige titles ranging from *🌱 Rookie Fighter* to *👑 Omnipotent God*.
- 🔔 **Celebratory Level-Up Alerts:**
  - Automatic celebratory announcements in chat when a member reaches a new level.
  - Can be toggled ON/OFF by group administrators via `/settings`.
- 🌐 **Global Cross-Group Hall of Fame (`/global`):**
  - Aggregates activity across all groups running the bot to establish a global Telegram leaderboard.
- 📢 **Admin Broadcast & Analytics Suite:**
  - High-speed concurrent message broadcaster with worker pools and automatic 429 FloodWait handling.
  - Real-time performance analytics (`/stats`) monitoring memory alloc, goroutines, and message velocity.
- 🔒 **7-Layer Forced Credit Protection:**
  - AGPL-3.0 attribution, zero-width steganographic watermarks, tamper-detection, and telemetry monitoring for STD DEEPANSHU & @STDBOTS.

---

## 🎮 XP Progression & Fighter Tiers

| Level Range | Prestige Title | Required XP |
| :--- | :--- | :--- |
| **Lv. 1 - 4** | 🌱 Rookie Fighter | 0 – 800 XP |
| **Lv. 5 - 9** | 🗡️ Apprentice | 800 – 3,200 XP |
| **Lv. 10 - 14** | 🎯 Warrior | 3,200 – 9,800 XP |
| **Lv. 15 - 24** | 🔥 Battle Knight | 9,800 – 28,800 XP |
| **Lv. 25 - 34** | 🛡️ Elite Veteran | 28,800 – 57,800 XP |
| **Lv. 35 - 49** | ⚔️ Mythic Gladiator | 57,800 – 120,050 XP |
| **Lv. 50 - 74** | ⚡ Grandmaster Fighter | 120,050 – 273,800 XP |
| **Lv. 75 - 99** | 🔱 Celestial Warlord | 273,800 – 490,050 XP |
| **Lv. 100+** | 👑 Omnipotent God | 490,050+ XP |

---

## 🚀 One-Click Deployments

### 🟣 Deploy to Heroku
Click the button below to deploy your instance to Heroku in under 60 seconds:

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://dashboard.heroku.com/new?template=https://github.com/StdBots/StdRankingBot)

### 🚂 Deploy on Railway
Click the button below to deploy on Railway with container support:

[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/template/new?template=https://github.com/StdBots/StdRankingBot)

### 🖥️ 1-Command VPS Deployment (Linux / Ubuntu / Debian)
Run this single command on your VPS as root:
```bash
curl -fsSL https://raw.githubusercontent.com/StdBots/StdRankingBot/main/scripts/install_vps.sh | bash
```

---

## 🐳 Docker & Manual VPS Setup

```bash
# 1. Clone repository
git clone https://github.com/StdBots/StdRankingBot.git
cd StdRankingBot

# 2. Configure environment
cp .env.example .env
nano .env

# 3. Start with Docker Compose
docker compose up -d --build
```

---

## ⚙️ Environment Variables

| Variable | Description | Required | Default |
| :--- | :--- | :--- | :--- |
| `BOT_TOKEN` | Telegram Bot Token from [@BotFather](https://t.me/BotFather) | **Yes** | — |
| `OWNER_ID` | Telegram Numeric User ID of the Bot Owner | **Yes** | `7394590844` |
| `MONGO_URI` | MongoDB Connection String (Atlas or Local) | **Yes** | `mongodb://localhost:27017/stdrankingbot` |
| `DB_NAME` | Database Name | No | `stdrankingbot` |
| `FORCE_SUB_CHANNEL` | Channel username for force subscribe (without `@`) | No | `StdBots` |
| `LOG_CHANNEL_ID` | Channel ID for audit & alert forwarding | No | `0` |
| `ANTI_SPAM_COOLDOWN_SECONDS` | Anti-farming cooldown per user in seconds | No | `3` |
| `MIN_MSG_LENGTH` | Minimum characters to award XP | No | `3` |
| `BASE_XP_PER_MSG` | Base XP awarded per valid message | No | `15` |
| `ENV` | Environment (`production` / `development`) | No | `production` |

---

## ⌨️ Bot Commands

| Command | Scope | Description |
| :--- | :--- | :--- |
| `/top` or `/ranking` | Group | Generates HD Leaderboard Banner with Daily, Weekly, Monthly, and All-Time buttons |
| `/me` or `/profile` | Group / PM | Generates your personalized 840x500 Gamer Profile Card |
| `/global` | Everywhere | Displays global cross-group Hall of Fame |
| `/settings` | Group Admins | Toggle Level-Up celebration alerts and view group stats |
| `/help` | Everywhere | Complete guide on XP mechanics, leveling curve, and commands |
| `/stats` | Owner | Real-time performance analytics, goroutines, and memory alloc |
| `/broadcast` | Owner | High-speed concurrent announcement engine across all groups |

---

## 🔒 7-Layer Forced Credit Protection

This repository is distributed under the **GNU Affero General Public License v3 (AGPL-3.0)** with strict attribution requirements:
1. **Source Code Header:** Preserved in every source and header file.
2. **Obfuscated Strings:** Dev identity encoded via multi-layer string obfuscation.
3. **Invisible Steganographic Watermarks:** Injected into Telegram messages using zero-width Unicode characters (`\u200B`, `\u200C`, `\u200D`).
4. **Runtime Integrity Guard:** Verifies developer signatures on startup; halts execution if tampered.
5. **UI & Banner Footers:** Active hyperlinks on all interactive keyboards and generated HD images.
6. **Telemetry Heartbeat:** Non-blocking telemetry reporting fork statuses.
7. **Strict AGPL-3.0 Additional Terms:** Commercial or modified instances MUST display "Powered by STD BOTS (@STDBOTS)".

---

## 👨‍💻 Developer & Support

- **Lead Architect:** [STD DEEPANSHU](https://deepanshu.in)
- **Telegram Channel:** [@STDBOTS](https://t.me/STDBOTS)
- **Personal Handle:** [@STD_DEEPANSHU](https://t.me/STD_DEEPANSHU)
- **Website:** [https://deepanshu.in](https://deepanshu.in)

Copyright (c) 2024-2026 STD DEEPANSHU & STD BOTS. All Rights Reserved.
