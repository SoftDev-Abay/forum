package models

import (
	"database/sql"
	"errors"
	"time"
	// ""
)

// id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
// token TEXT NOT NULL UNIQUE,
// user_id INTEGER NOT NULL UNIQUE,
// createdAt DATETIME NOT NULL,
// expiresAt DATETIME NOT NULL,
// FOREIGN KEY (user_id) REFERENCES Users(id)

type SessionModelInterface interface {
	GetById(id int) (*Session, error)
	Insert(token string, userId int) (int, error)
	GetLastUserSession(id int) (*Session, error)
	GetUserIDByToken(token string) (int, error)
	DeleteByToken(token string) error
}

type Session struct {
	ID        int
	Token     string
	UserID    int
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Define a UserModel type which wraps a sql.DB connection pool.
type SessionModel struct {
	DB *sql.DB
}

func (m *SessionModel) Insert(token string, userId int) (int, error) {
	// Write the SQL statement we want to execute. I've split it over two lines
	// for readability (which is why it's surrounded with backquotes instead
	// of normal double quotes).

	stmt := `INSERT INTO sessions (token, user_id, createdAt, expiresAt)
	VALUES(?, ?, datetime('now'), datetime('now',  '1 days'))`
	// Use the Exec() method on the embedded connection pool to execute the
	// statement. The first parameter is the SQL statement, followed by the
	// title, content and expiry values for the placeholder parameters. This
	// method returns a sql.Result type, which contains some basic
	// information about what happened when the statement was executed.

	result, err := m.DB.Exec(stmt, token, userId) // db.Exec is first creating a prepared statement
	// which is bascially sql query compiled but without paramentrs,
	//  this way parameters are treared as pure data, thus they cant change the intent of the request
	// this is better than just putting paraments into the sql string
	if err != nil {
		return 0, err
	}

	// Use the LastInsertId() method on the result to get the ID of our
	// newly inserted record in the users table.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	// The ID returned has the type int64, so we convert it to an int type
	// before returning.
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
