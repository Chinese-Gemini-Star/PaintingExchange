package model

// Image 图片
// @Description 图片
type Image struct {
	ID        string   `json:"id" bson:"id" example:"68c8d808-54f7-4cfc-94c9-015416033dc9"`                               // 图片id(UUID)
	Auth      string   `json:"auth" bson:"auth" example:"test"`                                                           // 图片作者用户名
	BigURI    string   `json:"bigURI" bson:"bigURI" example:"assert/images/big_68c8d808-54f7-4cfc-94c9-015416033dc9.jpg"` // 大图地址
	MediumURI string   `json:"mediumURI" bson:"mediumURI" example:"TODO"`                                                 // 中图地址
	SmallURI  string   `json:"smallURI" bson:"smallURI" example:"TODO"`                                                   // 小图地址
	Label     []string `json:"label" bson:"label"`                                                                        // 图片标签
	Intro     string   `json:"intro" bson:"intro"`                                                                        // 图片简介
	Like      int      `json:"like" bson:"like" example:"0"`                                                              // 收藏人数
}
