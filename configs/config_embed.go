package configs_embed

import "embed"

//go:embed config.yaml
var CFS embed.FS //上面go:embed和变量紧挨着，可以精确到文件; 如果多个可以：//go:embed excel/* templates/* images/*
