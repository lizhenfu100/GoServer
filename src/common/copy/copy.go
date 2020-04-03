package copy

import (
	"common/file"
	"net/url"
	"reflect"
	"strings"
)

//填充同名field，将url中的参数(小写)解析为结构体
func CopyForm(ptr interface{}, form url.Values) {
	val, typ := reflect.ValueOf(ptr).Elem(), reflect.TypeOf(ptr).Elem()
	for i := 0; i < typ.NumField(); i++ {
		if field := val.Field(i); field.CanSet() {
			name := strings.ToLower(typ.Field(i).Name)
			if vs, ok := form[name]; ok {
				if len(vs) > 0 {
					file.SetField(field, vs[0])
				} else {
					file.SetField(field, "")
				}
			}
		}
	}
}
func Form2Json(ptr interface{}, form url.Values) {
	val, typ := reflect.ValueOf(ptr).Elem(), reflect.TypeOf(ptr).Elem()
	for i := 0; i < typ.NumField(); i++ {
		if field := val.Field(i); field.CanSet() {
			name := typ.Field(i).Tag.Get("json")
			if vs, ok := form[name]; ok {
				if len(vs) > 0 {
					file.SetField(field, vs[0])
				} else {
					file.SetField(field, "")
				}
			}
		}
	}
}

//拷贝同名field
func CopySameField(pDest interface{}, pSrc interface{}) {
	val, typ := reflect.ValueOf(pDest).Elem(), reflect.TypeOf(pDest).Elem()
	src := reflect.ValueOf(pSrc).Elem()
	for i := 0; i < typ.NumField(); i++ {
		if field := val.Field(i); field.CanSet() {
			if v := src.FieldByName(typ.Field(i).Name); v.IsValid() {
				field.Set(v)
			}
		}
	}
}
