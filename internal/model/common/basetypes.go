package common

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONMap 是对 map[string]interface{} 的封装，
// 用于数据库 JSON 字段与 Go map 的自动序列化/反序列化。
type JSONMap map[string]interface{}

// Value 实现 driver.Valuer 接口，用于数据库写入时自动序列化为 JSON。
func (m *JSONMap) Value() (driver.Value, error) {
	if m == nil || *m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan 实现 sql.Scanner 接口，用于数据库读取时自动反序列化为 JSONMap。
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		// 保证空值时不为 nil，避免后续访问时 panic
		*m = make(map[string]interface{})
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("basetypes.JSONMap.Scan: invalid type %T (must be []byte or string)", value)
	}

	// 确保 m 不为 nil
	if *m == nil {
		*m = make(map[string]interface{})
	}

	return json.Unmarshal(data, m)
}

type TreeNode[T any] interface {
	GetChildren() []T
	SetChildren(children T)
	GetID() int
	GetParentID() int
}
