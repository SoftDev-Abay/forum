package models

type Users struct {
	ID       uint
	Username string
	Password string
	Email    string
	Enabled  bool
}
