package models

import (
	"database/sql"
	"time"
)

type ReportsModelInterface interface {
	CreateReport(moderatorID int, postID int, reasonID int, description string, dateCreated time.Time) error
	GetAllReports() ([]*Reports, error)
}

type Reports struct {
	ID             int
	ModeratorID    int
	PostID         int
	ReportReasonID int
	Description    string
	DateCreated    time.Time
	AdminID        *int
	AdminResponse  *string
}

type ReportsModel struct {
	DB *sql.DB
}

func (m *ReportsModel) CreateReport(moderatorID int, postID int, reasonID int, description string, dateCreated time.Time) error {
	stmt := `
        INSERT INTO Reports (moderator_id, post_id, report_reason_id, description, dateCreated)
        VALUES (?, ?, ?, ?, ?)
    `
	_, err := m.DB.Exec(stmt, moderatorID, postID, reasonID, description, dateCreated)
	return err
}

// In your ReportsModel:
func (m *ReportsModel) GetAllReports() ([]*Reports, error) {
	stmt := `
        SELECT id, moderator_id, post_id, report_reason_id, description, dateCreated, admin_id, admin_response
        FROM Reports
        ORDER BY dateCreated DESC
    `
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*Reports
	for rows.Next() {
		r := &Reports{}
		err := rows.Scan(
			&r.ID,
			&r.ModeratorID,
			&r.PostID,
			&r.ReportReasonID,
			&r.Description,
			&r.DateCreated,
			&r.AdminID,
			&r.AdminResponse,
		)
		if err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	return reports, rows.Err()
}
