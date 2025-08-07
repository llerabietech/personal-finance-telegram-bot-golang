# 💼 FinanceBot – Personal Finance Assistant for Telegram

A smart personal finance tracker built in **Go (Golang)**. Helps users manage expenses, set monthly limits, track income, and get insightful reports — all via a clean Telegram interface.

Perfect for budgeting, saving, and staying on top of your finances.

---

## 🚀 Features

-  **Expense Tracking**: Log spending with simple commands like `food 500`
-  **Income Tracking**: Record income sources like `salary 80000`
-  **Categories & Limits**: Create categories (e.g., food, transport) with monthly spending limits
-  **Spending Alerts**:
  - ⚠️ Warns at 80% of limit
  - ❌ Alerts when limit is exceeded
-  **Monthly Reports**: Automatic summary sent on the 1st of each month
-  **Analytics**: View monthly breakdown by category and balance (income vs expenses)
-  **Multi-Language Support**:
  - 🇷🇺 Russian
  - 🇬🇧 English
  - Language selected at `/start`
-  **Dynamic Menu System**:
  - Hierarchical navigation (Categories, Limits)
  - Back button and clean UX
-  **Data Management**:
  - SQLite for persistent storage
  - Redis for session state (FSM)
  - Auto-cleanup of expenses older than 3 months
-  **Currency Support**: Auto currency symbol (₽ for RU, $ for EN)

## 🛠️ Tech Stack

- **Language**: Go 1.21+
- **Telegram API**: [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- **Database**: SQLite (expenses, categories, incomes)
- **Session Storage**: Redis (user states, language, temp data)
- **Localization**: Custom i18n system with dynamic currency
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

### 3. Run Redis (via Docker)

```bash
docker run --name finance-redis -p 6379:6379 -d redis:alpine
```

### 4. Set environment variable

Get a token from @BotFather on Telegram. Create config.json in /configs

### ▶️ Run the Bot

```bash
go run main.go
```

Start the bot in Telegram with /start

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
├── configs/
│   ├── config.json
├── state/
│   └── state.go      # FSM & user language
├── i18n/
│   └── i18n.go       # Translations & currency
├── utils/
│   └── format.go       
│   └── text.go       # Title case with golang.org/x/text
└── README.md
```