package main

const (
	// OutTypeForPrint
	OutTypeForPrint = "print"
	// OutTypeForFile
	OutTypeForFile = "file"
)

const (
	// DefaultJsonFile 默认json文件
	DefaultJsonFile = "transferGoStruct.json"
	// DefaultOutType 默认输出方式
	DefaultOutType = "print"
	// DefaultOutFile 默认输出文件
	DefaultOutFile = "gen_json2go_types.go"
	// DefaultStructName 结构体名称
	DefaultStructName = "Json2GoAutoGenerate"
)

const (
	Xmap    = "map[string]interface {}"
	Xlist   = "[]interface {}"
	Xstring = "string"
	goBegin = "package transferGoStruct\n\n"
)
