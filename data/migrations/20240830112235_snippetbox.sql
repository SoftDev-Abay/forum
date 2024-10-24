-- +goose Up
-- +goose StatementBegin

-- Create table for Users
CREATE TABLE Users (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    enabled BOOLEAN NOT NULL
);

-- Create table for Categories
CREATE TABLE Categories (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE
);

-- Create table for Posts
CREATE TABLE Posts (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    imgUrl TEXT,
    createdAt DATETIME NOT NULL,
    category_id INTEGER NOT NULL,
    owner_id INTEGER NOT NULL,
    FOREIGN KEY (category_id) REFERENCES Categories(id),
    FOREIGN KEY (owner_id) REFERENCES Users(id)
);

-- Create table for Comments
CREATE TABLE Comments (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    text TEXT NOT NULL,
    FOREIGN KEY (post_id) REFERENCES Posts(id),
    FOREIGN KEY (user_id) REFERENCES Users(id)
);

-- Create table for Post Reactions (use TEXT for reaction type)
CREATE TABLE Post_Reactions (
    type TEXT CHECK(type IN ('like', 'dislike')) NOT NULL,
    user_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, post_id),
    FOREIGN KEY (user_id) REFERENCES Users(id),
    FOREIGN KEY (post_id) REFERENCES Post(id)
);

-- Create table for Comment Reactions (use TEXT for reaction type)
CREATE TABLE Comment_Reactions (
    type TEXT CHECK(type IN ('like', 'dislike')) NOT NULL,
    user_id INTEGER NOT NULL,
    comment_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, comment_id),
    FOREIGN KEY (user_id) REFERENCES Users(id),
    FOREIGN KEY (comment_id) REFERENCES Comments(id)
);

-- Create table for Sessions
CREATE TABLE Sessions (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    token TEXT NOT NULL UNIQUE,
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