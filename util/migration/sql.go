package migration

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
	"github.com/AndreeJait/GO-ANDREE-UTILITIES/logs"
	"github.com/AndreeJait/GO-ANDREE-UTILITIES/persistent"
)

func NewSqlMigration(orm persistent.ORM, migrations map[int64]*Script, logger logs.Logger) (Tool, error) {
	if orm == nil {
		return nil, errors.New("orm is required!")
	}

	if logger == nil {
		return nil, errors.New("logger is required!")
	}

	return &sql{orm: orm, migrations: migrations, logger: logger}, nil
}

func (s *sql) Up() error {
	if err := isMigrationTableExists(s); err != nil {
		return err
	}

	if isMigrationScriptsEmpty(s) {
		return nil
	}

	last, err := getLatestMigrationVersionFromDatabase(s)

	if err != nil {
		return err
	}

	if last == getLatestMigrationVersion(s) {
		s.logger.Infof("%s migration already up to date!", UpTag)
		return nil
	}

	keys := make([]int64, 0)

	for k := range s.migrations {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(l, r int) bool {
		return keys[l] < keys[r]
	})

	versions := make([]int64, 0)

	// - first migration or migration table truncated
	if last == 0 {
		versions = keys
	} else {
		start := 0
		for i, key := range keys {
			if key == last {
				start = i + 1
				break
			}
		}
		versions = keys[start:]
	}

	for _, version := range versions {
		var (
			script = s.migrations[version]
		)

		// - check migration script
		if len(script.Up) == 0 {
			return errors.New(fmt.Sprintf("%s migration script %d can't be blank!", UpTag, version))
		}

		s.logger.Infof("%s executing migration version %d", UpTag, version)

		if script.UsingTransaction {
			err := s.WithTransaction(func(tx persistent.ORM) error {
				rows, err := tx.RawSql(script.Up)

				if err != nil {
					return errors.WithStack(err)
				}

				if err := rows.Close(); err != nil {
					s.logger.Errorf("failed to close rows", err)
					return errors.WithStack(err)
				}

				if err := tx.Exec("INSERT INTO "+TableName+" VALUES(?)", version); err != nil {
					return errors.WithStack(err)
				}

				return nil
			})

			if err != nil {
				s.logger.Errorf("%s failed to execute migration script %d: %s", UpTag, version, err)
				return errors.WithStack(err)
			}

			s.logger.Infof("%s migration with version %d migrated!", UpTag, version)
			continue
		}

		// - raw sql to exec multiple statements
		var (
			tx = s.orm
		)

		rows, err := tx.RawSql(script.Up)

		if err != nil {
			s.logger.Errorf("%s failed to execute migration script %d: %s", UpTag, version, err)
			return errors.WithStack(err)
		}

		if err := rows.Close(); err != nil {
			s.logger.Errorf("failed to close rows", err)
		}

		if err := tx.Exec("INSERT INTO "+TableName+" VALUES(?)", version); err != nil {
			s.logger.Errorf("%s failed to execute migration script %d: %s", UpTag, version, err)
			return errors.WithStack(err)
		}

		s.logger.Infof("%s migration with version %d migrated!", UpTag, version)
	}

	return nil
}

func (s *sql) Down() error {
	if err := isMigrationTableExists(s); err != nil {
		return err
	}

	if isMigrationScriptsEmpty(s) {
		return nil
	}

	version, err := getLatestMigrationVersionFromDatabase(s)

	if err != nil {
		return err
	}

	if version == 0 {
		s.logger.Infof("%s migrations table is empty, nothing to do", DownTag)
		return nil
	}

	keys := make([]int64, 0)

	for k := range s.migrations {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(l, r int) bool {
		return keys[l] < keys[r]
	})

	script := s.migrations[version]

	// - check migration script
	if len(script.Down) == 0 {
		return errors.New(fmt.Sprintf("%s script with version %d can't be empty", DownTag, version))
	}

	s.logger.Infof("%s begin down migration %d version", DownTag, version)

	if script.UsingTransaction {
		err := s.WithTransaction(func(tx persistent.ORM) error {
			rows, err := tx.RawSql(script.Down)

			if err != nil {
				s.logger.Errorf("%s failed to execute migration down script with version %d: %s", DownTag, version, err)
				return errors.WithStack(err)
			}

			if err := rows.Close(); err != nil {
				s.logger.Errorf("failed to close rows", err)
				return errors.WithStack(err)
			}

			if err := tx.Exec("DELETE FROM migrations WHERE version >= ?", version); err != nil {
				s.logger.Errorf("%s failed to execute delete migration script %d: %s", DownTag, version, err)
				return errors.WithStack(err)
			}

			return nil
		})

		if err != nil {
			s.logger.Errorf("%s failed to execute migration script %d: %s", DownTag, version, err)
			return nil
		}

		s.logger.Infof("%s migration version %d succeeded", DownTag, version)
		return nil
	}

	// - execute migrations script
	rows, err := s.orm.RawSql(script.Down)

	if err != nil {
		s.logger.Errorf("%s failed to execute migration down script with version %d", DownTag, version)
		return errors.WithStack(err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Errorf("failed to close rows", err)
		}
	}()

	// - remove version greater than before from migrations table
	if err := s.orm.Exec("DELETE FROM migrations WHERE version >= ?", version); err != nil {
		s.logger.Errorf("%s failed to execute delete migration script %+v", DownTag, version)
		return errors.WithStack(err)
	}

	s.logger.Infof("%s migration version %d succeeded", DownTag, version)

	return nil
}

func (s *sql) Check() error {
	if err := isMigrationTableExists(s); err != nil {
		return err
	}

	if isMigrationScriptsEmpty(s) {
		return nil
	}

	version := getLatestMigrationVersion(s)

	if err := isAlreadyMigrated(s, version); err != nil {
		return err
	}

	return nil
}

func (s *sql) Truncate() error {
	if err := isMigrationTableExists(s); err != nil {
		return err
	}

	if err := s.orm.Exec("TRUNCATE TABLE " + TableName); err != nil {
		return errors.New(fmt.Sprintf("failed to truncate %s", TableName))
	}

	return nil
}

func (s *sql) Initialize() error {
	if err := isMigrationTableExists(s); err == nil {
		s.logger.Infof("%s table already exists, nothing to do!", TableName)
		return nil
	}

	err := s.WithTransaction(func(tx persistent.ORM) error {
		query := `CREATE TABLE migrations(version bigint not null)`

		if err := tx.Exec(query); err != nil {
			s.logger.Errorf("failed to create %s table!", TableName)
			return errors.WithStack(err)
		}

		return nil
	})

	if err != nil {
		s.logger.Errorf("failed to initialize table")
		return errors.WithStack(err)
	}

	s.logger.Infof("%s table created", TableName)
	return nil
}

// - private

func isMigrationScriptsEmpty(s *sql) bool {
	length := len(s.migrations)

	if length == 0 {
		s.logger.Info("migration script is empty, nothing to migrate!")
		return true
	}

	return false
}

func isMigrationTableExists(s *sql) error {
	query := fmt.Sprintf("SELECT 1 FROM %s", TableName)

	rows, err := s.orm.RawSql(query)

	if err != nil {
		return errors.Wrapf(err, "migration table %s not found!", TableName)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Errorf("failed to close rows", err)
		}
	}()

	return nil
}

func isAlreadyMigrated(s *sql, version int64) error {
	rows, err := s.orm.RawSql("SELECT version FROM "+TableName+" ORDER BY ?", ColumnName)

	if err != nil {
		return errors.Wrapf(err, "failed to check migration with version %d", version)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Errorf("failed to close rows", err)
		}
	}()

	err = nil
	id, total := int64(0), 0
	found := false

	for rows.Next() {
		total += 1
		err = rows.Scan(&id)

		if err != nil {
			break
		}

		if id == version {
			found = true
			break
		}
	}

	// - migrations table is empty but migration file is not empty
	if total == 0 {
		return errors.New(fmt.Sprintf("migration %d is not migrated!", version))
	}

	// - failed row.Scan()
	if err != nil {
		return errors.Wrap(err, "failed to scan version column")
	}

	// - migrations table not empty but could not find migration with specific version
	if !found {
		return errors.New(fmt.Sprintf("migration %d is not migrated!", version))
	}

	return nil
}

func getLatestMigrationVersion(s *sql) int64 {
	keys := make([]int64, 0)

	for k := range s.migrations {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(l, r int) bool {
		return keys[l] < keys[r]
	})

	return keys[len(keys)-1]
}

func getLatestMigrationVersionFromDatabase(s *sql) (int64, error) {
	query := fmt.Sprintf("SELECT %s FROM migrations ORDER BY %s DESC", ColumnName, ColumnName)

	rows, err := s.orm.RawSql(query)

	if err != nil {
		return 0, errors.Wrap(err, "failed to get latest migration version from database!")
	}

	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Errorf("failed to close rows ", err)
		}
	}()

	version, err := int64(0), nil

	for rows.Next() {
		err = rows.Scan(&version)

		if err != nil {
			break
		}
		break
	}

	if err != nil {
		return 0, errors.Wrap(err, "failed to scan version column!")
	}

	return version, nil
}

func (s *sql) WithTransaction(block func(tx persistent.ORM) error) error {
	var (
		txn = s.orm.Begin()
	)

	if err := block(txn); err != nil {
		if err := txn.Rollback(); err != nil {
			return errors.Wrap(err, "failed to rollback")
		}
		return errors.WithStack(err)
	}

	if err := txn.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit")
	}

	return nil
}
