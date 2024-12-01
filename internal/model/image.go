package model

type Image struct {
	ID        string   `json:"id" bson:"id"`               // 图片id(UUID)
	Auth      string   `json:"auth" bson:"auth"`           // 图片作者用户名
	BigURI    string   `json:"bigURI" bson:"bigURI"`       // 大图地址
	MediumURI string   `json:"mediumURI" bson:"mediumURI"` // 中图地址
	SmallURI  string   `json:"smallURI" bson:"smallURI"`   // 小图地址
	Label     []string `json:"label" bson:"label"`         // 图片标签
	Intro     string   `json:"intro" bson:"intro"`         // 图片简介
	Like      int      `json:"like" bson:"like"`           // 收藏人数
}
