package models

type Users struct {
	Id       uint
	Username string
	Password string
	Email    string
	Enabled  bool
}
