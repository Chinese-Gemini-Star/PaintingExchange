package service

import (
	"PaintingExchange/internal/model"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CheckPass(user model.User, database gorm.DB) error {
	var res model.User
	res.Username = user.Username
	// 判断管理员信息是否正确
	rows := database.Where(&res).Find(&res).RowsAffected
	if rows == 0 || bcrypt.CompareHashAndPassword([]byte(res.Password), []byte(user.Password)) != nil {
		// 不存在用户名或密码错误
		return errors.New("admin wrote wrong username or password")
	}
	return nil
}
