package models

import (
	"time"
	
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Employee represents an employee in the HR system
type Employee struct {
	ID           uuid.UUID      `json:"id" db:"id"`
	EmployeeID   string         `json:"employeeId" db:"employee_id"`
	FirstName    string         `json:"firstName" db:"first_name"`
	LastName     string         `json:"lastName" db:"last_name"`
	Email        string         `json:"email" db:"email"`
	PhoneNumber  *string        `json:"phoneNumber" db:"phone_number"`
	HireDate     time.Time      `json:"hireDate" db:"hire_date"`
	Status       EmployeeStatus `json:"status" db:"status"`
	PositionID   *uuid.UUID     `json:"positionId" db:"position_id"`
	DepartmentID *uuid.UUID     `json:"departmentId" db:"department_id"`
	ManagerID    *uuid.UUID     `json:"managerId" db:"manager_id"`
	CreatedAt    time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time      `json:"updatedAt" db:"updated_at"`
}

// Position represents a job position
type Position struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	Title           string        `json:"title" db:"title"`
	Description     *string       `json:"description" db:"description"`
	DepartmentID    *uuid.UUID    `json:"departmentId" db:"department_id"`
	MinSalary       *decimal.Decimal `json:"minSalary" db:"min_salary"`
	MaxSalary       *decimal.Decimal `json:"maxSalary" db:"max_salary"`
	SalaryCurrency  string        `json:"salaryCurrency" db:"salary_currency"`
	Requirements    []string      `json:"requirements" db:"requirements"`
	Responsibilities []string     `json:"responsibilities" db:"responsibilities"`
	IsActive        bool          `json:"isActive" db:"is_active"`
	CreatedAt       time.Time     `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time     `json:"updatedAt" db:"updated_at"`
}

// Department represents an organizational department
type Department struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	Name               string     `json:"name" db:"name"`
	Description        *string    `json:"description" db:"description"`
	ManagerID          *uuid.UUID `json:"managerId" db:"manager_id"`
	ParentDepartmentID *uuid.UUID `json:"parentDepartmentId" db:"parent_department_id"`
	IsActive           bool       `json:"isActive" db:"is_active"`
	CreatedAt          time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time  `json:"updatedAt" db:"updated_at"`
}

// SalaryRange represents salary range for a position
type SalaryRange struct {
	MinSalary decimal.Decimal `json:"minSalary"`
	MaxSalary decimal.Decimal `json:"maxSalary"`  
	Currency  string         `json:"currency"`
}

// EmployeeStatus represents the status of an employee
type EmployeeStatus string

const (
	EmployeeStatusActive     EmployeeStatus = "ACTIVE"
	EmployeeStatusInactive   EmployeeStatus = "INACTIVE" 
	EmployeeStatusTerminated EmployeeStatus = "TERMINATED"
	EmployeeStatusOnLeave    EmployeeStatus = "ON_LEAVE"
)

// Input types for GraphQL mutations

// CreateEmployeeInput represents input for creating an employee
type CreateEmployeeInput struct {
	EmployeeID   string         `json:"employeeId"`
	FirstName    string         `json:"firstName"`
	LastName     string         `json:"lastName"`
	Email        string         `json:"email"`
	PhoneNumber  *string        `json:"phoneNumber"`
	HireDate     time.Time      `json:"hireDate"`
	PositionID   *uuid.UUID     `json:"positionId"`
	DepartmentID *uuid.UUID     `json:"departmentId"`
	ManagerID    *uuid.UUID     `json:"managerId"`
}

// UpdateEmployeeInput represents input for updating an employee
type UpdateEmployeeInput struct {
	FirstName    *string         `json:"firstName"`
	LastName     *string         `json:"lastName"`
	Email        *string         `json:"email"`
	PhoneNumber  *string         `json:"phoneNumber"`
	PositionID   *uuid.UUID      `json:"positionId"`
	DepartmentID *uuid.UUID      `json:"departmentId"`
	ManagerID    *uuid.UUID      `json:"managerId"`
	Status       *EmployeeStatus `json:"status"`
}

// CreatePositionInput represents input for creating a position
type CreatePositionInput struct {
	Title           string           `json:"title"`
	Description     *string          `json:"description"`
	DepartmentID    *uuid.UUID       `json:"departmentId"`
	SalaryRange     *SalaryRangeInput `json:"salaryRange"`
	Requirements    []string         `json:"requirements"`
	Responsibilities []string        `json:"responsibilities"`
}

// UpdatePositionInput represents input for updating a position
type UpdatePositionInput struct {
	Title           *string          `json:"title"`
	Description     *string          `json:"description"`
	DepartmentID    *uuid.UUID       `json:"departmentId"`
	SalaryRange     *SalaryRangeInput `json:"salaryRange"`
	Requirements    []string         `json:"requirements"`
	Responsibilities []string        `json:"responsibilities"`
	IsActive        *bool            `json:"isActive"`
}

// CreateDepartmentInput represents input for creating a department  
type CreateDepartmentInput struct {
	Name               string     `json:"name"`
	Description        *string    `json:"description"`
	ManagerID          *uuid.UUID `json:"managerId"`
	ParentDepartmentID *uuid.UUID `json:"parentDepartmentId"`
}

// UpdateDepartmentInput represents input for updating a department
type UpdateDepartmentInput struct {
	Name               *string    `json:"name"`
	Description        *string    `json:"description"`
	ManagerID          *uuid.UUID `json:"managerId"`
	ParentDepartmentID *uuid.UUID `json:"parentDepartmentId"`
	IsActive           *bool      `json:"isActive"`
}

// SalaryRangeInput represents salary range input
type SalaryRangeInput struct {
	MinSalary decimal.Decimal `json:"minSalary"`
	MaxSalary decimal.Decimal `json:"maxSalary"`
	Currency  string         `json:"currency"`
}