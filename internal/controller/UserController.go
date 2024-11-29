package controller

import (
	"PaintingExchange/internal/model"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"gorm.io/gorm"
)

type UserController struct {
	Ctx iris.Context
	Db  *gorm.DB
}

// GetBy 获取指定用户名的用户对象(无密码)
func (c *AuthController) GetBy(username string) mvc.Result {
	var user model.User
	c.Db.Where("username=?", username).Find(&user)
	user.Password = ""
	return mvc.Response{
		Code:   iris.StatusOK,
		Object: user,
	}
}

func (c *AuthController) Put(user model.User) mvc.Result {
	// TODO 更新用户对象
	return mvc.Response{
		Code: iris.StatusNoContent,
	}
}

// GetStars 获取用户收藏
func (c *AuthController) GetStars() mvc.Result {
	// TODO 查询用户收藏的内容
	return mvc.Response{
		Code: iris.StatusOK,
	}
}
