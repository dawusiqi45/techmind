package controller

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestArticleRequestTagContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	createCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	createCtx.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"title":"t","content":"c","tags":["Go","SRE"]}`))
	createCtx.Request.Header.Set("Content-Type", "application/json")
	var createReq CreateArticleReq
	if err := createCtx.ShouldBindJSON(&createReq); err != nil {
		t.Fatalf("bind create request: %v", err)
	}
	assertTags(t, createReq.Tags)

	updateCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	updateCtx.Request = httptest.NewRequest("PUT", "/", strings.NewReader(`{"title":"t","content":"c","tags":["Go","SRE"]}`))
	updateCtx.Request.Header.Set("Content-Type", "application/json")
	var updateReq UpdateArticleReq
	if err := updateCtx.ShouldBindJSON(&updateReq); err != nil {
		t.Fatalf("bind update request: %v", err)
	}
	assertTags(t, updateReq.Tags)
}

func assertTags(t *testing.T, tags []string) {
	t.Helper()
	if len(tags) != 2 || tags[0] != "Go" || tags[1] != "SRE" {
		t.Fatalf("unexpected tags: %#v", tags)
	}
}
