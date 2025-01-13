package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type PostsModelInterface interface {
	Insert(title string, content string, imgUrl string, createdAt time.Time, categoryID int, ownerID int) (int, error)
	Get(id int) (*Posts, error)
	Latest() ([]*Posts, error)
	GetPostsByUserID(userID int) ([]*Posts, error)
	UpdatePostLikeDislikeCounts(postID int, likeCount int, dislikeCount int) error
	GetPostsByIDs(postIDs []int) ([]*Posts, error)
	GetFilteredPosts(categoryID, page, pageSize int) ([]*Posts, error)
	CountPosts(categoryID int) (int, error)
	DeletePostById(id int) error
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
func (m *PostModel) UpdatePostLikeDislikeCounts(postID int, likeCount int, dislikeCount int) error {
	// Update the post's like/dislike counts in the database
	stmt := `UPDATE Posts SET like_count = ?, dislike_count = ? WHERE id = ?`
	_, err := m.DB.Exec(stmt, likeCount, dislikeCount, postID)
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

func (m *PostModel) GetPostsByIDs(postIDs []int) ([]*Posts, error) {
	if len(postIDs) == 0 {
		// No liked post IDs, just return empty slice.
		return []*Posts{}, nil
	}

	// Build dynamic placeholders like (?, ?, ?) for the SQL IN clause.
	placeholders := make([]string, len(postIDs))
	args := make([]interface{}, len(postIDs))
	for i, id := range postIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
        SELECT id, title, content, imgUrl, createdAt, category_id, owner_id, like_count, dislike_count
        FROM Posts
        WHERE id IN (%s)
        ORDER BY createdAt ASC
    `, strings.Join(placeholders, ", "))

	rows, err := m.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Posts
	for rows.Next() {
		post := &Posts{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.ImgUrl, &post.CreatedAt,
			&post.CategoryID, &post.OwnerID, &post.LikeCount, &post.DislikeCount)
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

func (m *PostModel) GetFilteredPosts(categoryID, page, pageSize int) ([]*Posts, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// Base query
	query := `
        SELECT id, title, content, imgUrl, createdAt, category_id, owner_id, like_count, dislike_count
        FROM Posts
    `
	var args []interface{}
	var whereClauses []string

	// If filtering by category
	if categoryID > 0 {
		whereClauses = append(whereClauses, "category_id = ?")
		args = append(args, categoryID)
	}

	// If we have any WHERE clauses, add them
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += ` ORDER BY createdAt DESC`
	query += ` LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	rows, err := m.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Posts
	for rows.Next() {
		post := &Posts{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.ImgUrl, &post.CreatedAt,
			&post.CategoryID, &post.OwnerID, &post.LikeCount, &post.DislikeCount)
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

func (m *PostModel) CountPosts(categoryID int) (int, error) {
	query := `SELECT COUNT(*) FROM Posts`
	var args []interface{}
	var whereClauses []string

	// If filtering by category
	if categoryID > 0 {
		whereClauses = append(whereClauses, "category_id = ?")
		args = append(args, categoryID)
	}

	// If we have any WHERE clauses, add them
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var count int
	err := m.DB.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (m *PostModel) DeletePostById(id int) error {
	// Update the post's like/dislike counts in the database
	stmt := `DELETE FROM Posts WHERE id = ?`
	_, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}

	return nil
}
