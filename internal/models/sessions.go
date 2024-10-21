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
	Insert(token string, userId uint) (int, error)
}

type Session struct {
	ID        uint
	Token     string
	UserID    uint
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Define a UserModel type which wraps a sql.DB connection pool.
type SessionModel struct {
	DB *sql.DB
}

func (m *SessionModel) Insert(token string, userId uint) (int, error) {
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
	// Write the SQL statement we want to execute. Again, I've split it over two
	// lines for readability.
	// stmt := `SELECT id, title, content, created, expires FROM Users
	// WHERE expires > datetime('now') AND id = ?`

	stmt := `SELECT id, token, user_id, createdAt, expiresAt FROM sessions
	WHERE id = ?`

	// Use the QueryRow() method on the connection pool to execute our
	// SQL statement, passing in the untrusted id variable as the value for the
	// placeholder parameter. This returns a pointer to a sql.Row object which
	// holds the result from the database.
	// Initialize a pointer to a new zeroed User struct.
	s := &Session{}
	// Use row.Scan() to copy the values from each field in sql.Row to the
	// corresponding field in the User struct. Notice that the arguments
	// to row.Scan are *pointers* to the place you want to copy the data into,
	// and the number of arguments must be exactly the same as the number of
	// columns returned by your statement.
	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.Token, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		// If the query returns no rows, then row.Scan() will return a
		// sql.ErrNoRows error. We use the errors.Is() function check for that
		// error specifically, and return our own ErrNoRecord error
		// instead (we'll create this in a moment).
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	// If everything went OK then return the Snippet object.
	return s, nil
}


func (m *SessionModel) GetLastUserSession(id int) (*Session, error) {
	// Write the SQL statement we want to execute. Again, I've split it over two
	// lines for readability.
	// stmt := `SELECT id, title, content, created, expires FROM Users
	// WHERE expires > datetime('now') AND id = ?`

	stmt := `SELECT id, token, user_id, createdAt, expiresAt FROM sessions
	WHERE user_id = ?
	order by expiresAt
	limit 1
	`

	// Use the QueryRow() method on the connection pool to execute our
	// SQL statement, passing in the untrusted id variable as the value for the
	// placeholder parameter. This returns a pointer to a sql.Row object which
	// holds the result from the database.
	// Initialize a pointer to a new zeroed User struct.
	s := &Session{}
	// Use row.Scan() to copy the values from each field in sql.Row to the
	// corresponding field in the User struct. Notice that the arguments
	// to row.Scan are *pointers* to the place you want to copy the data into,
	// and the number of arguments must be exactly the same as the number of
	// columns returned by your statement.
	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.Token, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		// If the query returns no rows, then row.Scan() will return a
		// sql.ErrNoRows error. We use the errors.Is() function check for that
		// error specifically, and return our own ErrNoRecord error
		// instead (we'll create this in a moment).
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	// If everything went OK then return the Snippet object.
	return s, nil
}
