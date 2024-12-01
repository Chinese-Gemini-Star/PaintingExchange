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
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
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

	// 允许跨域
	app.UseRouter(cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
	}))

	// swaggerAPI界面
	swaggerUI := swagger.Handler(swaggerFiles.Handler,
		swagger.URL("/swagger/swagger.json"),
		swagger.DeepLinking(true),
		swagger.Prefix("/swagger"),
	)
	app.Get("/swagger", swaggerUI)
	app.Get("/swagger/{any:path}", swaggerUI)

	// 数据库连接
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", env.GetEnv("dbUserName", "paintingExchange"), env.GetEnv("dbPassword", "1234567"), env.GetEnv("dbName", "paintingExchange")),
	}), &gorm.Config{})
	if err != nil {
		log.Fatalln("数据库连接失败")
	} else {
		db.AutoMigrate(&model.User{})
	}

	// 绑定依赖和路由
	mvc.Configure(app, func(application *mvc.Application) {
		application.Register(db)
		application.Party("/user").Handle(new(controller.AuthController))
		application.Party("/user", service.JWTMiddleware).Handle(new(controller.UserController))
	})

	if err := app.Listen(":8880"); err != nil {
		log.Fatalln("启动失败")
	}
}
