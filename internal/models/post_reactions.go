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
	AddReaction(userID uint, postID uint, reactionType string) error
	UpdateReaction(userID uint, postID uint, reactionType string) error
	DeleteReaction(userID uint, postID uint) error
	GetReaction(userID uint, postID uint) (*PostReaction, error)
	GetReactionCount(postID uint, reactionType string) (int, error)
}

type PostReaction struct {
	UserID uint
	PostID uint
	Type   string // 'like' or 'dislike'
}

type PostReactionsModel struct {
	DB *sql.DB
}

// AddReaction adds a new reaction (like or dislike) for a user on a post
func (m *PostReactionsModel) AddReaction(userID uint, postID uint, reactionType string) error {
	// Ensure valid reaction type
	if reactionType != "like" && reactionType != "dislike" {
		return errors.New("invalid reaction type")
	}

	// Check if the user has already reacted to this post
	existingReaction, err := m.GetReaction(userID, postID)
	if err != nil && err != ErrNoReaction {
		return err
	}

	// If a reaction exists, return error (because they should not be adding another reaction)
	if existingReaction != nil {
		return ErrReactionAlreadyExists // user already reacted to this post
	}

	// Insert the new reaction
	stmt := `INSERT INTO Post_Reactions (user_id, post_id, type) VALUES (?, ?, ?)`
	_, err = m.DB.Exec(stmt, userID, postID, reactionType)
	if err != nil {
		return err
	}

	return nil
}

func (m *PostReactionsModel) UpdateReaction(userID uint, postID uint, reactionType string) error {
	// Ensure valid reaction type
	if reactionType != "like" && reactionType != "dislike" {
		return errors.New("invalid reaction type")
	}

	// Check if the user has already reacted to this post
	_, err := m.GetReaction(userID, postID)
	if err != nil {
		if err == ErrNoReaction {
			return errors.New("cannot update reaction: no reaction found")
		}
		return err
	}

	// Update the reaction if it exists
	stmt := `UPDATE Post_Reactions SET type = ? WHERE user_id = ? AND post_id = ?`
	_, err = m.DB.Exec(stmt, reactionType, userID, postID)
	if err != nil {
		return err
	}

	return nil
}

// DeleteReaction removes a reaction (like or dislike) from a user on a post
func (m *PostReactionsModel) DeleteReaction(userID uint, postID uint) error {
	stmt := `DELETE FROM Post_Reactions WHERE user_id = ? AND post_id = ?`
	_, err := m.DB.Exec(stmt, userID, postID)
	if err != nil {
		return err
	}
	return nil
}

// GetReaction retrieves the reaction of a user on a specific post
func (m *PostReactionsModel) GetReaction(userID uint, postID uint) (*PostReaction, error) {
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

// GetReactionCount gets the count of reactions (like or dislike) for a post
func (m *PostReactionsModel) GetReactionCount(postID uint, reactionType string) (int, error) {
	stmt := `SELECT COUNT(*) FROM Post_Reactions WHERE post_id = ? AND type = ?`
	var count int
	err := m.DB.QueryRow(stmt, postID, reactionType).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}