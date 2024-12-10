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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	// app.Logger().SetLevel("debug")

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
		DSN: fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", env.GetEnv("dbUserName", "paintingExchange"), env.GetEnv("dbPassword", "1234567"), env.GetEnv("dbHost", "localhost"), env.GetEnv("dbName", "paintingExchange")),
	}), &gorm.Config{})
	if err != nil {
		log.Fatalln("MySQL数据库连接失败")
	} else {
		db.AutoMigrate(&model.User{})
		db.AutoMigrate(&model.Star{})
		db.AutoMigrate(&model.Admin{})
		db.AutoMigrate(&model.Message{})
	}
	app.Use(func(ctx iris.Context) {
		ctx.Values().Set("db", db)
		ctx.Next()
	})

	// mongoDB数据库连接
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:27017", env.GetEnv("mgHost", "localhost")))
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

	// 算法端连接
	var algo service.SearchServiceClient
	if conn, err := grpc.NewClient(fmt.Sprintf("%s:8881", env.GetEnv("algoHost", "localhost")), grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		log.Fatalln("算法层gRPC连接失败", err)
	} else {
		defer conn.Close()
		algo = service.NewSearchServiceClient(conn)
	}

	// 创建图片缓存目录并绑定路由
	err = os.MkdirAll(env.GetImgDir(), os.ModePerm)
	err = os.MkdirAll(env.GetAvatarDir(), os.ModePerm)
	if err != nil {
		log.Fatalln("创建图片缓存目录失败:", err)
	}
	app.HandleDir("/assert/images", env.GetImgDir())
	app.HandleDir("/assert/avatars", env.GetAvatarDir())

	// 后台(测试用)
	app.HandleDir("/back", "webapp/back")
	app.HandleDir("/back/page", "webapp/back")

	// 绑定依赖和路由
	mvc.Configure(app, func(application *mvc.Application) {
		application.Register(db)
		application.Register(mg)
		application.Register(algo)
		application.Party("/").Handle(new(controller.AuthController))
		application.Party("/user", service.JWTMiddleware).Handle(new(controller.UserController))
		application.Party("/image", service.JWTMiddleware).Handle(new(controller.ImageController))
		application.Party("/back", service.JWTMiddleware, service.CheckIsAdmin).Handle(new(controller.BackController))
	})

	// 绑定websocket
	app.Get("/chat", service.BeginWsRequest, controller.HandleWebsocket)

	if err := app.Listen(":8880"); err != nil {
		log.Fatalln("启动失败")
	}
}
