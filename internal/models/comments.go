package models

import (
	"database/sql"
	"time"
)

type CommentsModelInterface interface {
	Insert(postID int, userID int, text string) (int, error)
	GetAllByPostID(postID int) ([]*Comments, error)
}

type Comments struct {
	ID          int
	PostID      int
	UserID      int
	Text        string
	LikeCount   int
	DislikeCout int
	CreatedAt   time.Time
}

type CommentsModel struct {
	DB *sql.DB
}

func (m *CommentsModel) Insert(postID int, userID int, text string) (int, error) {
	stmt := `INSERT INTO Comments (post_id, user_id, text, like_count, dislike_count)
	VALUES (?, ?, ?, 0, 0)`

	result, err := m.DB.Exec(stmt, postID, userID, text)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *CommentsModel) GetAllByPostID(UserID int) ([]*Comments, error) {
	stmt := `SELECT id, post_id, user_id, text, like_count, dislike_count, created_at
			FROM Comments
			WHERE post_id = ?
			ORDER BY created_at ASC`

	rows, err := m.DB.Query(stmt, UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comments

	for rows.Next() {
		comment := &Comments{}
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Text, &comment.LikeCount, &comment.DislikeCout, &comment.CreatedAt)
		if err != nil {
			return nil, err
		}

		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}
