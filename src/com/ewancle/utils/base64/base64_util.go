package base64util

import (
	"encoding/base64"
)

// EncodeString 字符串 -> Base64字符串
func EncodeString(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// DecodeString Base64字符串 -> 字符串
func DecodeString(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// EncodeBytes []byte -> Base64字符串
func EncodeBytes(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBytes Base64字符串 -> []byte
func DecodeBytes(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// URLEncodeString URL安全Base64编码
func URLEncodeString(s string) string {
	return base64.URLEncoding.EncodeToString([]byte(s))
}

// URLDecodeString URL安全Base64解码
func URLDecodeString(s string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MustDecodeString 生产环境一般会加 Must 方法：
func MustDecodeString(s string) string {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return string(data)
}
