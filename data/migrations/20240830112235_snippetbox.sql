-- +goose Up
-- +goose StatementBegin

-- Create table for Users
CREATE TABLE Users (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    role VARCHAR(20) CHECK(role IN ('user', 'moderator', 'admin')) NOT NULL DEFAULT 'user',
    enabled BOOLEAN NOT NULL DEFAULT 1
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
    like_count INTEGER NOT NULL,
    dislike_count INTEGER NOT NULL,
    FOREIGN KEY (category_id) REFERENCES Categories(id),
    FOREIGN KEY (owner_id) REFERENCES Users(id)
);

-- Create table for Comments
CREATE TABLE Comments (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    text TEXT NOT NULL,
    like_count INTEGER NOT NULL,
    dislike_count INTEGER NOT NULL,
    FOREIGN KEY (post_id) REFERENCES Posts(id),
    FOREIGN KEY (user_id) REFERENCES Users(id)
);

-- Create table for Post_Reactions (use TEXT for reaction type)
CREATE TABLE Post_Reactions (
    type TEXT CHECK(type IN ('like', 'dislike')) NOT NULL,
    user_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, post_id),
    FOREIGN KEY (user_id) REFERENCES Users(id),
    FOREIGN KEY (post_id) REFERENCES Posts(id)
);

-- Create table for Comment_Reactions (use TEXT for reaction type)
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

-- Create table for Report_Reasons
CREATE TABLE Report_Reasons (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    text TEXT NOT NULL
);

-- Create table for Reports
CREATE TABLE Reports (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    moderator_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    report_reason_id INTEGER NOT NULL,
    description TEXT NOT NULL,
    dateCreated DATETIME NOT NULL,
    admin_id INTEGER,

    FOREIGN KEY (moderator_id) REFERENCES Users(id),
    FOREIGN KEY (post_id) REFERENCES Posts(id),
    FOREIGN KEY (report_reason_id) REFERENCES Report_Reasons(id),
    FOREIGN KEY (admin_id) REFERENCES Users(id)
);

-- Create table for Promotion_Requests
CREATE TABLE Promotion_Requests (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    description TEXT,
    status  VARCHAR(20) CHECK(status IN ('pending', 'approved', 'rejected')) NOT NULL DEFAULT 'pending',
    FOREIGN KEY (user_id) REFERENCES Users(id)
);

CREATE TABLE Notifications (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    type TEXT CHECK(type IN ('post_like', 'post_dislike', 'comment', 'comment_like', 'comment_dislike')) NOT NULL,
    actor_id INTEGER NOT NULL,
    recipient_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    comment_id INTEGER,
    created_at DATETIME NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT 0,

    FOREIGN KEY (actor_id) REFERENCES Users(id),
    FOREIGN KEY (recipient_id) REFERENCES Users(id),
    FOREIGN KEY (post_id) REFERENCES Posts(id),
    FOREIGN KEY (comment_id) REFERENCES Comments(id)
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS Promotion_Requests;
DROP TABLE IF EXISTS Reports;
DROP TABLE IF EXISTS Report_Reasons;
DROP TABLE IF EXISTS Sessions;
DROP TABLE IF EXISTS Comment_Reactions;
DROP TABLE IF EXISTS Post_Reactions;
DROP TABLE IF EXISTS Comments;
DROP TABLE IF EXISTS Posts;
DROP TABLE IF EXISTS Categories;
DROP TABLE IF EXISTS Users;

-- +goose StatementEnd