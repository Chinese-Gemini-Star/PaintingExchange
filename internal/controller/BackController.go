package controller

import (
	"PaintingExchange/internal/model"
	"PaintingExchange/internal/service"
	"context"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	"log"
)

type BackController struct {
	Ctx  iris.Context
	Db   *gorm.DB
	Mg   *mongo.Client
	Algo service.SearchServiceClient
}

// GetUser 获取所有用户
// @Summary 获取所有用户列表
// @Description 返回所有用户的信息
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {array} model.User "用户列表"
// @Failure 401 {object} string "未授权，用户未登录或会话失效"
// @Router /back/user [get]
// @Security BearerAuth
func (c *BackController) GetUser() mvc.Result {
	var users []model.User
	c.Db.Find(&users)

	return mvc.Response{
		Code:   iris.StatusOK,
		Object: users,
	}
}

// PostUserBan 封禁用户
// @Summary 封禁用户
// @Description 封禁指定用户账号(只有username有用)
// @Tags admin
// @Accept json
// @Produce json
// @Param user body model.User true "用户信息，包含用户名"
// @Success 204 {string} string "封禁成功，无返回内容"
// @Failure 401 {object} string "未授权，用户未登录或会话失效"
// @Failure 400 {string} string "用户不存在"
// @Router /back/user/ban [post]
// @Security BearerAuth
func (c *BackController) PostUserBan(user model.User) mvc.Result {
	// 获取原用户对象
	var prevUser model.User
	rows := c.Db.Where("username = ?", user.Username).First(&prevUser).RowsAffected
	if rows == 0 {
		// 用户不存在
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "用户不存在",
		}
	}

	// 标记已封禁
	prevUser.IsBan = true
	c.Db.Where("username=?", prevUser.Username).Updates(&prevUser)

	// 其图片标记作者已封禁
	if err := c.changAuthBan(prevUser.Username, true); err != nil {
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}

	return mvc.Response{
		Code: iris.StatusNoContent,
	}
}

// PostUserUnban 解除封禁
// @Summary 解除封禁用户
// @Description 解除指定用户的封禁状态(只有username有用)
// @Tags admin
// @Accept json
// @Produce json
// @Param user body model.User true "用户信息，包含用户名"
// @Success 204 {string} string "解除封禁成功，无返回内容"
// @Failure 401 {object} string "未授权，用户未登录或会话失效"
// @Failure 400 {string} string "用户不存在"
// @Router /back/user/unban [post]
// @Security BearerAuth
func (c *BackController) PostUserUnban(user model.User) mvc.Result {
	// 获取原用户对象
	var prevUser model.User
	rows := c.Db.Where("username = ?", user.Username).First(&prevUser).RowsAffected
	if rows == 0 {
		// 用户不存在
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "用户不存在",
		}
	}

	// 标记未封禁
	prevUser.IsBan = false
	log.Println(prevUser)
	c.Db.Where("username = ?", prevUser.Username).Select("is_ban").Updates(&prevUser)

	// 其图片标记作者未封禁
	if err := c.changAuthBan(prevUser.Username, false); err != nil {
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}

	return mvc.Response{
		Code: iris.StatusNoContent,
	}
}

// GetImage 获取所有图片
// @Summary 获取所有图片信息
// @Description 管理员获取所有图片的详细信息
// @Tags admin
// @Success 200 {array} model.Image "所有图片对象列表"
// @Failure 401 {object} string "未授权，用户未登录或会话失效"
// @Failure 500 {string} string "服务器错误"
// @Router /back/image [get]
// @Security BearerAuth
func (c *BackController) GetImage() mvc.Result {
	images := c.Mg.Database("PaintingExchange").Collection("Images")

	// 查询所有记录
	cursor, err := images.Find(nil, bson.D{})
	if err != nil {
		log.Println("图片查询失败")
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}
	defer cursor.Close(nil)

	// 转换为Image对象
	var res []model.Image
	if err := cursor.All(nil, &res); err != nil {
		log.Println("图片转换失败")
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}

	return mvc.Response{
		Code:   iris.StatusOK,
		Object: res,
	}
}

// PostImageBan 封禁图片
// @Summary 封禁图片
// @Description 管理员封禁指定图片(仅id有效)
// @Tags admin
// @Security BearerAuth
// @Param image body model.Image true "图片对象，其中包含图片ID"
// @Success 204 "操作成功，无返回内容"
// @Failure 400 {string} string "图片不存在"
// @Failure 401 {object} string "未授权，用户未登录或会话失效"
// @Failure 500 {string} string "服务器错误"
// @Router /back/image/ban [post]
// @Security BearerAuth
func (c *BackController) PostImageBan(image model.Image) mvc.Result {
	images := c.Mg.Database("PaintingExchange").Collection("Images")

	// 获取原图片对象
	var prev model.Image
	filter := bson.D{{"_id", image.ID}}
	if res := images.FindOne(nil, filter); res.Err() != nil {
		log.Println("图片", image.ID, "查找失败", res.Err().Error())
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "图片不存在",
		}
	} else {
		res.Decode(&prev)
	}

	// 设置封禁
	prev.IsBan = true

	// 存入数据库
	update := bson.D{{"$set", prev}}
	if _, err := images.UpdateOne(nil, filter, update); err != nil {
		log.Println("图片", image.ID, "封禁信息写入失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}

	// 修改向量记录
	if _, err := c.Algo.UpdateImage(context.Background(), &service.Image{
		Id:    prev.ID,
		Title: prev.Title,
		Label: prev.Label,
		IsBan: prev.IsBan,
	}); err != nil {
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}

	return mvc.Response{
		Code: iris.StatusNoContent,
	}
}

// PostImageUnban 解禁图片
// @Summary 解除图片封禁
// @Description 解除对该图片的封禁状态(仅id有效)
// @Tags admin
// @Accept json
// @Produce json
// @Param image body model.Image true "图片信息,包含id"
// @Success 204 "解除封禁成功，无返回内容"
// @Failure 400 {object} string "请求错误，图片ID无效或图片不存在"
// @Failure 401 {object} string "未授权，用户未登录或会话失效"
// @Failure 500 {object} string "服务器内部错误"
// @Router /back/image/unban [post]
// @Security BearerAuth
func (c *BackController) PostImageUnban(image model.Image) mvc.Result {
	images := c.Mg.Database("PaintingExchange").Collection("Images")

	// 获取原图片对象
	var prev model.Image
	filter := bson.D{{"_id", image.ID}}
	if res := images.FindOne(nil, filter); res.Err() != nil {
		log.Println("图片", image.ID, "查找失败", res.Err().Error())
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "图片不存在",
		}
	} else {
		res.Decode(&prev)
	}

	// 解除封禁
	prev.IsBan = false

	// 存入数据库
	update := bson.D{{"$set", prev}}
	if _, err := images.UpdateOne(nil, filter, update); err != nil {
		log.Println("图片", image.ID, "封禁信息写入失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}

	// 修改向量记录
	if _, err := c.Algo.UpdateImage(context.Background(), &service.Image{
		Id:    prev.ID,
		Title: prev.Title,
		Label: prev.Label,
		IsBan: prev.IsBan,
	}); err != nil {
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}

	return mvc.Response{
		Code: iris.StatusNoContent,
	}
}

func (c *BackController) changAuthBan(username string, isBan bool) error {
	images := c.Mg.Database("PaintingExchange").Collection("Images")

	// 更新封禁信息
	filter := bson.D{{"auth", username}}
	update := bson.D{{"$set", bson.M{"authIsBan": isBan}}}
	if _, err := images.UpdateMany(nil, filter, update); err != nil {
		return err
	}
	return nil
}
