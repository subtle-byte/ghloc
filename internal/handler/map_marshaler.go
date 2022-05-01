package handler

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
)

type SortedMap struct {
	Map            interface{}
	TransformValue func(value interface{}) interface{}
	Less           func(value1 interface{}, value2 interface{}) bool
}

type pair struct {
	key, value interface{}
}

func (m SortedMap) MarshalJSON() ([]byte, error) {
	if m.TransformValue == nil {
		m.TransformValue = func(value interface{}) interface{} { return value }
	}

	mapValue := reflect.ValueOf(m.Map)
	array := make([]pair, mapValue.Len())
	mapRange := mapValue.MapRange()
	for i := 0; mapRange.Next(); i++ {
		array[i].key = mapRange.Key().Interface()
		array[i].value = m.TransformValue(mapRange.Value().Interface())
	}
	sort.Slice(array, func(i, j int) bool { return m.Less(array[i].value, array[j].value) })

	buf := bytes.Buffer{}
	buf.WriteByte('{')
	for i, pair := range array {
		if i != 0 {
			buf.WriteByte(',')
		}
		keyBytes, err := json.Marshal(pair.key)
		if err != nil {
			return nil, err
		}
		valueBytes, err := json.Marshal(pair.value)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')
		buf.Write(valueBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}
