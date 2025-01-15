package models

import (
	"database/sql"
)

type PromotionRequests struct {
	ID          int
	UserID      int
	Description string
	Status      string
}

type PromotionRequestsModelInterface interface {
	Insert(user_id int, description string, status string) (int, error)
	GetByID(id int) (*PromotionRequests, error)
	GetAll() ([]*PromotionRequests, error)
	UpdateStatus(id int, status string) error
}

type PromotionRequestsModel struct {
	DB *sql.DB
}

func (m *PromotionRequestsModel) Insert(user_id int, description string, status string) (int, error) {
	stmt := `INSERT INTO Promotion_Requests (user_id, description, status)
			 VALUES (?, ?, ?)`

	result, err := m.DB.Exec(stmt, user_id, description, status)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *PromotionRequestsModel) GetByID(id int) (*PromotionRequests, error) {
	stmt := `SELECT id, user_id, description, status FROM Promotion_Requests WHERE id = ?`
	row := m.DB.QueryRow(stmt, id)

	var pr PromotionRequests
	err := row.Scan(&pr.ID, &pr.UserID, &pr.Description, &pr.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &pr, nil
}

func (m *PromotionRequestsModel) GetAll() ([]*PromotionRequests, error) {
	stmt := `SELECT id, user_id, description, status FROM Promotion_Requests`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*PromotionRequests
	for rows.Next() {
		var pr PromotionRequests
		err := rows.Scan(&pr.ID, &pr.UserID, &pr.Description, &pr.Status)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &pr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

func (m *PromotionRequestsModel) UpdateStatus(id int, status string) error {
	stmt := `UPDATE Promotion_Requests SET status = ? WHERE id = ?`
	_, err := m.DB.Exec(stmt, status, id)
	return err
}
