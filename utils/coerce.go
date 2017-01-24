package utils

import (
	"fmt"
	"reflect"
	"strconv"
)

func CoerceFloat(value interface{}) (float32, bool) {
	switch result := value.(type) {
	case int:
		return float32(result), true
	case int32:
		return float32(result), true
	case int64:
		return float32(result), true
	case uint:
		return float32(result), true
	case uint8:
		return float32(result), true
	case uint16:
		return float32(result), true
	case uint32:
		return float32(result), true
	case uint64:
		return float32(result), true
	case float32:
		return result, true
	case float64:
		return float32(result), true
	case bool:
		if result == true {
			return 1.0, true
		}
		return 0.0, true
	case string:
		val, err := strconv.ParseFloat(result, 64)
		if err != nil {
			return 0.0, false
		}
		return float32(val), true
	default:
		v := reflect.ValueOf(value)
		kind := v.Kind()
		switch kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return float32(v.Int()), true
		case reflect.Float32, reflect.Float64:
			return float32(v.Float()), true
		}
		return 0.0, false
	}
}

func CoerceInt(value interface{}) (int32, bool) {
	switch result := value.(type) {
	case int:
		return int32(result), true
	case int32:
		return result, true
	case int64:
		return int32(result), true
	case uint:
		return int32(result), true
	case uint8:
		return int32(result), true
	case uint16:
		return int32(result), true
	case uint32:
		return int32(result), true
	case uint64:
		return int32(result), true
	case float32:
		return int32(result), true
	case float64:
		return int32(result), true
	case bool:
		if result == true {
			return 1, true
		}
		return 0, true
	case string:
		val, err := strconv.ParseInt(result, 10, 64)
		if err != nil {
			return 0, false
		}
		return int32(val), true
	default:
		v := reflect.ValueOf(value)
		kind := v.Kind()
		switch kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int32(v.Int()), true
		case reflect.Float32, reflect.Float64:
			return int32(v.Float()), true
		}
		return 0, false
	}
}

func CoerceBoolean(value interface{}) (bool, bool) {
	switch result := value.(type) {
	case int:
		if result != 0 {
			return true, true
		}
		return false, true
	case int32:
		if result != 0 {
			return true, true
		}
		return false, true
	case int64:
		if result != 0 {
			return true, true
		}
		return false, true
	case float32:
		if result != 0.0 {
			return true, true
		}
		return false, true
	case float64:
		if result != 0.0 {
			return true, true
		}
		return false, true
	case uint:
		if result != 0 {
			return true, true
		}
		return false, true
	case uint8:
		if result != 0 {
			return true, true
		}
		return false, true
	case uint16:
		if result != 0 {
			return true, true
		}
		return false, true
	case uint32:
		if result != 0 {
			return true, true
		}
		return false, true
	case uint64:
		if result != 0 {
			return true, true
		}
		return false, true
	case bool:
		return result, true
	case string:
		if result == "false" {
			return false, true
		} else if result == "true" {
			return true, true
		} else if result == "" {
			return false, true
		}
		return true, true
	default:
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Bool {
			return v.Bool(), true
		}
		return false, false
	}
}

func CoerceString(value interface{}) (string, bool) {
	switch result := value.(type) {
	case int:
		return strconv.FormatInt(int64(result), 10), true
	case int32:
		return strconv.FormatInt(int64(result), 10), true
	case int64:
		return strconv.FormatInt(result, 10), true
	case uint:
		return strconv.FormatInt(int64(result), 10), true
	case uint8:
		return strconv.FormatInt(int64(result), 10), true
	case uint16:
		return strconv.FormatInt(int64(result), 10), true
	case uint32:
		return strconv.FormatInt(int64(result), 10), true
	case uint64:
		return strconv.FormatInt(int64(result), 10), true
	case float32:
		return strconv.FormatFloat(float64(result), 'f', -1, 64), true
	case float64:
		return strconv.FormatFloat(result, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(result), true
	case string:
		return result, true
	case fmt.Stringer:
		return result.String(), true
	default:
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.String {
			return v.String(), true
		}
		return "", false
	}
}

func CoerceEnum(value interface{}) (string, bool) {
	switch result := value.(type) {
	case int:
		return strconv.FormatInt(int64(result), 32), true
	case int32:
		return strconv.FormatInt(int64(result), 32), true
	case int64:
		return strconv.FormatInt(result, 32), true
	case float32:
		return strconv.FormatFloat(float64(result), 'f', -1, 32), true
	case float64:
		return strconv.FormatFloat(result, 'f', -1, 32), true
	case bool:
		return strconv.FormatBool(result), true
	case string:
		return result, true
	default:
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.String {
			return v.String(), true
		}
		return "", false
	}
}
