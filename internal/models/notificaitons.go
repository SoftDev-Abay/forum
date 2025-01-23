package models

import (
	"database/sql"
	"time"
)

type Notifications struct {
	ID           int
	Type         string
	Actor_ID     int
	Recipient_ID int
	Post_ID      int
	Comment_ID   sql.NullInt64 
	Created_at   time.Time
	Is_read      bool
}

type NotificationsModelInterface interface {
	Insert(notificationType string, actorID, recipientID, postID int, commentID *int) (int, error)
	GetAllByRecipient(userID int) ([]*Notifications, error)
	MarkAsRead(notificationID int) error
	GetUsersNorificationsCount(recipientID int) (int, error)
	MarkAllAsReadByUser(userID int) error
}

type NotificationsModel struct {
	DB *sql.DB
}

func (m *NotificationsModel) Insert(notificationType string, actorID, recipientID, postID int, commentID *int) (int, error) {
	stmt := `
        INSERT INTO Notifications 
            (type, actor_id, recipient_id, post_id, comment_id, created_at, is_read)
        VALUES (?, ?, ?, ?, ?, datetime('now'), 0)
    `

	var commentArg interface{}
	if commentID == nil {
		commentArg = nil
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
			Comment_ID: sql.NullInt64{}, 
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
