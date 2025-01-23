package models

import (
	"database/sql"
	"fmt"
	"time"
)

type ReportsModelInterface interface {
	Get(reportID int) (*Reports, error)
	CreateReport(moderatorID int, postID int, reasonID int, description string, dateCreated time.Time) error
	GetAllReports() ([]*Reports, error)
	UpdateAdminResponse(reportID, adminID int, adminResponse string) error
	DeleteReportByID(reportID int) error
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

func (m *ReportsModel) Get(reportID int) (*Reports, error) {
	stmt := `
        SELECT id, moderator_id, post_id, report_reason_id, description, dateCreated
        FROM Reports
        WHERE id = ?
    `
	row := m.DB.QueryRow(stmt, reportID)

	r := &Reports{}
	err := row.Scan(
		&r.ID,
		&r.ModeratorID,
		&r.PostID,
		&r.ReportReasonID,
		&r.Description,
		&r.DateCreated,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNoRecord 
	} else if err != nil {
		return nil, err
	}
	return r, nil
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

func (m *ReportsModel) UpdateAdminResponse(reportID, adminID int, adminResponse string) error {
	stmt := `
        UPDATE Reports
        SET admin_id = ?, admin_response = ?
        WHERE id = ?
    `
	result, err := m.DB.Exec(stmt, adminID, adminResponse, reportID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no report found with id = %d", reportID)
	}
	return nil
}

func (m *ReportsModel) DeleteReportByID(reportID int) error {
	stmt := `DELETE FROM Reports WHERE id = ?`
	result, err := m.DB.Exec(stmt, reportID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no report found with id = %d", reportID)
	}
	return nil
}
