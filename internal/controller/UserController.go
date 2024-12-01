package controller

import (
	"PaintingExchange/internal/model"
	"PaintingExchange/internal/service"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
)

// UserController 用户相关操作控制器
type UserController struct {
	Ctx iris.Context
	Db  *gorm.DB
}

// GetBy 获取指定用户名的用户对象(无密码)
// @Summary 获取指定用户名的用户对象(无密码)
// @Description 根据用户名获取用户详细信息(无密码)，需要JWT验证
// @Tags User
// @Param username path string true "用户名"
// @Success 200 {object} model.User "用户对象(无密码)"
// @Failure 401 {object} string "未授权错误"
// @Failure 404 {object} string "用户不存在"
// @Router /user/{username} [get]
// @Security BearerAuth
func (c *UserController) GetBy(username string) mvc.Result {
	log.Println("查询用户", username)
	var user model.User
	c.Db.Where("username=?", username).Find(&user)
	if user.Password == "" {
		log.Println("查询用户", username, "不存在")
		return mvc.Response{
			Code: iris.StatusNotFound,
			Text: "用户不存在",
		}
	}
	user.Password = ""

	log.Println("查询用户", username, "成功")
	return mvc.Response{
		Code:   iris.StatusOK,
		Object: user,
	}
}

// Put 更新用户对象(仅限自己)
// @Summary 更新用户信息
// @Description 允许已登录的用户更新自己的信息，包括密码。如果没有提供密码，密码保持不变。
// @Tags User
// @Accept json
// @Produce json
// @Param user body model.User true "用户信息"
// @Success 204 {object} nil "用户信息更新成功，无返回内容"
// @Failure 401 {object} string "未授权，用户未登录或会话失效"
// @Failure 403 {object} string "禁止操作，尝试修改非自己的信息"
// @Failure 500 {object} string "服务器内部错误"
// @Router /user [put]
// @Security BearerAuth
func (c *UserController) Put(user model.User) mvc.Result {
	// 验证用户是否是自己
	loginUser, err := c.Ctx.User().GetRaw()
	if err != nil {
		return mvc.Response{
			Code: iris.StatusUnauthorized,
			Text: iris.StatusText(iris.StatusUnauthorized),
		}
	}
	loginUserName := loginUser.(iris.SimpleUser).Username
	log.Println(loginUserName, "修改用户信息")
	if user.Username != loginUserName {
		log.Println("非法修改非个人信息")
		return mvc.Response{
			Code: iris.StatusForbidden,
			Text: "只能修改自己的用户信息",
		}
	}
	// 设置密码
	var prevUser model.User
	c.Db.Find(&prevUser, "username=?", user.Username)
	if user.Password == "" || service.CheckPass(user, *c.Db) {
		user.Password = prevUser.Password
	} else {
		// 修改了密码
		log.Println("修改了密码")
		if password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost); err != nil {
			log.Println("用户", user.Username, "修改密码加密失败")
			return mvc.Response{
				Code: iris.StatusInternalServerError,
				Text: iris.StatusText(iris.StatusInternalServerError),
			}
		} else {
			user.Password = string(password)
		}
	}

	// 更新用户
	c.Db.Where("username=?", user.Username).Updates(&user)
	log.Println(user.Username, "用户信息更新完成")
	return mvc.Response{
		Code: iris.StatusNoContent,
	}
}

// GetStars 获取用户收藏
func (c *UserController) GetStars() mvc.Result {
	// TODO 查询用户收藏的内容
	return mvc.Response{
		Code: iris.StatusOK,
	}
}
