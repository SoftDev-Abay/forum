package models

import (
	"database/sql"
	"errors"
)

var (
	ErrReactionAlreadyExists = errors.New("models: reaction already exists for this user on the post")
	ErrNoReaction            = errors.New("models: no reaction found for this user on the post")
)

type PostReactionModelInterface interface {
	AddReaction(userID int, postID int, reactionType string) error
	UpdateReaction(userID int, postID int, reactionType string) error
	DeleteReaction(userID int, postID int) error
	GetReaction(userID int, postID int) (*PostReaction, error)
	GetReactionCount(postID int, reactionType string) (int, error)
	GetReactionByUserID(userID int) (*PostReaction, error)
	GetLikedPostIDsByUserID(userID int) ([]int, error)
	DeleteReactionsByPostId(postID int) error
}

type PostReaction struct {
	UserID int
	PostID int
	Type   string
}

type PostReactionsModel struct {
	DB *sql.DB
}

func (m *PostReactionsModel) AddReaction(userID int, postID int, reactionType string) error {
	if reactionType != "like" && reactionType != "dislike" {
		return errors.New("invalid reaction type")
	}

	existingReaction, err := m.GetReaction(userID, postID)
	if err != nil && err != ErrNoReaction {
		return err
	}

	if existingReaction != nil {
		return ErrReactionAlreadyExists
	}

	stmt := `INSERT INTO Post_Reactions (user_id, post_id, type) VALUES (?, ?, ?)`
	_, err = m.DB.Exec(stmt, userID, postID, reactionType)
	if err != nil {
		return err
	}

	return nil
}

func (m *PostReactionsModel) UpdateReaction(userID int, postID int, reactionType string) error {
	if reactionType != "like" && reactionType != "dislike" {
		return errors.New("invalid reaction type")
	}

	_, err := m.GetReaction(userID, postID)
	if err != nil {
		if err == ErrNoReaction {
			return errors.New("cannot update reaction: no reaction found")
		}
		return err
	}

	stmt := `UPDATE Post_Reactions SET type = ? WHERE user_id = ? AND post_id = ?`
	_, err = m.DB.Exec(stmt, reactionType, userID, postID)
	if err != nil {
		return err
	}

	return nil
}

func (m *PostReactionsModel) DeleteReaction(userID int, postID int) error {
	stmt := `DELETE FROM Post_Reactions WHERE user_id = ? AND post_id = ?`
	_, err := m.DB.Exec(stmt, userID, postID)
	if err != nil {
		return err
	}
	return nil
}

func (m *PostReactionsModel) GetReaction(userID int, postID int) (*PostReaction, error) {
	stmt := `SELECT type FROM Post_Reactions WHERE user_id = ? AND post_id = ?`
	row := m.DB.QueryRow(stmt, userID, postID)

	var reactionType string
	err := row.Scan(&reactionType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoReaction
		}
		return nil, err
	}

	return &PostReaction{
		UserID: userID,
		PostID: postID,
		Type:   reactionType,
	}, nil
}

func (m *PostReactionsModel) GetReactionByUserID(userID int) (*PostReaction, error) {
	stmt := `SELECT type FROM Post_Reactions WHERE user_id = ?`
	row := m.DB.QueryRow(stmt, userID)

	var reactionType string
	err := row.Scan(&reactionType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoReaction
		}
		return nil, err
	}

	return &PostReaction{
		UserID: userID,
		Type:   reactionType,
	}, nil
}

func (m *PostReactionsModel) GetReactionCount(postID int, reactionType string) (int, error) {
	stmt := `SELECT COUNT(*) FROM Post_Reactions WHERE post_id = ? AND type = ?`
	var count int
	err := m.DB.QueryRow(stmt, postID, reactionType).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *PostReactionsModel) GetLikedPostIDsByUserID(userID int) ([]int, error) {
	stmt := `SELECT post_id FROM Post_Reactions WHERE user_id = ? AND type = 'like'`
	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var postIDs []int
	for rows.Next() {
		var postID int
		if err := rows.Scan(&postID); err != nil {
			return nil, err
		}
		postIDs = append(postIDs, postID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return postIDs, nil
}

func (m *PostReactionsModel) DeleteReactionsByPostId(postID int) error {
	stmt := `DELETE FROM Post_Reactions WHERE  post_id = ?`
	_, err := m.DB.Exec(stmt, postID)
	if err != nil {
		return err
	}
	return nil
}
