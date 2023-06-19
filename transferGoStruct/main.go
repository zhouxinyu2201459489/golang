package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	ozlog "github.com/usthooz/oozlog/go"
)

const (
	// 主命令
	exec = "transferGoStruct"
	// version 当前版本
	version = "v1.0"
)

var (
	// command 命令
	command string
	// workPath current work path
	workPath string
	// jsonFile json文件名称
	jsonFile string
	// outputFile 输出文件名称
	outFile string
	// outType 输出类型
	outType string
)

var (
	// commandsMap 命令集
	commandMap map[string]*Command
)

var (
	// 用户名
	username string
	// 密码
	password string
	// ip端口
	host string
	// 编码格式
	charset string
	// 数据库类型
	dbType string
	// 数据库名
	dbName string
	// 表名
	tableName string
)

// Command
type Command struct {
	Name   string
	Detail string
	Func   func(name, detail string)
}

func init() {
	flag.StringVar(&jsonFile, "json_file", "transferGoStruct.json", "json file.")
	flag.StringVar(&outType, "out_type", "print", "struct out type.")
	flag.StringVar(&outFile, "out_file", "gen_json2go_types.go", "output file.")

	flag.StringVar(&username, "user_name", "root", "check username.")
	flag.StringVar(&password, "password", "", "check password.")
	flag.StringVar(&host, "host", "127.0.0.1:3306", "check host.")
	flag.StringVar(&charset, "charset", "utf8mb4", "check charset.")
	flag.StringVar(&dbType, "db_type", "mysql", "check dbType.")
	flag.StringVar(&dbName, "db_name", "", "check dbName.")
	flag.StringVar(&tableName, "table_name", "", "check tableName.")
}

// initCommands
func initCommands() {
	for i, v := range os.Args {
		switch i {
		case 1:
			command = v
		}
	}

	// 初始化命令列表
	commandMap = map[string]*Command{
		"v": &Command{
			Name:   "v",
			Detail: "查看当前版本号",
			Func:   getVersion,
		},
		"help": &Command{
			Name:   "help",
			Detail: "查看帮助信息",
			Func:   getHelp,
		},
		"json2go": &Command{
			Name:   "json2go",
			Detail: "根据json文件自动生成struct",
			Func:   json2goStruct,
		},
		"sql2go": &Command{
			Name:   "sql2go",
			Detail: "根据db_sql连接自动生成struct",
			Func:   sql2goStruct,
		},
	}
}

// getHelp get this project's help
func getHelp(name, detail string) {
	commands := make([]string, 0, len(commandMap))
	for _, v := range commandMap {
		commands = append(commands, fmt.Sprintf("%s\t%s", v.Name, v.Detail))
	}
	outputHelp(fmt.Sprintf("Usage: %s <command>", exec), commands, []string{
		"-json_file\t json2go: json文件, 默认json文件为json2go.json",
		"-out_type\t 输出类型, 默认输出方式为输出到文件file/可选print、file",
		"-out_file\t 输出文件, 默认输出文件为gen_json2go_types.go",
		"-user_name\t sql2go: 用户名, 默认root",
		"-password\t sql2go: 密码, 默认''",
		"-host\t sql2go: db连接地址, 默认127.0.0.1:3306",
		"-charset\t sql2go: 编码集, 默认utf8mb4",
		"-db_type\t sql2go: 连接类型, 默认mysql",
		"-db_name\t sql2go: 数据库名称, 默认''",
		"-table_name\t sql2go: 表名, 默认''",
	}, []string{
		"transferGoStruct json2go",
		"transferGoStruct json2go -out_type=print",
		"transferGoStruct json2go -out_type=file",
		"transferGoStruct json2go -out_type=file -out_file=out_types.go",
		"transferGoStruct sql2go -out_type=print -user_name=root -password=root -db_name=test -table_name=test",
		"transferGoStruct sql2go -out_type=file -out_file=out_types.go -user_name=root -password=root -db_name=test -table_name=test",
	})
}

func outputHelp(usage string, commands, options, examples []string) {
	fmt.Println("\n", usage)
	if len(commands) > 0 {
		fmt.Println("\n Commands:")
		for _, s := range commands {
			fmt.Println(fmt.Sprintf("\t%s", s))
		}
	}
	if len(options) > 0 {
		fmt.Println("\n Options:")
		for _, s := range options {
			fmt.Println(fmt.Sprintf("\t%s", s))
		}
	}
	if len(examples) > 0 {
		fmt.Println("\n Examples:")
		for _, s := range examples {
			fmt.Println(fmt.Sprintf("\t%s", s))
		}
	}
	fmt.Println()
}

// getVersion 查看当前版本
func getVersion(name, detail string) {
	fmt.Println(version)
}

// json2goStruct
func json2goStruct(name, detail string) {
	ozlog.Infof("开始生成结构...")
	readJsonAndGen(jsonFile, outType, outFile)
	if outType == OutTypeForFile {
		ozlog.Infof("生成文件 %s", outFile)
	}
	ozlog.Infof("生成结构完成...")
}

// sql2goStruct
func sql2goStruct(name, detail string) {
	ozlog.Infof("开始连接db...")
	dbInfo := &DBInfo{
		DBType:    dbType,
		Host:      host,
		UserName:  username,
		Password:  password,
		Charset:   charset,
		DbName:    dbName,
		TableName: tableName,
	}
	sql2go(dbInfo)
	ozlog.Infof("生成结构完成...")
}

// checkArgs check common is nil?
func checkArgs() bool {
	if len(command) == 0 {
		getHelp("help", commandMap["help"].Detail)
		return false
	}
	return true
}

// getWorkDir get current work dir
func getWorkDir() {
	// get current dir
	currentDir, err := os.Getwd()
	if err != nil {
		ozlog.Fatalf("%s", err)
	}
	// gei this window all gopath
	pathList := strings.Split(os.Getenv("GOPATH"), ":")
	for _, path := range pathList {
		if strings.HasPrefix(currentDir, path) {
			var (
				prefix string
			)
			// check this path ends
			if strings.HasSuffix(path, "/") {
				// /Users/admin/work/goProject/
				prefix = path + "src/"
			} else {
				// /Users/admin/work/goProject
				prefix = path + "/src/"
			}
			// output: github.com/usthooz/oozgorm
			currentDir = currentDir[len(prefix):]
		}
		workPath = currentDir
	}
}

// 测试本地数据库时打开
//func main() {
//	dbInfo := &DBInfo{
//		DBType:    "mysql",
//		Host:      "127.0.0.1:3306",
//		UserName:  "root",
//		Password:  "Zhouxl950319!",
//		Charset:   "utf8mb4",
//		DbName:    "herman",
//		TableName: "admin",
//	}
//	sql2go(dbInfo)
//}

func main() {
	// 获取当前目录
	getWorkDir()
	// 初始化命令
	initCommands()
	if len(os.Args) < 2 {
		getHelp("help", commandMap["help"].Detail)
		return
	}
	flag.CommandLine.Parse(os.Args[2:])
	if !checkArgs() {
		return
	}
	c := commandMap[command]
	if c == nil {
		getHelp("help", commandMap["help"].Detail)
		return
	} else {
		c.Func(c.Name, c.Detail)
	}
}
