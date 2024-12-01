package controller

import (
	"PaintingExchange/internal/model"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	"io"
	"log"
	"os"
	"path/filepath"
)

// ImageController 用户相关操作控制器
type ImageController struct {
	Ctx iris.Context
	Db  *gorm.DB
	Mg  *mongo.Client
}

// GetBy 获取图片对象
// @Summary 获取指定ID的图片对象
// @Description 根据提供的图片ID，查找并返回该图片对象
// @Tags image
// @Accept json
// @Produce json
// @Param imageID path string true "图片ID"
// @Success 200 {object} model.Image "图片对象"
// @Failure 400 {object} string "请求错误，图片ID无效"
// @Failure 401 {object} string "未授权，用户未登录或会话失效"
// @Failure 404 {object} string "未找到图片，图片ID不存在"
// @Failure 500 {object} string "服务器内部错误"
// @Router /image/{imageID} [get]
// @Security BearerAuth
func (c *ImageController) GetBy(imageID string) mvc.Result {
	images := c.Mg.Database("PaintingExchange").Collection("Images")

	// 查找图片
	filter := bson.D{{"id", imageID}}
	if res := images.FindOne(nil, filter); res.Err() != nil {
		log.Println("图片", imageID, "查找失败", res.Err().Error())
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "图片不存在",
		}
	} else {
		var image model.Image
		res.Decode(&image)
		return mvc.Response{
			Code:   iris.StatusOK,
			Object: image,
		}
	}
}

// Post 创建图片(元信息)
// @Summary 创建图片元信息
// @Description 创建图片对象,需要先调用 /image/file [POST] 接口,获取到图片对象(包括图片id,作者用户名以及图片各个大小的地址),然后将其他元数据补充入此对象,再强求
// @Tags image
// @Accept json
// @Produce json
// @Param image body model.Image true "图片对象,在/image/file [POST] 接口的返回值上补充元数据所得"
// @Success 201 {string} string "图片ID"
// @Failure 401 {object} string "未授权，用户未登录或会话失效"
// @Failure 500 {object} string "服务器内部错误"
// @Router /image [post]
// @Security BearerAuth
func (c *ImageController) Post(image model.Image) mvc.Result {
	images := c.Mg.Database("PaintingExchange").Collection("Images")

	// 验证用户名
	loginUser, err := c.Ctx.User().GetRaw()
	if err != nil {
		return mvc.Response{
			Code: iris.StatusUnauthorized,
			Text: iris.StatusText(iris.StatusUnauthorized),
		}
	}
	loginUserName := loginUser.(iris.SimpleUser).Username
	if loginUserName != image.Auth {
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "图片文件非本人上传或请求数据中作者信息异常",
		}
	}

	// 创建图片
	image.Like = 0
	if res, err := images.InsertOne(nil, &image); err != nil {
		log.Println("图片创建失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	} else {
		return mvc.Response{
			Code: iris.StatusCreated,
			Text: res.InsertedID.(primitive.ObjectID).Hex(),
		}
	}
}

// PostFile 上传图片文件
// @Summary 上传图片文件
// @Description 上传图片文件，返回图片对象(包括图片id,作者用户名和地址).
// @Tags image
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "图片文件"
// @Success 201 {object} model.Image "图片上传成功，返回图片对象"
// @Failure 401 {object} string "未授权，用户未登录或会话失效"
// @Failure 500 {object} string "服务器内部错误"
// @Router /image/file [post]
// @Security BearerAuth
func (c *ImageController) PostFile() mvc.Result {
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
	log.Println(loginUserName, "上传图片文件")
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
	// TODO 生成不同尺码的图片
	// 创建对应文件
	bigURI := filepath.Join("./assert/images", "big_"+imageID+filepath.Ext(info.Filename))
	out, err := os.Create(bigURI)
	if err != nil {
		log.Println("图片文件创建失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}
	defer out.Close()

	// 写入图片
	_, err = io.Copy(out, file)
	if err != nil {
		log.Println("图片文件保存失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}

	// 封装成对象返回
	var image model.Image
	image.ID = imageID
	image.Auth = loginUserName
	image.BigURI = bigURI
	return mvc.Response{
		Code:   iris.StatusCreated,
		Object: image,
	}
}

// Put 修改图片
func (c *ImageController) Put(image model.Image) {
	// TODO
}

// DeleteBy 删除图片
func (c *ImageController) DeleteBy(imageID string) {
	// TODO
}
