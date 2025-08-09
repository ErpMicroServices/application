// Package pagination provides GraphQL-compatible pagination utilities for the ERP microservices system.
// It supports both cursor-based pagination (following GraphQL Cursor Connections Specification)
// and offset-based pagination for simpler use cases.
package pagination

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/erpmicroservices/common-go/pkg/uuid"
)

// PageInfo contains information about the current page in cursor-based pagination.
// This follows the GraphQL Cursor Connections Specification.
type PageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor,omitempty"`
	EndCursor       string `json:"endCursor,omitempty"`
}

// Connection represents a paginated collection following GraphQL conventions.
type Connection[T any] struct {
	Edges      []Edge[T] `json:"edges"`
	PageInfo   PageInfo  `json:"pageInfo"`
	TotalCount int       `json:"totalCount"`
}

// Edge represents an edge in a connection with a cursor and node.
type Edge[T any] struct {
	Cursor string `json:"cursor"`
	Node   T      `json:"node"`
}

// CursorArgs represents cursor-based pagination arguments.
type CursorArgs struct {
	First  *int    `json:"first,omitempty"`
	After  *string `json:"after,omitempty"`
	Last   *int    `json:"last,omitempty"`
	Before *string `json:"before,omitempty"`
}

// OffsetArgs represents offset-based pagination arguments.
type OffsetArgs struct {
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`
	Page   *int `json:"page,omitempty"`
	Size   *int `json:"size,omitempty"`
}

// Cursor represents a cursor for pagination.
type Cursor struct {
	Value string
	Type  CursorType
}

// CursorType defines the type of cursor value.
type CursorType string

const (
	CursorTypeID        CursorType = "id"
	CursorTypeTimestamp CursorType = "timestamp"
	CursorTypeOffset    CursorType = "offset"
	CursorTypeCustom    CursorType = "custom"
)

// Constants for pagination limits and defaults.
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
	MinPageSize     = 1
)

// NewConnection creates a new empty connection.
func NewConnection[T any]() *Connection[T] {
	return &Connection[T]{
		Edges:      make([]Edge[T], 0),
		PageInfo:   PageInfo{},
		TotalCount: 0,
	}
}

// AddEdge adds an edge to the connection.
func (c *Connection[T]) AddEdge(cursor string, node T) {
	edge := Edge[T]{
		Cursor: cursor,
		Node:   node,
	}
	c.Edges = append(c.Edges, edge)
}

// UpdatePageInfo updates the page info based on the current edges and pagination parameters.
func (c *Connection[T]) UpdatePageInfo(args CursorArgs, totalCount int, hasMore bool) {
	c.TotalCount = totalCount
	c.PageInfo.HasNextPage = hasMore
	c.PageInfo.HasPreviousPage = args.After != nil

	if len(c.Edges) > 0 {
		c.PageInfo.StartCursor = c.Edges[0].Cursor
		c.PageInfo.EndCursor = c.Edges[len(c.Edges)-1].Cursor
	}
}

// ValidateCursorArgs validates cursor-based pagination arguments.
func (args CursorArgs) Validate() error {
	// Cannot specify both first and last
	if args.First != nil && args.Last != nil {
		return fmt.Errorf("cannot specify both first and last")
	}

	// Cannot specify both after and before
	if args.After != nil && args.Before != nil {
		return fmt.Errorf("cannot specify both after and before")
	}

	// Validate first argument
	if args.First != nil {
		if *args.First < 0 {
			return fmt.Errorf("first must be non-negative")
		}
		if *args.First > MaxPageSize {
			return fmt.Errorf("first cannot exceed %d", MaxPageSize)
		}
	}

	// Validate last argument
	if args.Last != nil {
		if *args.Last < 0 {
			return fmt.Errorf("last must be non-negative")
		}
		if *args.Last > MaxPageSize {
			return fmt.Errorf("last cannot exceed %d", MaxPageSize)
		}
	}

	return nil
}

// GetLimit returns the effective limit for the query.
func (args CursorArgs) GetLimit() int {
	if args.First != nil {
		return *args.First
	}
	if args.Last != nil {
		return *args.Last
	}
	return DefaultPageSize
}

// IsForward returns true if this is forward pagination (first/after).
func (args CursorArgs) IsForward() bool {
	return args.First != nil || args.After != nil
}

// ValidateOffsetArgs validates offset-based pagination arguments.
func (args OffsetArgs) Validate() error {
	if args.Limit != nil {
		if *args.Limit <= 0 {
			return fmt.Errorf("limit must be positive")
		}
		if *args.Limit > MaxPageSize {
			return fmt.Errorf("limit cannot exceed %d", MaxPageSize)
		}
	}

	if args.Offset != nil && *args.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}

	if args.Page != nil && *args.Page <= 0 {
		return fmt.Errorf("page must be positive")
	}

	if args.Size != nil {
		if *args.Size <= 0 {
			return fmt.Errorf("size must be positive")
		}
		if *args.Size > MaxPageSize {
			return fmt.Errorf("size cannot exceed %d", MaxPageSize)
		}
	}

	return nil
}

// GetLimit returns the effective limit for offset-based pagination.
func (args OffsetArgs) GetLimit() int {
	if args.Limit != nil {
		return *args.Limit
	}
	if args.Size != nil {
		return *args.Size
	}
	return DefaultPageSize
}

// GetOffset returns the effective offset for the query.
func (args OffsetArgs) GetOffset() int {
	if args.Offset != nil {
		return *args.Offset
	}
	if args.Page != nil && args.Size != nil {
		return (*args.Page - 1) * *args.Size
	}
	if args.Page != nil {
		return (*args.Page - 1) * DefaultPageSize
	}
	return 0
}

// EncodeCursor encodes a cursor value with its type.
func EncodeCursor(value string, cursorType CursorType) string {
	combined := fmt.Sprintf("%s:%s", cursorType, value)
	return base64.StdEncoding.EncodeToString([]byte(combined))
}

// DecodeCursor decodes a cursor string into its value and type.
func DecodeCursor(cursor string) (Cursor, error) {
	if cursor == "" {
		return Cursor{}, fmt.Errorf("cursor cannot be empty")
	}

	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return Cursor{}, fmt.Errorf("invalid cursor format: %w", err)
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return Cursor{}, fmt.Errorf("invalid cursor structure")
	}

	return Cursor{
		Type:  CursorType(parts[0]),
		Value: parts[1],
	}, nil
}

// EncodeIDCursor creates a cursor for an ID-based pagination.
func EncodeIDCursor(id uuid.UUID) string {
	return EncodeCursor(id.String(), CursorTypeID)
}

// EncodeTimestampCursor creates a cursor for timestamp-based pagination.
func EncodeTimestampCursor(timestamp time.Time) string {
	return EncodeCursor(timestamp.UTC().Format(time.RFC3339Nano), CursorTypeTimestamp)
}

// EncodeOffsetCursor creates a cursor for offset-based pagination.
func EncodeOffsetCursor(offset int) string {
	return EncodeCursor(strconv.Itoa(offset), CursorTypeOffset)
}

// DecodeIDCursor decodes an ID cursor.
func DecodeIDCursor(cursor string) (uuid.UUID, error) {
	c, err := DecodeCursor(cursor)
	if err != nil {
		return uuid.UUID{}, err
	}

	if c.Type != CursorTypeID {
		return uuid.UUID{}, fmt.Errorf("expected ID cursor, got %s", c.Type)
	}

	return uuid.NewFromString(c.Value)
}

// DecodeTimestampCursor decodes a timestamp cursor.
func DecodeTimestampCursor(cursor string) (time.Time, error) {
	c, err := DecodeCursor(cursor)
	if err != nil {
		return time.Time{}, err
	}

	if c.Type != CursorTypeTimestamp {
		return time.Time{}, fmt.Errorf("expected timestamp cursor, got %s", c.Type)
	}

	return time.Parse(time.RFC3339Nano, c.Value)
}

// DecodeOffsetCursor decodes an offset cursor.
func DecodeOffsetCursor(cursor string) (int, error) {
	c, err := DecodeCursor(cursor)
	if err != nil {
		return 0, err
	}

	if c.Type != CursorTypeOffset {
		return 0, fmt.Errorf("expected offset cursor, got %s", c.Type)
	}

	return strconv.Atoi(c.Value)
}

// OffsetPage represents a page in offset-based pagination.
type OffsetPage[T any] struct {
	Data       []T  `json:"data"`
	TotalCount int  `json:"totalCount"`
	Page       int  `json:"page"`
	Size       int  `json:"size"`
	TotalPages int  `json:"totalPages"`
	HasNext    bool `json:"hasNext"`
	HasPrev    bool `json:"hasPrev"`
}

// NewOffsetPage creates a new offset-based page.
func NewOffsetPage[T any](data []T, totalCount, page, size int) *OffsetPage[T] {
	totalPages := (totalCount + size - 1) / size
	if totalPages == 0 {
		totalPages = 1
	}

	return &OffsetPage[T]{
		Data:       data,
		TotalCount: totalCount,
		Page:       page,
		Size:       size,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// PaginationBuilder helps build pagination queries with method chaining.
type PaginationBuilder struct {
	limit    int
	offset   int
	orderBy  string
	orderDir string
	filters  map[string]interface{}
}

// NewPaginationBuilder creates a new pagination builder.
func NewPaginationBuilder() *PaginationBuilder {
	return &PaginationBuilder{
		limit:   DefaultPageSize,
		offset:  0,
		filters: make(map[string]interface{}),
	}
}

// WithLimit sets the limit for pagination.
func (b *PaginationBuilder) WithLimit(limit int) *PaginationBuilder {
	if limit > 0 && limit <= MaxPageSize {
		b.limit = limit
	}
	return b
}

// WithOffset sets the offset for pagination.
func (b *PaginationBuilder) WithOffset(offset int) *PaginationBuilder {
	if offset >= 0 {
		b.offset = offset
	}
	return b
}

// WithOrderBy sets the ordering field.
func (b *PaginationBuilder) WithOrderBy(field string) *PaginationBuilder {
	b.orderBy = field
	return b
}

// WithOrderDirection sets the ordering direction.
func (b *PaginationBuilder) WithOrderDirection(direction string) *PaginationBuilder {
	if direction == "ASC" || direction == "DESC" {
		b.orderDir = direction
	}
	return b
}

// WithFilter adds a filter to the pagination query.
func (b *PaginationBuilder) WithFilter(key string, value interface{}) *PaginationBuilder {
	b.filters[key] = value
	return b
}

// Build returns the configured pagination parameters.
func (b *PaginationBuilder) Build() (limit, offset int, orderBy, orderDir string, filters map[string]interface{}) {
	return b.limit, b.offset, b.orderBy, b.orderDir, b.filters
}

// ApplyOffsetArgs applies offset arguments to the builder.
func (b *PaginationBuilder) ApplyOffsetArgs(args OffsetArgs) *PaginationBuilder {
	return b.WithLimit(args.GetLimit()).WithOffset(args.GetOffset())
}

// ApplyCursorArgs applies cursor arguments to the builder.
func (b *PaginationBuilder) ApplyCursorArgs(args CursorArgs) *PaginationBuilder {
	return b.WithLimit(args.GetLimit())
}
