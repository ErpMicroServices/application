# Invoice GraphQL API

A GraphQL API service for managing invoices, billing, and payments in the ERP microservices system.

## Features

- **Invoice Management**: Complete invoice lifecycle from creation to payment
- **Billing Operations**: Customer billing with multiple line items
- **Payment Processing**: Track payments and payment methods
- **Tax Calculation**: Automatic tax calculation and tracking  
- **Discount Management**: Apply various types of discounts and coupons
- **Multi-Currency Support**: Handle invoices in different currencies
- **Authentication**: JWT-based authentication with role-based access control
- **Authorization**: Role-based permissions (BILLING_ADMIN, BILLING_MANAGER)
- **Apollo Federation**: Compatible with GraphQL federation gateway
- **Health Checks**: Built-in health and readiness endpoints
- **Observability**: Structured logging with zerolog

## Quick Start

### Prerequisites

- Go 1.23 or later
- PostgreSQL 15+
- Redis (optional, for caching)

### Development Setup

1. **Clone and navigate to the project:**
   ```bash
   cd invoice-endpoint-graphql
   ```

2. **Install dependencies:**
   ```bash
   make deps
   ```

3. **Generate GraphQL code:**
   ```bash
   make generate
   ```

4. **Run the development server:**
   ```bash
   make run
   ```

   The API will be available at:
   - GraphQL Endpoint: http://localhost:8080/graphql
   - GraphQL Playground: http://localhost:8080/playground
   - Health Check: http://localhost:8080/health

### Docker Setup

```bash
make docker-up  # Build and run with Docker Compose
make docker-down # Stop services
```

## API Schema

### Key Types

- **Invoice**: Invoice header with totals and metadata
- **InvoiceItem**: Line items with products and quantities
- **InvoiceTax**: Tax calculations and rates
- **InvoiceDiscount**: Discounts and coupon codes
- **InvoicePayment**: Payment records and methods

### Example Queries

```graphql
query GetInvoices {
  invoices(filter: {status: SENT}) {
    edges {
      node {
        id
        invoiceNumber
        customer {
          name
          email
        }
        totalAmount
        balanceAmount
        dueDate
        status
        items {
          description
          quantity
          unitPrice
          totalPrice
        }
      }
    }
  }
}

query GetInvoice {
  invoice(id: "123") {
    invoiceNumber
    customer {
      name
    }
    invoiceDate
    dueDate
    totalAmount
    paidAmount
    balanceAmount
    status
    items {
      description
      quantity
      unitPrice
      totalPrice
    }
    payments {
      amount
      paymentDate
      paymentMethod
    }
  }
}
```

### Example Mutations

```graphql
mutation CreateInvoice {
  createInvoice(input: {
    customerId: "customer-123"
    dueDate: "2024-02-15T00:00:00Z"
    currency: "USD"
    items: [
      {
        description: "Professional Services"
        quantity: 10
        unitPrice: 150.00
        taxable: true
      }
    ]
  }) {
    id
    invoiceNumber
    totalAmount
    status
  }
}

mutation MarkInvoicePaid {
  markInvoicePaid(
    id: "invoice-123"
    input: {
      paymentMethod: CREDIT_CARD
      amount: 1500.00
      paymentDate: "2024-01-15T10:30:00Z"
      paymentReference: "CC-REF-12345"
    }
  ) {
    id
    status
    paidAmount
    balanceAmount
  }
}

mutation ApplyDiscount {
  applyDiscount(
    invoiceId: "invoice-123"
    input: {
      discountType: PERCENTAGE
      discountValue: 10.00
      description: "Early payment discount"
    }
  ) {
    id
    discountAmount
    totalAmount
  }
}
```

## Configuration

Key environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | Server port |
| `DATABASE_NAME` | `invoice_db` | Database name |
| `DATABASE_USER` | `invoice_user` | Database user |
| `DATABASE_PASSWORD` | `invoice_password` | Database password |
| `AUTH_ENABLED` | `false` | Enable authentication |
| `GRAPHQL_PLAYGROUND` | `true` | Enable GraphQL playground |

## Business Logic

### Invoice Lifecycle

1. **Draft** → Create invoice with items
2. **Pending** → Review and validate
3. **Sent** → Send to customer  
4. **Paid/Partial** → Record payments
5. **Overdue** → Past due date
6. **Cancelled** → Cancel if needed

### Calculations

- **Subtotal**: Sum of all line items
- **Tax Amount**: Applied based on tax rules
- **Discount Amount**: Various discount types supported
- **Total Amount**: Subtotal + Tax - Discount
- **Balance**: Total - Paid amounts

### Payment Handling

- Multiple payment methods supported
- Partial payments allowed
- Payment history tracked
- Automatic balance calculation

## Authentication & Authorization

Role-based access control:

- **BILLING_ADMIN**: Full access to all billing operations
- **BILLING_MANAGER**: Create and manage invoices
- **BILLING_USER**: Read-only access to invoices

## Development

### Project Structure

```
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── repositories/   # Data access layer
│   └── services/       # Business logic
├── pkg/
│   ├── directives/     # GraphQL directives
│   ├── middleware/     # HTTP middleware
│   └── models/         # Data models
└── schema.graphql      # GraphQL schema
```

### Available Commands

```bash
make help              # Show all available commands
make deps              # Install dependencies
make generate          # Generate GraphQL code
make build             # Build the application
make run               # Run the application
make test              # Run tests
make docker-build      # Build Docker image
```

## License

This project is licensed under the Apache License 2.0.