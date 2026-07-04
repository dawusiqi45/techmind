package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是统一的 JSON 响应结构
type Response struct {
	Code Code        `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// OK 返回成功响应，data 可为 nil
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  CodeSuccess.Msg(),
		Data: data,
	})
}

// Fail 返回业务失败响应，使用错误码默认消息
func Fail(c *gin.Context, code Code) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  code.Msg(),
	})
}

// FailWithMsg 返回业务失败响应，允许覆盖消息（用于参数校验等场景）
func FailWithMsg(c *gin.Context, code Code, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
	})
}

// AbortWithUnauthorized 中间件中止并返回 401
func AbortWithUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, Response{
		Code: CodeUnauthorized,
		Msg:  CodeUnauthorized.Msg(),
	})
}

// AbortWithRateLimited 中间件中止并返回 429
func AbortWithRateLimited(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusTooManyRequests, Response{
		Code: CodeRateLimited,
		Msg:  CodeRateLimited.Msg(),
	})
}
