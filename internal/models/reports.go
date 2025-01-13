package models

import "time"

type Reports struct {
	ID             int
	ModeratorID    int
	PostID         int
	ReportReasonID int
	Description    string
	DateCreated    time.Time
	AdminID        int
	AdminResponse  string
}
