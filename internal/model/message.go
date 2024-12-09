package model

import "time"

// Message 聊天消息
// @Description 聊天消息
type Message struct {
	ID      int       `json:"id" gorm:"primary_key" swaggerignore:"true"`
	From    string    `json:"from" example:"test"` // 发送者用户名
	To      string    `json:"to" example:"test1"`  // 接收者用户名
	Content string    `json:"content"`             // 消息内容
	Time    time.Time `json:"time"`                // 发送时间
}
