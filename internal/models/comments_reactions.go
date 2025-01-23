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
	DeleteReactioByCommentId(CommentID int) error
}

type CommentsReactions struct {
	UserID    int
	CommentID int
	Type      string
}

type CommentsReactionsModel struct {
	DB *sql.DB
}

func (m *CommentsReactionsModel) AddReaction(userID int, CommentID int, reactionType string) error {
	if reactionType != "like" && reactionType != "dislike" {
		return errors.New("invalid reaction type")
	}

	existingReaction, err := m.GetReaction(userID, CommentID)
	if err != nil && err != ErrNoReaction {
		return err
	}

	if existingReaction != nil {
		return ErrReactionAlreadyExists
	}

	stmt := `INSERT INTO Comment_Reactions (user_id, comment_id, type) VALUES (?, ?, ?)`
	_, err = m.DB.Exec(stmt, userID, CommentID, reactionType)
	if err != nil {
		return err
	}

	return nil
}

func (m *CommentsReactionsModel) UpdateReaction(userID int, CommentID int, reactionType string) error {
	if reactionType != "like" && reactionType != "dislike" {
		return errors.New("invalid reaction type")
	}

	_, err := m.GetReaction(userID, CommentID)
	if err != nil {
		if err == ErrNoReaction {
			return errors.New("cannot update reaction: no reaction found")
		}
		return err
	}

	stmt := `UPDATE Comment_Reactions SET type = ? WHERE user_id = ? AND comment_id = ?`
	_, err = m.DB.Exec(stmt, reactionType, userID, CommentID)
	if err != nil {
		return err
	}

	return nil
}

func (m *CommentsReactionsModel) DeleteReaction(userID int, CommentID int) error {
	stmt := `DELETE FROM Comment_Reactions WHERE user_id = ? AND comment_id = ?`
	_, err := m.DB.Exec(stmt, userID, CommentID)
	if err != nil {
		return err
	}
	return nil
}

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

func (m *CommentsReactionsModel) GetReactionCount(CommentID int, reactionType string) (int, error) {
	stmt := `SELECT COUNT(*) FROM Comment_Reactions WHERE comment_id = ? AND type = ?`
	var count int
	err := m.DB.QueryRow(stmt, CommentID, reactionType).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *CommentsReactionsModel) DeleteReactioByCommentId(CommentID int) error {
	stmt := `DELETE FROM Comment_Reactions WHERE comment_id = ?`
	_, err := m.DB.Exec(stmt, CommentID)
	if err != nil {
		return err
	}
	return nil
}
