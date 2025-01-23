package models

import (
	"database/sql"
	"errors"
	"time"
)

type SessionModelInterface interface {
	GetById(id int) (*Session, error)
	Insert(token string, userId int) (int, error)
	GetLastUserSession(id int) (*Session, error)
	GetUserIDByToken(token string) (int, error)
	DeleteByToken(token string) error
	DeleteByUserId(userId int) error
	GetByUserId(userId int) (*Session, error)
}

type Session struct {
	ID        int
	Token     string
	UserID    int
	CreatedAt time.Time
	ExpiresAt time.Time
}

type SessionModel struct {
	DB *sql.DB
}

func (m *SessionModel) Insert(token string, userId int) (int, error) {
	stmt := `INSERT INTO sessions (token, user_id, createdAt, expiresAt)
	VALUES(?, ?, datetime('now'), datetime('now',  '1 days'))`

	result, err := m.DB.Exec(stmt, token, userId)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *SessionModel) GetById(id int) (*Session, error) {
	stmt := `SELECT id, token, user_id, createdAt, expiresAt FROM sessions
	WHERE id = ?`

	s := &Session{}
	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.Token, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return s, nil
}

func (m *SessionModel) GetByUserId(userId int) (*Session, error) {
	stmt := `SELECT id, token, user_id, createdAt, expiresAt FROM sessions
	WHERE user_id = ?`

	s := &Session{}
	err := m.DB.QueryRow(stmt, userId).Scan(&s.ID, &s.Token, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return s, nil
}

func (m *SessionModel) GetLastUserSession(id int) (*Session, error) {
	stmt := `SELECT id, token, user_id, createdAt, expiresAt FROM sessions
	WHERE user_id = ?
	order by expiresAt
	limit 1
	`
	s := &Session{}

	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.Token, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return s, nil
}

func (m *SessionModel) GetUserIDByToken(token string) (int, error) {
	var userID int
	stmt := `SELECT user_id FROM Sessions WHERE token = ? AND expiresAt > CURRENT_TIMESTAMP`

	err := m.DB.QueryRow(stmt, token).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("session not found or expired")
		}
		return 0, err
	}

	return userID, nil
}

func (m *SessionModel) DeleteByToken(token string) error {
	stmt := `DELETE FROM Sessions WHERE token = ?`

	_, err := m.DB.Exec(stmt, token)
	if err != nil {
		return err
	}

	return nil
}


func (m *SessionModel) DeleteByUserId(userId int) error {
	stmt := `DELETE FROM Sessions WHERE user_id = ?`

	_, err := m.DB.Exec(stmt, userId)
	if err != nil {
		return err
	}

	return nil
}