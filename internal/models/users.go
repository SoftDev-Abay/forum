package models

import (
	"database/sql"
	"errors"
)

type UserModelInterface interface {
	GetById(id int) (*User, error)
	GetByToken(token string) (*User, error)
	GetAll() ([]*User, error)
	Insert(email string, username string, password string, enabled bool) (int, error)
	GetByUsernameOrEmail(column string) (*User, error)
}

type User struct {
	ID       int
	Username string
	Password string
	Email    string
	Role     string
	Enabled  bool
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) GetByUsernameOrEmail(column string) (*User, error) {
	stmt := `SELECT id, email, username, password, enabled, role FROM users
	WHERE username = ? OR email = ?`

	u := &User{}

	err := m.DB.QueryRow(stmt, column, column).Scan(&u.ID, &u.Email, &u.Username, &u.Password, &u.Enabled, &u.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return u, nil
}

func (m *UserModel) GetById(id int) (*User, error) {
	stmt := `SELECT id, email, username, password, enabled, role FROM users
	WHERE id = ?`

	u := &User{}

	err := m.DB.QueryRow(stmt, id).Scan(&u.ID, &u.Email, &u.Username, &u.Password, &u.Enabled, &u.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return u, nil
}

func (m *UserModel) GetByToken(token string) (*User, error) {
	stmt := `SELECT u.id, u.email, u.username, u.password, u.enabled, u.role
	FROM users u
	INNER JOIN Sessions s on s.user_id = u.id
	WHERE s.token = ? and s.expiresAt > datetime('now')`

	u := &User{}

	err := m.DB.QueryRow(stmt, token).Scan(&u.ID, &u.Email, &u.Username, &u.Password, &u.Enabled, &u.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return u, nil
}

func (m *UserModel) GetAll() ([]*User, error) {
	stmt := `SELECT id, email, username, password, enabled, &u.Role FROM users`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []*User{}

	for rows.Next() {
		u := &User{}

		err = rows.Scan(&u.ID, &u.Email, &u.Username, &u.Password, &u.Enabled, &u.Role)
		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Check if the email exists in the database
func (m *UserModel) EmailExists(email string) (bool, error) {
	stmt := `SELECT EXISTS (SELECT 1 FROM users WHERE email = ?);`

	var exists bool
	err := m.DB.QueryRow(stmt, email).Scan(&exists)
	if err != nil {
		return false, err // return the error if any
	}

	return exists, nil
}

// Check if the username exists in the database
func (m *UserModel) UsernameExists(username string) (bool, error) {
	stmt := `SELECT EXISTS (SELECT 1 FROM users WHERE username = ?);`

	var exists bool
	err := m.DB.QueryRow(stmt, username).Scan(&exists)
	if err != nil {
		return false, err // return the error if any
	}

	return exists, nil
}

// Insert a new user into the database
func (m *UserModel) Insert(email string, username string, password string, enabled bool) (int, error) {
	// Check if email or username already exists
	emailExists, err := m.EmailExists(email)
	if err != nil {
		return 0, err // return the error from EmailExists if any
	}
	if emailExists {
		return 0, ErrDuplicateEmail // return duplicate email error if found
	}

	usernameExists, err := m.UsernameExists(username)
	if err != nil {
		return 0, err // return the error from UsernameExists if any
	}
	if usernameExists {
		return 0, ErrDuplicateUsername // return duplicate username error if found
	}

	// Proceed with insertion if no duplicates found
	stmt := `INSERT INTO users (email, username, password, enabled)
	VALUES(?, ?, ?, ?)`

	result, err := m.DB.Exec(stmt, email, username, password, enabled)
	if err != nil {
		return 0, err // return error from Exec if any
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err // return error from LastInsertId if any
	}

	return int(id), nil // return the user ID after successful insertion
}
