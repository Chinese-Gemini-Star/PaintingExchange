package model

// User 用户
// @Description 用户
type User struct {
	Username  string `gorm:"primary_key" json:"username" example:"test"`                                  // 用户名
	Password  string `gorm:"not null" json:"password" example:"123456"`                                   // 密码
	Nickname  string `json:"nickname" example:"test"`                                                     // 昵称
	AvatarURI string `json:"avatarURI" example:"assert/avatars/d18b9c4b-8d7f-407f-a630-cf2596bd7511.jpg"` // 头像地址
	Intro     string `json:"intro" example:"我是test"`                                                      // 描述
	IsBan     bool   `json:"isBan" example:"false"`                                                       // 是否被封禁
}
