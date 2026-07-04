package validator

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	once  sync.Once
	valid *validator.Validate
)

func init() {
	once.Do(func() {
		valid = validator.New()
		valid.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	})
}

// Validate 返回全局 validator 实例
func Validate() *validator.Validate {
	return valid
}

// Struct 校验结构体，返回第一条翻译后的错误消息，通过则返回空字符串
func Struct(s any) string {
	if err := valid.Struct(s); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return translate(e)
		}
	}
	return ""
}

// Var 校验单个变量，通过则返回空字符串
func Var(field any, tag string) string {
	if err := valid.Var(field, tag); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			return translate(e)
		}
	}
	return ""
}

func translate(e validator.FieldError) string {
	field := e.Field()
	switch e.Tag() {
	case "required":
		return field + "不能为空"
	case "min":
		return field + "长度不能小于" + e.Param()
	case "max":
		return field + "长度不能大于" + e.Param()
	case "len":
		return field + "长度必须为" + e.Param()
	case "email":
		return "邮箱格式不正确"
	case "url":
		return field + "格式不正确"
	case "oneof":
		return field + "取值必须为: " + e.Param()
	default:
		return field + "校验失败: " + e.Tag()
	}
}
