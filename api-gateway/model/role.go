package model

type Role struct {
	ID       int64  `gorm:"column:id" json:"id"`
	UserName string `gorm:"column:user_name" json:"user_name"`
	Password string `gorm:"column:password" json:"password"`
	Status   bool   `gorm:"column:status" json:"status"`
	IsAdmin  bool   `gorm:"column:is_admin" json:"is_admin"`
}
