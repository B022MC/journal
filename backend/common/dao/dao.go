package dao

import (
	"fmt"
	"sync"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// connMap stores named database connections.
// Each RPC service registers its connection during initialization.
var connMap = sync.Map{}

// Register stores a database connection with the given key.
// Typically called in ServiceContext initialization.
func Register(key string, conf sqlx.SqlConf) {
	connMap.Store(key, sqlx.MustNewConn(conf))
}

// GetConn retrieves a database connection by key.
// Panics if the key is not registered.
func GetConn(key string) sqlx.SqlConn {
	v, ok := connMap.Load(key)
	if !ok {
		panic(fmt.Sprintf("dao: unknown connection key %q, did you forget to call dao.Register?", key))
	}
	return v.(sqlx.SqlConn)
}

// MustGetConn is an alias for GetConn for readability.
func MustGetConn(key string) sqlx.SqlConn {
	return GetConn(key)
}
