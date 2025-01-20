package models

import (
	"database/sql"
	"errors"
)

type CategoriesModelInterface interface {
	Insert(name string) (int, error)
	Get(id int) (*Categories, error)
	GetAll() ([]*Categories, error)
	Delete(id int) error
}

type Categories struct {
	ID   int
	Name string
}

type CategoriesModel struct {
	DB *sql.DB
}

func (m *CategoriesModel) Insert(name string) (int, error) {
	stmt := `INSERT INTO Categories (name) 
			 VALUES (?)`

	result, err := m.DB.Exec(stmt, name)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *CategoriesModel) Get(id int) (*Categories, error) {
	stmt := `SELECT id, name
			 FROM Categories
			 WHERE id = ?`
	row := m.DB.QueryRow(stmt, id)

	categories := &Categories{}

	err := row.Scan(&categories.ID, &categories.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return categories, nil
}

func (m *CategoriesModel) GetAll() ([]*Categories, error) {
	stmt := `SELECT id, name
	         FROM Categories
	         ORDER BY name DESC
	         LIMIT 10`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Categories

	for rows.Next() {
		category := &Categories{}
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}


// delete a category

func (m *CategoriesModel) Delete(id int) error {
	stmt := `DELETE FROM Categories WHERE id = ?`
	_, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}
	return nil
}