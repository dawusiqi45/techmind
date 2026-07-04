package response

// Code 是业务错误码类型
type Code int

const (
	CodeSuccess      Code = 1000
	CodeInvalidParam Code = 1001
	CodeServerError  Code = 1002
	CodeUnauthorized Code = 1003
	CodeForbidden    Code = 1004
	CodeNotFound     Code = 1005
	CodeRateLimited  Code = 1006

	// 用户相关
	CodeUserExist    Code = 2001
	CodeUserNotExist Code = 2002
	CodeWrongPassword Code = 2003

	// 文章相关
	CodeArticleNotExist Code = 3001
)

var codeMsg = map[Code]string{
	CodeSuccess:         "success",
	CodeInvalidParam:    "invalid params",
	CodeServerError:     "internal server error",
	CodeUnauthorized:    "unauthorized",
	CodeForbidden:       "forbidden",
	CodeNotFound:        "not found",
	CodeUserExist:       "user already exists",
	CodeUserNotExist:    "user not found",
	CodeWrongPassword:   "wrong password",
	CodeArticleNotExist: "article not found",
	CodeRateLimited:     "rate limit exceeded",
}

// Msg 返回错误码对应的默认消息
func (c Code) Msg() string {
	if msg, ok := codeMsg[c]; ok {
		return msg
	}
	return "unknown error"
}
