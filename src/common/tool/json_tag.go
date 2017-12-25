/***********************************************************************
* @ produce struct tag code for object
* @ brief
	1、为结构体中的变量，生成json的tag
	2、把单词用下划线连(通过大写字母来区分)
	3、带有tag的struct声明，可用json.Unmarshal(buf, &obj)直接解析

* @ tools: json -> go【https://mholt.github.io/json-to-go/】

* @ example
		type MyStruct struct {
			Name      string
			MaxHeight int
		}
		var s MyStruct
		fmt.Prinln( tool.ProduceStructTag(s, "json"))

* @ result
		type MyStruct struct {
			Name      string `json:"name"`
			MaxHeight int    `json:"max_height"`
		}

		b := []byte("name": "zhoumf" ,"max_height": 233)
		if err := json.Unmarshal(b, &s); err != nil {
	        fmt.Println(err)
	        return
	    }

* @ author zhoumf
* @ date 2017-5-3
***********************************************************************/
package tool

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

func ProduceStructTag(obj interface{}, tag string) string {
	var newDefineCode string
	s := reflect.ValueOf(obj)
	newDefineCode = fmt.Sprintf("type %s struct {\n", s.Type().String())
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		n := s.Type().Field(i).Name
		newDefineCode = fmt.Sprintf("%s\t%s\t%s\t\t%s\n",
			newDefineCode,
			n,
			f.Type(),
			getTagName(n, tag))
	}
	newDefineCode = fmt.Sprintf("%s}\n", newDefineCode)
	return newDefineCode
}
func getTagName(name, tag string) (retName string) {
	isFirst := true
	for _, r := range name {
		if unicode.IsUpper(r) {
			if isFirst {
				retName = fmt.Sprintf("%s%s", retName, strings.ToLower(string(r)))
				isFirst = false
			} else {
				retName = fmt.Sprintf("%s_%s", retName, strings.ToLower(string(r)))
			}
		} else {
			retName = fmt.Sprintf("%s%s", retName, string(r))
		}
	}
	retName = fmt.Sprintf("`%s:\"%s\"`", tag, retName)
	return
}
