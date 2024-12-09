package controller

import (
	"PaintingExchange/internal/env"
	"PaintingExchange/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"io"
	"log"
	"os"
	"path/filepath"
)

// UserController 用户相关操作控制器
type UserController struct {
	Ctx iris.Context
	Db  *gorm.DB
}

// GetBy 获取指定用户名的用户对象(无密码)
// @Summary 获取指定用户名的用户对象(无密码)
// @Description 根据用户名获取用户详细信息(无密码)，需要JWT验证
// @Tags user
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

// PostAvatar 上传用户头像
// @Summary 上传用户头像
// @Description 上传用户自己的头像文件,后续需要再请求一次/user [put]来更新用户信息
// @Tags user
// @Accept multipart/form-data
// @Produce text/plain
// @Param image formData file true "用户上传的头像文件"
// @Success 201 {string} string "返回头像存储路径"
// @Failure 401 {string} string "用户未授权"
// @Failure 500 {string} string "服务器内部错误"
// @Router /user/avatar [post]
// @Security BearerAuth
func (c *UserController) PostAvatar() mvc.Result {
	// 获取用户名
	loginUser, err := c.Ctx.User().GetRaw()
	if err != nil {
		return mvc.Response{
			Code: iris.StatusUnauthorized,
			Text: iris.StatusText(iris.StatusUnauthorized),
		}
	}
	loginUserName := loginUser.(iris.SimpleUser).Username
	// 读取图片
	log.Println(loginUserName, "上传头像")
	file, info, err := c.Ctx.FormFile("image")
	if err != nil {
		log.Println("图片文件上传失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}

	// 生成图片id
	imageID := uuid.New().String()
	// 创建对应文件
	uri := filepath.Join(env.GetAvatarDir(), imageID+filepath.Ext(info.Filename))
	out, err := os.Create(uri)
	if err != nil {
		log.Println("头像文件创建失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}
	defer out.Close()

	// 写入图片
	_, err = io.Copy(out, file)
	if err != nil {
		log.Println("头像文件保存失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}

	log.Println("头像文件保存成功")
	// 返回头像地址
	return mvc.Response{
		Code: iris.StatusCreated,
		Text: uri,
	}
}

// Put 更新用户对象(仅限自己)
// @Summary 更新用户信息
// @Description 允许已登录的用户更新自己的信息，包括密码。如果没有提供密码，密码保持不变。
// @Tags user
// @Accept json
// @Produce json
// @Param user body model.User true "用户信息"
// @Success 204 {object} nil "用户信息更新成功，无返回内容"
// @Failure 401 {object} string "未授权错误"
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
	if user.Password == "" {
		user.Password = prevUser.Password
	} else {
		// 修改了密码
		log.Println("修改了密码")
		if password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost); err != nil {
			log.Println("用户", user.Username, "修改密码加密失败")
			return mvc.Response{
				Code: iris.StatusInternalServerError,
				Text: err.Error(),
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

// GetStar 获取用户的收藏信息
// @Summary 获取用户的收藏信息
// @Description 查询用户自己的所有收藏信息
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {array} model.Star "返回用户的所有收藏记录"
// @Failure 401 {object} string "未授权，用户未登录或会话失效"
// @Router /user/star [get]
// @Security BearerAuth
func (c *UserController) GetStar() mvc.Result {
	// 获取用户名
	loginUser, err := c.Ctx.User().GetRaw()
	if err != nil {
		return mvc.Response{
			Code: iris.StatusUnauthorized,
			Text: iris.StatusText(iris.StatusUnauthorized),
		}
	}
	loginUserName := loginUser.(iris.SimpleUser).Username
	log.Println("用户", loginUserName, "查询收藏")

	// 读取所有收藏信息
	var stars []model.Star
	c.Db.Where("username=?", loginUserName).Find(&stars)

	return mvc.Response{
		Code:   iris.StatusOK,
		Object: stars,
	}
}

// PostStar 用户收藏图片
// @Summary 用户收藏图片
// @Description 用户收藏指定的图片
// @Tags user
// @Accept json
// @Produce json
// @Param star body model.Star true "收藏信息,只需要图片ID"
// @Success 204 {object} nil "图片收藏成功，无返回内容"
// @Failure 400 {object} string "请求错误，图片不存在"
// @Failure 401 {object} string "未授权，用户未登录或会话失效"
// @Failure 403 {object} string "收藏的图片被封禁"
// @Failure 500 {object} string "服务器内部错误"
// @Router /user/star [post]
// @Security BearerAuth
func (c *UserController) PostStar(star model.Star) mvc.Result {
	star.ID = 0

	// 获取用户名
	loginUser, err := c.Ctx.User().GetRaw()
	if err != nil {
		return mvc.Response{
			Code: iris.StatusUnauthorized,
			Text: iris.StatusText(iris.StatusUnauthorized),
		}
	}
	loginUserName := loginUser.(iris.SimpleUser).Username
	log.Println("用户", loginUserName, "收藏图片", star.ImageID)
	star.Username = loginUserName

	// 验证图片是否存在
	resp, err := resty.New().R().SetHeader("Authorization", loginUser.(iris.SimpleUser).Authorization).Get("http://localhost:8880/image/" + star.ImageID)
	if err != nil {
		log.Println("请求图片信息失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: iris.StatusText(iris.StatusInternalServerError),
		}
	}
	if resp.StatusCode() == 400 {
		log.Println("图片不存在")
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "收藏的图片不存在",
		}
	} else if resp.StatusCode() == 403 {
		log.Println("图片被封禁")
		return mvc.Response{
			Code: iris.StatusForbidden,
			Text: "收藏的图片被封禁",
		}
	}

	// 记录收藏信息
	c.Db.Create(&star)
	return mvc.Response{
		Code: iris.StatusNoContent,
	}
}

// DeleteStar 取消收藏图片
// @Summary 取消用户的图片收藏
// @Description 取消自己收藏的图片
// @Tags user
// @Accept json
// @Produce json
// @Param star body model.Star true "收藏信息,只需要图片ID"
// @Success 204 {object} nil "取消收藏成功，无返回内容"
// @Failure 400 {object} string "请求错误，收藏记录不存在"
// @Failure 401 {object} string "未授权，用户未登录或会话失效"
// @Router /user/star [delete]
// @Security BearerAuth
func (c *UserController) DeleteStar(star model.Star) mvc.Result {
	// 获取用户名
	loginUser, err := c.Ctx.User().GetRaw()
	if err != nil {
		return mvc.Response{
			Code: iris.StatusUnauthorized,
			Text: iris.StatusText(iris.StatusUnauthorized),
		}
	}
	loginUserName := loginUser.(iris.SimpleUser).Username
	log.Println("用户", loginUserName, "取消收藏", star.ImageID)
	star.Username = loginUserName

	// 从数据库中删除
	star.ID = 0
	if c.Db.Where(&star).Delete(&star).RowsAffected == 0 {
		log.Println("收藏记录不存在")
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "收藏记录不存在",
		}
	}

	return mvc.Response{
		Code: iris.StatusNoContent,
	}
}
