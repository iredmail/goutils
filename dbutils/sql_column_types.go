package dbutils

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
)

// NullString 用于字符串类型的 SQL 字段可能出现 null 的情况。
type NullString struct {
	value       string
	emptyToNull bool
}

func NewNullString(value string) NullString {
	return NullString{value: value}
}

func NewNullStringWithEmptyToNull(value string) NullString {
	return NullString{value: value, emptyToNull: true}
}

func (ns NullString) String() string {
	return ns.value
}

func (ns *NullString) Scan(value interface{}) error {
	if value == nil {
		ns.value = ""

		return nil
	}

	v, err := driver.String.ConvertValue(value)
	if err != nil {
		return err
	}

	switch s := v.(type) {
	case string:
		ns.value = s
	case []byte:
		ns.value = string(s)
	}

	return nil
}

// Value 在使用 struct 指定对应的 database 字段为 NullString 类型时，会调用此方法转换相应的值。
func (ns NullString) Value() (driver.Value, error) {
	if ns.emptyToNull && ns.value == "" {
		return nil, nil
	}

	return ns.value, nil
}

func (ns *NullString) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &ns.value)
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	return []byte(`"` + ns.value + `"`), nil
}

type NullFloat32 struct {
	value  float32
	isNull bool
}

func NewNullFloat32(v ...float32) NullFloat32 {
	nnf := NullFloat32{}
	if len(v) > 0 {
		nnf.value = v[0]
	} else {
		nnf.isNull = true
	}

	return nnf
}

func (nf NullFloat32) Float32() float32 {
	return nf.value
}

func (nf *NullFloat32) Scan(value interface{}) error {
	if value == nil {
		nf.isNull = true

		return nil
	}

	switch s := value.(type) {
	case float32:
		nf.value = s
	}

	return nil
}

func (nf NullFloat32) Value() (driver.Value, error) {
	if nf.isNull {
		return nil, nil
	}

	return nf.value, nil
}

func (nf *NullFloat32) UnmarshalJSON(data []byte) error {
	if bytes.Contains(data, []byte("null")) ||
		bytes.Contains(data, []byte("\"\"")) {
		nf.isNull = true

		return nil
	}

	return json.Unmarshal(data, &nf.value)
}

func (nf NullFloat32) MarshalJSON() ([]byte, error) {
	if nf.isNull {
		return json.Marshal(nil)
	}

	return json.Marshal(nf.value)
}

// IntBool 用于将 SQL 字段类型为整形的值转换为 bool。
// 值为 1 表示 true，0 表示 false。
type IntBool bool

func (ib IntBool) Bool() bool {
	return bool(ib)
}

func (ib *IntBool) Scan(value interface{}) error {
	v, err := driver.Bool.ConvertValue(value)
	if err != nil {
		return err
	}

	*ib = IntBool(v.(bool))

	return nil
}

func (ib IntBool) Value() (driver.Value, error) {
	switch ib {
	case true:
		return 1, nil
	default:
		return 0, nil
	}
}

// CharBool 用于将 SQL 字段类型为整形的值转换为 bool。
// 值为 'y' 表示 true，'n' 表示 false。
type CharBool bool

func (cb CharBool) Bool() bool {
	return bool(cb)
}

func (cb *CharBool) Scan(value interface{}) error {
	switch x := value.(type) {
	// mysql
	case []uint8:
		*cb = string(x) == "y"
	// pgsql
	case string:
		*cb = x == "y"
	}

	return nil
}

func (cb CharBool) Value() (driver.Value, error) {
	switch cb {
	case true:
		return "y", nil
	default:
		return "n", nil
	}
}
