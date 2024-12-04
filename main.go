package main

import (
	_ "PaintingExchange/docs"
	"PaintingExchange/internal/controller"
	"PaintingExchange/internal/env"
	"PaintingExchange/internal/model"
	"PaintingExchange/internal/service"
	"fmt"
	"github.com/iris-contrib/middleware/cors"
	"github.com/iris-contrib/swagger"
	"github.com/iris-contrib/swagger/swaggerFiles"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
)

// @title 绘画交流平台
// @version 1.0
// @description 绘画交流平台的后端API文档

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @host localhost:8880
// @BasePath /
func main() {
	app := iris.New()
	//app.Logger().SetLevel("debug")

	// 允许跨域
	app.UseRouter(cors.AllowAll())

	// swaggerAPI界面
	swaggerUI := swagger.Handler(swaggerFiles.Handler,
		swagger.URL("/swagger/swagger.json"),
		swagger.DeepLinking(true),
		swagger.Prefix("/swagger"),
	)
	app.Get("/swagger", swaggerUI)
	app.Get("/swagger/{any:path}", swaggerUI)

	// MySQL数据库连接
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", env.GetEnv("dbUserName", "paintingExchange"), env.GetEnv("dbPassword", "1234567"), env.GetEnv("dbName", "paintingExchange")),
	}), &gorm.Config{})
	if err != nil {
		log.Fatalln("MySQL数据库连接失败")
	} else {
		db.AutoMigrate(&model.User{})
		db.AutoMigrate(&model.Star{})
	}

	// mongoDB数据库连接
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	// 连接到MongoDB
	mg, err := mongo.Connect(nil, clientOptions)
	if err != nil {
		log.Fatalln("mongoDB数据库连接失败", err)
	} else {
		defer mg.Disconnect(nil)
	}
	if err := mg.Ping(nil, nil); err != nil {
		log.Fatalln("mongoDB数据库连接失败", err)
	}

	// 创建图片缓存目录并绑定路由
	err = os.MkdirAll(env.GetImgDir(), os.ModePerm)
	if err != nil {
		log.Fatalln("创建图片缓存目录失败:", err)
	}
	app.HandleDir("/assert/images", env.GetImgDir())

	// 绑定依赖和路由
	mvc.Configure(app, func(application *mvc.Application) {
		application.Register(db)
		application.Register(mg)
		application.Party("/user").Handle(new(controller.AuthController))
		application.Party("/user", service.JWTMiddleware).Handle(new(controller.UserController))
		application.Party("/image", service.JWTMiddleware).Handle(new(controller.ImageController))
	})

	if err := app.Listen(":8880"); err != nil {
		log.Fatalln("启动失败")
	}
}
