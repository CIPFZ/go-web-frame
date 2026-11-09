package apispec

import (
	"fmt"
	"slices"
	"strings"
)

// SortClause 是结构化排序项。
type SortClause struct {
	Field string `json:"field" validate:"required"`
	Order string `json:"order" validate:"required,oneof=asc desc"`
}

// SortSpec 声明一个 List 方法允许的排序字段（响应字段名 → SQL 列）与默认排序。
type SortSpec struct {
	Columns    map[string]string
	Default    []SortClause
	TieBreaker string // 稳定次序键，默认 "id"
}

// OrderBy 校验排序项并生成 ORDER BY 子句；总是追加 TieBreaker 作为稳定的次序键。
func (s SortSpec) OrderBy(clauses []SortClause) (string, error) {
	if len(clauses) == 0 {
		clauses = s.Default
	}
	parts := make([]string, 0, len(clauses)+1)
	for i, cl := range clauses {
		col, ok := s.Columns[cl.Field]
		if !ok {
			allowed := make([]string, 0, len(s.Columns))
			for k := range s.Columns {
				allowed = append(allowed, k)
			}
			slices.Sort(allowed)
			return "", InvalidParameter("sort.field_invalid", "sort field is not supported", FieldViolation{
				Field: fmt.Sprintf("sort[%d].field", i), Reason: "enum.undefined", Issue: "field cannot be used for sorting",
				ExpectedFormat: "one of: " + strings.Join(allowed, ", "),
			})
		}
		if cl.Order != "asc" && cl.Order != "desc" {
			return "", InvalidParameter("sort.order_invalid", "sort order must be asc or desc", FieldViolation{
				Field: fmt.Sprintf("sort[%d].order", i), Reason: "enum.undefined", Issue: "order is required", ExpectedFormat: "one of: asc, desc",
			})
		}
		parts = append(parts, col+" "+strings.ToUpper(cl.Order))
	}
	tie := s.TieBreaker
	if tie == "" {
		tie = "id"
	}
	parts = append(parts, tie+" ASC")
	return strings.Join(parts, ", "), nil
}
