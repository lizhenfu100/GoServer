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
				ID     int
				Des    string
				Item   IntPair
				Card   []IntPair
				ArrInt []int
				ArrStr []string
			}
			var G_MapCsv map[int]*TTestCsv = nil  	// map结构读表，首列作Key
			var G_SliceCsv []TTestCsv = nil 		// 数组结构读表，注册【&G_SliceCsv】到G_Csv_Map

			var G_Csv_Map = map[string]interface{}{
				"test": &G_MapCsv,
				// "test": &G_SliceCsv,
			}
}
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
	"strings"
)

type IntPair struct {
	ID  int
	Cnt int
}
type StrError struct {
	Str string
	Err error
}

func (self *StrError) Error() string {
	if self.Err == nil {
		return self.Str
	} else {
		return self.Str + " " + self.Err.Error()
	}
}

func GetExePath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	path = string(path[0 : 1+strings.LastIndex(path, "\\")])
	return path
}
func IsDirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}
	return true
}

//////////////////////////////////////////////////////////////////////
// 载入策划配表
var G_Csv_Map map[string]interface{} = nil

func LoadAllCsv() {
	pattern := GetExePath() + "csv/*.csv"
	names, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Printf("LoadAllCsv error : %s", err.Error())
	}
	for _, name := range names {
		_LoadOneCsv(name)
	}
}
func ReloadCsv(csvName string) {
	name := fmt.Sprintf("%scsv/%s.csv", GetExePath(), csvName)
	_LoadOneCsv(name)
}
func _LoadOneCsv(name string) {
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

	if ptr, ok := G_Csv_Map[strings.TrimSuffix(fstate.Name(), ".csv")]; ok {
		ParseRefCsv(records, ptr)
	} else {
		fmt.Printf("Csv not regist in G_Csv_Map: %s", name)
	}
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
				table.SetMapIndex(data.Field(0), data.Addr())
			}
		}
	}
}
func ParseRefCsvBySlice(records [][]string, pSlice interface{}) { // slice可减少对象数量，降低gc
	slice := reflect.ValueOf(pSlice).Elem() // 这里slice是nil
	typ := reflect.TypeOf(pSlice).Elem()

	// 表的数组，从1起始
	total, idx := _GetRecordsValidCnt(records), 1
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
func _parseHead(record []string) (ret int64) { // 不读的列：没命名/前缀"(c)"
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

		switch field.Kind() {
		case reflect.Int:
			{
				field.SetInt(int64(CheckAtoiName(s)))
			}
		case reflect.String:
			{
				field.SetString(s)
			}
		case reflect.Struct:
			{
				vec := ParseStringToPair(s)
				field.Set(reflect.ValueOf(vec[0]))
			}
		case reflect.Slice:
			{
				switch field.Type().Elem().Kind() {
				case reflect.Int:
					{
						vec := ParseStringToArrInt(s)
						field.Set(reflect.ValueOf(vec))
					}
				case reflect.String:
					{
						vec := strings.Split(s, "|")
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
func _GetRecordsValidCnt(records [][]string) (ret int) {
	for _, v := range records {
		if strings.Index(v[0], "#") == -1 { // "#"起始的不读
			ret++
		}
	}
	return ret
}

//////////////////////////////////////////////////////////////////////
//
func LoadCsv(path string) ([][]string, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	fstate, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if fstate.IsDir() {
		return nil, &StrError{"LoadCsv is dir!", nil}
	}

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}
func UpdateCsv(path string, records [][]string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	defer file.Close()
	if err != nil {
		return err
	}

	fstate, err := file.Stat()
	if err != nil {
		return err
	}
	if fstate.IsDir() {
		return &StrError{"UpdateCsv is dir!", nil}
	}

	csvWriter := csv.NewWriter(file)
	csvWriter.UseCRLF = true
	return csvWriter.WriteAll(records)
}
func AppendCsv(path string, record []string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND, os.ModePerm)
	defer file.Close()
	if err != nil {
		return err
	}

	fstate, err := file.Stat()
	if err != nil {
		return err
	}
	if fstate.IsDir() {
		return &StrError{"AppendCsv is dir!", nil}
	}

	csvWriter := csv.NewWriter(file)
	csvWriter.UseCRLF = true
	if err := csvWriter.Write(record); err != nil {
		return err
	}
	csvWriter.Flush()
	return nil
}
