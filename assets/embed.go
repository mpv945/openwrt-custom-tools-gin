package assets

import "embed"

//go:embed excel/*
var FS embed.FS //上面go:embed和变量紧挨着，可以精确到文件; 如果多个可以：//go:embed excel/* templates/* images/*
