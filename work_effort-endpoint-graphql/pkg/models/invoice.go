package models

import (
	"time"
	
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// WorkEffort represents an work_effort in the system
type WorkEffort struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	WorkEffortNumber   string          `json:"work_effortNumber" db:"work_effort_number"`
	CustomerID      uuid.UUID       `json:"customerId" db:"customer_id"`
	OrderID         *uuid.UUID      `json:"orderId" db:"order_id"`
	WorkEffortDate     time.Time       `json:"work_effortDate" db:"work_effort_date"`
	DueDate         time.Time       `json:"dueDate" db:"due_date"`
	Status          WorkEffortStatus   `json:"status" db:"status"`
	
	Subtotal        decimal.Decimal `json:"subtotal" db:"subtotal"`
	TaxAmount       decimal.Decimal `json:"taxAmount" db:"tax_amount"`
	DiscountAmount  decimal.Decimal `json:"discountAmount" db:"discount_amount"`
	TotalAmount     decimal.Decimal `json:"totalAmount" db:"total_amount"`
	PaidAmount      decimal.Decimal `json:"paidAmount" db:"paid_amount"`
	BalanceAmount   decimal.Decimal `json:"balanceAmount" db:"balance_amount"`
	Currency        string          `json:"currency" db:"currency"`
	
	BillingAddress  *Address        `json:"projectAddress"`
	ShippingAddress *Address        `json:"projectAddress"`
	
	Terms               *string    `json:"terms" db:"terms"`
	Notes               *string    `json:"notes" db:"notes"`
	InternalNotes       *string    `json:"internalNotes" db:"internal_notes"`
	
	SentAt              *time.Time `json:"sentAt" db:"sent_at"`
	PaidAt              *time.Time `json:"paidAt" db:"paid_at"`
	CancelledAt         *time.Time `json:"cancelledAt" db:"cancelled_at"`
	CancellationReason  *string    `json:"cancellationReason" db:"cancellation_reason"`
	
	CreatedBy       uuid.UUID      `json:"createdBy" db:"created_by"`
	CreatedAt       time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time      `json:"updatedAt" db:"updated_at"`
}

// WorkEffortItem represents a line item in an work_effort
type WorkEffortItem struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	WorkEffortID        uuid.UUID       `json:"work_effortId" db:"work_effort_id"`
	ProductID        *uuid.UUID      `json:"productId" db:"product_id"`
	OrderItemID      *uuid.UUID      `json:"orderItemId" db:"order_item_id"`
	
	Description      string          `json:"description" db:"description"`
	Quantity         decimal.Decimal `json:"quantity" db:"quantity"`
	UnitPrice        decimal.Decimal `json:"unitPrice" db:"unit_price"`
	TotalPrice       decimal.Decimal `json:"totalPrice" db:"total_price"`
	
	Taxable          bool            `json:"taxable" db:"taxable"`
	DiscountEligible bool            `json:"discountEligible" db:"discount_eligible"`
	
	CreatedAt        time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time       `json:"updatedAt" db:"updated_at"`
}

// WorkEffortTax represents tax applied to an work_effort
type WorkEffortTax struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	WorkEffortID     uuid.UUID       `json:"work_effortId" db:"work_effort_id"`
	TaxType       string          `json:"taxType" db:"tax_type"`
	TaxRate       decimal.Decimal `json:"taxRate" db:"tax_rate"`
	TaxableAmount decimal.Decimal `json:"taxableAmount" db:"taxable_amount"`
	TaxAmount     decimal.Decimal `json:"taxAmount" db:"tax_amount"`
	Description   *string         `json:"description" db:"description"`
}

// WorkEffortDiscount represents discount applied to an work_effort
type WorkEffortDiscount struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	WorkEffortID      uuid.UUID       `json:"work_effortId" db:"work_effort_id"`
	DiscountType   DiscountType    `json:"discountType" db:"discount_type"`
	DiscountValue  decimal.Decimal `json:"discountValue" db:"discount_value"`
	DiscountAmount decimal.Decimal `json:"discountAmount" db:"discount_amount"`
	Description    *string         `json:"description" db:"description"`
	CouponCode     *string         `json:"couponCode" db:"coupon_code"`
}

// WorkEffortPayment represents a payment made towards an work_effort
type WorkEffortPayment struct {
	ID               uuid.UUID     `json:"id" db:"id"`
	WorkEffortID        uuid.UUID     `json:"work_effortId" db:"work_effort_id"`
	PaymentMethod    PaymentMethod `json:"paymentMethod" db:"payment_method"`
	PaymentReference *string       `json:"paymentReference" db:"payment_reference"`
	Amount           decimal.Decimal `json:"amount" db:"amount"`
	PaymentDate      time.Time     `json:"paymentDate" db:"payment_date"`
	Notes            *string       `json:"notes" db:"notes"`
}

// Address represents project/project address
type Address struct {
	Line1      string  `json:"line1"`
	Line2      *string `json:"line2"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	PostalCode string  `json:"postalCode"`
	Country    string  `json:"country"`
}

// WorkEffortTotals represents aggregated work_effort totals
type WorkEffortTotals struct {
	TotalWorkEfforts  int             `json:"totalWorkEfforts"`
	TotalAmount    decimal.Decimal `json:"totalAmount"`
	PaidAmount     decimal.Decimal `json:"paidAmount"`
	UnpaidAmount   decimal.Decimal `json:"unpaidAmount"`
	OverdueAmount  decimal.Decimal `json:"overdueAmount"`
	Currency       string          `json:"currency"`
}

// Enums

// WorkEffortStatus represents the status of an work_effort
type WorkEffortStatus string

const (
	WorkEffortStatusDraft       WorkEffortStatus = "DRAFT"
	WorkEffortStatusPending     WorkEffortStatus = "PENDING"
	WorkEffortStatusSent        WorkEffortStatus = "SENT"
	WorkEffortStatusPaid        WorkEffortStatus = "PAID"
	WorkEffortStatusPartialPaid WorkEffortStatus = "PARTIAL_PAID"
	WorkEffortStatusOverdue     WorkEffortStatus = "OVERDUE"
	WorkEffortStatusCancelled   WorkEffortStatus = "CANCELLED"
	WorkEffortStatusRefunded    WorkEffortStatus = "REFUNDED"
)

// DiscountType represents the type of discount
type DiscountType string

const (
	DiscountTypePercentage  DiscountType = "PERCENTAGE"
	DiscountTypeFixedAmount DiscountType = "FIXED_AMOUNT"
	DiscountTypeBuyXGetY    DiscountType = "BUY_X_GET_Y"
)

// PaymentMethod represents payment method
type PaymentMethod string

const (
	PaymentMethodCash           PaymentMethod = "CASH"
	PaymentMethodCheck          PaymentMethod = "CHECK"
	PaymentMethodCreditCard     PaymentMethod = "CREDIT_CARD"
	PaymentMethodDebitCard      PaymentMethod = "DEBIT_CARD"
	PaymentMethodBankTransfer   PaymentMethod = "BANK_TRANSFER"
	PaymentMethodPayPal         PaymentMethod = "PAYPAL"
	PaymentMethodCryptocurrency PaymentMethod = "CRYPTOCURRENCY"
	PaymentMethodOther          PaymentMethod = "OTHER"
)

// Input types for GraphQL mutations

// CreateWorkEffortInput represents input for creating an work_effort
type CreateWorkEffortInput struct {
	CustomerID      uuid.UUID               `json:"customerId"`
	OrderID         *uuid.UUID              `json:"orderId"`
	WorkEffortDate     *time.Time              `json:"work_effortDate"`
	DueDate         time.Time               `json:"dueDate"`
	Currency        string                  `json:"currency"`
	BillingAddress  *AddressInput           `json:"projectAddress"`
	ShippingAddress *AddressInput           `json:"projectAddress"`
	Terms           *string                 `json:"terms"`
	Notes           *string                 `json:"notes"`
	InternalNotes   *string                 `json:"internalNotes"`
	Items           []CreateWorkEffortItemInput `json:"items"`
}

// UpdateWorkEffortInput represents input for updating an work_effort
type UpdateWorkEffortInput struct {
	DueDate         *time.Time    `json:"dueDate"`
	BillingAddress  *AddressInput `json:"projectAddress"`
	ShippingAddress *AddressInput `json:"projectAddress"`
	Terms           *string       `json:"terms"`
	Notes           *string       `json:"notes"`
	InternalNotes   *string       `json:"internalNotes"`
}

// CreateWorkEffortItemInput represents input for creating an work_effort item
type CreateWorkEffortItemInput struct {
	ProductID        *uuid.UUID      `json:"productId"`
	OrderItemID      *uuid.UUID      `json:"orderItemId"`
	Description      string          `json:"description"`
	Quantity         decimal.Decimal `json:"quantity"`
	UnitPrice        decimal.Decimal `json:"unitPrice"`
	Taxable          *bool           `json:"taxable"`
	DiscountEligible *bool           `json:"discountEligible"`
}

// UpdateWorkEffortItemInput represents input for updating an work_effort item
type UpdateWorkEffortItemInput struct {
	Description      *string          `json:"description"`
	Quantity         *decimal.Decimal `json:"quantity"`
	UnitPrice        *decimal.Decimal `json:"unitPrice"`
	Taxable          *bool            `json:"taxable"`
	DiscountEligible *bool            `json:"discountEligible"`
}

// AddWorkEffortItemInput represents input for adding an item to an work_effort
type AddWorkEffortItemInput struct {
	WorkEffortID        uuid.UUID       `json:"work_effortId"`
	ProductID        *uuid.UUID      `json:"productId"`
	OrderItemID      *uuid.UUID      `json:"orderItemId"`
	Description      string          `json:"description"`
	Quantity         decimal.Decimal `json:"quantity"`
	UnitPrice        decimal.Decimal `json:"unitPrice"`
	Taxable          *bool           `json:"taxable"`
	DiscountEligible *bool           `json:"discountEligible"`
}

// PaymentInput represents input for recording a payment
type PaymentInput struct {
	PaymentMethod    PaymentMethod   `json:"paymentMethod"`
	PaymentReference *string         `json:"paymentReference"`
	Amount           decimal.Decimal `json:"amount"`
	PaymentDate      time.Time       `json:"paymentDate"`
	Notes            *string         `json:"notes"`
}

// DiscountInput represents input for applying a discount
type DiscountInput struct {
	DiscountType  DiscountType    `json:"discountType"`
	DiscountValue decimal.Decimal `json:"discountValue"`
	Description   *string         `json:"description"`
	CouponCode    *string         `json:"couponCode"`
}

// TaxInput represents input for applying tax
type TaxInput struct {
	TaxType     string          `json:"taxType"`
	TaxRate     decimal.Decimal `json:"taxRate"`
	Description *string         `json:"description"`
}

// AddressInput represents address input
type AddressInput struct {
	Line1      string  `json:"line1"`
	Line2      *string `json:"line2"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	PostalCode string  `json:"postalCode"`
	Country    string  `json:"country"`
}

// WorkEffortFilter represents filter criteria for work_efforts
type WorkEffortFilter struct {
	Status     *WorkEffortStatus   `json:"status"`
	CustomerID *uuid.UUID       `json:"customerId"`
	StartDate  *time.Time       `json:"startDate"`
	EndDate    *time.Time       `json:"endDate"`
	MinAmount  *decimal.Decimal `json:"minAmount"`
	MaxAmount  *decimal.Decimal `json:"maxAmount"`
	SearchTerm *string          `json:"searchTerm"`
}

// PaginationInput represents pagination parameters
type PaginationInput struct {
	First  *int    `json:"first"`
	After  *string `json:"after"`
	Last   *int    `json:"last"`
	Before *string `json:"before"`
}