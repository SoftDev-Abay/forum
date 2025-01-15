package models

import "database/sql"

type ReportsReasonsModelInterface interface {
	GetAllReasons() ([]*ReportReasons, error)
}

type ReportReasons struct {
	ID   int
	Text string
}

type ReportReasonsModel struct {
	DB *sql.DB
}

func (m *ReportReasonsModel) GetAllReasons() ([]*ReportReasons, error) {
	stmt := `SELECT id, text FROM Report_Reasons`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reasons []*ReportReasons
	for rows.Next() {
		r := &ReportReasons{}
		if err := rows.Scan(&r.ID, &r.Text); err != nil {
			return nil, err
		}
		reasons = append(reasons, r)
	}
	return reasons, rows.Err()
}
