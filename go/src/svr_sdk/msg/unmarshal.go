package msg

import (
	"common"
	"net/url"
	"reflect"
	"strings"
)

func Unmarshal(ptr interface{}, form url.Values) {
	val, typ := reflect.ValueOf(ptr).Elem(), reflect.TypeOf(ptr).Elem()
	for i := 0; i < typ.NumField(); i++ {
		name := strings.ToLower(typ.Field(i).Name)
		if v, ok := form[name]; ok {
			common.SetField(val.Field(i), v[0])
		}
	}
}
