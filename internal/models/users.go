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
	// If everything went OK then return the Snippet object.
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
	// If everything went OK then return the Snippet object.
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
	// If everything went OK then return the Snippet object.
	return u, nil
}

func (m *UserModel) GetAll() ([]*User, error) {
	// Write the SQL statement we want to execute.
	stmt := `SELECT id, email, username, password, enabled FROM users`

	// stmt := `SELECT id, title, content, created, expires FROM users
	// WHERE expires > datetime('now') ORDER BY id DESC LIMIT 10`
	// Use the Query() method on the connection pool to execute our
	// SQL statement. This returns a sql.Rows resultset containing the result of
	// our query.
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	// We defer rows.Close() to ensure the sql.Rows resultset is
	// always properly closed before the Latest() method returns. This defer
	// statement should come *after* you check for an error from the Query()
	// method. Otherwise, if Query() returns an error, you'll get a panic
	// trying to close a nil resultset.
	defer rows.Close()
	// Initialize an empty slice to hold the User structs.
	users := []*User{}
	// Use rows.Next to iterate through the rows in the resultset. This
	// prepares the first (and then each subsequent) row to be acted on by the
	// rows.Scan() method. If iteration over all the rows completes then the
	// resultset automatically closes itself and frees-up the underlying
	// database connection.
	for rows.Next() {
		// Create a pointer to a new zeroed User struct.
		u := &User{}
		// Use rows.Scan() to copy the values from each field in the row to the
		// new User object that we created. Again, the arguments to row.Scan()
		// must be pointers to the place you want to copy the data into, and the
		// number of arguments must be exactly the same as the number of
		// columns returned by your statement.
		err = rows.Scan(&u.ID, &u.Email, &u.Username, &u.Password, &u.Enabled)
		if err != nil {
			return nil, err
		}
		// Append it to the slice of users.
		users = append(users, u)
	}
	// When the rows.Next() loop has finished we call rows.Err() to retrieve any
	// error that was encountered during the iteration. It's important to
	// call this - don't assume that a successful iteration was completed
	// over the whole resultset.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// If everything went OK then return the users slice.
	return users, nil
}

func (m *UserModel) Insert(email string, username string, password string, enabled bool) (int, error) {
	// Write the SQL statement we want to execute. I've split it over two lines
	// for readability (which is why it's surrounded with backquotes instead
	// of normal double quotes).

	stmt := `INSERT INTO users (email, username, password, enabled)
	VALUES(?, ?, ?, ?)`
	// Use the Exec() method on the embedded connection pool to execute the
	// statement. The first parameter is the SQL statement, followed by the
	// title, content and expiry values for the placeholder parameters. This
	// method returns a sql.Result type, which contains some basic
	// information about what happened when the statement was executed.

	result, err := m.DB.Exec(stmt, email, username, password, enabled) // db.Exec is first creating a prepared statement
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
