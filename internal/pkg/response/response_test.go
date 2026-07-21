package response

import (
	"net/http"
	"testing"
)

func TestHTTPStatus(t *testing.T) {
	tests := map[Code]int{
		CodeInvalidParam:    http.StatusBadRequest,
		CodeUnauthorized:    http.StatusUnauthorized,
		CodeForbidden:       http.StatusForbidden,
		CodeNotFound:        http.StatusNotFound,
		CodeUserExist:       http.StatusConflict,
		CodeServerError:     http.StatusInternalServerError,
		CodeArticleNotExist: http.StatusNotFound,
	}
	for code, want := range tests {
		if got := httpStatus(code); got != want {
			t.Errorf("httpStatus(%d)=%d, want %d", code, got, want)
		}
	}
}
