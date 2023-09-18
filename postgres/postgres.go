package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fio_service/config"
	"fio_service/structs"
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresqlDatabase struct {
	conn *pgxpool.Pool
}

var ctx context.Context = context.Background()

var DB PostgresqlDatabase

// Connect to database
func InitDB(url string, migrate_version uint) error {
	var err error

	DB.conn, err = pgxpool.Connect(context.Background(), url)
	if err != nil {
		return err
	}
	log.Println("Connected to database")
	// down if set in config
	if config.Conf.DOWN_OLD_DB_EVERYTIME {
		if err = dropDatabase(url); err != nil {
			return err
		}
	}
	// migrate
	err = migrateDB(url, migrate_version)
	if err != nil {
		return err
	}

	return DB.conn.Ping(ctx)
}

// Drop database
func dropDatabase(url string) error {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()
	log.Println("Database migration...")
	// Connect
	db, err := sql.Open("postgres", url)
	if err != nil {
		return err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	var m *migrate.Migrate = new(migrate.Migrate)
	m, err = migrate.NewWithDatabaseInstance(
		"file://postgres/migrations",
		"postgres", driver)
	if err != nil || m == nil {
		return err
	}
	defer m.Close()
	// Drop
	if err = m.Drop(); err.Error() != "no change" {
		return err
	}
	return nil
}

// Migrate database
func migrateDB(url string, version uint) error {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()
	log.Println("Database migration...")
	// Connect
	db, err := sql.Open("postgres", url)
	if err != nil {
		return err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	var m *migrate.Migrate = new(migrate.Migrate)
	m, err = migrate.NewWithDatabaseInstance(
		"file://postgres/migrations",
		"postgres", driver)
	if err != nil || m == nil {
		return err
	}
	defer m.Close()
	// Get version
	if version == 0 {
		// Up to latest version
		if err = m.Up(); err.Error() != "no change" {
			return err
		}
		return nil
	}
	// Migrate to fixed version
	if err = m.Migrate(version); err.Error() != "no change" {
		return err
	}
	return nil
}

// Log errors
func error_logging(err error) {
	if err != nil {
		pc := make([]uintptr, 10)
		n := runtime.Callers(2, pc)
		frames := runtime.CallersFrames(pc[:n])
		frame, _ := frames.Next()
		// fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
		log.Printf("[Postgres] error on %s: %s", frame.Function, err)
	}
}

// Add user
func (db *PostgresqlDatabase) AddUser(u structs.FIO) error {
	_, err := db.conn.Exec(ctx, "insert into fio (name, surname, patronymic, age, gender, nationality) values ($1, $2, $3, $4, $5, $6)",
		u.Name, u.Surname, u.Patronymic, u.Age, u.Gender, u.Nationality,
	)
	error_logging(err)
	return err
}

func avoidSQLInjection(a ...string) error {
	for _, s := range a {
		if strings.ContainsAny(s, "';\"") {
			return errors.New("contains forbidden symbols")
		}
	}
	return nil
}

func OpUsed(op string, sql string) string {
	if strings.Contains(sql, op) {
		return "and"
	}
	return op
}

// Get users
func (db *PostgresqlDatabase) GetUsers(offset, limit int, sd structs.SearchUserData) ([]structs.FIO, error) {
	// avoid sql injection
	err := avoidSQLInjection(sd.Name.String, sd.Patronymic.String, sd.Surname.String, sd.Gender.String, sd.Nationality.String)
	if err != nil {
		return nil, err
	}
	// create sql
	var sql string = "select id, name, surname, patronymic, age, gender, nationality from fio"
	// name filter
	if sd.Name.Valid {
		sql = fmt.Sprintf("%s %s name = '%s'", sql, OpUsed("where", sql), sd.Name.String)
	}
	// surname filter
	if sd.Surname.Valid {
		sql = fmt.Sprintf("%s %s surname = '%s'", sql, OpUsed("where", sql), sd.Surname.String)
	}
	// patronymic filter
	if sd.Patronymic.Valid {
		sql = fmt.Sprintf("%s %s patronymic = '%s'", sql, OpUsed("where", sql), sd.Patronymic.String)
	}
	// gender filter
	if sd.Gender.Valid {
		sql = fmt.Sprintf("%s %s gender = '%s'", sql, OpUsed("where", sql), sd.Gender.String)
	}
	// nationality filter
	if sd.Nationality.Valid {
		sql = fmt.Sprintf("%s %s nationality = '%s'", sql, OpUsed("where", sql), sd.Nationality.String)
	}
	// age filter
	if sd.Age_down.Valid {
		sql = fmt.Sprintf("%s %s age >= %d", sql, OpUsed("where", sql), sd.Age_down.Int64)
	}
	if sd.Age_up.Valid {
		sql = fmt.Sprintf("%s %s age <= %d", sql, OpUsed("where", sql), sd.Age_up.Int64)
	}
	// offset limit
	sql = fmt.Sprintf("%s offset %d limit %d", sql, offset, limit)
	// db query
	rows, err := db.conn.Query(ctx, sql)
	var fios []structs.FIO
	if err != nil {
		error_logging(err)
		return fios, err
	}
	defer rows.Close()
	for rows.Next() {
		var f structs.FIO
		rows.Scan(
			&f.ID,
			&f.Name,
			&f.Surname,
			&f.Patronymic,
			&f.Age,
			&f.Gender,
			&f.Nationality,
		)
		fios = append(fios, f)
	}
	return fios, err
}

// delete user
func (db *PostgresqlDatabase) DelUser(id int) error {
	_, err := db.conn.Exec(ctx, "delete from fio where id = $1", id)
	error_logging(err)
	return err
}

// update user
func (db *PostgresqlDatabase) EditUser(ef structs.EditFIO) error {
	_, err := db.conn.Exec(ctx, `
		update fio set
		name = COALESCE($2, name),
		surname = COALESCE($3, surname),
		patronymic = COALESCE($4, patronymic),
		age = COALESCE($5, age),
		gender = COALESCE($6, gender),
		nationality = COALESCE($7, nationality)
		where id = $1
	`, ef.ID, ef.Name, ef.Surname, ef.Patronymic, ef.Age, ef.Gender, ef.Nationality)
	error_logging(err)
	return err
}
