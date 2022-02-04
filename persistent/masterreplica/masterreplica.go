package masterreplica

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/AndreeJait/GO-ANDREE-UTILITIES/persistent"
)

// DB is logical database object with Master as master physical database
// and Replica as slave database with loadbalancer
type DB struct {
	// Master is master physical database
	Master persistent.ORM

	// Replica can be a slave physical database, but for more than 1 slave Replica
	// you can put loadbalancer on top of your Replica sets that handle the load
	// distribution, which can be round robin or others
	Replica persistent.ORM
}

// New create a new DB.
func New(master, replica persistent.ORM) persistent.ORM {
	return &DB{Master: master, Replica: replica}
}

// Ping send a ping to both master and replica to make sure all database
// connections are alive.
func (db *DB) Ping() error {
	if err := db.Master.Ping(); err != nil {
		return errors.Wrap(err, "master")
	}
	err := db.Replica.Ping()
	return errors.Wrap(err, "replica")
}

// Close close current db connection in master and replica.
func (db *DB) Close() error {
	if err := db.Master.Close(); err != nil {
		return errors.Wrap(err, "master")
	}
	err := db.Replica.Close()
	return errors.Wrap(err, "replica")
}

// Set set setting by name, which could be used in callbacks, will clone
// and return a new master db.
func (db *DB) Set(name string, value interface{}) persistent.ORM {
	return db.Master.Set(name, value)
}

// Error check error in both master and replica db.
func (db *DB) Error() error {
	if err := db.Master.Error(); err != nil {
		return errors.Wrap(err, "master")
	}

	err := db.Replica.Error()
	return errors.Wrap(err, "replica")
}

// Where return a new relation, filter records with given conditions, accepts
// `map`, `struct` or `string` as conditions. Clone both master and replica db.
func (db *DB) Where(query interface{}, args ...interface{}) persistent.ORM {
	return &DB{
		Master:  db.Master.Where(query, args...),
		Replica: db.Replica.Where(query, args...),
	}
}

// First find first record that match given conditions, order by primary key.
// Executed using replica db.
func (db *DB) First(object interface{}) error {
	return errors.Wrap(db.Replica.First(object), "replica")
}

// All find all records that match given conditions, order by primary key.
// Executed using replica db.
func (db *DB) All(object interface{}) error {
	return errors.Wrap(db.Replica.All(object), "replica")
}

// Order specify order when retrieve records from database.
//     db.Order("name DESC")
// Return cloned persistent.ORM from replica db.
func (db *DB) Order(args interface{}) persistent.ORM {
	return db.Replica.Order(args)
}

// Limit specify the number of records to be retrieved.
// Return cloned persistent.ORM from replica db.
func (db *DB) Limit(args interface{}) persistent.ORM {
	return db.Replica.Limit(args)
}

// Offset specify the number of records to skip before starting to return
// the records. Return cloned persistent.ORM from replica db.
func (db *DB) Offset(args interface{}) persistent.ORM {
	return db.Replica.Offset(args)
}

// Create insert the value into database.
func (db *DB) Create(object interface{}) error {
	return errors.Wrap(db.Master.Create(object), "master")
}

// Update update the object in database.
func (db *DB) Update(object interface{}) error {
	return errors.Wrap(db.Master.Update(object), "master")
}

// Delete delete object in database.
func (db *DB) Delete(object interface{}) error {
	return errors.Wrap(db.Master.Delete(object), "master")
}

// BulkDelete bulk delete data from given table.
func (db *DB) BulkDelete(tableName string, data []interface{}) error {
	return errors.Wrap(db.Master.BulkDelete(tableName, data), "master")
}

// SoftDelete soft deleting data.
func (db *DB) SoftDelete(object interface{}) error {
	return errors.Wrap(db.Master.SoftDelete(object), "master")
}

// Exec execute given query.
func (db *DB) Exec(sql string, args ...interface{}) error {
	return errors.Wrap(db.Master.Exec(sql, args...), "master")
}

// RawSqlWithObject execute given raw query.
func (db *DB) RawSqlWithObject(sql string, object interface{}, args ...interface{}) error {
	return errors.Wrap(db.Master.RawSqlWithObject(sql, object, args...), "master")
}

// RawSql execute given raw query.
func (db *DB) RawSql(sql string, args ...interface{}) (*sql.Rows, error) {
	rows, err := db.Master.RawSql(sql, args...)

	return rows, errors.Wrap(err, "master")
}

// BulkUpsert bulk upsert data per chunkSize.
func (db *DB) BulkUpsert(tableName string, chunkSize int, data []interface{}) error {
	return errors.Wrap(db.Master.BulkUpsert(tableName, chunkSize, data), "master")
}

// Search find data with spesific field to select and criteria.
func (db *DB) Search(tableName string, selectField []string, criteria []persistent.Criteria, results interface{}) error {
	return errors.Wrap(db.Replica.Search(tableName, selectField, criteria, results), "replica")
}

// HasTable return true if given table's name is exist in db.
func (db *DB) HasTable(tableName string) bool {
	return db.Replica.HasTable(tableName)
}

// CreateTable create a new table.
func (db *DB) CreateTable(data interface{}) error {
	return errors.Wrap(db.Master.CreateTable(data), "master")
}

// CreateTableWithName create a new table with given name.
func (db *DB) CreateTableWithName(tableName string, data interface{}) error {
	return errors.Wrap(db.Master.CreateTableWithName(tableName, data), "master")
}

// DropTable drop table if exist.
func (db *DB) DropTable(data interface{}) error {
	return errors.Wrap(db.Master.DropTable(data), "master")
}

// DropTableWithName drop table with spesific name if exist.
func (db *DB) DropTableWithName(tableName string, data interface{}) error {
	return errors.Wrap(db.Master.DropTableWithName(tableName, data), "master")
}

// Table specify the table you would like to run db operations. Return both
// cloned master and replica db.
func (db *DB) Table(tableName string) persistent.ORM {
	return &DB{
		Master:  db.Master.Table(tableName),
		Replica: db.Replica.Table(tableName),
	}
}

// Begin begins a transaction in master db.
func (db *DB) Begin() persistent.ORM {
	return db.Master.Begin()
}

// Commit commit a transaction in master db.
func (db *DB) Commit() error {
	return errors.Wrap(db.Master.Commit(), "master")
}

// Rollback rollback a transaction in master db.
func (db *DB) Rollback() error {
	return errors.Wrap(db.Master.Rollback(), "master")
}

// UnderlyingDB return master db connection.
func (db *DB) UnderlyingDB() *sql.DB {
	return db.Master.UnderlyingDB()
}
