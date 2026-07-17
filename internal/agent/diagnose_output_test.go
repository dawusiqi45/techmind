package agent

import "testing"

func TestParseRawDiagnoseBuildsSafeStructuredGuidance(t *testing.T) {
	raw := "```json\n" + `{
  "summary":"API 延迟升高",
  "impact":"部分请求超时",
  "root_cause":"高置信度：副本不足，依据 Deployment 与延迟指标",
  "suggestions":["先确认当前副本和资源使用率"],
  "verification_commands":[
    {"purpose":"检查副本","command":"kubectl -n techmind get deployment techmind-server","expected":"AVAILABLE 等于 DESIRED","risk":"high","approval_required":true},
    {"purpose":"读取敏感信息","command":"kubectl -n techmind get secret llm-key","expected":"获得密钥","risk":"low","approval_required":false}
  ],
  "change_plan":[
    {"target":"Deployment/techmind-server","instruction":"经容量确认后将副本调整为 3","command_or_patch":"kubectl -n techmind scale deployment techmind-server --replicas=3","risk":"medium","preconditions":["确认节点容量"],"validation":"可用副本达到 3","rollback":"恢复原副本数","approval_required":false}
  ],
  "validation_commands":[
    {"purpose":"确认发布状态","command":"kubectl -n techmind rollout status deployment/techmind-server","expected":"successfully rolled out","risk":"low","approval_required":false}
  ],
  "rollback_commands":[
    {"purpose":"回滚版本","command":"kubectl -n techmind rollout undo deployment/techmind-server","expected":"恢复上一版本","risk":"medium","approval_required":false},
    {"purpose":"删除异常 Pod","command":"kubectl delete pod bad-pod","expected":"Pod 被删除","risk":"high","approval_required":true}
  ]
}` + "\n```"

	result := parseRawDiagnose(raw, "")
	if result.Summary != "API 延迟升高" {
		t.Fatalf("unexpected summary: %q", result.Summary)
	}
	if len(result.VerificationCommands) != 1 {
		t.Fatalf("expected only the safe read-only command, got %#v", result.VerificationCommands)
	}
	if result.VerificationCommands[0].ApprovalRequired || result.VerificationCommands[0].Risk != "low" {
		t.Fatalf("read-only command policy was not enforced: %#v", result.VerificationCommands[0])
	}
	if len(result.ChangePlan) != 1 || !result.ChangePlan[0].ApprovalRequired {
		t.Fatalf("change plan must require approval: %#v", result.ChangePlan)
	}
	if len(result.RollbackCommands) != 1 || !result.RollbackCommands[0].ApprovalRequired {
		t.Fatalf("unsafe rollback should be filtered and safe rollback should require approval: %#v", result.RollbackCommands)
	}
}

func TestParseRawDiagnoseLegacySuggestionsDoNotIncludeEvidence(t *testing.T) {
	raw := "摘要：测试告警\n影响：无\n根因：证据不足\n证据：\n- CPU 升高\n建议：\n- 查询副本状态\n- 检查最近发布"
	result := parseRawDiagnose(raw, "")
	if len(result.Suggestions) != 2 || result.Suggestions[0] != "查询副本状态" {
		t.Fatalf("legacy suggestions were parsed incorrectly: %#v", result.Suggestions)
	}
}

func TestSafeSingleCommandRejectsCompositionAndSecrets(t *testing.T) {
	for _, command := range []string{
		"kubectl get pods | Select-String Error",
		"kubectl get pods; kubectl delete pod api",
		"kubectl get secret api-key",
		"curl -H 'Authorization: Bearer credential' https://example.invalid",
	} {
		if safeSingleCommand(command) {
			t.Errorf("unsafe command was accepted: %q", command)
		}
	}
}

func TestReadOnlyCommandRejectsHiddenMutation(t *testing.T) {
	for _, command := range []string{
		"kubectl set image deployment/api api=v2 --record=get",
		"helm upgrade techmind ./chart --description get",
		"redis-cli SET diagnostic-command xinfo",
		"docker compose restart ps",
		"curl -XPOST https://example.invalid/health",
	} {
		if readOnlyCommand(command) {
			t.Errorf("mutating command was classified as read-only: %q", command)
		}
	}
}
