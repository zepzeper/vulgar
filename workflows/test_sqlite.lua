-- SQLite Module Test Script
-- Run with: vulgar test_sqlite.lua

local sqlite = require("integrations.sqlite")

print("=== SQLite Module Test ===\n")

-- Open database
local db, err = sqlite.open(".data/test.db")
if err then
    print("ERROR opening database:", err)
    return
end
print("[OK] Opened database: .data/test.db")

-- Create table
local result, err = sqlite.exec(db, [[
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        email TEXT UNIQUE,
        age INTEGER
    )
]])
if err then
    print("ERROR creating table:", err)
else
    print("[OK] Created users table")
end

-- Clear existing data
sqlite.exec(db, "DELETE FROM users")

-- Insert data
local id1, err = sqlite.insert(db, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", {"Alice", "alice@example.com", 28})
local id2, err = sqlite.insert(db, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", {"Bob", "bob@example.com", 35})
local id3, err = sqlite.insert(db, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", {"Charlie", "charlie@example.com", 42})
print("[OK] Inserted 3 users with IDs:", id1, id2, id3)

-- Query all rows
local rows, err = sqlite.query(db, "SELECT * FROM users ORDER BY name")
if err then
    print("ERROR querying:", err)
else
    print("[OK] Query returned", #rows, "rows:")
    for i, row in ipairs(rows) do
        print("    ", row.id, row.name, row.email, row.age)
    end
end

-- Query one row
local user, err = sqlite.query_one(db, "SELECT * FROM users WHERE name = ?", {"Bob"})
if user then
    print("[OK] Found user:", user.name, "age:", user.age)
end

-- Update
local count, err = sqlite.update(db, "UPDATE users SET age = ? WHERE name = ?", {36, "Bob"})
print("[OK] Updated", count, "row(s)")

-- Verify update
local user, err = sqlite.query_one(db, "SELECT age FROM users WHERE name = ?", {"Bob"})
print("[OK] Bob's new age:", user.age)

-- Delete
local count, err = sqlite.delete(db, "DELETE FROM users WHERE name = ?", {"Charlie"})
print("[OK] Deleted", count, "row(s)")

-- Transaction test
print("\n=== Transaction Test ===")
local tx, err = sqlite.begin(db)
if err then
    print("ERROR beginning transaction:", err)
else
    local result, err = sqlite.tx_exec(tx, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", {"Diana", "diana@example.com", 30})
    if err then
        sqlite.rollback(tx)
        print("[ROLLBACK] Transaction rolled back:", err)
    else
        sqlite.commit(tx)
        print("[OK] Transaction committed")
    end
end

-- Final count
local rows, err = sqlite.query(db, "SELECT COUNT(*) as count FROM users")
print("[OK] Final user count:", rows[1].count)

-- Close
local err = sqlite.close(db)
if not err then
    print("\n[OK] Database closed successfully")
end

print("\n=== All SQLite tests passed! ===")
