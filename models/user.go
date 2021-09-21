package models

type Users struct {
	UserId   string `json:"userId"`
	UserName string `json:"userName"`
	Password string `json:"Password"`
	IsAdmin  bool   `json:"isAdmin"`
}
