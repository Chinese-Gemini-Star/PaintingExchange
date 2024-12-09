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
	checkJwtFrom(ctx, "Authorization")
}

func checkJwtFrom(ctx iris.Context, header string) {
	// 从上下文获取数据库对象
	db := ctx.Values().Get("db").(*gorm.DB)

	// 从请求头中获取 Authorization 字段
	authHeader := ctx.GetHeader(header)
	if authHeader == "" {
		log.Println("缺少", header, "请求头")
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
	var roles []string
	// 验证用户是否是管理员
	if isAdmin, exist := claims["isAdmin"]; exist && isAdmin.(bool) {
		roles = append(roles, "admin")
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
		Roles:         roles,
	})
	ctx.Next()
}

func CheckIsAdmin(ctx iris.Context) {
	// 查询用户
	loginUser, err := ctx.User().GetRaw()
	if err != nil {
		log.Println("管理员验证失败,用户对象不存在")
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.Text(iris.StatusText(iris.StatusUnauthorized))
		return
	}

	// 检查身份是否包含管理员
	isAdmin := false
	for _, item := range loginUser.(iris.SimpleUser).Roles {
		if item == "admin" {
			isAdmin = true
		}
	}

	if !isAdmin {
		log.Println("管理员验证失败,用户不是管理员")
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.Text(iris.StatusText(iris.StatusUnauthorized))
		return
	}

	ctx.Next()
}

// BeginWsRequest Websocket连接鉴权
func BeginWsRequest(ctx iris.Context) {
	token := ctx.GetHeader("Sec-WebSocket-Protocol")
	ctx.Header("Sec-WebSocket-Protocol", token)
	checkJwtFrom(ctx, "Sec-WebSocket-Protocol")
}
