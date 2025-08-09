// Package logging provides structured logging utilities for the ERP microservices system
// using zerolog with ERP-specific context and formatting.
package logging

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/erpmicroservices/common-go/pkg/uuid"
	"github.com/rs/zerolog"
)

// LogLevel represents the logging level.
type LogLevel string

const (
	TraceLevel LogLevel = "TRACE"
	DebugLevel LogLevel = "DEBUG"
	InfoLevel  LogLevel = "INFO"
	WarnLevel  LogLevel = "WARN"
	ErrorLevel LogLevel = "ERROR"
	FatalLevel LogLevel = "FATAL"
	PanicLevel LogLevel = "PANIC"
)

// LogFormat represents the logging output format.
type LogFormat string

const (
	JSONFormat    LogFormat = "JSON"
	ConsoleFormat LogFormat = "CONSOLE"
)

// Context keys for logging context
type contextKey string

const (
	CorrelationIDKey contextKey = "correlation_id"
	RequestIDKey     contextKey = "request_id"
	UserIDKey        contextKey = "user_id"
	TraceIDKey       contextKey = "trace_id"
	SpanIDKey        contextKey = "span_id"
)

// Config represents logging configuration.
type Config struct {
	Level          LogLevel  `json:"level"`
	Format         LogFormat `json:"format"`
	ServiceName    string    `json:"serviceName"`
	ServiceVersion string    `json:"serviceVersion"`
	Environment    string    `json:"environment"`
	Output         io.Writer `json:"-"`
	TimeFormat     string    `json:"timeFormat"`
	CallerEnabled  bool      `json:"callerEnabled"`
	StackTrace     bool      `json:"stackTrace"`
}

// DefaultConfig returns a default logging configuration.
func DefaultConfig() *Config {
	return &Config{
		Level:          InfoLevel,
		Format:         JSONFormat,
		ServiceName:    "erp-service",
		ServiceVersion: "unknown",
		Environment:    "development",
		Output:         os.Stdout,
		TimeFormat:     time.RFC3339,
		CallerEnabled:  true,
		StackTrace:     false,
	}
}

// Logger wraps zerolog.Logger with ERP-specific functionality.
type Logger struct {
	logger zerolog.Logger
	config *Config
}

// NewLogger creates a new logger with the given service name.
func NewLogger(serviceName string) *Logger {
	config := DefaultConfig()
	config.ServiceName = serviceName
	return NewLoggerWithConfig(config)
}

// NewLoggerWithConfig creates a new logger with the given configuration.
func NewLoggerWithConfig(config *Config) *Logger {
	if config == nil {
		config = DefaultConfig()
	}

	// Configure zerolog
	if config.Format == ConsoleFormat {
		config.Output = zerolog.ConsoleWriter{
			Out:        config.Output,
			TimeFormat: config.TimeFormat,
			NoColor:    false,
		}
	}

	zerolog.TimeFieldFormat = config.TimeFormat
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return fmt.Sprintf("%s:%d", file, line)
	}

	logger := zerolog.New(config.Output).
		Level(convertLogLevel(config.Level)).
		With().
		Timestamp().
		Str("service", config.ServiceName).
		Str("version", config.ServiceVersion).
		Str("environment", config.Environment)

	if config.CallerEnabled {
		logger = logger.Caller()
	}

	return &Logger{
		logger: logger.Logger(),
		config: config,
	}
}

// convertLogLevel converts our LogLevel to zerolog.Level.
func convertLogLevel(level LogLevel) zerolog.Level {
	switch level {
	case TraceLevel:
		return zerolog.TraceLevel
	case DebugLevel:
		return zerolog.DebugLevel
	case InfoLevel:
		return zerolog.InfoLevel
	case WarnLevel:
		return zerolog.WarnLevel
	case ErrorLevel:
		return zerolog.ErrorLevel
	case FatalLevel:
		return zerolog.FatalLevel
	case PanicLevel:
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// With returns a logger with additional context fields.
func (l *Logger) With() zerolog.Context {
	return l.logger.With()
}

// Trace returns a trace event.
func (l *Logger) Trace() *zerolog.Event {
	return l.logger.Trace()
}

// Debug returns a debug event.
func (l *Logger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

// Info returns an info event.
func (l *Logger) Info() *zerolog.Event {
	return l.logger.Info()
}

// Warn returns a warning event.
func (l *Logger) Warn() *zerolog.Event {
	return l.logger.Warn()
}

// Error returns an error event.
func (l *Logger) Error() *zerolog.Event {
	return l.logger.Error()
}

// Fatal returns a fatal event.
func (l *Logger) Fatal() *zerolog.Event {
	return l.logger.Fatal()
}

// Panic returns a panic event.
func (l *Logger) Panic() *zerolog.Event {
	return l.logger.Panic()
}

// WithCorrelationID adds correlation ID to the logger context.
func (l *Logger) WithCorrelationID(id string) zerolog.Logger {
	return l.logger.With().Str("correlation_id", id).Logger()
}

// WithRequestID adds request ID to the logger context.
func (l *Logger) WithRequestID(id string) zerolog.Logger {
	return l.logger.With().Str("request_id", id).Logger()
}

// WithUserID adds user ID to the logger context.
func (l *Logger) WithUserID(userID uuid.UUID) zerolog.Logger {
	return l.logger.With().Str("user_id", userID.String()).Logger()
}

// WithTraceID adds trace ID to the logger context.
func (l *Logger) WithTraceID(traceID string) zerolog.Logger {
	return l.logger.With().Str("trace_id", traceID).Logger()
}

// WithError adds error information to the logger context.
func (l *Logger) WithError(err error) zerolog.Logger {
	return l.logger.With().AnErr("error", err).Logger()
}

// WithFields adds multiple fields to the logger context.
func (l *Logger) WithFields(fields map[string]interface{}) zerolog.Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return ctx.Logger()
}

// ContextLogger extracts context fields and returns a logger with that context.
func (l *Logger) ContextLogger(ctx context.Context) zerolog.Logger {
	logger := l.logger

	if correlationID := getStringFromContext(ctx, CorrelationIDKey); correlationID != "" {
		logger = logger.With().Str("correlation_id", correlationID).Logger()
	}

	if requestID := getStringFromContext(ctx, RequestIDKey); requestID != "" {
		logger = logger.With().Str("request_id", requestID).Logger()
	}

	if userID := getUUIDFromContext(ctx, UserIDKey); !userID.IsNil() {
		logger = logger.With().Str("user_id", userID.String()).Logger()
	}

	if traceID := getStringFromContext(ctx, TraceIDKey); traceID != "" {
		logger = logger.With().Str("trace_id", traceID).Logger()
	}

	if spanID := getStringFromContext(ctx, SpanIDKey); spanID != "" {
		logger = logger.With().Str("span_id", spanID).Logger()
	}

	return logger
}

// SetLevel sets the logging level.
func (l *Logger) SetLevel(level LogLevel) {
	l.logger = l.logger.Level(convertLogLevel(level))
	l.config.Level = level
}

// GetLevel returns the current logging level.
func (l *Logger) GetLevel() LogLevel {
	return l.config.Level
}

// Close closes the logger if it has resources to clean up.
func (l *Logger) Close() error {
	// zerolog doesn't require explicit closing, but we provide this
	// for compatibility with other logging frameworks
	return nil
}

// Helper functions for context

// WithCorrelationID adds correlation ID to context.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// WithRequestID adds request ID to context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithUserIDContext adds user ID to context.
func WithUserIDContext(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithTraceID adds trace ID to context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithSpanID adds span ID to context.
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, SpanIDKey, spanID)
}

// GetCorrelationIDFromContext extracts correlation ID from context.
func GetCorrelationIDFromContext(ctx context.Context) string {
	return getStringFromContext(ctx, CorrelationIDKey)
}

// GetRequestIDFromContext extracts request ID from context.
func GetRequestIDFromContext(ctx context.Context) string {
	return getStringFromContext(ctx, RequestIDKey)
}

// GetUserIDFromContext extracts user ID from context.
func GetUserIDFromContext(ctx context.Context) uuid.UUID {
	return getUUIDFromContext(ctx, UserIDKey)
}

// GetTraceIDFromContext extracts trace ID from context.
func GetTraceIDFromContext(ctx context.Context) string {
	return getStringFromContext(ctx, TraceIDKey)
}

// GetSpanIDFromContext extracts span ID from context.
func GetSpanIDFromContext(ctx context.Context) string {
	return getStringFromContext(ctx, SpanIDKey)
}

// Helper functions
func getStringFromContext(ctx context.Context, key contextKey) string {
	if val := ctx.Value(key); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getUUIDFromContext(ctx context.Context, key contextKey) uuid.UUID {
	if val := ctx.Value(key); val != nil {
		if id, ok := val.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.UUID{}
}

// Performance logging utilities

// Timer provides timing functionality for performance logging.
type Timer struct {
	start  time.Time
	logger *Logger
	event  *zerolog.Event
}

// StartTimer starts a new timer for performance monitoring.
func (l *Logger) StartTimer(operation string) *Timer {
	return &Timer{
		start:  time.Now(),
		logger: l,
		event:  l.Debug().Str("operation", operation),
	}
}

// End completes the timer and logs the duration.
func (t *Timer) End(msg string) {
	duration := time.Since(t.start)
	t.event.Dur("duration", duration).Msg(msg)
}

// EndWithError completes the timer and logs the duration with an error.
func (t *Timer) EndWithError(err error, msg string) {
	duration := time.Since(t.start)
	if err != nil {
		t.logger.Error().
			Err(err).
			Dur("duration", duration).
			Msg(msg)
	} else {
		t.event.Dur("duration", duration).Msg(msg)
	}
}

// Middleware logging utilities

// RequestLogFields represents common fields for request logging.
type RequestLogFields struct {
	Method        string        `json:"method"`
	URL           string        `json:"url"`
	UserAgent     string        `json:"userAgent"`
	IP            string        `json:"ip"`
	StatusCode    int           `json:"statusCode"`
	ResponseSize  int64         `json:"responseSize"`
	Duration      time.Duration `json:"duration"`
	CorrelationID string        `json:"correlationId"`
	RequestID     string        `json:"requestId"`
	UserID        string        `json:"userId"`
}

// LogRequest logs HTTP request information.
func (l *Logger) LogRequest(fields RequestLogFields) {
	event := l.Info()

	if fields.StatusCode >= 500 {
		event = l.Error()
	} else if fields.StatusCode >= 400 {
		event = l.Warn()
	}

	event.
		Str("method", fields.Method).
		Str("url", fields.URL).
		Str("user_agent", fields.UserAgent).
		Str("ip", fields.IP).
		Int("status_code", fields.StatusCode).
		Int64("response_size", fields.ResponseSize).
		Dur("duration", fields.Duration).
		Str("correlation_id", fields.CorrelationID).
		Str("request_id", fields.RequestID).
		Str("user_id", fields.UserID).
		Msg("HTTP request processed")
}

// Global logger functions for convenience

var globalLogger *Logger

// init initializes the global logger.
func init() {
	globalLogger = NewLogger("erp-system")
}

// SetGlobalLogger sets the global logger instance.
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance.
func GetGlobalLogger() *Logger {
	return globalLogger
}

// Global logging functions
func Trace() *zerolog.Event { return globalLogger.Trace() }
func Debug() *zerolog.Event { return globalLogger.Debug() }
func Info() *zerolog.Event  { return globalLogger.Info() }
func Warn() *zerolog.Event  { return globalLogger.Warn() }
func Error() *zerolog.Event { return globalLogger.Error() }
func Fatal() *zerolog.Event { return globalLogger.Fatal() }
func Panic() *zerolog.Event { return globalLogger.Panic() }

// Configuration from environment
func LoadConfigFromEnv() *Config {
	config := DefaultConfig()

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Level = LogLevel(strings.ToUpper(level))
	}

	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.Format = LogFormat(strings.ToUpper(format))
	}

	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		config.ServiceName = serviceName
	}

	if serviceVersion := os.Getenv("SERVICE_VERSION"); serviceVersion != "" {
		config.ServiceVersion = serviceVersion
	}

	if environment := os.Getenv("ENVIRONMENT"); environment != "" {
		config.Environment = environment
	}

	if timeFormat := os.Getenv("LOG_TIME_FORMAT"); timeFormat != "" {
		config.TimeFormat = timeFormat
	}

	if caller := os.Getenv("LOG_CALLER"); caller == "true" {
		config.CallerEnabled = true
	} else if caller == "false" {
		config.CallerEnabled = false
	}

	if stack := os.Getenv("LOG_STACK_TRACE"); stack == "true" {
		config.StackTrace = true
	}

	return config
}
