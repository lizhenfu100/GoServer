/***********************************************************************
* @ 反射解析表结构
* @ brief
	1、表数据格式：
			数  值：	1234
			字符串：	zhoumf
			数  组：	[10,20,30]  [[1,2], [2,3]]
			数值对：	同结构体				旧格式：(24|1)(11|1)...
			结构体：	{"ID": 233, "Cnt": 1}	新格式：Json
			Map：	{"key1": 1, "key2": 2}  转换为JSON的Object，key必须是string

			物品权重表，可配成两列：[]IntPair + []int

	2、代码数据格式
			type TTestCsv struct { // 字段须与csv表格的顺序一致
				Num  int
				Str  string
				Arr1 []int
				Arr2 []string
				Arr3 [][]int
				St   struct {
					ID  int
					Cnt int
				}
				Sts []struct {
					ID  int
					Cnt int
				}
				M map[string]int
			}

	3、首次出现的有效行(非注释的)，即为表头

	4、行列注释："#"开头的行，没命名/前缀"_"的列    有些列仅client显示用的

	5、使用方式：
			var G_MapCsv map[int]*TTestCsv = nil  	// map结构读表，首列作Key
			var G_SliceCsv []TTestCsv = nil 		// 数组结构读表，注册【&G_SliceCsv】到G_Csv_Map

			var G_Csv_Map = map[string]interface{}{
				"test": &G_MapCsv,
				// "test": &G_SliceCsv,
			}
* @ author zhoumf
* @ date 2016-6-22
***********************************************************************/
package file

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var G_Csv_Map map[string]interface{}

func LoadAllCsv() {
	if names, err := WalkDir("csv/", ""); err == nil {
		for _, name := range names {
			LoadOneCsv(name)
		}
	} else {
		fmt.Printf("LoadAllCsv error : %s\n", err.Error())
	}
}
func LoadOneCsv(fullName string) {
	if ptr := getRegPtr(fullName); ptr == nil {
		fmt.Printf("%s not regist in G_Csv_Map\n", fullName)
	} else if records, err := ReadCsv(fullName); err != nil {
		fmt.Printf("ReadCsv error : %s\n", err.Error())
	} else {
		ParseRefCsv(records, ptr)
	}
}
func getRegPtr(fullName string) interface{} {
	name := strings.TrimSuffix(fullName, ".csv")
	name = strings.TrimPrefix(name, "csv/")
	if ptr, ok := G_Csv_Map[name]; ok {
		return ptr
	}
	return nil
}

// -------------------------------------
// 反射解析
func ParseRefCsv(records [][]string, ptr interface{}) {
	switch reflect.TypeOf(ptr).Elem().Kind() {
	case reflect.Map:
		ParseRefCsvByMap(records, ptr)
	case reflect.Slice:
		ParseRefCsvBySlice(records, ptr)
	case reflect.Struct:
		ParseRefCsvByStruct(records, ptr)
	default:
		fmt.Printf("Csv Type Error: TypeName:%s\n", reflect.TypeOf(ptr).Elem().String())
	}
}
func ParseRefCsvByMap(records [][]string, pMap interface{}) {
	table := reflect.ValueOf(pMap).Elem()
	typ := table.Type().Elem().Elem() // map内保存的指针，第二次Elem()得到所指对象类型
	table.Set(reflect.MakeMap(table.Type()))

	total, idx := _GetRecordsValidCnt(records), 0
	slice := reflect.MakeSlice(reflect.SliceOf(typ), total, total) // 避免多次new对象，直接new数组，拆开用

	bParsedName, nilFlag := false, int64(0)
	for _, v := range records {
		if strings.Index(v[0], "#") == -1 { // "#"起始的不读
			if !bParsedName {
				nilFlag = _parseHead(v)
				bParsedName = true
			} else {
				// data := reflect.New(typ).Elem()
				data := slice.Index(idx)
				idx++
				_parseData(v, nilFlag, data)
				if table.MapIndex(data.Field(0)).IsValid() {
					panic("csv map key is repeated !!!\n")
				} else {
					table.SetMapIndex(data.Field(0), data.Addr())
				}
			}
		}
	}
}
func ParseRefCsvBySlice(records [][]string, pSlice interface{}) { // slice可减少对象数量，降低gc
	slice := reflect.ValueOf(pSlice).Elem() // 这里slice是nil
	typ := reflect.TypeOf(pSlice).Elem()

	total, idx := _GetRecordsValidCnt(records), 0
	slice.Set(reflect.MakeSlice(typ, total, total))

	bParsedName, nilFlag := false, int64(0)
	for _, v := range records {
		if strings.Index(v[0], "#") == -1 { // "#"起始的不读
			if !bParsedName {
				nilFlag = _parseHead(v)
				bParsedName = true
			} else {
				data := slice.Index(idx)
				idx++
				_parseData(v, nilFlag, data)
			}
		}
	}
}
func ParseRefCsvByStruct(records [][]string, pStruct interface{}) {
	st := reflect.ValueOf(pStruct).Elem()
	for _, v := range records {
		if strings.Index(v[0], "#") == -1 { // "#"起始的不读
			SetField(st.FieldByName(v[0]), v[1])
		}
	}
}
func _parseHead(record []string) (nilFlag int64) { // 不读的列：没命名/前缀"_"
	length := len(record)
	if length > 64 {
		panic("csv column is over to 64 !!!\n")
	}
	for i := 0; i < length; i++ {
		if record[i] == "" || strings.Index(record[i], "_") == 0 {
			nilFlag |= (1 << uint(i))
		}
	}
	return nilFlag
}
func _parseData(record []string, nilFlag int64, data reflect.Value) {
	idx := 0
	for i, s := range record {
		if nilFlag&(1<<uint(i)) > 0 { // 跳过没命名的列
			continue
		}

		field := data.Field(idx)
		idx++

		if s == "" { // 没填的就不必解析了，跳过，idx还是要自增哟
			continue
		}
		SetField(field, s)
	}
}
func SetField(field reflect.Value, s string) {
	if !field.IsValid() || !field.CanSet() {
		return
	}
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		{
			if v, err := strconv.ParseInt(s, 0, field.Type().Bits()); err == nil {
				field.SetInt(v)
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		{
			if v, err := strconv.ParseUint(s, 0, field.Type().Bits()); err == nil {
				field.SetUint(v)
			}
		}
	case reflect.Float32, reflect.Float64:
		{
			if v, err := strconv.ParseFloat(s, field.Type().Bits()); err == nil {
				field.SetFloat(v)
			}
		}
	case reflect.String:
		{
			field.SetString(s)
		}
	case reflect.Struct, reflect.Map:
		{
			if err := json.Unmarshal([]byte(s), field.Addr().Interface()); err != nil {
				fmt.Errorf("Field Parse Error: (content=%v): %v", s, err)
			}
		}
	case reflect.Slice:
		{
			switch field.Type().Elem().Kind() {
			case reflect.String:
				{
					sFix := strings.Trim(strings.Replace(s, " ", "", -1), "[]")
					vec := strings.Split(sFix, ",")
					field.Set(reflect.ValueOf(vec))
				}
			case reflect.Int, reflect.Uint32, reflect.Float32, reflect.Struct, reflect.Slice:
				{
					if err := json.Unmarshal([]byte(s), field.Addr().Interface()); err != nil {
						fmt.Errorf("Field Parse Error: (content=%v): %v", s, err)
					}
				}
			default:
				{
					fmt.Printf("Field Type Error: TypeName:%s", field.Type().String())
				}
			}
		}
	default:
		fmt.Printf("Field Type Error: TypeName:%s\n", field.Type().String())
	}
}
func _GetRecordsValidCnt(records [][]string) (ret int) {
	for _, v := range records {
		if strings.Index(v[0], "#") == -1 { // "#"起始的不读
			ret++
		}
	}
	return ret - 1 //减掉表头那一行
}

// -------------------------------------
// 读写csv文件
func ReadCsv(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	r := csv.NewReader(file)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}
func UpdateCsv(dir, name string, records [][]string) error {
	file, err := CreateFile(dir, name, os.O_TRUNC|os.O_WRONLY)
	if err != nil {
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	return w.WriteAll(records)
}
func AppendCsv(dir, name string, record []string) error {
	file, err := CreateFile(dir, name, os.O_APPEND|os.O_WRONLY)
	if err != nil {
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	if err := w.Write(record); err != nil {
		return err
	}
	w.Flush()
	return nil
}
