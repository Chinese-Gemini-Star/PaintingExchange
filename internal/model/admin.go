package model

// Admin 管理员
// @Description 管理员
type Admin struct {
	Username string `json:"username" example:"admin"`
	Password string `json:"password" example:"admin"`
}
