package common

import (
	"common/file"
	"net/url"
	"reflect"
	"strings"
)

//填充同名field，将url中的参数(小写)解析为结构体
func Unmarshal(ptr interface{}, form url.Values) {
	val, typ := reflect.ValueOf(ptr).Elem(), reflect.TypeOf(ptr).Elem()
	for i := 0; i < typ.NumField(); i++ {
		if val.Field(i).CanSet() {
			name := strings.ToLower(typ.Field(i).Name)
			if v := form.Get(name); v != "" {
				file.SetField(val.Field(i), v)
			}
		}
	}
}

//拷贝同名field
func CopySameField(pDest interface{}, pSrc interface{}) {
	typ1 := reflect.TypeOf(pDest).Elem()
	val1 := reflect.ValueOf(pDest).Elem()
	val2 := reflect.ValueOf(pSrc).Elem()
	for i := 0; i < typ1.NumField(); i++ {
		if v := val2.FieldByName(typ1.Field(i).Name); v.IsValid() && v.CanSet() {
			val1.Field(i).Set(v)
		}
	}
}
