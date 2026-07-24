package handler

import "os"

func lookupEnv(key string) string {
	return os.Getenv(key)
}
