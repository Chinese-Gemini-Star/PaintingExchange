package service

import (
	"PaintingExchange/internal/env"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kataras/iris/v12"
	"log"
	"strings"
)

// JWTMiddleware JWT验证中间件
func JWTMiddleware(ctx iris.Context) {
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

	// 将用户信息存入上下文
	ctx.SetUser(iris.SimpleUser{
		Username:      claims["username"].(string),
		Authorization: authHeader,
	})
	ctx.Next()
}
