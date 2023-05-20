package paas

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"

	"unicode"

	"github.com/jinzhu/copier"
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

func GetDateTime(t time.Time) time.Time {
	return time.Now()
}

func FormatDateString(t time.Time) string {
	return t.Format("2006-01-02")
}

func FormatTimeString(t time.Time) string {
	return t.Format("15:04:05")
}

func FormatDateTimeString(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func FormatDateStringNumber(t time.Time) string {
	return t.Format("20060102")
}

func FormatTimeStringNumber(t time.Time) string {
	return t.Format("150405")
}

func FormatDateTimeStringNumber(t time.Time) string {
	return t.Format("20060102150405")
}

func ParseDate(ymd_ string) time.Time {
	t, _ := time.Parse("2006-01-02", ymd_)
	return t
}

func ParseTime(hms_ string) time.Time {
	t, _ := time.Parse("15:04:05", hms_)
	return t
}

func ParseDateTime(ymd_ string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", ymd_)
	return t
}

func ParseDateNumber(ymd string) time.Time {
	t, _ := time.Parse("20060102", ymd)
	return t
}

func ParseTimeNumber(hms string) time.Time {
	t, _ := time.Parse("150405", hms)
	return t
}

func ParseDateTimeNumber(ymd string) time.Time {
	t, _ := time.Parse("20060102150405", ymd)
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

func IsEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

func IsNotEmpty(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() != 0
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() != 0
	case reflect.Ptr, reflect.Interface:
		return !v.IsNil()
	default:
		return false
	}
}

func IsArray(value interface{}) bool {
	valueType := reflect.TypeOf(value)
	return valueType.Kind() == reflect.Array
}
