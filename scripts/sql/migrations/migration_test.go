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
