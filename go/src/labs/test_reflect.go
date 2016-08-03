package main

import (
	"fmt"
	"reflect"
)

//////////////////////////////////////////////////////////////////////
// Google Go example
//////////////////////////////////////////////////////////////////////
func TestDish() {
	// iterate through the attributes of a Data Model instance
	for name, mtype := range attributes(&Dish{}) {
		fmt.Printf("Name: %s, Type %s\n", name, mtype.Name())
	}
}

// Data Model
type Dish struct {
	Id     int
	Name   string
	Origin string
	Query  func()
}

// Example of how to use Go's reflection
// Print the attributes of a Data Model
func attributes(m interface{}) map[string]reflect.Type {
	typ := reflect.TypeOf(m)
	// if a pointer to a struct is passed, get the type of the dereferenced object
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// create an attribute data structure as a map of types keyed by a string.
	attrs := make(map[string]reflect.Type)
	// Only structs are supported so return an empty result if the passed object
	// isn't a struct
	if typ.Kind() != reflect.Struct {
		fmt.Printf("%v type can't have attributes inspected\n", typ.Kind())
		return attrs
	}

	// loop through the struct's fields and set the map
	for i := 0; i < typ.NumField(); i++ {
		p := typ.Field(i)
		if !p.Anonymous {
			attrs[p.Name] = p.Type
		}
	}

	return attrs
}

//////////////////////////////////////////////////////////////////////
// 反射 struct
//////////////////////////////////////////////////////////////////////
type TRef struct {
	S       string
	N       int
	private string // 私有变量，无法用反射、空interface编辑
}

func test_SetStruct() {
	t := TRef{}
	SetStruct(&t)
	fmt.Println("=>", t)
}
func SetStruct(ptr interface{}) {
	data := reflect.ValueOf(ptr).Elem()
	fmt.Println(data.Type())
	fmt.Println(data.Field(0), data.Field(0).Type())
	fmt.Println(data.Field(1), data.Field(1).Type())
	fmt.Println(data.CanSet())

	fmt.Println(data.Field(0).CanSet()) // 可设置
	fmt.Println(data.Field(2).CanSet()) // 小写开头的私有成员：不可设置
	data.Field(0).SetString("渣渣")
	data.Field(1).SetInt(22222)

	fmt.Println("-------------------------")
}

//////////////////////////////////////////////////////////////////////
// 反射 map
//////////////////////////////////////////////////////////////////////
type IntPair struct {
	ID  int
	Cnt int
}
type TMap struct {
	ID    int
	Des   string
	Item  []IntPair
	Card  []IntPair
	Array []int
}

func test_SetMap() {
	m := make(map[int]*TMap)
	SetMap(&m)
	fmt.Println("=>", *m[0])
}
func SetMap(m interface{}) {
	table := reflect.ValueOf(m).Elem()
	typ := table.Type().Elem().Elem()

	data := reflect.New(typ).Elem()

	for i := 0; i < data.NumField(); i++ {
		field := data.Field(i)
		switch field.Kind() {
		case reflect.Int:
			{
				field.SetInt(22222)
			}
		case reflect.String:
			{
				field.SetString("渣渣")
			}
		case reflect.Slice:
			{
				switch field.Type().Elem().Kind() {
				case reflect.Int:
					{
						vec := []int{1, 1, 1, 1, 1}
						field.Set(reflect.ValueOf(vec))
					}
				case reflect.Struct:
					{
					}
				default:
					{
						fmt.Sprintf("Csv Type Error: TypeName:%s", data.Field(i).Type().String())
					}
				}
			}
		default:
			{
				fmt.Sprintf("Csv Type Error: TypeName:%s", data.Field(i).Type().String())
			}
		}
	}
	//data.Field(0).Set(reflect.ValueOf("渣渣"))

	//	fmt.Println(data.Field(3).Type().Elem())

	table.SetMapIndex(reflect.ValueOf(0), data.Addr())
}
