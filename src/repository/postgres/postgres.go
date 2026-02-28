package postgres

import (
	"InstantWellnessKits/src/config"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/cloudsqlconn/postgres/pgxv5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	cloudDriverName = "cloudsql-postgres"
	localDriverName = "pgx"
	migrationsPath  = "file://src/migrations"
	taxRatesCsvPath = "tax_rates.csv"
)

func InitDb(cfg *config.Config) (*sql.DB, error) {
	var dsn string
	var driverName string
	if cfg.Env == "PROD" {
		cleanup, err := pgxv5.RegisterDriver(cloudDriverName)
		if err != nil {
			return nil, err
		}
		_ = cleanup

		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.Database.ConnectionName, cfg.Database.User,
			cfg.Database.UserPassword, cfg.Database.Name)

		driverName = cloudDriverName
	} else {
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.Database.User, cfg.Database.UserPassword, cfg.Database.Host,
			cfg.Database.Port, cfg.Database.Name)

		driverName = localDriverName
	}

	conn, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)

	return conn, nil
}

func ApplyMigrations(conn *sql.DB) error {
	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath, "postgres", driver)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply database migrations: %w", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("No database migrations to apply.")
	} else {
		log.Println("Database migrations applied.")
	}

	return nil
}

func SeedTaxRates(db *sql.DB) error {
	file, err := os.Open(taxRatesCsvPath)
	if err != nil {
		return fmt.Errorf("failed to open csv file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	if _, err = reader.Read(); err != nil {
		return fmt.Errorf("failed to read csv headers: %w", err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read csv records: %w", err)
	}

	query := `
		INSERT INTO tax_rates (
			jurisdiction_type, jurisdiction_name, composite_rate, 
			state_rate, county_rate, city_rate, special_rate, special_name
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (jurisdiction_type, jurisdiction_name) DO NOTHING;
	`

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var insertedCount int
	for _, row := range records {
		var specialName *string
		if row[7] != "" {
			specialName = &row[7]
		}

		res, err := stmt.Exec(row[0], row[1], row[2],
			row[3], row[4], row[5], row[6], specialName)
		if err != nil {
			return fmt.Errorf("failed to insert row %v: %w", row, err)
		}

		rowsAffected, _ := res.RowsAffected()
		insertedCount += int(rowsAffected)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	if insertedCount > 0 {
		log.Printf("Seeder: Successfully inserted %d new tax rates!", insertedCount)
	} else {
		log.Println("Seeder: Tax rates are already up to date.")
	}

	return nil
}
