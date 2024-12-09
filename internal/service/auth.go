package service

import (
	"PaintingExchange/internal/env"
	"PaintingExchange/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
	"log"
	"strings"
)

// JWTMiddleware JWT验证中间件
func JWTMiddleware(ctx iris.Context) {
	// 从上下文获取数据库对象
	db := ctx.Values().Get("db").(*gorm.DB)

	// 从请求头中获取 Authorization 字段
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		log.Println("缺少Authorization请求头")
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.Text(iris.StatusText(iris.StatusUnauthorized))
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		log.Println("Authorization请求头中缺少token")
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.Text(iris.StatusText(iris.StatusUnauthorized))
		return
	}

	// 解析 JWT Token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return env.GetJWTKey(), nil
	})
	if err != nil || !token.Valid {
		log.Println("jwt验证失败,token非法:", err)
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.Text(iris.StatusText(iris.StatusUnauthorized))
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("jwt验证失败,token解析失败")
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.Text(iris.StatusText(iris.StatusUnauthorized))
		return
	}

	user := model.User{
		Username: claims["username"].(string),
	}
	// 验证用户是否是管理员
	if isAdmin, exist := claims["isAdmin"]; exist && isAdmin.(bool) {
		// 不作处理
	} else {
		// 验证用户是否存在或被封禁
		rows := db.Where(&user).Find(&user).RowsAffected
		if rows == 0 || user.IsBan {
			log.Println("jwt验证失败,用户名不存在或被封禁")
			ctx.StatusCode(iris.StatusUnauthorized)
			ctx.Text(iris.StatusText(iris.StatusUnauthorized))
			return
		}
	}

	// 将用户信息存入上下文
	ctx.SetUser(iris.SimpleUser{
		Username:      user.Username,
		Authorization: authHeader,
	})
	ctx.Next()
}
