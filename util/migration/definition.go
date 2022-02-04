package migration

import (
	"github.com/AndreeJait/GO-ANDREE-UTILITIES/logs"
	"github.com/AndreeJait/GO-ANDREE-UTILITIES/persistent"
	"github.com/AndreeJait/GO-ANDREE-UTILITIES/persistent/mongo"
)

const (
	TableName  = "migrations"
	ColumnName = "version"
	UpTag      = "[MIGRATION-UP] -"
	DownTag    = "[MIGRATION-DOWN] -"

	NoSqlUpTag      = "[MIGRATION-UP-NOSQL] -"
	NoSqlDownTag    = "[MIGRATION-DOWN-NOSQL] -"
)

type (
	Tool interface {
		Up() error
		Down() error
		Check() error
		Truncate() error
		Initialize() error
	}

	Script struct {
		Up, Down string
		UsingTransaction bool
	}

	NoSqlScript struct {
		Up, Down func(mongo.Mongo) error
	}

	sql struct {
		orm        persistent.ORM
		migrations map[int64]*Script
		logger     logs.Logger
	}

	nosql struct {
		orm        mongo.Mongo
		migrations map[int64]*NoSqlScript
		logger     logs.Logger
	}

	nosqlcollection struct {
		Version int64 `bson:"version"`
	}
)
