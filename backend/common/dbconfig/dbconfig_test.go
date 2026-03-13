package dbconfig

import "testing"

func TestToSqlConfHonorsReadWriteSplit(t *testing.T) {
	conf := Config{
		ReadWriteSplit: false,
		DataSource:     "root:pwd@tcp(localhost:3306)/journal",
		Replicas:       []string{"root:pwd@tcp(localhost:3307)/journal"},
	}

	sqlConf := conf.ToSqlConf()
	if len(sqlConf.Replicas) != 0 {
		t.Fatalf("expected replicas to be ignored when read-write split is disabled, got %d", len(sqlConf.Replicas))
	}

	conf.ReadWriteSplit = true
	sqlConf = conf.ToSqlConf()
	if len(sqlConf.Replicas) != 1 {
		t.Fatalf("expected replicas to be kept when read-write split is enabled, got %d", len(sqlConf.Replicas))
	}
}

func TestValidateRequiresReplicasWhenSplitEnabled(t *testing.T) {
	conf := Config{
		ReadWriteSplit: true,
		DataSource:     "root:pwd@tcp(localhost:3306)/journal",
	}

	if err := conf.Validate("DB"); err == nil {
		t.Fatal("expected validation error when replicas are missing")
	}
}
