package controller

import (
	"PaintingExchange/internal/env"
	"PaintingExchange/internal/model"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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
// @Failure 500 {object} string "服务器内部错误"
// @Router /image/{imageID} [get]
// @Security BearerAuth
func (c *ImageController) GetBy(imageID string) mvc.Result {
	images := c.Mg.Database("PaintingExchange").Collection("Images")

	// 查找图片
	filter := bson.D{{"_id", imageID}}
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
// @Description 创建图片对象,需要先调用 /image/file [POST] 接口,获取到图片对象(包括图片id,作者用户名以及图片各个大小的地址),然后将其他元数据补充入此对象,再请求
// @Tags image
// @Accept json
// @Produce json
// @Param image body model.Image true "图片对象,在/image/file [POST] 接口的返回值上补充元数据所得"
// @Success 201 {object} model.Image "图片对象"
// @Failure 400 {object} string "请求数据异常"
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
	log.Println("用户", loginUserName, "上传图片", image.ID)
	if loginUserName != image.Auth {
		log.Println("图片非用户", loginUserName, "本人上传")
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "图片文件非本人上传或请求数据中作者信息异常",
		}
	}

	// 验证图片元数据是否已经存在
	log.Println("尝试查询图片,以验证是否重复创建(下一行的查找错误为正常)")
	if c.GetBy(image.ID).(mvc.Response).Code != iris.StatusBadRequest {
		log.Println("重复创建图片")
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "图片信息已存在",
		}
	}

	// 验证文件名
	// TODO 验证另外两个文件名
	if !checkImageFile(image.BigURI, image.ID, "big") {
		log.Println("图片id或路径异常")
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "请求数据中id或图片路径信息异常",
		}
	}

	// 创建图片
	image.Like = 0
	image.CreatedAt = time.Now()
	if _, err := images.InsertOne(nil, &image); err != nil {
		log.Println("图片创建失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	} else {
		log.Println("图片创建成功")
		return mvc.Response{
			Code:   iris.StatusCreated,
			Object: image,
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

	log.Println("图片文件保存成功")
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
// @Summary 修改图片信息
// @Description 修改自己上传的图片信息(仅标题,简介和标签允许修改,其他均以数据库已有信息为准)
// @Tags image
// @Accept json
// @Produce json
// @Param image body model.Image true "图片信息"
// @Success 201 {object} model.Image "图片信息更新成功，返回更新后的图片信息"
// @Failure 400 {object} string "请求数据异常"
// @Failure 401 {object} string "未授权，用户未登录或会话失效"
// @Failure 500 {object} string "服务器内部错误"
// @Router /image [put]
// @Security BearerAuth
func (c *ImageController) Put(image model.Image) mvc.Result {
	images := c.Mg.Database("PaintingExchange").Collection("Images")

	// 查询用户名
	loginUser, err := c.Ctx.User().GetRaw()
	if err != nil {
		return mvc.Response{
			Code: iris.StatusUnauthorized,
			Text: iris.StatusText(iris.StatusUnauthorized),
		}
	}
	loginUserName := loginUser.(iris.SimpleUser).Username
	log.Println("用户", loginUserName, "修改图片", image.ID)

	// 查询原图片对象
	prevImageRes := c.GetBy(image.ID).(mvc.Response)
	if prevImageRes.Code != iris.StatusOK {
		log.Println("图片不存在")
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "图片不存在",
		}
	}
	prevImage := prevImageRes.Object.(model.Image)

	// 验证是否为本人操作
	if loginUserName != prevImage.Auth {
		log.Println("图片非用户", loginUserName, "本人上传")
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "图片文件非本人上传",
		}
	}

	// 更新图片信息
	prevImage.Title = image.Title
	prevImage.Intro = image.Intro
	prevImage.Label = image.Label
	filter := bson.D{{"_id", prevImage.ID}}
	update := bson.D{{"$set", prevImage}}
	if _, err := images.UpdateOne(nil, filter, update); err != nil {
		log.Println("图片更新失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	} else {
		log.Println("图片更新成功")
		return mvc.Response{
			Code:   iris.StatusCreated,
			Object: prevImage,
		}
	}
}

// DeleteBy 删除图片
// @Summary 删除指定ID的图片
// @Description 删除自己上传的指定ID的图片
// @Tags image
// @Accept json
// @Produce json
// @Param imageID path string true "图片ID"
// @Success 204 {object} nil "图片删除成功，无返回内容"
// @Failure 400 {object} string "请求错误"
// @Failure 401 {object} string "未授权，用户未登录或会话失效"
// @Failure 404 {object} string "未找到图片，图片ID不存在"
// @Failure 500 {object} string "服务器内部错误"
// @Router /image/{imageID} [delete]
// @Security BearerAuth
func (c *ImageController) DeleteBy(imageID string) mvc.Result {
	images := c.Mg.Database("PaintingExchange").Collection("Images")

	// 查询用户名
	loginUser, err := c.Ctx.User().GetRaw()
	if err != nil {
		return mvc.Response{
			Code: iris.StatusUnauthorized,
			Text: iris.StatusText(iris.StatusUnauthorized),
		}
	}
	loginUserName := loginUser.(iris.SimpleUser).Username
	log.Println("用户", loginUserName, "删除图片", imageID)

	// 查询原图片对象
	prevImageRes := c.GetBy(imageID).(mvc.Response)
	if prevImageRes.Code != iris.StatusOK {
		log.Println("图片不存在")
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "图片不存在",
		}
	}
	prevImage := prevImageRes.Object.(model.Image)

	// 验证是否为本人操作
	if loginUserName != prevImage.Auth {
		log.Println("图片非用户", loginUserName, "本人上传")
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "图片文件非本人上传",
		}
	}

	// 删除图片信息
	filter := bson.D{{"_id", prevImage.ID}}
	if _, err := images.DeleteOne(nil, filter); err != nil {
		log.Println("图片删除失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	} else {
		log.Println("图片删除成功")
		return mvc.Response{
			Code: iris.StatusNoContent,
		}
	}
}

// getFilenameWithoutExt 获取纯文件名
func getFilenameWithoutExt(path string) string {
	filename := filepath.Base(path)
	filenameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	return filenameWithoutExt
}

// fileExists 确认文件是否存在
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// checkImageFile 检查文件地址是否正确
func checkImageFile(filePath string, id string, size string) bool {
	// 检查文件名
	filename := strings.SplitN(getFilenameWithoutExt(filePath), "_", 2)
	if len(filename) < 2 || id != filename[1] || size != filename[0] {
		return false
	}

	// 检查文件路径
	if filepath.Dir(filePath) != env.GetImgDir() {
		return false
	}

	// 检查文件是否存在
	if !fileExists(filePath) {
		return false
	}

	return true
}
