package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // 必须导入,否则出问题
	ozlog "github.com/usthooz/oozlog/go"
	"log"
	"os"
	"strings"
	"text/template"
	"unicode"
)

type DBModel struct {
	DBEngine *sql.DB
	DBInfo   *DBInfo
}

type DBInfo struct {
	DBType    string
	Host      string
	UserName  string
	Password  string
	Charset   string
	DbName    string
	TableName string
}

type TableColumn struct {
	ColumnName    string
	DataType      string
	IsNullable    string
	ColumnKey     string
	ColumnType    string
	ColumnComment string
}

// 表字段类型映射
var DBTypeToStructType = map[string]string{
	"int":        "int32",
	"tinyint":    "int8",
	"smallint":   "int",
	"mediumint":  "int64",
	"bigint":     "int64",
	"bit":        "int",
	"bool":       "bool",
	"enum":       "string",
	"set":        "string",
	"varchar":    "string",
	"char":       "string",
	"tinytext":   "string",
	"mediumtext": "string",
	"text":       "string",
	"longtext":   "string",
	"blob":       "string",
	"tinyblob":   "string",
	"mediumblob": "string",
	"longblob":   "string",
	"date":       "time.Time",
	"datetime":   "time.Time",
	"timestamp":  "time.Time",
	"time":       "time.Time",
	"float":      "float64",
	"double":     "float64",
}

func NewDBModel(info *DBInfo) *DBModel {
	return &DBModel{DBInfo: info}
}

func (m *DBModel) Connect() error {
	var err error
	s := "%s:%s@tcp(%s)/information_schema?" +
		"charset=%s&parseTime=True&loc=Local"
	dns := fmt.Sprintf(
		s,
		m.DBInfo.UserName,
		m.DBInfo.Password,
		m.DBInfo.Host,
		m.DBInfo.Charset,
	)
	m.DBEngine, err = sql.Open(m.DBInfo.DBType, dns)
	if err != nil {
		return err
	}
	return nil
}

// function GetColumns returns the info of columns of specified table
func (m *DBModel) GetColumns(dbName, tableName string) ([]*TableColumn, error) {
	query := "SELECT COLUMN_NAME, DATA_TYPE, COLUMN_KEY, " +
		"IS_NULLABLE, COLUMN_TYPE, COLUMN_COMMENT " +
		"FROM COLUMNS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? "
	rows, err := m.DBEngine.Query(query, dbName, tableName)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return nil, errors.New("表中没有数据")
	}
	defer rows.Close()

	var columns []*TableColumn
	for rows.Next() {
		var column TableColumn
		err := rows.Scan(&column.ColumnName, &column.DataType, &column.ColumnKey, &column.IsNullable, &column.ColumnType, &column.ColumnComment)
		if err != nil {
			return nil, err
		}

		columns = append(columns, &column)
	}
	return columns, nil
}

const printStructTpl = `type {{.TableName | ToCamelCase}} struct {
{{range .Columns}}	{{ $length := len .Comment}} {{ if gt $length 0 }}// {{.Comment}} {{else}}// {{.Name}} {{ end }}
	{{ $typeLen := len .Type }} {{ if gt $typeLen 0 }}{{.Name | ToCamelCase}}	{{.Type}}	{{.Tag}}{{ else }}{{.Name}}{{ end }}
{{end}}}

func (model {{.TableName | ToCamelCase}}) TableName() string {
	return "{{.TableName}}"
}`

const fileStructTpl = `type {{.TableName | ToCamelCase}} struct {
{{range .Columns}}	{{ $length := len .Comment}} {{ if gt $length 0 }}// {{.Comment}} {{else}}// {{.Name}} {{ end }}
	{{ $typeLen := len .Type }} {{ if gt $typeLen 0 }}{{.Name | ToCamelCase}}	{{.Type}}	{{.Tag}}{{ else }}{{.Name}}{{ end }}
{{end}}}`

type StructTemplate struct {
	structTpl string
}

type StructColumn struct {
	Name    string
	Type    string
	Tag     string
	Comment string
}

type StructTemplateDB struct {
	TableName string
	Columns   []*StructColumn
}

func NewStructTemplate() *StructTemplate {
	if outType == OutTypeForPrint {
		return &StructTemplate{structTpl: printStructTpl}
	} else if outType == OutTypeForFile {
		return &StructTemplate{structTpl: fileStructTpl}
	}
	return nil
}

func (t *StructTemplate) AssemblyColumns(tbColumns []*TableColumn) []*StructColumn {

	//fmt.Println("------------------templateColumns start-----------------")
	//for _, v := range tbColumns {
	//	fmt.Printf("ColumnName: %s, ColumnType: %s, ColumnKey: %s, ColumnComment: %s, DataType: %s, IsNullable: %s \n", v.ColumnName, v.ColumnType, v.ColumnKey, v.ColumnComment, v.DataType, v.IsNullable)
	//}
	//fmt.Println("------------------templateColumns end-----------------")

	tplColumns := make([]*StructColumn, 0, len(tbColumns))
	for _, column := range tbColumns {
		var tag string
		if len(column.ColumnKey) > 0 && column.ColumnKey == "PRI" {
			tag = fmt.Sprintf("`"+"json:"+"\"%s\" "+"gorm:"+"\"column:%s;primary_key;comment:%s;\""+"`", column.ColumnName, column.ColumnName, column.ColumnComment)
		} else {
			tag = fmt.Sprintf("`"+"json:"+"\"%s\" "+"gorm:"+"\"column:%s;comment:%s;\""+"`", column.ColumnName, column.ColumnName, column.ColumnComment)
		}
		tplColumns = append(tplColumns, &StructColumn{
			Name:    column.ColumnName,
			Type:    DBTypeToStructType[column.DataType],
			Tag:     tag,
			Comment: column.ColumnComment,
		})
	}

	return tplColumns
}

func (t *StructTemplate) Generate(tableName string, tplColumns []*StructColumn) (string, error) {
	tpl := template.Must(template.New("sql2struct").Funcs(template.FuncMap{
		"ToCamelCase": UnderscoreToUpperCamelCase,
	}).Parse(t.structTpl))

	var buf bytes.Buffer

	tplDB := StructTemplateDB{
		TableName: tableName,
		Columns:   tplColumns,
	}
	err := tpl.Execute(&buf, tplDB)
	result := buf.String()
	if err != nil {
		return "", err
	}
	return result, nil
}

// 单词全部转为大写/小写
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

func ToLower(s string) string {
	return strings.ToLower(s)
}

// 下划线单词转大写驼峰单词
func UnderscoreToUpperCamelCase(s string) string {
	s = strings.Replace(s, "_", " ", -1)
	s = strings.Title(s)
	return strings.Replace(s, " ", "", -1)
}

// 下划线单词转小写驼峰单词
func UnderscoreToLowerCamelCase(s string) string {
	s = UnderscoreToUpperCamelCase(s)
	return string(unicode.ToLower(rune(s[0]))) + s[1:]
}

// 驼峰单词转下划线单词
func CamelCaseToUnderscore(s string) string {
	var output []rune
	for i, r := range s {
		if i == 0 {
			output = append(output, unicode.ToLower(r))
			continue
		}
		if unicode.IsUpper(r) {
			output = append(output, '_')
		}
		output = append(output, unicode.ToLower(r))
	}
	return string(output)
}

func sql2go(dbInfo *DBInfo) {

	dbModel := NewDBModel(dbInfo)

	err := dbModel.Connect()
	if err != nil {
		log.Fatalf("dbModel.Connect err: %v", err)
	}

	columns, err := dbModel.GetColumns(dbInfo.DbName, dbInfo.TableName)
	if err != nil {
		log.Fatalf("dbModel.GetColumns err: %v", err)
	}

	template := NewStructTemplate()

	templateColumns := template.AssemblyColumns(columns)
	str, err := template.Generate(tableName, templateColumns)
	if err != nil {
		log.Fatalf("template.Generate err: %v", err)
	}
	sn := SqlNew(str, outFile)
	sn.write2file()
}

type xsql struct {
	Msg string
	// 输出文件，默认json2go_types.go
	OutFile string
}

// New returns a new xsql
func SqlNew(msg, outFile string) *xsql {
	return &xsql{
		Msg:     msg,
		OutFile: outFile,
	}
}

func (xl *xsql) write2file() {
	if outType == OutTypeForFile {
		if len(xl.OutFile) == 0 {
			xl.OutFile = DefaultOutFile
		}
		file, err := os.OpenFile(xl.OutFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			log.Fatalf("os OpenFile err: %v", err)
			return
		}
		file.WriteString(goBegin)
		file.WriteString(xl.Msg + "\n")
		ozlog.Infof("生成文件 %s", outFile)
	} else if outType == OutTypeForPrint {
		if _, err := os.Stdout.WriteString(xl.Msg); err != nil {
			log.Fatalf("write stdout err: %v", err)
		}
	}
}
