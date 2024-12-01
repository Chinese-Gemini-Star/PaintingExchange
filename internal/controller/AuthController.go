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

// AuthController 用户登录注册控制器
type AuthController struct {
	Ctx iris.Context
	Db  *gorm.DB
}

// PostLogin 登录
// @Summary 用户登录
// @Description 用户通过用户名和密码登录，成功后返回 JWT Token
// @Tags auth
// @Accept json
// @Produce plain
// @Param user body model.User true "用户登录信息(只需要username和password)"
// @Success 200 {string} string "JWT Token"
// @Failure 403 {string} string "用户名或密码错误"
// @Failure 500 {string} string "服务器内部错误"
// @Router /user/login [post]
func (c *AuthController) PostLogin(user model.User) mvc.Result {
	log.Println("[登录注册] 用户", user.Username, "登录")

	// 验证密码
	if !service.CheckPass(user, *c.Db) {
		log.Println("[登录注册] 用户", user.Username, "密码错误")
		return mvc.Response{
			Code: iris.StatusForbidden,
			Text: "用户名或密码错误",
		}
	}

	// 签发jwt
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
			Text: iris.StatusText(iris.StatusInternalServerError),
		}
	}

	log.Println("[登录注册] 用户", user.Username, "登录成功")
	return mvc.Response{
		Code: iris.StatusOK,
		Text: tokenString,
	}
}

// PostRegister 注册
// @Summary 用户注册
// @Description 用户进行注册，成功后返回 JWT Token
// @Tags auth
// @Accept json
// @Produce plain
// @Param user body model.User true "用户注册信息(只需要username和password)"
// @Success 201 {string} string "JWT Token"
// @Failure 403 {string} string "用户名已存在"
// @Failure 500 {string} string "服务器内部错误"
// @Router /user/register [post]
func (c *AuthController) PostRegister(user model.User) mvc.Result {
	log.Println("[登录注册] 用户", user.Username, "注册")

	// 验证用户名是否存在
	var tmp model.User
	if c.Db.Where("username=?", user.Username).Find(&tmp); tmp.Password != "" {
		log.Println("[登录注册] 用户", user.Username, "重复注册")
		return mvc.Response{
			Code: iris.StatusForbidden,
			Text: "用户名已存在",
		}
	}

	// 加密密码
	if password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost); err != nil {
		log.Println("[登录注册] 用户", user.Username, "密码加密失败")
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: iris.StatusText(iris.StatusInternalServerError),
		}
	} else {
		user.Password = string(password)
	}

	// 写入数据库
	c.Db.Create(&user)

	// 签发jwt
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
			Text: iris.StatusText(iris.StatusInternalServerError),
		}
	}

	log.Println("[登录注册] 用户", user.Username, "注册成功")
	return mvc.Response{
		Code: iris.StatusCreated,
		Text: tokenString,
	}
}
