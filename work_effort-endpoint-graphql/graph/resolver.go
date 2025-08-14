package graph

import (
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

// Resolver is the root GraphQL resolver
type Resolver struct {
	DB     *sqlx.DB
	Logger zerolog.Logger
}

// Query resolver
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

// Mutation resolver  
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// Employee resolver
func (r *Resolver) Employee() EmployeeResolver {
	return &employeeResolver{r}
}

// Position resolver
func (r *Resolver) Position() PositionResolver {
	return &positionResolver{r}
}

// Department resolver
func (r *Resolver) Department() DepartmentResolver {
	return &departmentResolver{r}
}