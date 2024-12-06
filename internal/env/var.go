package env

import "os"

// GetEnv 获取环境变量
func GetEnv(key string, value string) string {
	env, err := os.LookupEnv(key)
	if !err {
		env = value
	}
	return env
}

// GetJWTKey 获取jwt密钥
func GetJWTKey() []byte {
	return []byte("PaintingExchange")
}

func GetImgDir() string {
	return "assert/images"
}

func GetAvatarDir() string {
	return "assert/avatars"
}
