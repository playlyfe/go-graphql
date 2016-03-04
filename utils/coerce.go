package utils

import (
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
	case float32:
		return result, true
	case float64:
		return float32(result), true
	case bool:
		if result == true {
			return 1.0, true
		} else {
			return 0.0, true
		}
	case string:
		val, err := strconv.ParseFloat(result, 64)
		if err != nil {
			return 0.0, false
		}
		return float32(val), true
	default:
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
	case float32:
		return int32(result), true
	case float64:
		return int32(result), true
	case bool:
		if result == true {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		val, err := strconv.ParseInt(result, 10, 64)
		if err != nil {
			return 0, false
		}
		return int32(val), true
	default:
		return 0, false
	}
}

func CoerceBoolean(value interface{}) (bool, bool) {
	switch result := value.(type) {
	case int:
		if result != 0 {
			return true, true
		} else {
			return false, true
		}
	case int32:
		if result != 0 {
			return true, true
		} else {
			return false, true
		}
	case int64:
		if result != 0 {
			return true, true
		} else {
			return false, true
		}
	case float32:
		if result != 0.0 {
			return true, true
		} else {
			return false, true
		}
	case float64:
		if result != 0.0 {
			return true, true
		} else {
			return false, true
		}
	case bool:
		return result, true
	case string:
		if result == "false" {
			return false, true
		} else if result == "true" {
			return true, true
		} else if result == "" {
			return false, true
		} else {
			return true, true
		}
	default:
		return false, false
	}
}

func CoerceString(value interface{}) (string, bool) {
	switch result := value.(type) {
	case int:
		println(result, "--------", strconv.FormatInt(int64(result), 10))
		return strconv.FormatInt(int64(result), 10), true
	case int32:
		return strconv.FormatInt(int64(result), 10), true
	case int64:
		return strconv.FormatInt(result, 10), true
	case float32:
		return strconv.FormatFloat(float64(result), 'f', -1, 64), true
	case float64:
		return strconv.FormatFloat(result, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(result), true
	case string:
		return result, true
	default:
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
		return "", false
	}
}
