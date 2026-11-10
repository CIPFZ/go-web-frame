package apispec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// MaxRequestBytes 是契约接口的请求体上限。
const MaxRequestBytes = 1 << 20

var (
	validateOnce sync.Once
	validate     *validator.Validate
)

func validatorInstance() *validator.Validate {
	validateOnce.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())
		validate.RegisterTagNameFunc(func(f reflect.StructField) string {
			name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			if name == "-" {
				return ""
			}
			if name == "" {
				if f.Anonymous {
					return embeddedField + f.Name // JSON 中展开的嵌入结构体：错误路径里去掉这一段
				}
				return f.Name
			}
			return name
		})
		_ = validate.RegisterValidation("sid", func(fl validator.FieldLevel) bool {
			return IsSID(fl.Param(), fl.Field().String())
		})
	})
	return validate
}

// embeddedField 标记嵌入结构体的字段名（JSON 中展开，不出现在字段路径里）。
const embeddedField = "\x00"

// RegisterValidation 注册业务模块的声明式校验标签；须在处理请求前（启动时）调用。
func RegisterValidation(tag string, fn func(value string) bool) {
	if err := validatorInstance().RegisterValidation(tag, func(fl validator.FieldLevel) bool {
		return fn(fl.Field().String())
	}); err != nil {
		panic(err)
	}
}

// Decode 严格解码请求体并执行声明式校验：
// 解码失败 → BadRequest；未知字段、类型不符、校验失败 → InvalidParameter / MissingParameter。
// 空请求体按 {} 处理。
func Decode(c *gin.Context, dst any) error {
	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, MaxRequestBytes))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return NewError(CodePayloadTooLarge, "request.too_large", fmt.Sprintf("request body exceeds %d bytes", MaxRequestBytes))
		}
		return NewError(CodeBadRequest, "request.unreadable", "request body could not be read")
	}
	if len(bytes.TrimSpace(body)) == 0 {
		body = []byte("{}")
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return decodeError(err)
	}
	if dec.More() {
		return NewError(CodeBadRequest, "request.malformed_json", "request body must contain a single JSON object")
	}
	return Validate(dst)
}

// Validate 对已解码的结构体执行 validate 标签校验。
func Validate(v any) error {
	err := validatorInstance().Struct(v)
	if err == nil {
		return nil
	}
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return Internal(err)
	}
	violations := make([]FieldViolation, 0, len(verrs))
	allRequired := true
	for _, fe := range verrs {
		fv := violationFor(fe)
		if fv.Reason != "field.required" {
			allRequired = false
		}
		violations = append(violations, fv)
	}
	if allRequired {
		e := NewError(CodeMissingParameter, "request.validation_failed", "required parameters are missing")
		e.FieldViolations = violations
		return e
	}
	return InvalidParameter("request.validation_failed", "request parameters are invalid", violations...)
}

func violationFor(fe validator.FieldError) FieldViolation {
	parts := strings.Split(fe.Namespace(), ".")[1:] // 去掉顶层结构体名
	parts = slices.DeleteFunc(parts, func(p string) bool { return strings.HasPrefix(p, embeddedField) })
	field := strings.Join(parts, ".")
	fv := FieldViolation{Field: field}
	switch fe.Tag() {
	case "required", "required_if", "required_unless", "required_with", "required_without":
		fv.Reason, fv.Issue = "field.required", "field is required"
	case "oneof":
		fv.Reason, fv.Issue = "enum.undefined", "value is not an allowed enum value"
		fv.ExpectedFormat = "one of: " + strings.ReplaceAll(fe.Param(), " ", ", ")
	case "min", "max", "gte", "lte", "gt", "lt", "len":
		if fe.Kind() == reflect.String || fe.Kind() == reflect.Slice || fe.Kind() == reflect.Map {
			fv.Reason, fv.Issue = "length.out_of_range", "length is out of range"
		} else {
			fv.Reason, fv.Issue = "value.out_of_range", "value is out of range"
		}
		fv.ExpectedFormat = fe.Tag() + " " + fe.Param()
	case "sid":
		fv.Reason, fv.Issue = "format.invalid", "value is not a valid identifier"
		fv.ExpectedFormat = fe.Param() + "_<ulid>"
	case "excluded_with", "excluded_unless", "excluded_if":
		fv.Reason, fv.Issue = "oneof.conflict", "field conflicts with another field"
	default:
		fv.Reason, fv.Issue = "format.invalid", "value has an invalid format"
		if fe.Param() != "" {
			fv.ExpectedFormat = fe.Tag() + " " + fe.Param()
		} else {
			fv.ExpectedFormat = fe.Tag()
		}
	}
	return fv
}

func decodeError(err error) error {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &tooLarge):
		return NewError(CodePayloadTooLarge, "request.too_large", "request body is too large")
	case errors.As(err, &syntaxErr), errors.Is(err, io.ErrUnexpectedEOF):
		return NewError(CodeBadRequest, "request.malformed_json", "request body is not valid JSON")
	case errors.As(err, &typeErr):
		return InvalidParameter("request.validation_failed", "request parameters are invalid", FieldViolation{
			Field: typeErr.Field, Reason: "type.mismatch", Issue: "value has the wrong type", ExpectedFormat: typeErr.Type.String(),
		})
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		name := strings.Trim(strings.TrimPrefix(err.Error(), "json: unknown field "), `"`)
		return InvalidParameter("request.unknown_field", "request contains an undefined field", FieldViolation{
			Field: name, Reason: "field.unknown", Issue: "field is not defined by the schema",
		})
	default:
		var ae *Error
		if errors.As(err, &ae) {
			return ae // 自定义 UnmarshalJSON（如 oneof、枚举）返回的契约错误
		}
		return NewError(CodeBadRequest, "request.malformed_json", "request body could not be decoded")
	}
}
