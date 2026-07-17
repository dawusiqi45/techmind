package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestSREMigrationCreatesRequiredIncidentTables(t *testing.T) {
	data, err := os.ReadFile("001_sre_agent_audit_and_incident.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(data)
	for _, table := range []string{"incident", "incident_alert", "ops_tool_call"} {
		statement := "CREATE TABLE IF NOT EXISTS `" + table + "`"
		if !strings.Contains(sql, statement) {
			t.Errorf("migration is missing %s", statement)
		}
	}
}

func TestSREReliabilityMigrationAddsTaskKey(t *testing.T) {
	data, err := os.ReadFile("002_sre_agent_reliability.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(data)
	for _, fragment := range []string{"task_key", "uk_task_key", "legacy:"} {
		if !strings.Contains(sql, fragment) {
			t.Errorf("reliability migration is missing %q", fragment)
		}
	}
}

func TestSREActionGuidanceMigrationAddsStructuredFields(t *testing.T) {
	data, err := os.ReadFile("003_sre_action_guidance.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(data)
	for _, column := range []string{"verification_commands", "change_plan", "validation_commands", "rollback_commands"} {
		if !strings.Contains(sql, column) {
			t.Errorf("action guidance migration is missing %q", column)
		}
	}
}
