package models

import (
	"database/sql"
	"errors"
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
	GetAllCommentsReactionsByPostID(postID int) ([]*CommentReaction, error)
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

type CommentReaction struct {
	Comment
	IsLiked    bool
	IsDisliked bool
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

		// Check if the reaction is valid before switching
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
			// If the reaction is null, set both to false
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



// updatePostLikeDislikeCounts recalculates the like/dislike counts for a post
func (m *CommentsModel) UpdateCommentLikeDislikeCounts(commentID int, likeCount int, dislikeCount int) error {
	// Update the post's like/dislike counts in the database
	stmt := `UPDATE Comments SET like_count = ?, dislike_count = ? WHERE id = ?`
	_, err := m.DB.Exec(stmt, likeCount, dislikeCount, commentID)
	if err != nil {
		return err
	}

	return nil
}

// so I need to write a query that will identify get whether a comment was liked or not
// find a comment reaction
// exists:
// 1) is liked
// 2) is disliked
// doesnt exist:
// 1) is null

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

func (m *CommentsModel) GetAllCommentsReactionsByPostID(postID int) ([]*CommentReaction, error) {
	stmt := `SELECT c.id, c.post_id, c.user_id, c.text, c.like_count, c.dislike_count, c.created_at,
                    cr.type as reaction, cr.user_id as reaction_user_id
             FROM Comments c
             LEFT JOIN Comment_Reactions cr ON cr.comment_id = c.id
             WHERE c.post_id = ?
             ORDER BY c.created_at ASC`

	rows, err := m.DB.Query(stmt, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commentReactions []*CommentReaction

	for rows.Next() {
		var reaction sql.NullString
		var reactionUserID sql.NullInt64

		comment := &CommentReaction{}

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Text,
			&comment.LikeCount,
			&comment.DislikeCount,
			&comment.CreatedAt,
			&reaction,
			&reactionUserID,
		)
		if err != nil {
			return nil, err
		}

		// Initialize IsLiked and IsDisliked based on the reaction
		comment.IsLiked = reaction.Valid && reaction.String == "like"
		comment.IsDisliked = reaction.Valid && reaction.String == "dislike"

		commentReactions = append(commentReactions, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return commentReactions, nil
}


// updatePostLikeDislikeCounts recalculates the like/dislike counts for a post
func (m *CommentsModel) DeleteCommentsByPostId(postID int) error {
	// Update the post's like/dislike counts in the database
	stmt := `DELETE FROM Comments WHERE post_id = ?`
	_, err := m.DB.Exec(stmt, postID)
	if err != nil {
		return err
	}

	return nil
}

// updatePostLikeDislikeCounts recalculates the like/dislike counts for a post
func (m *CommentsModel) DeleteCommentById(id int) error {
	// Update the post's like/dislike counts in the database
	stmt := `DELETE FROM Comments WHERE id = ?`
	_, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}

	return nil
}
