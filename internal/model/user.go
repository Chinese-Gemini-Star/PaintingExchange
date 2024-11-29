package model

type User struct {
	Username  string `gorm:"primary_key" json:"username"`
	Password  string `gorm:"not null" json:"password"`
	AvatarURI string `json:"avatarURI"`
	Intro     string `json:"intro"`
}
