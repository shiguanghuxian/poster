package common

import (
	"errors"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GetRootDir 获取执行路径
func GetRootDir() string {
	// 文件不存在获取执行路径
	file, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		file = fmt.Sprintf(".%s", string(os.PathSeparator))
	} else {
		file = fmt.Sprintf("%s%s", file, string(os.PathSeparator))
	}
	return file
}

// PathExists 判断文件或目录是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// HexToColor 16进制颜色转Color
func HexToColor(str string) (color.Color, error) {
	if len(str) < 6 {
		return nil, errors.New("Illegal hexadecimal color")
	}
	if strings.HasPrefix(str, "#") == true {
		str = str[1:]
	}
	if len(str) != 6 {
		return nil, errors.New("Illegal hexadecimal color")
	}

	r, err := strconv.ParseInt(str[:2], 16, 10)
	if err != nil {
		return nil, err
	}
	g, err := strconv.ParseInt(str[2:4], 16, 10)
	if err != nil {
		return nil, err
	}
	b, err := strconv.ParseInt(str[4:], 16, 10)
	if err != nil {
		return nil, err
	}

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 255,
	}, nil
}
