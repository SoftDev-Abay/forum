package models

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"
)

type CommentsModelInterface interface {
	Insert(postID int, userID int, text string, created_at time.Time) (int, error)
	GetAllByPostIdAndUserId(userId int, postId int) ([]*CommentReaction, error)
	UpdateCommentLikeDislikeCounts(commentID int, likeCount int, dislikeCount int) error
	Get(id int) (*Comment, error)
	DeleteCommentsByPostId(postID int) error
	GetAllByPostId(postId int) ([]*Comment, error)
	DeleteCommentById(id int) error
	GetAllCommentsReactionsByPostID(postID int, userID int) ([]*CommentReaction, error)
	GetAllByUserId(userId int) ([]*CommentPostAddition, error)
}

type Comment struct {
	ID           int
	PostID       int
	UserID       int
	Text         string
	LikeCount    int
	DislikeCount int
	CreatedAt    time.Time
}

type CommentAdditionals struct {
	Username string
}

type CommentReaction struct {
	Comment
	CommentAdditionals
	IsLiked    bool
	IsDisliked bool
}

type CommentPostAddition struct {
	Comment
	PostTitle string
}

type CommentsModel struct {
	DB *sql.DB
}

func (m *CommentsModel) Get(id int) (*Comment, error) {
	stmt := `SELECT id, post_id, user_id, text, like_count, dislike_count, created_at
	         FROM Comments
	         WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)

	comment := &Comment{}

	err := row.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Text, &comment.LikeCount, &comment.DislikeCount, &comment.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return comment, nil
}

func (m *CommentsModel) Insert(postID int, userID int, text string, created_at time.Time) (int, error) {
	stmt := `INSERT INTO Comments (post_id, user_id, text, like_count, dislike_count, created_at)
	VALUES (?, ?, ?, 0, 0, ?)`

	result, err := m.DB.Exec(stmt, postID, userID, text, created_at)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *CommentsModel) GetAllByPostIdAndUserId(userId int, postId int) ([]*CommentReaction, error) {
	stmt := `SELECT c.id, c.post_id, c.user_id, c.text, c.like_count, c.dislike_count, c.created_at, cr.type as reaction
			FROM Comments c
			LEFT JOIN Comment_Reactions cr on cr.comment_id = c.id
			WHERE c.post_id = ? AND c.user_id = ?
			ORDER BY c.created_at ASC`

	rows, err := m.DB.Query(stmt, postId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*CommentReaction

	for rows.Next() {
		comment := &CommentReaction{}
		var reaction sql.NullString

		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Text, &comment.LikeCount, &comment.DislikeCount, &comment.CreatedAt, &reaction)
		if err != nil {
			return nil, err
		}

		if reaction.Valid {
			switch reaction.String {
			case "like":
				comment.IsDisliked = false
				comment.IsLiked = true
			case "dislike":
				comment.IsDisliked = true
				comment.IsLiked = false
			default:
				comment.IsDisliked = false
				comment.IsLiked = false
			}
		} else {

			comment.IsDisliked = false
			comment.IsLiked = false
		}

		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (m *CommentsModel) UpdateCommentLikeDislikeCounts(commentID int, likeCount int, dislikeCount int) error {
	stmt := `UPDATE Comments SET like_count = ?, dislike_count = ? WHERE id = ?`
	_, err := m.DB.Exec(stmt, likeCount, dislikeCount, commentID)
	if err != nil {
		return err
	}

	return nil
}

func (m *CommentsModel) GetAllByPostId(postId int) ([]*Comment, error) {
	stmt := `SELECT c.id, c.post_id, c.user_id, c.text, c.like_count, c.dislike_count, c.created_at
			FROM Comments c
			WHERE c.post_id = ? `

	rows, err := m.DB.Query(stmt, postId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment

	for rows.Next() {
		comment := &Comment{}

		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Text, &comment.LikeCount, &comment.DislikeCount, &comment.CreatedAt)
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

func (m *CommentsModel) GetAllCommentsReactionsByPostID(postID int, userID int) ([]*CommentReaction, error) {
	stmt := `SELECT 
				c.id AS comment_id, 
				c.post_id, 
				c.user_id, 
				u.username,  
				c.text, 
				c.like_count, 
				c.dislike_count, 
				c.created_at,
				GROUP_CONCAT(cr.type || ':' || cr.user_id, ', ') AS reactions
			FROM 
				Comments c
			INNER JOIN 
				Users u ON u.id = c.user_id
			LEFT JOIN 
				Comment_Reactions cr ON cr.comment_id = c.id
			WHERE 
				c.post_id = ?
			GROUP BY 
				c.id, c.post_id, c.user_id, u.username, c.text, c.like_count, c.dislike_count, c.created_at
			ORDER BY 
				c.created_at ASC;
`

	rows, err := m.DB.Query(stmt, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commentReactions []*CommentReaction

	for rows.Next() {
		var reactions sql.NullString

		comment := &CommentReaction{}

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Username,
			&comment.Text,
			&comment.LikeCount,
			&comment.DislikeCount,
			&comment.CreatedAt,
			&reactions,
		)
		if err != nil {
			return nil, err
		}

		comment.IsLiked = false
		comment.IsDisliked = false

		if reactions.Valid {
			reactionEntries := strings.Split(reactions.String, ", ")
			for _, entry := range reactionEntries {
				parts := strings.Split(entry, ":")
				if len(parts) == 2 {
					reactionType := parts[0]
					reactionUserId, err := strconv.Atoi(parts[1])
					
					if err != nil {
						return nil, err
					}

					if reactionUserId == userID {
						if reactionType == "like" {
							comment.IsLiked = true
						} else if reactionType == "dislike" {
							comment.IsDisliked = true
						}
					}
				}
			}
		}

		commentReactions = append(commentReactions, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return commentReactions, nil
}

func (m *CommentsModel) DeleteCommentsByPostId(postID int) error {
	stmt := `DELETE FROM Comments WHERE post_id = ?`
	_, err := m.DB.Exec(stmt, postID)
	if err != nil {
		return err
	}

	return nil
}

func (m *CommentsModel) DeleteCommentById(id int) error {
	stmt := `DELETE FROM Comments WHERE id = ?`
	_, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *CommentsModel) GetAllByUserId(userId int) ([]*CommentPostAddition, error) {
	stmt := `SELECT c.id, c.post_id, c.user_id, c.text, c.like_count, c.dislike_count, c.created_at, p.title
			FROM Comments c
			INNER JOIN Posts p ON p.id = c.post_id
			WHERE user_id = ?`

	rows, err := m.DB.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*CommentPostAddition

	for rows.Next() {
		comment := &CommentPostAddition{}

		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Text, &comment.LikeCount, &comment.DislikeCount, &comment.CreatedAt, &comment.PostTitle)
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
