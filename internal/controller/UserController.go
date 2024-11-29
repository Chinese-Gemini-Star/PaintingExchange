package controller

import (
	"PaintingExchange/internal/env"
	"PaintingExchange/internal/model"
	"PaintingExchange/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"time"
)

type UserController struct {
	Ctx iris.Context
	Db  *gorm.DB
}

// PostLogin 登录
func (c *UserController) PostLogin(user model.User) mvc.Result {
	log.Println("[登录注册] 用户", user.Username, "登录")
	if service.CheckPass(user, *c.Db) != nil {
		log.Println("[登录注册] 用户", user.Username, "密码错误")
		return mvc.Response{
			Code: iris.StatusForbidden,
			Text: "用户名或密码错误",
		}
	}
	log.Println("[登录注册] 用户", user.Username, "验证通过,签发jwt")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24小时后过期
		"username": user.Username,                         // 用户名
	})

	tokenString, err := token.SignedString(env.GetJWTKey())
	if err != nil {
		log.Println("[登录注册] 用户", user.Username, "jwt签发失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
		}
	}

	log.Println("[登录注册] 用户", user.Username, "登录成功")
	return mvc.Response{
		Code: iris.StatusOK,
		Text: tokenString,
	}
}

// PostRegister 注册
func (c *UserController) PostRegister(user model.User) mvc.Result {
	log.Println("[登录注册] 用户", user.Username, "注册")
	var tmp model.User
	if c.Db.Where("username=?", user.Username).Find(&tmp); tmp.Password != "" {
		log.Println("[登录注册] 用户", user.Username, "重复注册")
		return mvc.Response{
			Code: iris.StatusForbidden,
			Text: "用户名已存在",
		}
	}
	log.Println("[登录注册] 用户", user.Username, "验证通过,签发jwt")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24小时后过期
		"username": user.Username,                         // 用户名
	})

	tokenString, err := token.SignedString(env.GetJWTKey())
	if err != nil {
		log.Println("[登录注册] 用户", user.Username, "jwt签发失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
		}
	}

	// 加密密码
	if password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost); err != nil {
		log.Println("[登录注册] 用户", user.Username, "密码加密失败")
		return mvc.Response{
			Code: iris.StatusInternalServerError,
		}
	} else {
		user.Password = string(password)
	}

	// 写入数据库
	c.Db.Create(&user)

	log.Println("[登录注册] 用户", user.Username, "注册成功")
	return mvc.Response{
		Code: iris.StatusCreated,
		Text: tokenString,
	}
}

// GetBy 获取用户对象(无密码)
func (c *UserController) GetBy(username string) mvc.Result {
	var user model.User
	c.Db.Where("username=?", username).Find(&user)
	user.Password = ""
	return mvc.Response{
		Code:   iris.StatusOK,
		Object: user,
	}
}

func (c *UserController) Put(user model.User) mvc.Result {
	// TODO 更新用户对象
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
