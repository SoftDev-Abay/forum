package models

import (
	"database/sql"
	"time"
)

// Notifications represents a row in the Notifications table.
type Notifications struct {
	ID           int
	Type         string
	Actor_ID     int
	Recipient_ID int
	Post_ID      int
	Comment_ID   sql.NullInt64 // or *int if you prefer
	Created_at   time.Time
	Is_read      bool
}

// NotificationsModelInterface defines the methods we'll implement in the model.
type NotificationsModelInterface interface {
	Insert(notificationType string, actorID, recipientID, postID int, commentID *int) (int, error)
	GetAllByRecipient(userID int) ([]*Notifications, error)
	MarkAsRead(notificationID int) error
	GetUsersNorificationsCount(recipientID int) (int, error)
	MarkAllAsReadByUser(userID int) error
}

// NotificationsModel implements NotificationsModelInterface.
type NotificationsModel struct {
	DB *sql.DB
}

// Insert creates a new notification row.
// commentID is optional; if it's nil, we insert NULL in the comment_id column.
func (m *NotificationsModel) Insert(notificationType string, actorID, recipientID, postID int, commentID *int) (int, error) {
	stmt := `
        INSERT INTO Notifications 
            (type, actor_id, recipient_id, post_id, comment_id, created_at, is_read)
        VALUES (?, ?, ?, ?, ?, datetime('now'), 0)
    `

	var commentArg interface{}
	if commentID == nil {
		commentArg = nil // Insert NULL for comment_id
	} else {
		commentArg = *commentID
	}

	result, err := m.DB.Exec(stmt, notificationType, actorID, recipientID, postID, commentArg)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// GetAllByRecipient fetches all notifications for a given user.
// You might want ORDER BY created_at DESC or filter unread only, etc.
func (m *NotificationsModel) GetAllByRecipient(userID int) ([]*Notifications, error) {
	stmt := `
        SELECT 
            id, type, actor_id, recipient_id, post_id, comment_id, created_at, is_read
        FROM Notifications
        WHERE recipient_id = ?
        ORDER BY created_at DESC
    `
	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*Notifications
	for rows.Next() {
		n := &Notifications{
			Comment_ID: sql.NullInt64{}, // by default
		}
		err := rows.Scan(
			&n.ID,
			&n.Type,
			&n.Actor_ID,
			&n.Recipient_ID,
			&n.Post_ID,
			&n.Comment_ID,
			&n.Created_at,
			&n.Is_read,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

// MarkAsRead sets is_read=1 for a specific notification ID.
func (m *NotificationsModel) MarkAsRead(notificationID int) error {
	stmt := `UPDATE Notifications SET is_read = 1 WHERE id = ?`
	_, err := m.DB.Exec(stmt, notificationID)
	return err
}

func (m *NotificationsModel) GetUsersNorificationsCount(recipientID int) (int, error) {
	stmt := `
		SELECT COUNT(*) FROM Notifications
		WHERE recipient_id = ? AND is_read = 0
	`
	var count int
	err := m.DB.QueryRow(stmt, recipientID).Scan(&count)

	return count, err
}

func (m *NotificationsModel) MarkAllAsReadByUser(userID int) error {
    stmt := `UPDATE Notifications SET is_read = 1 WHERE recipient_id = ?`
    _, err := m.DB.Exec(stmt, userID)
    return err
}
