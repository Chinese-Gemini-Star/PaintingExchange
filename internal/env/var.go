package env

import "os"

func GetEnv(key string, value string) string {
	env, err := os.LookupEnv(key)
	if !err {
		env = value
	}
	return env
}

func GetJWTKey() []byte {
	return []byte("PaintingExchange")
}
