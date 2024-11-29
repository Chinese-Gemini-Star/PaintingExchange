package model

// User 用户
// @Description 用户
type User struct {
	Username  string `gorm:"primary_key" json:"username" example:"test"` // 用户名
	Password  string `gorm:"not null" json:"password" example:"123456"`  // 密码
	AvatarURI string `json:"avatarURI" example:"TODO"`                   // 头像地址
	Intro     string `json:"intro" example:"我是test"`                     // 描述
}
