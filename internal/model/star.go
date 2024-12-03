package model

// Star 收藏信息
// @Description 收藏信息
type Star struct {
	ID       uint   `gorm:"primary_key" swaggerignore:"true"`                       // 主键
	Username string `json:"username" example:"test"`                                // 用户名
	ImageID  string `json:"imageID" example:"68c8d808-54f7-4cfc-94c9-015416033dc9"` // 图片ID
}
