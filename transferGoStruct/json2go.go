package main

var (
	Xstr     = "%s"
	Xbegin   = "\ntype %s struct {"
	Xend     = "}\n"
	Xkeyv    = "%s %s %s"
	Xkeyvori = "%s %s %s"
	Xkeyvtag = "%s %s `json:\"%s\"`"
)

type xjson struct {
	Name      string
	Msg       string
	LowerCase bool
	UpperCase bool
	JSONTag   bool
	MapTag    map[string]string
	Parent    map[string]interface{}
	Sub       []map[int]string
	Out       []string
	// json文件，默认json2go.json
	JsonFile string
	// 输出类型[print file]，默认print
	OutType string
	// 输出文件，默认json2go_types.go
	OutFile string
}

// JsonNew returns a new xjson
func JsonNew(msg, jsonFile, outType, outFile string) *xjson {
	return &xjson{
		Name:     DefaultStructName,
		Msg:      msg,
		JSONTag:  true,
		MapTag:   map[string]string{},
		JsonFile: jsonFile,
		OutType:  outType,
		OutFile:  outFile,
	}
}

// reloade
func (xj *xjson) Flush() {
	*xj = *&xjson{
		Name:      xj.Name,
		Msg:       xj.Msg,
		MapTag:    map[string]string{},
		Parent:    map[string]interface{}{},
		Sub:       []map[int]string{},
		Out:       []string{},
		JSONTag:   true,
		LowerCase: false,
		UpperCase: false,
	}
}
