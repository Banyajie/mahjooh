package utils

import (
	"bytes"
	"encoding/gob"
	"errors"
	"reflect"
	"strconv"
	"strings"
)

func Encode(data interface{}) ([]byte, error) {
	buff := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buff)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func Decode(data []byte, to interface{}) error {
	buff := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buff)

	return dec.Decode(to)
}

func Contain(obj interface{}, target interface{}) (bool, error) {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true, nil
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true, nil
		}
	}

	return false, errors.New("not in array")
}

func ByteToString(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[0:i])
		}
	}
	return string(p)
}

/*
	将一维数组的表示转化为字符串形式
*/
func ReverToString(data []int) string {
	if len(data) == 0 {
		return ""
	}
	var str []string
	for i := 0; i < len(data); i++ {
		str = append(str, strconv.Itoa(data[i]))
	}

	return strings.Join(str, ",")
}
