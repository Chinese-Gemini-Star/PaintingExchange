package controller

import (
	"PaintingExchange/internal/env"
	"PaintingExchange/internal/model"
	"PaintingExchange/internal/service"
	"context"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gocv.io/x/gocv"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ImageController 用户相关操作控制器
type ImageController struct {
	Ctx  iris.Context
	Db   *gorm.DB
	Mg   *mongo.Client
	Algo service.SearchServiceClient
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
// @Failure 403 {object} string "图片被封禁"
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

		// 验证图片是否被封
		if image.IsBan || image.AuthIsBan {
			return mvc.Response{
				Code: iris.StatusForbidden,
				Text: "图片被封禁",
			}
		}

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
	if !checkImageFile(image.BigURI, image.ID, "big") || !checkImageFile(image.MidURI, image.ID, "mid") {
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
	}
	log.Println("图片创建成功")

	// 调用算法层向量化
	_, err = c.Algo.CreateImage(context.Background(), &service.Image{
		Id:    image.ID,
		Title: image.Title,
		Label: image.Label,
		IsBan: image.IsBan,
	})
	if err != nil {
		log.Println("算法层gRPC调用失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}

	return mvc.Response{
		Code:   iris.StatusCreated,
		Object: image,
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
	img, err := service.FileToMat(file)
	if err != nil {
		log.Println("图片文件读取失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}

	// 生成图片id
	imageID := uuid.New().String()

	originalWidth := img.Cols()
	originalHeight := img.Rows()

	// 保存大尺寸图片
	bigImg := service.ResizeImage(img, originalWidth, originalHeight, service.BigSize)
	bigURI := filepath.Join("./assert/images", "big_"+imageID+filepath.Ext(info.Filename))
	gocv.IMWrite(bigURI, bigImg)

	// 保存中尺寸图片
	midImg := service.ResizeImage(img, originalWidth, originalHeight, service.MidSize)
	midURI := filepath.Join("./assert/images", "mid_"+imageID+filepath.Ext(info.Filename))
	gocv.IMWrite(midURI, midImg)
	log.Println("图片文件保存成功")

	// 封装成对象返回
	var image model.Image
	image.ID = imageID
	image.Auth = loginUserName
	image.BigURI = bigURI
	image.MidURI = midURI
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
// @Failure 403 {object} string "图片被封禁"
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
	if prevImageRes.Code == iris.StatusBadRequest {
		log.Println("图片不存在")
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "图片不存在",
		}
	} else if prevImageRes.Code == iris.StatusForbidden {
		log.Println("图片被封禁")
		return mvc.Response{
			Code: iris.StatusForbidden,
			Text: "图片被封禁",
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
	}
	log.Println("图片更新成功")

	// 调用算法层向量化
	c.Algo.UpdateImage(context.Background(), &service.Image{
		Id:    image.ID,
		Title: image.Title,
		Label: image.Label,
		IsBan: image.IsBan,
	})

	return mvc.Response{
		Code:   iris.StatusCreated,
		Object: prevImage,
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
	}

	// 调用算法层删除向量
	c.Algo.DeleteImage(context.Background(), &service.Image{
		Id:    imageID,
		Title: "",
		Label: nil,
		IsBan: false,
	})

	log.Println("图片删除成功")
	return mvc.Response{
		Code: iris.StatusNoContent,
	}

}

// GetNewest 获取最新9个图片
// @Summary 获取最新的9张图片
// @Description 获取最新的9张图片
// @Tags image
// @Accept json
// @Produce json
// @Success 200 {array} model.Image "返回最新的9张图片信息"
// @Failure 500 {object} string "服务器内部错误"
// @Router /image/newest [get]
// @Security BearerAuth
func (c *ImageController) GetNewest() mvc.Result {
	images := c.Mg.Database("PaintingExchange").Collection("Images")

	log.Println("获取最新9个图片")
	// 获取最新9个图片
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}}) // 按创建时间降序排序
	cursor, err := images.Find(nil, bson.D{}, findOptions)
	if err != nil {
		log.Println("最新图片查询失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}
	defer cursor.Close(nil)

	// 组装返回结果,保留9个
	var res []model.Image
	for cursor.Next(nil) && len(res) < 9 {
		var image model.Image
		if err := cursor.Decode(&image); err != nil {
			log.Println("最新图片对象读取失败", err)
			return mvc.Response{
				Code: iris.StatusInternalServerError,
				Text: err.Error(),
			}
		}

		// 跳过已封禁
		if image.IsBan || image.AuthIsBan {
			continue
		}

		res = append(res, image)
	}
	//if err := cursor.All(nil, &res); err != nil {
	//	log.Println("最新图片对象读取失败", err)
	//	return mvc.Response{
	//		Code: iris.StatusInternalServerError,
	//		Text: err.Error(),
	//	}
	//}

	return mvc.Response{
		Code:   iris.StatusOK,
		Object: res,
	}
}

// GetFromBy 获取指定用户上传的所有图片
// @Summary 获取指定用户上传的所有图片
// @Description 查询指定用户名上传的所有图片
// @Tags image
// @Accept json
// @Produce json
// @Param username path string true "用户名"
// @Success 200 {array} model.Image "返回指定用户上传的所有图片信息"
// @Failure 500 {object} string "服务器内部错误"
// @Router /image/from/{username} [get]
// @Security BearerAuth
func (c *ImageController) GetFromBy(username string) mvc.Result {
	images := c.Mg.Database("PaintingExchange").Collection("Images")

	log.Println("查询用户", username, "上传的图片")

	// 查询用户上传的图片
	filter := bson.M{
		"auth": username,
	}
	cursor, err := images.Find(nil, filter)
	if err != nil {
		log.Println("查询用户上传图片失败", err)
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}
	defer cursor.Close(nil)

	// 组装返回对象
	var res []model.Image
	for cursor.Next(nil) {
		var image model.Image
		if err := cursor.Decode(&image); err != nil {
			log.Println("用户上传的图片对象读取失败", err)
			return mvc.Response{
				Code: iris.StatusInternalServerError,
				Text: err.Error(),
			}
		}

		// 跳过已封禁
		if image.IsBan || image.AuthIsBan {
			continue
		}

		res = append(res, image)
	}
	//if err := cursor.All(nil, &res); err != nil {
	//	log.Println("用户上传的图片对象读取失败", err)
	//	return mvc.Response{
	//		Code: iris.StatusInternalServerError,
	//		Text: err.Error(),
	//	}
	//}

	return mvc.Response{
		Code:   iris.StatusOK,
		Object: res,
	}
}

// GetSearch 查询图片
// @Summary 查询图片
// @Description 查询图片，进行标签匹配和标题模糊匹配
// @Tags image
// @Accept json
// @Produce json
// @Param search query string true "查询内容"
// @Success 200 {array} model.Image "返回符合查询条件的图片信息"
// Failure 400 {object} string "缺少请求参数"
// @Failure 500 {object} string "服务器内部错误"
// @Router /image/search [get]
// @Security BearerAuth
func (c *ImageController) GetSearch() mvc.Result {
	images := c.Mg.Database("PaintingExchange").Collection("Images")

	// 获取请求参数
	search := c.Ctx.URLParam("search")
	if search == "" {
		return mvc.Response{
			Code: iris.StatusBadRequest,
			Text: "缺少请求参数",
		}
	}

	log.Println("查询图片,内容:", search)

	// 并发查询
	var res []model.Image
	var mu sync.Mutex
	var g errgroup.Group

	// 智能检索
	g.Go(func() error {
		// 调用算法层向量化检索
		aiRes, err := c.Algo.SearchImage(context.Background(), &service.Search{
			Search: search,
		})
		if err != nil {
			log.Println("算法层gRPC调用失败", err)
			return err
		}
		ids := aiRes.ImageIds
		if ids == nil {
			return nil
		}

		// 查询图片对象
		filter := bson.D{{"_id", bson.D{{"$in", ids}}}}
		cursor, err := images.Find(nil, filter)
		if err != nil {
			log.Println("id查询图片失败", err)
			return err
		}

		// 读取图片对象
		for cursor.Next(nil) {
			var image model.Image
			if err := cursor.Decode(&image); err != nil {
				log.Println("id读取图片失败", err)
				return err
			}

			// 跳过已封禁
			if image.IsBan || image.AuthIsBan {
				continue
			}

			mu.Lock()
			res = append(res, image)
			mu.Unlock()
		}
		return nil
	})

	// 空格分割关键字
	keywords := strings.Split(search, " ")
	// 匹配关键字至标签
	g.Go(func() error {
		// 查询标签
		filter := bson.M{
			"label": bson.M{"$in": keywords},
		}
		cursor, err := images.Find(nil, filter)
		if err != nil {
			log.Println("标签查询图片失败")
			return err
		}
		defer cursor.Close(nil)

		// 保存结果
		for cursor.Next(nil) {
			var image model.Image
			if err := cursor.Decode(&image); err != nil {
				log.Println("标签查询图片对象读取失败", err)
				return err
			}

			// 跳过已封禁
			if image.IsBan || image.AuthIsBan {
				continue
			}

			mu.Lock()
			res = append(res, image)
			mu.Unlock()
		}

		//var tagRes []model.Image
		//if err := cursor.All(nil, &tagRes); err != nil {
		//	log.Println("标签查询图片读取失败", err)
		//	return err
		//}
		return nil
	})
	// 模糊匹配标题
	g.Go(func() error {
		// 查询标题
		filter := bson.M{
			"title": bson.M{
				"$regex":   search, // 模糊匹配
				"$options": "i",    // 忽略大小写
			},
		}
		cursor, err := images.Find(nil, filter)
		if err != nil {
			log.Println("标题查询图片失败", err)
			return err
		}
		defer cursor.Close(nil)

		// 保存结果
		for cursor.Next(nil) {
			log.Println(1)
			var image model.Image
			if err := cursor.Decode(&image); err != nil {
				log.Println("标题查询图片对象读取失败", err)
				return err
			}

			// 跳过已封禁
			if image.IsBan || image.AuthIsBan {
				continue
			}

			mu.Lock()
			res = append(res, image)
			mu.Unlock()
		}

		//var titleRes []model.Image
		//if err := cursor.All(nil, &titleRes); err != nil {
		//	log.Println("标题查询图片读取失败", err)
		//	return mvc.Response{
		//		Code: iris.StatusInternalServerError,
		//		Text: err.Error(),
		//	}
		//}
		return nil
	})

	//res := append(tagRes, titleRes...)
	if err := g.Wait(); err != nil {
		return mvc.Response{
			Code: iris.StatusInternalServerError,
			Text: err.Error(),
		}
	}
	// 按上传时间降序去重
	res = uniqueSortedImages(res)

	return mvc.Response{
		Code:   iris.StatusOK,
		Object: res,
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

// uniqueSortedImages 图片切片排序并去重
func uniqueSortedImages(images []model.Image) []model.Image {
	if len(images) <= 1 {
		return images
	}

	// 排序
	sort.Slice(images, func(i, j int) bool {
		return images[i].CreatedAt.After(images[j].CreatedAt)
	})

	// 去重
	res := []model.Image{images[0]}
	for i := 1; i < len(images); i++ {
		if images[i].ID != images[i-1].ID {
			res = append(res, images[i])
		}
	}

	return res
}
