/***********************************************************************
* @ 反射解析表结构
* @ brief
	1、表数据格式：
			数  值：1234
			字符串：zhoumf
			数值对：(24|1)(11|1)...
			数  组：10|20|30...

	2、首次出现的有效行(非注释的)，即为表头

	3、行列注释："#"开头的行，没命名/前缀"(c)"的列    有些列仅client显示用的

	4、使用方式如下：
			type TTestCsv struct { // 字段须与csv表格的顺序一致
				ID    int
				Des   string
				Item  []IntPair
				Card  []IntPair
				Array []string
			}
			var G_MapCsv = make(map[int]*TTestCsv)  // map结构读表，将【&G_MapCsv】注册进G_ReflectParserMap即可自动读取
			var G_SliceCsv []TTestCsv = nil 		// 数组结构读表，注册【&G_SliceCsv】
* @ author zhoumf
* @ date 2016-6-22
***********************************************************************/
package common

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

//////////////////////////////////////////////////////////////////////
// 测试数据
//////////////////////////////////////////////////////////////////////
type TTestCsv struct {
	ID    int
	Des   string
	Item  []IntPair
	Card  []IntPair
	Array []int
}
type IntPair struct {
	ID  int
	Cnt int
}

var G_MapCsv = make(map[int]*TTestCsv)
var G_SliceCsv []TTestCsv = nil

//////////////////////////////////////////////////////////////////////
// 开服调用
//////////////////////////////////////////////////////////////////////
var G_CsvParserMap map[string]interface{}

func InitReflectParser() {
	G_CsvParserMap = map[string]interface{}{
		"test": &G_MapCsv,
		// "test": &G_SliceCsv,
	}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////
func GetCurrPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	path = string(path[0 : 1+strings.LastIndex(path, "\\")])
	return path
}
func LoadAllFiles() {
	pattern := GetCurrPath() + "csv/*.csv"
	names, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Printf("LoadAllFiles error : %s", err.Error())
	}
	for _, name := range names {
		LoadCsv(name)
	}
}

func LoadCsv(name string) {
	file, err := os.Open(name)
	defer file.Close()
	if err != nil {
		fmt.Printf("LoadCsv Open() error : %s", err.Error())
		return
	}

	fstate, err := file.Stat()
	if err != nil {
		fmt.Printf("LoadCsv Stat() error : %s", err.Error())
		return
	}
	if fstate.IsDir() == true {
		fmt.Printf("LoadCsv is dir : %s", name)
		return
	}

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		fmt.Printf("LoadCsv ReadAll() error : %s", err.Error())
		return
	}

	ptr, ok := G_CsvParserMap[strings.TrimSuffix(name, ".csv")]
	if !ok {
		fmt.Printf("Csv not regist in G_CsvParserMap: %s", name)
		return
	}
	ParseRefCsv(records, ptr)
}
func ParseRefCsv(records [][]string, ptr interface{}) {
	switch reflect.TypeOf(ptr).Elem().Kind() {
	case reflect.Map:
		ParseRefCsvByMap(records, ptr)
	case reflect.Slice:
		ParseRefCsvBySlice(records, ptr)
	default:
		fmt.Printf("Csv Type Error: TypeName:%s", reflect.TypeOf(ptr).Elem().String())

	}
}
func ParseRefCsvByMap(records [][]string, pMap interface{}) {
	table := reflect.ValueOf(pMap).Elem()
	typ := table.Type().Elem().Elem() // map内保存的指针，第二次Elem()得到所指对象类型

	total, idx := GetRecordsValidCnt(records), 0
	slice := reflect.MakeSlice(reflect.SliceOf(typ), total, total) // 避免多次new对象，直接new数组，拆开用

	bParsedName, nilFlag := false, int64(0)
	for _, v := range records {
		if strings.Index(v[0], "#") == -1 { // "#"起始的不读
			if !bParsedName {
				nilFlag = parseRefName(v)
				bParsedName = true
			} else {
				// data := reflect.New(typ).Elem()
				data := slice.Index(idx)
				idx++
				parseRefData(v, nilFlag, data)
				table.SetMapIndex(data.Field(0), data.Addr())
			}
		}
	}
}
func ParseRefCsvBySlice(records [][]string, pSlice interface{}) { // slice可减少对象数量，降低gc
	slice := reflect.ValueOf(pSlice).Elem() // 这里slice是nil
	typ := reflect.TypeOf(pSlice).Elem()

	// 表的数组，从1起始
	idx := 1
	total := GetRecordsValidCnt(records) + 1
	slice.Set(reflect.MakeSlice(typ, total, total))

	bParsedName, nilFlag := false, int64(0)
	for _, v := range records {
		if strings.Index(v[0], "#") == -1 { // "#"起始的不读
			if !bParsedName {
				nilFlag = parseRefName(v)
				bParsedName = true
			} else {
				data := slice.Index(idx)
				idx++
				parseRefData(v, nilFlag, data)
			}
		}
	}
}
func parseRefName(record []string) (ret int64) { // 不读的列：没命名/前缀"(c)"
	length := len(record)
	if length > 64 {
		fmt.Printf("csv column is over to 64 !!!")
	}
	for i := 0; i < length; i++ {
		if record[i] == "" || strings.Index(record[i], "(c)") == 0 {
			ret |= (1 << uint(i))
		}
	}
	return ret
}
func parseRefData(record []string, nilFlag int64, data reflect.Value) {
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

		switch field.Kind() {
		case reflect.Int:
			{
				field.SetInt(int64(CheckAtoiName(s, s)))
			}
		case reflect.String:
			{
				field.SetString(s)
			}
		case reflect.Slice:
			{
				switch field.Type().Elem().Kind() {
				case reflect.Int:
					{
						vec := ParseStringToArray(s)
						field.Set(reflect.ValueOf(vec))
					}
				case reflect.Struct:
					{
						vec := ParseStringToPair(s)
						field.Set(reflect.ValueOf(vec))
					}
				default:
					{
						fmt.Printf("Csv Type Error: TypeName:%s", data.Field(i).Type().String())
					}
				}
			}
		default:
			{
				fmt.Printf("Csv Type Error: TypeName:%s", data.Field(i).Type().String())
			}
		}
	}
}
func GetRecordsValidCnt(records [][]string) (ret int) {
	for _, v := range records {
		if strings.Index(v[0], "#") == -1 { // "#"起始的不读
			ret++
		}
	}
	return ret
}

func CheckAtoiName(s string, name string) int {
	if len(s) <= 0 {
		fmt.Printf("field: %s is empty", name)
		return 0
	}
	ret, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("field: %s text can't convert to int", name)
	}
	return ret
}

// 格式：(id1|num1)(id2|num2)
func ParseStringToPair(str string) []IntPair {
	sFix := strings.Trim(str, "()")
	slice := strings.Split(sFix, ")(")
	items := make([]IntPair, len(slice))
	for i, v := range slice {
		pv := strings.Split(v, "|")
		if len(pv) != 2 {
			fmt.Printf("ParseStringToPair : %s", str)
			return items
		}
		items[i].ID = CheckAtoiName(pv[0], pv[0])
		items[i].Cnt = CheckAtoiName(pv[1], pv[1])
	}
	return items
}

// 格式：32400|43200|64800|75600
func ParseStringToArray(str string) []int {
	slice := strings.Split(str, "|")
	nums := make([]int, len(slice))
	for i, v := range slice {
		nums[i] = CheckAtoiName(v, v)
	}
	return nums
}
