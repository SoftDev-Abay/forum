package models

import (
	"database/sql"
	"errors"
	"time"
)

type PostsModelInterface interface {
	Insert(title string, content string, imgUrl string, createdAt time.Time, categoryID int, ownerID int) (int, error)
	Get(id int) (*Posts, error)
	Latest() ([]*Posts, error)
	GetPostsByUserID(userID int) ([]*Posts, error)
}

type Posts struct {
	ID           int
	Title        string
	Content      string
	ImgUrl       string
	CreatedAt    time.Time
	CategoryID   int
	OwnerID      int
	LikeCount    int
	DislikeCount int
	IsLiked      bool // Tracks whether the logged-in user has liked the post
	IsDisliked   bool // Tracks whether the logged-in user has disliked the post
}

type PostModel struct {
	DB                 *sql.DB
	PostReactionsModel *PostReactionsModel // Inject PostReactionsModel into PostModel
}

func (m *PostModel) Insert(title string, content string, imgUrl string, createdAt time.Time, categoryID int, ownerID int) (int, error) {
	stmt := `INSERT INTO Posts (title, content, imgUrl, createdAt, category_id, owner_id, like_count, dislike_count)
	         VALUES (?, ?, ?, ?, ?, ?, 0, 0)`

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
	stmt := `SELECT id, title, content, imgUrl, createdAt, category_id, owner_id, like_count, dislike_count
	         FROM Posts
	         WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)

	post := &Posts{}

	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.ImgUrl, &post.CreatedAt, &post.CategoryID, &post.OwnerID, &post.LikeCount, &post.DislikeCount)
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
	stmt := `SELECT id, title, content, imgUrl, createdAt, category_id, owner_id, like_count, dislike_count
	         FROM Posts
	         ORDER BY createdAt ASC
	         LIMIT 10`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Posts

	for rows.Next() {
		post := &Posts{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.ImgUrl, &post.CreatedAt, &post.CategoryID, &post.OwnerID, &post.LikeCount, &post.DislikeCount)
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

// updatePostLikeDislikeCounts recalculates the like/dislike counts for a post
func (m *PostModel) updatePostLikeDislikeCounts(postID int) error {
	likes, err := m.PostReactionsModel.GetReactionCount(postID, "like")
	if err != nil {
		return err
	}

	dislikes, err := m.PostReactionsModel.GetReactionCount(postID, "dislike")
	if err != nil {
		return err
	}

	// Update the post's like/dislike counts in the database
	stmt := `UPDATE Posts SET like_count = ?, dislike_count = ? WHERE id = ?`
	_, err = m.DB.Exec(stmt, likes, dislikes, postID)
	if err != nil {
		return err
	}

	return nil
}

func (m *PostModel) GetPostsByUserID(userID int) ([]*Posts, error) {
	stmt := `SELECT id, title, content, imgUrl, createdAt, category_id, owner_id, like_count, dislike_count
			FROM Posts 
			WHERE owner_id = ?
			ORDER BY createdAt ASC`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Posts

	for rows.Next() {
		post := &Posts{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.ImgUrl, &post.CreatedAt, &post.CategoryID, &post.OwnerID, &post.LikeCount, &post.DislikeCount)
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
