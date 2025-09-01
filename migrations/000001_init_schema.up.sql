CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY
    );
CREATE TABLE IF NOT EXISTS categories (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        user_id INTEGER,
        limit_sum REAL,
        FOREIGN KEY(user_id) REFERENCES users(id)
    );
CREATE TABLE IF NOT EXISTS expenses (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        category_id INTEGER,
        amount REAL,
        date TEXT,
        FOREIGN KEY(category_id) REFERENCES categories(id)
    );
CREATE TABLE IF NOT EXISTS incomes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        source TEXT,
        amount REAL,
        date TEXT
    );


