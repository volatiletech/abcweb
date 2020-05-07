package cmd

import "testing"

func TestPostgresConnStr(t *testing.T) {
	t.Parallel()

	var cfg migrateConfig

	cfg.Host = "localhost"
	cfg.Port = 3306
	cfg.DBName = "mydb"
	cfg.User = "bob"

	connStr := postgresConnStr(cfg)
	if connStr != `user=bob host=localhost port=3306 dbname=mydb` {
		t.Errorf("mismatch, got %s", connStr)
	}

	cfg.Pass = "pass"
	cfg.SSLMode = "true"

	connStr = postgresConnStr(cfg)
	if connStr != `user=bob password=pass host=localhost port=3306 dbname=mydb sslmode=true` {
		t.Errorf("mismatch, got %s", connStr)
	}
}
