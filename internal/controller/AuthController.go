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
	"path/filepath"
	"time"
)

// AuthController 用户登录注册控制器
type AuthController struct {
	Ctx iris.Context
	Db  *gorm.DB
}

// PostUserLogin 登录
// @Summary 用户登录
// @Description 用户通过用户名和密码登录，成功后返回 JWT Token
// @Tags auth
// @Accept json
// @Produce plain
// @Param user body model.User true "用户登录信息(只需要username和password)"
// @Success 200 {string} string "JWT Token"
// @Failure 400 {string} string "用户名或密码错误"
// @Failure 403 {string} string "用户名被封禁"
// @Failure 500 {string} string "服务器内部错误"
// @Router /user/login [post]
func (c *AuthController) PostUserLogin(user model.User) mvc.Result {
	log.Println("[登录注册] 用户", user.Username, "登录")

	// 验证密码
	if res, err := service.CheckPass(user, *c.Db); err != nil {
		log.Println("[登录注册] 用户", user.Username, "被封禁")
		return mvc.Response{
			Code: iris.StatusForbidden,
			Text: "用户被封禁",
		}

	} else if !res {
		log.Println("[登录注册] 用户", user.Username, "密码错误")
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "用户名或密码错误",
		}
	}

	// 签发jwt
	log.Println("[登录注册] 用户", user.Username, "验证通过,签发jwt")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24小时后过期
		"username": user.Username,                         // 用户名
		"isAdmin":  false,                                 // 是否为管理员
	})
	tokenString, err := token.SignedString(env.GetJWTKey())
	if err != nil {
		log.Println("[登录注册] 用户", user.Username, "jwt签发失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}

	log.Println("[登录注册] 用户", user.Username, "登录成功")
	return mvc.Response{
		Code: iris.StatusOK,
		Text: tokenString,
	}
}

// PostUserRegister 注册
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
func (c *AuthController) PostUserRegister(user model.User) mvc.Result {
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
			Text: err.Error(),
		}
	} else {
		user.Password = string(password)
	}

	// 写入数据库
	user.Nickname = user.Username
	user.AvatarURI = filepath.Join(env.GetAvatarDir(), "0.png")
	c.Db.Create(&user)

	// 签发jwt
	log.Println("[登录注册] 用户", user.Username, "验证通过,签发jwt")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24小时后过期
		"username": user.Username,                         // 用户名
		"isAdmin":  false,                                 // 是否为管理员
	})
	tokenString, err := token.SignedString(env.GetJWTKey())
	if err != nil {
		log.Println("[登录注册] 用户", user.Username, "jwt签发失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}

	log.Println("[登录注册] 用户", user.Username, "注册成功")
	return mvc.Response{
		Code: iris.StatusCreated,
		Text: tokenString,
	}
}

// GetJwtTest 测试jwt
// @Description 测试JWT是否有效，验证用户是否具有访问权限
// @Tags auth
// @Success 200 {object} string "JWT验证通过"
// @Failure 401 {object} string "未授权，JWT无效或已过期"
// @Router /jwt/test [get]
// @Security BearerAuth
func (c *AuthController) GetJwtTest() {
	service.JWTMiddleware(c.Ctx)
}

// PostBackLogin 管理员登录
// @Summary 管理员登录
// @Description 管理员登录
// @Tags admin
// @Accept json
// @Produce json
// @Param admin body model.Admin true "管理员登录信息"
// @Success 200 {object} map[string]string "JWT 令牌"
// @Failure 403 {string} string "用户名或密码错误"
// @Failure 500 {string} string "服务器错误"
// @Router /back/login [post]
func (c *AuthController) PostBackLogin(admin model.Admin) mvc.Result {
	log.Println("管理员", admin.Username, "登录")

	// 验证密码
	if !service.CheckAdmin(admin, *c.Db) {
		log.Println("[登录注册] 管理员", admin.Username, "密码错误")
		return mvc.Response{
			Code: iris.StatusForbidden,
			Text: "用户名或密码错误",
		}
	}

	// 签发jwt
	log.Println("[登录注册] 管理员", admin.Username, "验证通过,签发jwt")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24小时后过期
		"username": admin.Username,                        // 用户名
		"isAdmin":  true,                                  // 是否为管理员
	})
	tokenString, err := token.SignedString(env.GetJWTKey())
	if err != nil {
		log.Println("[登录注册] 用户", admin.Username, "jwt签发失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}

	log.Println("[登录注册] 用户", admin.Username, "登录成功")
	return mvc.Response{
		Code: iris.StatusOK,
		Text: tokenString,
	}
}
