package service

import (
	"PaintingExchange/internal/model"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CheckPass 验证密码(因为循环依赖原因,无法放于user对象中)
func CheckPass(user model.User, database gorm.DB) (bool, error) {
	var res model.User
	res.Username = user.Username
	// 确认存在用户
	rows := database.Where(&res).Find(&res).RowsAffected
	if rows == 0 || bcrypt.CompareHashAndPassword([]byte(res.Password), []byte(user.Password)) != nil {
		// 不存在用户名或密码错误
		return false, nil
	}

	if res.IsBan {
		return false, errors.New("ban")
	}

	return true, nil
}

// CheckAdmin 验证管理员密码
func CheckAdmin(admin model.Admin, database gorm.DB) bool {
	var res model.Admin
	res.Username = admin.Username
	// 确认存在用户
	rows := database.Where(&res).Find(&res).RowsAffected
	if rows == 0 || res.Password != admin.Password {
		// 不存在用户名或密码错误
		return false
	}
	return true
}
