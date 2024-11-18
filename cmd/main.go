package main

import (
	"PaintingExchange/internal/controller"
	"PaintingExchange/internal/env"
	"PaintingExchange/internal/model"
	"fmt"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

func main() {
	app := iris.New()
	app.UseRouter(cors.New(cors.Options{AllowedOrigins: []string{"*"}}))

	// 数据库连接
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", env.GetEnv("dbUserName", "paintingExchange"), env.GetEnv("dbPassword", "1234567"), env.GetEnv("dbName", "paintingExchange")),
	}), &gorm.Config{})

	if err != nil {
		log.Fatalln("数据库连接失败")
	} else {
		db.AutoMigrate(&model.User{})
	}

	mvc.Configure(app, func(application *mvc.Application) {
		application.Register(db)
		application.Party("/user").Handle(new(controller.UserController))
	})

	if err := app.Listen(":8880"); err != nil {
		log.Fatalln("启动失败")
	}
}
