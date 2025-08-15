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

- **Language**: Go 1.24+
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

> Get a token from [@BotFather](https://t.me/BotFather) on Telegram.
>
> This file is auto-created by `make up` if missing.

### ▶️ Run with Docker (Recommended)

We use Docker Compose to run the bot and Redis together.

### Build and start
```bash
make up
```
> This will:
>
> - Build the bot image
> - Start Redis with persistence and password
> - Connect bot to Redis via service name `redis:6379`
> - Load environment from `.env`

### Other Make commands

```bash
make build    # Build image only
make logs     # View bot logs
make down     # Stop containers
make clean    # Stop and remove containers
make env      # Show .env variables
make help     # Show all commands
```

## ⚙️ Configuration

The bot is highly configurable through environment variables. Key settings include:

### Limit Thresholds
- `LIMIT_WARNING_THRESHOLD`: Percentage for spending warnings (default: 80%)
- `LIMIT_OVERLOAD_THRESHOLD`: Percentage for limit exceeded alerts (default: 100%)
- `BALANCE_WARNING_THRESHOLD`: Percentage for balance warnings (default: 10%)

### Display & Formatting
- `CURRENCY_SYMBOL`: Currency symbol (default: ₽)
- `DATE_FORMAT`: Date format (default: 2006-01-02)
- `MONTH_FORMAT`: Month format (default: 2006-01)
- `TIME_FORMAT`: Time format (default: 15:04)

### Status Emojis
- `EMOJI_SUCCESS`: Success indicator (default: ✅)
- `EMOJI_WARNING`: Warning indicator (default: 🟡)
- `EMOJI_ERROR`: Error indicator (default: ❌)
- `EMOJI_BALANCE_GOOD`: Good balance (default: 🟢)
- `EMOJI_BALANCE_WARNING`: Balance warning (default: 🟡)
- `EMOJI_BALANCE_BAD`: Poor balance (default: 🔴)

### Confirmation Words
- `CONFIRMATION_WORDS`: Comma-separated list of confirmation words (default: да,yes)

### Language Configuration
- `LANGUAGES`: Comma-separated list of supported languages (default: ru,en)
- `DEFAULT_LANGUAGE`: Default language for fallback (default: en)

See .example.env` for all available options.

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
├── Dockerfile        
├── docker-compose.yml 
├── Makefile          # Dev commands
├── .env              # Auto-created
└── README.md
```