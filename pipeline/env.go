package pipeline

import (
	"errors"
	"os"
	"strconv"
)

func GetEnv(k string) (string, error) {
	logger := Logger().With("f", "GetEnv")
	s, ok := os.LookupEnv(k)
	logger.Debug("os.LookupEnv", "k", k, "s", s, "ok", ok)
	if ok {
		return s, nil
	}
	return s, errors.New("Environment variable not defined")
}

func GetEnvAsInt(k string) (int, error) {
	s, err := GetEnv(k)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
}

// https://appliedgo.com/blog/a-tip-and-a-trick-when-working-with-generics
// func Env[T any](k string, or T) (T, error) {
// 	s, ok := os.LookupEnv(workMaxEnv)
// 	if ok {
// 		switch any(or).(type) {
// 		case string:

// 		}

// 	}
// }
