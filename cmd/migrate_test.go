package cmd

import "testing"

func TestMysqlConnStr(t *testing.T) {
	t.Parallel()

	var cfg migrateConfig

	cfg.Host = "localhost"
	cfg.Port = 3306
	cfg.DBName = "mydb"
	cfg.User = "bob"

	connStr := mysqlConnStr(cfg)
	if connStr != `bob@tcp(localhost:3306)/mydb` {
		t.Errorf("mismatch, got %s", connStr)
	}

	cfg.Pass = "pass"
	cfg.SSLMode = "true"

	connStr = mysqlConnStr(cfg)
	if connStr != `bob:pass@tcp(localhost:3306)/mydb?tls=true` {
		t.Errorf("mismatch, got %s", connStr)
	}
}

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
