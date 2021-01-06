/*
 * Copyright (C) @2021 Webank Group Holding Limited
 * <p>
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * <p>
 * http://www.apache.org/licenses/LICENSE-2.0
 * <p>
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 */

package base

import (
	"encoding/json"
	"fmt"
)

func ParseBoolean(val interface{}) (bool, error) {
	switch val {
	case nil:
		return false, nil
	case "1":
		return true, nil
	case 1:
		return true, nil
	case "0":
		return false, nil
	case 0:
		return false, nil
	default:
		return false, fmt.Errorf("unable to casting %v (type %T)", val, val)
	}
}

func ParseFloat64(val interface{}) (float64, error) {
	switch val.(type) {
	case nil:
		return 0, nil
	case json.Number:
		return val.(json.Number).Float64()
	default:
		return 0, fmt.Errorf("unable to casting number %v (type %T)", val, val)
	}
}

func ParseFloat32(val interface{}) (float32, error) {
	number, err := ParseFloat64(val)
	if err != nil {
		return float32(0), err
	}
	return float32(number), nil
}

func ParseInt(val interface{}) (int, error) {
	number, err := ParseFloat64(val)
	if err != nil {
		return 0, err
	}
	return int(number), nil
}

func ParseInt8(val interface{}) (int8, error) {
	number, err := ParseFloat64(val)
	if err != nil {
		return 0, err
	}
	return int8(number), err
}

func ParseInt16(val interface{}) (int16, error) {
	number, err := ParseFloat64(val)
	if err != nil {
		return 0, err
	}
	return int16(number), err
}

func ParseInt32(val interface{}) (int32, error) {
	number, err := ParseFloat64(val)
	if err != nil {
		return 0, err
	}
	return int32(number), nil
}

func ParseInt64(val interface{}) (int64, error) {
	number, err := ParseFloat64(val)
	if err != nil {
		return 0, err
	}
	return int64(number), nil
}

func ParseString(value interface{}) (string, error) {
	switch value.(type) {
	case string:
		return value.(string), nil
	default:
		return "", fmt.Errorf("unable to casting number %v (type %T)", value, value)
	}
}

func ParseUint(val interface{}) (uint, error) {
	number, err := ParseFloat64(val)
	if err != nil {
		return uint(0), err
	}
	return uint(number), nil
}

func ParseUint8(val interface{}) (uint8, error) {
	number, err := ParseFloat64(val)
	if err != nil {
		return uint8(0), err
	}
	return uint8(number), nil
}

func ParseUint16(val interface{}) (uint16, error) {
	number, err := ParseFloat64(val)
	if err != nil {
		return uint16(0), err
	}
	return uint16(number), nil
}

func ParseUint32(val interface{}) (uint32, error) {
	number, err := ParseFloat64(val)
	if err != nil {
		return uint32(0), err
	}
	return uint32(number), nil
}

func ParseUint64(val interface{}) (uint64, error) {
	number, err := ParseFloat64(val)
	if err != nil {
		return uint64(0), err
	}
	return uint64(number), nil
}

func ParseUintptr(val interface{}) (uintptr, error) {
	number, err := ParseFloat64(val)
	if err != nil {
		return uintptr(0), err
	}
	return uintptr(number), nil
}

func ParseStringArray(val interface{}) ([]string, error) {
	switch val.(type) {
	case []interface{}:
		obj := val.([]interface{})
		res := make([]string, 0)
		for _, k := range obj {
			res = append(res, k.(string))
		}
		return res, nil
	}
	return nil, fmt.Errorf("invalid array format")
}

func ParseIntArray(val interface{}) ([]int, error) {
	switch val.(type) {
	case []int:
		if op, ok := val.([]int); ok {
			return op, nil
		}
	}
	return nil, fmt.Errorf("invalid array format")
}
