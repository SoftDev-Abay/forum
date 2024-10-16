-- +goose Up
-- +goose StatementBegin

-- Create table for Users
CREATE TABLE Users (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    enabled BOOLEAN NOT NULL
);

-- Create table for Categories
CREATE TABLE Categories (
    name VARCHAR(100) NOT NULL PRIMARY KEY,
    description TEXT
);

-- Create table for Posts
CREATE TABLE Post (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    title VARCHAR(255) NOT NULL,
    desc TEXT NOT NULL,
    imgUrl TEXT,
    createdAt DATETIME NOT NULL,
    category VARCHAR(100) NOT NULL,
    owner_id INTEGER NOT NULL,
    like_count INTEGER DEFAULT 0,
    dislike_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    FOREIGN KEY (category) REFERENCES Categories(name),
    FOREIGN KEY (owner_id) REFERENCES Users(id)
);

-- Create table for Comments
CREATE TABLE Comments (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    text TEXT NOT NULL,
    like_count INTEGER DEFAULT 0,
    dislike_count INTEGER DEFAULT 0,
    FOREIGN KEY (post_id) REFERENCES Post(id),
    FOREIGN KEY (user_id) REFERENCES Users(id)
);

-- Create table for Post Reactions (use TEXT for reaction type)
CREATE TABLE Post_Reactions (
    type TEXT CHECK(type IN ('like', 'dislike')) NOT NULL,
    user_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES Users(id),
    FOREIGN KEY (post_id) REFERENCES Post(id)
);

-- Create table for Comment Reactions (use TEXT for reaction type)
CREATE TABLE Comment_Reactions (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    type TEXT CHECK(type IN ('like', 'dislike')) NOT NULL,
    user_id INTEGER NOT NULL,
    comment_id INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES Users(id),
    FOREIGN KEY (comment_id) REFERENCES Comments(id)
);

-- Create table for Sessions
CREATE TABLE Sessions (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    token TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    createdAt DATETIME NOT NULL,
    expiresAt DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES Users(id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS Sessions;
DROP TABLE IF EXISTS Comment_Reactions;
DROP TABLE IF EXISTS Post_Reactions;
DROP TABLE IF EXISTS Comments;
DROP TABLE IF EXISTS Post;
DROP TABLE IF EXISTS Categories;
DROP TABLE IF EXISTS Users;
-- +goose StatementEnd