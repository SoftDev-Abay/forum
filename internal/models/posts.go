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
	Get(id int) (*Post, error)
	Latest() ([]*Post, error)
	UpdatePostLikeDislikeCounts(postID int, likeCount int, dislikeCount int) error
	GetPostsByUserID(userID int) ([]*Post, error)
	GetPostsByIDs(postIDs []int) ([]*Post, error)
	GetFilteredPosts(userID, categoryID, page, pageSize int) ([]*PostByUser, error)
	CountPosts(categoryID int) (int, error)
	DeletePostById(id int) error
	UpdatePost(id int, title, content, imgUrl string, categoryID int) error
}

type Post struct {
	ID           int
	Title        string
	Content      string
	ImgUrl       string
	CreatedAt    time.Time
	CategoryID   int
	OwnerID      int
	LikeCount    int
	DislikeCount int
}

type PostAdditionals struct {
	CommentCount int
	CategoryName string
	OwnerName    string
}

type PostUserAdditionals struct {
	IsLiked    bool
	IsDisliked bool
}

type PostByUser struct {
	Post
	PostAdditionals
	PostUserAdditionals
}

type PostModel struct {
	DB                 *sql.DB
	PostReactionsModel *PostReactionsModel 
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

func (m *PostModel) Get(id int) (*Post, error) {
	stmt := `SELECT id, title, content, imgUrl, createdAt, category_id, owner_id, like_count, dislike_count
	         FROM Posts
	         WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)

	post := &Post{}

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

func (m *PostModel) Latest() ([]*Post, error) {
	stmt := `SELECT id, title, content, imgUrl, createdAt, category_id, owner_id, like_count, dislike_count
	         FROM Posts
	         ORDER BY createdAt ASC
	         LIMIT 10`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post

	for rows.Next() {
		post := &Post{}
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

func (m *PostModel) UpdatePostLikeDislikeCounts(postID int, likeCount int, dislikeCount int) error {
	stmt := `UPDATE Posts SET like_count = ?, dislike_count = ? WHERE id = ?`
	_, err := m.DB.Exec(stmt, likeCount, dislikeCount, postID)
	if err != nil {
		return err
	}

	return nil
}

func (m *PostModel) GetPostsByUserID(userID int) ([]*Post, error) {
	stmt := `SELECT id, title, content, imgUrl, createdAt, category_id, owner_id, like_count, dislike_count
			FROM Posts 
			WHERE owner_id = ?
			ORDER BY createdAt ASC`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post

	for rows.Next() {
		post := &Post{}
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

func (m *PostModel) GetPostsByIDs(postIDs []int) ([]*Post, error) {
	if len(postIDs) == 0 {
		return []*Post{}, nil
	}

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

	var posts []*Post
	for rows.Next() {
		post := &Post{}
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

func (m *PostModel) GetFilteredPosts(userID int, categoryID, page, pageSize int) ([]*PostByUser, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	query := `
		SELECT p.id, p.title, p.content, p.imgUrl, p.createdAt, p.category_id, cat.name as category_name, u.id AS owner_id, u.username AS owner_name,
			   p.like_count, p.dislike_count, 
			   CASE WHEN ? > 0 AND pr.type = 'like' THEN 1 ELSE 0 END AS is_liked,
			   CASE WHEN ? > 0 AND pr.type = 'dislike' THEN 1 ELSE 0 END AS is_disliked,
			   COUNT(c.id) AS comment_count
		FROM Posts AS p
		INNER JOIN Users AS u ON p.owner_id = u.id
		INNER JOIN Categories AS cat ON p.category_id = cat.id
		LEFT JOIN Comments AS c ON p.id = c.post_id
		LEFT JOIN Post_Reactions AS pr ON p.id = pr.post_id AND pr.user_id = ?
	`

	var args []interface{}
	args = append(args, userID, userID, userID) 
	var whereClauses []string

	if categoryID > 0 {
		whereClauses = append(whereClauses, "p.category_id = ?")
		args = append(args, categoryID)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += `
		GROUP BY p.id, p.title, p.content, p.imgUrl, p.createdAt, p.category_id, 
		         p.owner_id, p.like_count, p.dislike_count
		ORDER BY p.createdAt DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := m.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*PostByUser
	for rows.Next() {
		post := &PostByUser{}
		var isLiked, isDisliked int
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.ImgUrl, &post.CreatedAt,
			&post.CategoryID, &post.CategoryName, &post.OwnerID, &post.OwnerName, &post.LikeCount, &post.DislikeCount,
			&isLiked, &isDisliked, &post.CommentCount)
		if err != nil {
			return nil, err
		}
		post.IsLiked = isLiked == 1
		post.IsDisliked = isDisliked == 1

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

	if categoryID > 0 {
		whereClauses = append(whereClauses, "category_id = ?")
		args = append(args, categoryID)
	}

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
	stmt := `DELETE FROM Posts WHERE id = ?`
	_, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *PostModel) UpdatePost(id int, title, content, imgUrl string, categoryID int) error {
	query := "UPDATE Posts SET "
	args := []interface{}{}

	if title != "" {
		query += "title = ?, "
		args = append(args, title)
	}
	if content != "" {
		query += "content = ?, "
		args = append(args, content)
	}
	if imgUrl != "" {
		query += "imgUrl = ?, "
		args = append(args, imgUrl)
	}
	if categoryID != 0 {
		query += "category_id = ?, "
		args = append(args, categoryID)
	}

	query = query[:len(query)-2]

	query += " WHERE id = ?"
	args = append(args, id)

	_, err := m.DB.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}
