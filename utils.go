package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Args map[string]string

func getExeDir() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	if strings.Contains(exePath, "go-build") {
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			log.Fatal("Path not found")
		}
		return filepath.Dir(filename)
	}
	if path, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = path
	}
	return filepath.Dir(exePath)
}

func resolveFile(filename string) (string, error) {
	exeDir := getExeDir()
	paths := []string{
		filepath.Join(filename),
		filepath.Join(exeDir, filename),
		filepath.Join(".", filename),
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return filename, fmt.Errorf("file not found: %s", filename)
}

func parseArgs() Args {
	args := os.Args[1:]
	params := make(Args)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// 仅处理以 "-" 开头的参数
		if len(arg) > 1 && arg[0] == '-' {
			// 移除所有前导的 "-"
			keyPart := strings.TrimLeft(arg, "-")
			if keyPart == "" {
				// 忽略无效参数（如单独的 "--"）
				continue
			}

			// 分割键和值（支持等号分隔符）
			parts := strings.SplitN(keyPart, "=", 2)
			key := parts[0]
			var val string

			if len(parts) == 2 {
				// 情况 1: 等号分隔（-key=value）
				val = parts[1]
			} else {
				// 情况 2: 空格分隔（-key value）
				if i+1 < len(args) && !isFlag(args[i+1]) {
					val = args[i+1]
					i++ // 跳过已处理的值
				}
			}

			// 存储键值对（允许空值）
			params[key] = val
		}
	}
	return params
}

// 判断参数是否为标志（以 "-" 开头）
func isFlag(s string) bool {
	return len(s) > 0 && s[0] == '-'
}
