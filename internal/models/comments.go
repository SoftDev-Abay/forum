package models

import "database/sql"

type CommentsModelInterface interface {

}

type Comments struct {
	ID     uint
	PostID uint
	UserID uint
	Text string
}

type CommentsModel struct {
	DB *sql.DB
}

func (m *CommentsModel) Insert(postID uint, userID uint, text string) (uint, error) {
	stmt := `INSERT INTO Comments (post_id, user_id, text)
	VALUES (?, ?, ?)`

	result, err := m.DB.Exec(stmt, postID, userID, text)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}