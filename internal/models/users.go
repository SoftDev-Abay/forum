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
	UpdateRole(id int, role string) error
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
	stmt := `SELECT id, email, username, password, enabled, role FROM users`

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

func (m *UserModel) EmailExists(email string) (bool, error) {
	stmt := `SELECT EXISTS (SELECT 1 FROM users WHERE email = ?);`

	var exists bool
	err := m.DB.QueryRow(stmt, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (m *UserModel) UsernameExists(username string) (bool, error) {
	stmt := `SELECT EXISTS (SELECT 1 FROM users WHERE username = ?);`

	var exists bool
	err := m.DB.QueryRow(stmt, username).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (m *UserModel) Insert(email string, username string, password string, enabled bool) (int, error) {
	emailExists, err := m.EmailExists(email)
	if err != nil {
		return 0, err
	}
	if emailExists {
		return 0, ErrDuplicateEmail
	}

	usernameExists, err := m.UsernameExists(username)
	if err != nil {
		return 0, err
	}
	if usernameExists {
		return 0, ErrDuplicateUsername
	}

	stmt := `INSERT INTO users (email, username, password, enabled)
	VALUES(?, ?, ?, ?)`

	result, err := m.DB.Exec(stmt, email, username, password, enabled)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *UserModel) UpdateRole(id int, role string) error {
	stmt := `UPDATE users SET role = ? WHERE id = ?`

	_, err := m.DB.Exec(stmt, role, id)
	if err != nil {
		return err
	}

	return nil
}
