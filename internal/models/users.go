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
	ID       uint
	Username string
	Password string
	Email    string
	Enabled  bool
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) GetByUsernameOrEmail(column string) (*User, error) {
	stmt := `SELECT id, email, username, password, enabled FROM users
	WHERE username = ? OR email = ?`

	u := &User{}

	err := m.DB.QueryRow(stmt, column, column).Scan(&u.ID, &u.Email, &u.Username, &u.Password, &u.Enabled)
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
	stmt := `SELECT id, email, username, password, enabled FROM users
	WHERE id = ?`

	u := &User{}

	err := m.DB.QueryRow(stmt, id).Scan(&u.ID, &u.Email, &u.Username, &u.Password, &u.Enabled)
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
	stmt := `SELECT u.id, u.email, u.username, u.password, u.enabled 
	FROM users u
	INNER JOIN Sessions s on s.user_id = u.id
	WHERE s.token = ? and s.expiresAt > datetime('now')`

	u := &User{}

	err := m.DB.QueryRow(stmt, token).Scan(&u.ID, &u.Email, &u.Username, &u.Password, &u.Enabled)
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
	stmt := `SELECT id, email, username, password, enabled FROM users`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []*User{}

	for rows.Next() {
		u := &User{}

		err = rows.Scan(&u.ID, &u.Email, &u.Username, &u.Password, &u.Enabled)
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

func (m *UserModel) Insert(email string, username string, password string, enabled bool) (int, error) {
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
