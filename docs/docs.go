// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/user/login": {
            "post": {
                "description": "用户通过用户名和密码登录，成功后返回 JWT Token",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "用户登录",
                "parameters": [
                    {
                        "description": "用户登录信息(只需要username和password)",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.User"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "JWT Token",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "403": {
                        "description": "用户名或密码错误",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "服务器内部错误",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user/register": {
            "post": {
                "description": "用户进行注册，成功后返回 JWT Token",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "用户注册",
                "parameters": [
                    {
                        "description": "用户注册信息(只需要username和password)",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.User"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "JWT Token",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "403": {
                        "description": "用户名已存在",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "服务器内部错误",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model.User": {
            "description": "用户",
            "type": "object",
            "properties": {
                "avatarURI": {
                    "description": "头像地址",
                    "type": "string",
                    "example": "TODO"
                },
                "intro": {
                    "description": "描述",
                    "type": "string",
                    "example": "我是test"
                },
                "password": {
                    "description": "密码",
                    "type": "string",
                    "example": "123456"
                },
                "username": {
                    "description": "用户名",
                    "type": "string",
                    "example": "test"
                }
            }
        }
    },
    "securityDefinitions": {
        "jwt": {
            "type": "apiKey",
            "name": "Bearer",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8880",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "绘画交流平台",
	Description:      "绘画交流平台的后端API文档",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
