package models

import (
	"database/sql"
	"errors"
)

type CommentsReactionsModelInterface interface {
	AddReaction(userID int, CommentID int, reactionType string) error
	UpdateReaction(userID int, CommentID int, reactionType string) error
	DeleteReaction(userID int, CommentID int) error
	GetReaction(userID int, CommentID int) (*CommentsReactions, error)
	GetReactionCount(CommentID int, reactionType string) (int, error)
	GetReactionByUserID(userID int) (*CommentsReactions, error)
}

type CommentsReactions struct {
	UserID    int
	CommentID int
	Type      string // 'like' or 'dislike'
}

type CommentsReactionsModel struct {
	DB *sql.DB
}

// AddReaction adds a new reaction (like or dislike) for a user on a post
func (m *CommentsReactionsModel) AddReaction(userID int, CommentID int, reactionType string) error {
	// Ensure valid reaction type
	if reactionType != "like" && reactionType != "dislike" {
		return errors.New("invalid reaction type")
	}

	// Check if the user has already reacted to this post
	existingReaction, err := m.GetReaction(userID, CommentID)
	if err != nil && err != ErrNoReaction {
		return err
	}

	// If a reaction exists, return error (because they should not be adding another reaction)
	if existingReaction != nil {
		return ErrReactionAlreadyExists // user already reacted to this post
	}

	// Insert the new reaction
	stmt := `INSERT INTO Comment_Reactions (user_id, comment_id, type) VALUES (?, ?, ?)`
	_, err = m.DB.Exec(stmt, userID, CommentID, reactionType)
	if err != nil {
		return err
	}

	return nil
}

func (m *CommentsReactionsModel) UpdateReaction(userID int, CommentID int, reactionType string) error {
	// Ensure valid reaction type
	if reactionType != "like" && reactionType != "dislike" {
		return errors.New("invalid reaction type")
	}

	// Check if the user has already reacted to this post
	_, err := m.GetReaction(userID, CommentID)
	if err != nil {
		if err == ErrNoReaction {
			return errors.New("cannot update reaction: no reaction found")
		}
		return err
	}

	// Update the reaction if it exists
	stmt := `UPDATE Comment_Reactions SET type = ? WHERE user_id = ? AND comment_id = ?`
	_, err = m.DB.Exec(stmt, reactionType, userID, CommentID)
	if err != nil {
		return err
	}

	return nil
}

// DeleteReaction removes a reaction (like or dislike) from a user on a post
func (m *CommentsReactionsModel) DeleteReaction(userID int, CommentID int) error {
	stmt := `DELETE FROM Comment_Reactions WHERE user_id = ? AND comment_id = ?`
	_, err := m.DB.Exec(stmt, userID, CommentID)
	if err != nil {
		return err
	}
	return nil
}

// GetReaction retrieves the reaction of a user on a specific post
func (m *CommentsReactionsModel) GetReaction(userID int, CommentID int) (*CommentsReactions, error) {
	stmt := `SELECT type FROM Comment_Reactions WHERE user_id = ? AND comment_id = ?`
	row := m.DB.QueryRow(stmt, userID, CommentID)

	var reactionType string
	err := row.Scan(&reactionType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoReaction
		}
		return nil, err
	}

	return &CommentsReactions{
		UserID:    userID,
		CommentID: CommentID,
		Type:      reactionType,
	}, nil
}

func (m *CommentsReactionsModel) GetReactionByUserID(userID int) (*CommentsReactions, error) {
	stmt := `SELECT type FROM Comment_Reactions WHERE user_id = ?`
	row := m.DB.QueryRow(stmt, userID)

	var reactionType string
	err := row.Scan(&reactionType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoReaction
		}
		return nil, err
	}

	return &CommentsReactions{
		UserID: userID,
		Type:   reactionType,
	}, nil
}

// GetReactionCount gets the count of reactions (like or dislike) for a post
func (m *CommentsReactionsModel) GetReactionCount(CommentID int, reactionType string) (int, error) {
	stmt := `SELECT COUNT(*) FROM Comment_Reactions WHERE comment_id = ? AND type = ?`
	var count int
	err := m.DB.QueryRow(stmt, CommentID, reactionType).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
