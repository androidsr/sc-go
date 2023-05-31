package sc

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"

	"unicode"

	"github.com/jinzhu/copier"
)

var (
	CstZone = time.FixedZone("CST", 8*3600)
)

// 判断切片中是否包含指定值
func Contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// 将字符串中大写字母前加“_”
func GetUnderscore(s string) string {
	var result string
	for i, c := range s {
		if i > 0 && unicode.IsUpper(c) {
			result += "_"
			result += strings.ToLower(string(c))
		} else {
			result += string(c)
		}
	}
	return strings.ToLower(result)
}

func GetDateTime() time.Time {
	return time.Now().In(CstZone)
}

func FormatDateString(t time.Time) string {
	return t.In(CstZone).Format("2006-01-02")
}

func FormatTimeString(t time.Time) string {
	return t.In(CstZone).Format("15:04:05")
}

func FormatDateTimeString(t time.Time) string {
	return t.In(CstZone).Format("2006-01-02 15:04:05")
}

func FormatDateStringNumber(t time.Time) string {
	return t.In(CstZone).Format("20060102")
}

func FormatTimeStringNumber(t time.Time) string {
	return t.In(CstZone).Format("150405")
}

func FormatDateTimeStringNumber(t time.Time) string {
	return t.In(CstZone).Format("20060102150405")
}

func ParseDate(ymd_ string) time.Time {
	t, _ := time.ParseInLocation("2006-01-02", ymd_, CstZone)
	return t
}

func ParseTime(hms_ string) time.Time {
	t, _ := time.ParseInLocation("15:04:05", hms_, CstZone)
	return t.In(CstZone)
}

func ParseDateTime(ymd_ string) time.Time {
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", ymd_, CstZone)
	return t.In(CstZone)
}

func ParseDateNumber(ymd string) time.Time {
	t, _ := time.ParseInLocation("20060102", ymd, CstZone)
	return t.In(CstZone)
}

func ParseTimeNumber(hms string) time.Time {
	t, _ := time.ParseInLocation("150405", hms, CstZone)
	return t.In(CstZone)
}

func ParseDateTimeNumber(ymd string) time.Time {
	t, _ := time.ParseInLocation("20060102150405", ymd, CstZone)
	return t
}

func Copy[T any](from interface{}) (T, error) {
	var to T
	copier.CopyWithOption(&to, &from, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	return to, nil
}

func CopySlice[T any](from interface{}) ([]T, error) {
	to := make([]T, 0)
	copier.CopyWithOption(&to, &from, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	return to, nil
}

func GetIP(prefix string) string {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	var ips []net.IP
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				fmt.Println(err)
				continue
			}
			for _, addr := range addrs {
				ip, _, err := net.ParseCIDR(addr.String())
				if err != nil {
					fmt.Println(err)
					continue
				}
				ips = append(ips, ip)
			}
		}
	}

	for _, ip := range ips {
		if ip.To4() != nil {
			ipStr := ip.To4().String()
			if prefix != "" && strings.HasPrefix(ipStr, prefix) {
				return ipStr
			} else {
				return ipStr
			}
		}
	}
	return ""
}

func IsSlice(value interface{}) bool {
	valueType := reflect.ValueOf(value)
	return valueType.Kind() == reflect.Slice
}

func SliceToInterface[T any](vs []T) []interface{} {
	result := make([]interface{}, 0)
	for _, v := range vs {
		result = append(result, v)
	}
	return result
}

func AssertSliceType(value interface{}) []interface{} {
	vStr, okStr := value.([]string)
	if okStr {
		return SliceToInterface(vStr)
	}
	vInt, okInt := value.([]int)
	if okInt {
		return SliceToInterface(vInt)
	}
	vInt8, okInt8 := value.([]int8)
	if okInt8 {
		return SliceToInterface(vInt8)
	}
	vInt16, okInt16 := value.([]int16)
	if okInt16 {
		return SliceToInterface(vInt16)
	}
	vInt32, okInt32 := value.([]int32)
	if okInt32 {
		return SliceToInterface(vInt32)
	}
	vInt64, okInt64 := value.([]int64)
	if okInt64 {
		return SliceToInterface(vInt64)
	}
	vuInt8, okuInt8 := value.([]uint8)
	if okuInt8 {
		return SliceToInterface(vuInt8)
	}
	vuInt16, okuInt16 := value.([]uint16)
	if okuInt16 {
		return SliceToInterface(vuInt16)
	}
	vuInt32, okuInt32 := value.([]uint32)
	if okuInt32 {
		return SliceToInterface(vuInt32)
	}
	vuInt64, okuInt64 := value.([]uint64)
	if okuInt64 {
		return SliceToInterface(vuInt64)
	}
	vBool, okBool := value.([]bool)
	if okBool {
		return SliceToInterface(vBool)
	}
	vF32, okF32 := value.([]float32)
	if okF32 {
		return SliceToInterface(vF32)
	}
	vF64, okF64 := value.([]float64)
	if okF64 {
		return SliceToInterface(vF64)
	}
	vComplex64, okComplex64 := value.([]complex64)
	if okComplex64 {
		return SliceToInterface(vComplex64)
	}
	vComplex128, okComplex128 := value.([]complex128)
	if okComplex128 {
		return SliceToInterface(vComplex128)
	}
	inter, interOk := value.([]interface{})
	if interOk {
		return SliceToInterface(inter)
	}
	return nil
}
