# 💼 FinanceBot – Personal Finance Assistant for Telegram

A smart personal finance tracker built in **Go (Golang)**. Helps users manage expenses, set monthly limits, track income, and get insightful reports — all via a clean Telegram interface.

Perfect for budgeting, saving, and staying on top of your finances.

---

## 🚀 Features

- ✅ **Expense & Income Tracking**: Log with simple commands like `food 500` or `salary 80000`
- ✅ **Categories & Monthly Limits**: Create categories with spending limits
- ✅ **Spending Alerts**:
  - ⚠️ Warns at 80% of limit
  - ❌ Alerts when limit is exceeded
- ✅ **Analytics**: View monthly breakdown by category and balance (income vs expenses)
- ✅ **Monthly Reports**: Automatic summary sent on the 1st of each month
- ✅ **Multi-Language Support**:
  - 🇷🇺 Russian
  - 🇬🇧 English
  - Language selected at `/start`
- ✅ **Dynamic Menu System**:
  - Hierarchical navigation (Categories, Limits)
  - Back button and clean UX
- ✅ **Data Management**:
  - SQLite for persistent storage
  - Redis for session state (FSM)
  - Auto-cleanup of expenses older than 3 months
- ✅ **Currency Support**: Auto currency symbol (₽ for RU, $ for EN)
- ✅ **Docker & Makefile**: Easy deployment and local development

---

## 🛠️ Tech Stack

- **Language**: Go 1.21+
- **Telegram API**: [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- **Database**: SQLite (expenses, categories, incomes)
- **Session Storage**: Redis (user states, language, temp data)
- **Localization**: Custom i18n system with dynamic currency
- **Build & Deploy**: Docker, Docker Compose, Makefile
- **Architecture**: Modular (commands, db, state, i18n)

## 📦 Installation

### 1. Clone the repo

```bash
git clone git@github.com:llerabietech/personal-finance-telegram-bot-golang.git
cd personal-finance-telegram-bot-golang
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Create .env file

```bash
  TELEGRAM_TOKEN=your_telegram_bot_token_here
  REDIS_PASSWORD=strongpassword123
```
Get a token from @BotFather on Telegram. 

This file is auto-created by make up if missing. 

### ▶️ Run with Docker (Recommended)

We use Docker Compose to run the bot and Redis together.

## Build and start
```bash
make up
```
This will: 

Build the bot image
Start Redis with persistence and password
Connect bot to Redis via service name redis:6379
Load environment from .env
Start the bot in Telegram with /start

## Other Make commands

```bash
make build    # Build image only
make logs     # View bot logs
make down     # Stop containers
make clean    # Stop and remove containers
make env      # Show .env variables
make help     # Show all commands
```

### 📂 Project Structure

```
personal-finance-telegram-bot-golang/
├── main.go
├── bot/
│   ├── bot.go
│   └── scheduler.go
├── db/
│   ├── db.go         # SQLite
│   └── redis.go      # Redis client
│   ├── models.go      
├── commands/
│   ├── commands.go
│   └── monthly_report.go
├── state/
│   └── state.go      # FSM & user language
├── i18n/
│   └── i18n.go       # Translations & currency
├── utils/
│   └── format.go       
│   └── text.go       # Title case with golang.org/x/text
|   └── month.go      # Localized month names
├── Dockerfile        # Multi-stage build with CGO_ENABLED=1
├── docker-compose.yml # Redis + bot, with password
├── Makefile          # Dev commands
├── .env              # Auto-created
└── README.md
```