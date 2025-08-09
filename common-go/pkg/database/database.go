// Package database provides database interaction helpers for the ERP microservices system.
// It includes transaction management, null handling, connection pooling, and common query patterns.
package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/erpmicroservices/common-go/pkg/errors"
	"github.com/erpmicroservices/common-go/pkg/logging"
	"github.com/erpmicroservices/common-go/pkg/uuid"
)

// Config represents database connection configuration.
type Config struct {
	Host                  string        `json:"host"`
	Port                  int           `json:"port"`
	Database              string        `json:"database"`
	Username              string        `json:"username"`
	Password              string        `json:"password"`
	SSLMode               string        `json:"sslMode"`
	MaxConnections        int           `json:"maxConnections"`
	MaxIdleConnections    int           `json:"maxIdleConnections"`
	ConnectionMaxLifetime time.Duration `json:"connectionMaxLifetime"`
	ConnectionMaxIdleTime time.Duration `json:"connectionMaxIdleTime"`
	ConnectTimeout        time.Duration `json:"connectTimeout"`
	QueryTimeout          time.Duration `json:"queryTimeout"`
	EnableLogging         bool          `json:"enableLogging"`
	SlowQueryThreshold    time.Duration `json:"slowQueryThreshold"`
}

// DefaultConfig returns a default database configuration.
func DefaultConfig() *Config {
	return &Config{
		Host:                  "localhost",
		Port:                  5432,
		Database:              "erp_db",
		Username:              "erp_user",
		Password:              "erp_password",
		SSLMode:               "require",
		MaxConnections:        25,
		MaxIdleConnections:    5,
		ConnectionMaxLifetime: 5 * time.Minute,
		ConnectionMaxIdleTime: 2 * time.Minute,
		ConnectTimeout:        10 * time.Second,
		QueryTimeout:          30 * time.Second,
		EnableLogging:         true,
		SlowQueryThreshold:    1 * time.Second,
	}
}

// DSN returns the data source name for the database connection.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s connect_timeout=%d",
		c.Host, c.Port, c.Database, c.Username, c.Password, c.SSLMode,
		int(c.ConnectTimeout.Seconds()),
	)
}

// Connection wraps a sql.DB with additional functionality.
type Connection struct {
	db     *sql.DB
	config *Config
	logger *logging.Logger
}

// NewConnection creates a new database connection with the given configuration.
func NewConnection(config *Config) (*Connection, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Note: In a real implementation, you'd import a specific driver like "github.com/lib/pq"
	db, err := sql.Open("postgres", config.DSN())
	if err != nil {
		return nil, errors.DatabaseError("open", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetConnMaxLifetime(config.ConnectionMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnectionMaxIdleTime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, errors.DatabaseError("ping", err)
	}

	conn := &Connection{
		db:     db,
		config: config,
		logger: logging.NewLogger("database"),
	}

	return conn, nil
}

// Close closes the database connection.
func (c *Connection) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// DB returns the underlying sql.DB instance.
func (c *Connection) DB() *sql.DB {
	return c.db
}

// Ping tests the database connection.
func (c *Connection) Ping(ctx context.Context) error {
	if err := c.db.PingContext(ctx); err != nil {
		return errors.DatabaseError("ping", err)
	}
	return nil
}

// Stats returns database connection statistics.
func (c *Connection) Stats() sql.DBStats {
	return c.db.Stats()
}

// Transaction management

// TxFunc represents a function that operates within a transaction.
type TxFunc func(*sql.Tx) error

// WithTransaction executes a function within a database transaction.
func (c *Connection) WithTransaction(ctx context.Context, fn TxFunc) error {
	return c.WithTransactionOptions(ctx, nil, fn)
}

// WithTransactionOptions executes a function within a database transaction with options.
func (c *Connection) WithTransactionOptions(ctx context.Context, opts *sql.TxOptions, fn TxFunc) error {
	tx, err := c.db.BeginTx(ctx, opts)
	if err != nil {
		return errors.DatabaseError("begin transaction", err)
	}

	defer func() {
		if p := recover(); p != nil {
			c.rollback(tx)
			panic(p) // Re-throw panic after Rollback
		} else if err != nil {
			c.rollback(tx) // err is non-nil; don't change it
		} else {
			err = c.commit(tx) // err is nil; if Commit returns error update err
		}
	}()

	err = fn(tx)
	return err
}

func (c *Connection) commit(tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
		return errors.DatabaseError("commit transaction", err)
	}
	return nil
}

func (c *Connection) rollback(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil {
		c.logger.Error().Err(err).Msg("Failed to rollback transaction")
	}
}

// Query execution helpers

// QueryRow executes a query that returns a single row with logging and timeout.
func (c *Connection) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	ctx, cancel := c.withQueryTimeout(ctx)
	defer cancel()

	start := time.Now()
	row := c.db.QueryRowContext(ctx, query, args...)
	c.logQuery(query, args, time.Since(start), nil)

	return row
}

// Query executes a query that returns multiple rows with logging and timeout.
func (c *Connection) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, cancel := c.withQueryTimeout(ctx)
	defer cancel()

	start := time.Now()
	rows, err := c.db.QueryContext(ctx, query, args...)
	c.logQuery(query, args, time.Since(start), err)

	if err != nil {
		return nil, errors.DatabaseError("query", err)
	}

	return rows, nil
}

// Exec executes a query that doesn't return rows with logging and timeout.
func (c *Connection) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := c.withQueryTimeout(ctx)
	defer cancel()

	start := time.Now()
	result, err := c.db.ExecContext(ctx, query, args...)
	c.logQuery(query, args, time.Since(start), err)

	if err != nil {
		return nil, errors.DatabaseError("exec", err)
	}

	return result, nil
}

// Transaction query helpers

// TxQueryRow executes a query within a transaction that returns a single row.
func (c *Connection) TxQueryRow(tx *sql.Tx, ctx context.Context, query string, args ...interface{}) *sql.Row {
	ctx, cancel := c.withQueryTimeout(ctx)
	defer cancel()

	start := time.Now()
	row := tx.QueryRowContext(ctx, query, args...)
	c.logQuery(query, args, time.Since(start), nil)

	return row
}

// TxQuery executes a query within a transaction that returns multiple rows.
func (c *Connection) TxQuery(tx *sql.Tx, ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, cancel := c.withQueryTimeout(ctx)
	defer cancel()

	start := time.Now()
	rows, err := tx.QueryContext(ctx, query, args...)
	c.logQuery(query, args, time.Since(start), err)

	if err != nil {
		return nil, errors.DatabaseError("tx query", err)
	}

	return rows, nil
}

// TxExec executes a query within a transaction that doesn't return rows.
func (c *Connection) TxExec(tx *sql.Tx, ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := c.withQueryTimeout(ctx)
	defer cancel()

	start := time.Now()
	result, err := tx.ExecContext(ctx, query, args...)
	c.logQuery(query, args, time.Since(start), err)

	if err != nil {
		return nil, errors.DatabaseError("tx exec", err)
	}

	return result, nil
}

func (c *Connection) withQueryTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.config.QueryTimeout > 0 {
		return context.WithTimeout(ctx, c.config.QueryTimeout)
	}
	return context.WithCancel(ctx)
}

func (c *Connection) logQuery(query string, args []interface{}, duration time.Duration, err error) {
	if !c.config.EnableLogging {
		return
	}

	event := c.logger.Debug()
	if err != nil {
		event = c.logger.Error().Err(err)
	} else if duration > c.config.SlowQueryThreshold {
		event = c.logger.Warn()
	}

	event.
		Str("query", query).
		Interface("args", args).
		Dur("duration", duration).
		Msg("Database query executed")
}

// Null types for handling nullable database fields

// NullString wraps sql.NullString with JSON marshaling.
type NullString struct {
	sql.NullString
}

// NewNullString creates a NullString from a string pointer.
func NewNullString(s *string) NullString {
	if s == nil {
		return NullString{sql.NullString{Valid: false}}
	}
	return NullString{sql.NullString{String: *s, Valid: true}}
}

// StringPtr returns a string pointer or nil if null.
func (ns NullString) StringPtr() *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// MarshalJSON implements json.Marshaler.
func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return []byte(`"` + ns.String + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (ns *NullString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ns.Valid = false
		return nil
	}

	ns.Valid = true
	// Remove quotes from JSON string
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		ns.String = string(data[1 : len(data)-1])
	} else {
		ns.String = string(data)
	}
	return nil
}

// NullInt64 wraps sql.NullInt64 with JSON marshaling.
type NullInt64 struct {
	sql.NullInt64
}

// NewNullInt64 creates a NullInt64 from an int64 pointer.
func NewNullInt64(i *int64) NullInt64 {
	if i == nil {
		return NullInt64{sql.NullInt64{Valid: false}}
	}
	return NullInt64{sql.NullInt64{Int64: *i, Valid: true}}
}

// Int64Ptr returns an int64 pointer or nil if null.
func (ni NullInt64) Int64Ptr() *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}

// MarshalJSON implements json.Marshaler.
func (ni NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("%d", ni.Int64)), nil
}

// NullFloat64 wraps sql.NullFloat64 with JSON marshaling.
type NullFloat64 struct {
	sql.NullFloat64
}

// NewNullFloat64 creates a NullFloat64 from a float64 pointer.
func NewNullFloat64(f *float64) NullFloat64 {
	if f == nil {
		return NullFloat64{sql.NullFloat64{Valid: false}}
	}
	return NullFloat64{sql.NullFloat64{Float64: *f, Valid: true}}
}

// Float64Ptr returns a float64 pointer or nil if null.
func (nf NullFloat64) Float64Ptr() *float64 {
	if !nf.Valid {
		return nil
	}
	return &nf.Float64
}

// MarshalJSON implements json.Marshaler.
func (nf NullFloat64) MarshalJSON() ([]byte, error) {
	if !nf.Valid {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("%g", nf.Float64)), nil
}

// NullBool wraps sql.NullBool with JSON marshaling.
type NullBool struct {
	sql.NullBool
}

// NewNullBool creates a NullBool from a bool pointer.
func NewNullBool(b *bool) NullBool {
	if b == nil {
		return NullBool{sql.NullBool{Valid: false}}
	}
	return NullBool{sql.NullBool{Bool: *b, Valid: true}}
}

// BoolPtr returns a bool pointer or nil if null.
func (nb NullBool) BoolPtr() *bool {
	if !nb.Valid {
		return nil
	}
	return &nb.Bool
}

// MarshalJSON implements json.Marshaler.
func (nb NullBool) MarshalJSON() ([]byte, error) {
	if !nb.Valid {
		return []byte("null"), nil
	}
	if nb.Bool {
		return []byte("true"), nil
	}
	return []byte("false"), nil
}

// NullTime wraps sql.NullTime with JSON marshaling.
type NullTime struct {
	sql.NullTime
}

// NewNullTime creates a NullTime from a time.Time pointer.
func NewNullTime(t *time.Time) NullTime {
	if t == nil {
		return NullTime{sql.NullTime{Valid: false}}
	}
	return NullTime{sql.NullTime{Time: *t, Valid: true}}
}

// TimePtr returns a time.Time pointer or nil if null.
func (nt NullTime) TimePtr() *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

// MarshalJSON implements json.Marshaler.
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return []byte(`"` + nt.Time.Format(time.RFC3339) + `"`), nil
}

// NullUUID represents a nullable UUID.
type NullUUID struct {
	UUID  uuid.UUID
	Valid bool
}

// NewNullUUID creates a NullUUID from a UUID pointer.
func NewNullUUID(u *uuid.UUID) NullUUID {
	if u == nil || u.IsNil() {
		return NullUUID{Valid: false}
	}
	return NullUUID{UUID: *u, Valid: true}
}

// UUIDPtr returns a UUID pointer or nil if null.
func (nu NullUUID) UUIDPtr() *uuid.UUID {
	if !nu.Valid {
		return nil
	}
	return &nu.UUID
}

// Scan implements sql.Scanner interface.
func (nu *NullUUID) Scan(value interface{}) error {
	if value == nil {
		nu.UUID, nu.Valid = uuid.UUID{}, false
		return nil
	}

	nu.Valid = true
	return nu.UUID.Scan(value)
}

// Value implements driver.Valuer interface.
func (nu NullUUID) Value() (driver.Value, error) {
	if !nu.Valid {
		return nil, nil
	}
	return nu.UUID.Value()
}

// MarshalJSON implements json.Marshaler.
func (nu NullUUID) MarshalJSON() ([]byte, error) {
	if !nu.Valid {
		return []byte("null"), nil
	}
	return nu.UUID.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (nu *NullUUID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nu.Valid = false
		return nil
	}

	nu.Valid = true
	return nu.UUID.UnmarshalJSON(data)
}

// Repository pattern helpers

// BaseRepository provides common repository functionality.
type BaseRepository struct {
	conn   *Connection
	table  string
	logger *logging.Logger
}

// NewBaseRepository creates a new base repository.
func NewBaseRepository(conn *Connection, table string) *BaseRepository {
	return &BaseRepository{
		conn:   conn,
		table:  table,
		logger: logging.NewLogger("repository"),
	}
}

// Connection returns the database connection.
func (r *BaseRepository) Connection() *Connection {
	return r.conn
}

// Table returns the table name.
func (r *BaseRepository) Table() string {
	return r.table
}

// ExistsById checks if a record exists by ID.
func (r *BaseRepository) ExistsById(ctx context.Context, id uuid.UUID) (bool, error) {
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1)", r.table)

	var exists bool
	err := r.conn.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, errors.DatabaseError("exists check", err)
	}

	return exists, nil
}

// CountAll returns the total count of records in the table.
func (r *BaseRepository) CountAll(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.table)

	var count int64
	err := r.conn.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, errors.DatabaseError("count", err)
	}

	return count, nil
}

// Health check utilities

// HealthChecker provides database health checking functionality.
type HealthChecker struct {
	conn *Connection
}

// NewHealthChecker creates a new database health checker.
func NewHealthChecker(conn *Connection) *HealthChecker {
	return &HealthChecker{conn: conn}
}

// HealthCheck performs a comprehensive database health check.
func (hc *HealthChecker) HealthCheck(ctx context.Context) map[string]interface{} {
	health := map[string]interface{}{
		"status": "healthy",
		"checks": map[string]interface{}{},
	}

	// Connection test
	if err := hc.conn.Ping(ctx); err != nil {
		health["status"] = "unhealthy"
		health["checks"].(map[string]interface{})["connection"] = map[string]interface{}{
			"status": "fail",
			"error":  err.Error(),
		}
	} else {
		health["checks"].(map[string]interface{})["connection"] = map[string]interface{}{
			"status": "pass",
		}
	}

	// Connection pool stats
	stats := hc.conn.Stats()
	health["checks"].(map[string]interface{})["pool"] = map[string]interface{}{
		"open_connections":    stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}

	return health
}
