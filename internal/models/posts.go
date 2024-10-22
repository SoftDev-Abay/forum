package models

import (
	"database/sql"
	"errors"
	"time"
)

type PostsModelInterface interface {
	Insert(title string, content string, imgUrl string, createdAt time.Time, categoryID uint, ownerID uint) (int, error)
	Get(id int) (*Posts, error)
	Latest() ([]*Posts, error)
}

type Posts struct {
	ID         uint
	Title      string
	Content    string
	ImgUrl     string
	CreatedAt  time.Time
	CategoryID uint
	OwnerID    uint
}

type PostModel struct {
	DB *sql.DB
}

func (m *PostModel) Insert(title string, content string, imgUrl string, createdAt time.Time, categoryID uint, ownerID uint) (int, error) {
	stmt := `INSERT INTO Posts (title, content, imgUrl, createdAt, category_id, owner_id)
	         VALUES (?, ?, ?, ?, ?, ?)`

	result, err := m.DB.Exec(stmt, title, content, imgUrl, createdAt, categoryID, ownerID)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *PostModel) Get(id int) (*Posts, error) {
	stmt := `SELECT id, title, content, imgUrl, createdAt, category_id, owner_id
	         FROM Posts
	         WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)

	post := &Posts{}

	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.ImgUrl, &post.CreatedAt, &post.CategoryID, &post.OwnerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return post, nil
}

func (m *PostModel) Latest() ([]*Posts, error) {
	stmt := `SELECT id, title, content, imgUrl, createdAt, category_id, owner_id
	         FROM Posts
	         ORDER BY createdAt DESC
	         LIMIT 10`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Posts

	for rows.Next() {
		post := &Posts{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.ImgUrl, &post.CreatedAt, &post.CategoryID, &post.OwnerID)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
