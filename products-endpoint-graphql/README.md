# Products GraphQL API with AI Integration

A revolutionary product management GraphQL API built in Go, featuring advanced AI capabilities for product categorization, intelligent recommendations, and autonomous inventory optimization.

## 🚀 Revolutionary Features

### AI-Powered Product Management
- **Computer Vision Categorization**: Automatically categorize products from images using advanced computer vision
- **Intelligent Recommendations**: Machine learning-powered product recommendations with collaborative filtering
- **Dynamic Pricing**: AI-driven pricing optimization based on market conditions and demand forecasting  
- **Smart Inventory**: Autonomous inventory management with predictive reordering
- **Competitor Analysis**: Real-time competitive intelligence and market positioning

### Modern Architecture
- **GraphQL Federation**: Apollo Federation v2 compatible subgraph
- **Pure Go Implementation**: High-performance, concurrent, and memory-efficient
- **Advanced Caching**: Multi-layer caching with Redis and in-memory strategies
- **Real-time Updates**: GraphQL subscriptions for live product data
- **Comprehensive Observability**: Metrics, tracing, and structured logging

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   GraphQL       │    │   AI Services    │    │   Database      │
│   Federation    │◄──►│   - Categorization│◄──►│   PostgreSQL    │
│   Gateway       │    │   - Recommendations│   │   + Extensions  │
└─────────────────┘    │   - Vision Analysis│   └─────────────────┘
                       │   - Demand Forecast│
                       │   - Pricing Optim. │    ┌─────────────────┐
                       └──────────────────┘    │   Cache Layer    │
                                               │   Redis          │
┌─────────────────┐    ┌──────────────────┐    │   + Persistence  │
│   Auth/RBAC     │◄──►│   Business       │◄──►└─────────────────┘
│   OAuth2/JWT    │    │   Logic          │
│   Role-based    │    │   Domain Models  │    ┌─────────────────┐
└─────────────────┘    └──────────────────┘    │   External APIs  │
                                               │   - ML Models    │
                       ┌──────────────────┐    │   - Image CDN    │
                       │   Event Stream   │◄──►│   - Price APIs   │
                       │   Change Data    │    │   - Competitors  │
                       │   Capture (CDC)  │    └─────────────────┘
                       └──────────────────┘
```

## 🛠️ Technology Stack

### Core Technologies
- **Go 1.21**: Modern, concurrent, high-performance backend
- **GraphQL**: Type-safe API with gqlgen
- **PostgreSQL 15**: Advanced relational database with JSON support
- **Redis 7**: High-performance caching and session storage

### AI/ML Stack  
- **Computer Vision**: Product image analysis and feature extraction
- **NLP**: Text analysis for product descriptions and reviews
- **Recommendation Engine**: Collaborative filtering and content-based algorithms
- **Time Series Analysis**: Demand forecasting and inventory optimization

### Infrastructure
- **Docker**: Containerized deployment
- **Docker Compose**: Local development environment
- **Kubernetes**: Production orchestration (configs included)
- **Prometheus & Grafana**: Monitoring and alerting

## 🚦 Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 15 (for local development)
- Redis 7
- OpenAI API Key (for AI features)

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd products-endpoint-graphql
   ```

2. **Install dependencies**
   ```bash
   make deps
   make install-tools
   ```

3. **Generate GraphQL code**
   ```bash
   make generate
   ```

4. **Start dependencies**
   ```bash
   make docker-compose-up
   ```

5. **Run the application**
   ```bash
   make run
   # or with hot reload
   make dev
   ```

6. **Access GraphQL Playground**
   Open http://localhost:8080/playground

### Docker Development

```bash
# Start everything with Docker
docker-compose up -d

# View logs
docker-compose logs -f products-api

# Stop everything
docker-compose down
```

## 📊 Testing Strategy

### BDD Testing with Cucumber
The project follows Behavior-Driven Development with comprehensive Gherkin scenarios:

```gherkin
Feature: AI Product Categorization
  Scenario: Auto-categorize product from image
    Given I have a product with an image but no category
    When I run AI categorization
    Then the product should be assigned the correct category
    And the confidence score should be above 80%
```

### Test Coverage Requirements
- **Unit Tests**: 85%+ coverage required
- **Integration Tests**: Full GraphQL API coverage
- **BDD Tests**: End-to-end feature validation
- **Performance Tests**: Load testing with k6
- **AI Model Tests**: Accuracy and performance validation

### Running Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests
make test-integration

# BDD tests (Cucumber)
make test-bdd

# Load testing
make load-test

# Test coverage report
make test-coverage
```

## 🧠 AI Features Deep Dive

### Product Categorization
Automatically categorizes products using:
- **Computer Vision**: Analyzes product images to identify features
- **NLP Processing**: Extracts insights from product descriptions
- **Machine Learning**: Learns from human corrections and feedback

```go
// Example: AI Categorization
result, err := categorizer.CategorizeFromImage(ctx, imageURL, productID)
if err != nil {
    return nil, err
}

// Confidence-based assignment
if result.EstimatedCategories[0].Confidence > 0.8 {
    product.CategoryID = result.EstimatedCategories[0].CategoryID
}
```

### Intelligent Recommendations
Multi-algorithm recommendation engine:
- **Collaborative Filtering**: "Customers like you also purchased"
- **Content-Based**: Similar features and specifications
- **Hybrid Approach**: Combines multiple signals for optimal results

### Dynamic Pricing
AI-powered pricing optimization:
- **Market Analysis**: Real-time competitor pricing monitoring
- **Demand Forecasting**: Predictive analytics for optimal pricing
- **Elasticity Modeling**: Price sensitivity analysis

### Inventory Intelligence
Autonomous inventory management:
- **Demand Prediction**: ML-based demand forecasting
- **Seasonal Adjustments**: Automatic seasonal inventory planning
- **Reorder Optimization**: Intelligent reorder points and quantities

## 📡 GraphQL API

### Core Queries

```graphql
# Get product with AI insights
query GetProduct($id: UUID!) {
  product(id: $id) {
    id
    name
    description
    aiMetadata {
      autoCategory
      confidence
      recommendations {
        productId
        score
        explanation
      }
      pricingSuggestions {
        suggestedPrice
        reasoning
      }
    }
    inventory @hasRole(roles: ["INVENTORY_READ", "ADMIN"]) {
      quantityOnHand
      aiOptimization {
        optimalStockLevel
        reorderRecommendation
      }
    }
  }
}

# Search products with AI-powered relevance
query SearchProducts($query: String!, $limit: Int = 20) {
  searchProducts(query: $query, limit: $limit, includeAI: true) {
    id
    name
    aiMetadata {
      tags
      categoryPredictions {
        categoryName
        confidence
      }
    }
  }
}

# Get recommendations for a customer
query GetRecommendations($customerId: UUID!, $limit: Int = 10) {
  recommendedProducts(
    customerId: $customerId
    limit: $limit
    algorithm: "hybrid"
  ) {
    id
    name
    recommendationScore
    recommendationReason
  }
}
```

### AI-Enhanced Mutations

```graphql
# Trigger AI analysis
mutation AnalyzeProducts($input: AIAnalysisInput!) {
  analyzeProductWithAI(input: $input) @hasRole(roles: ["AI_ANALYSIS", "ADMIN"]) {
    id
    aiMetadata {
      autoCategory
      confidence
      imageAnalysis {
        detectedObjects {
          label
          confidence
        }
        dominantColors {
          hexCode
          colorName
        }
      }
    }
  }
}

# Generate pricing suggestions  
mutation OptimizePricing($productId: UUID!) {
  generatePricingSuggestions(productId: $productId) 
  @hasRole(roles: ["PRICING_GENERATE", "ADMIN"]) {
    suggestedPrice
    priceRange {
      minPrice
      maxPrice
      optimalPrice
    }
    factors {
      factor
      impact
      description
    }
  }
}
```

## 🔐 Security & Authorization

### Role-Based Access Control (RBAC)
Fine-grained permissions with GraphQL directives:

```graphql
type Product @auth {
  inventory: InventoryInfo @hasRole(roles: ["INVENTORY_READ", "ADMIN"])
  pricing: [ProductPricing!]! @hasRole(roles: ["PRICING_READ", "ADMIN"])
}

type Mutation {
  createProduct(input: CreateProductInput!): Product! 
    @hasRole(roles: ["PRODUCT_CREATE", "ADMIN"])
  
  analyzeProductWithAI(input: AIAnalysisInput!): [Product!]! 
    @hasRole(roles: ["AI_ANALYSIS", "ADMIN"])
}
```

### Authentication Methods
- **OAuth2/OpenID Connect**: Integration with existing authorization server
- **JWT Tokens**: Stateless authentication with role claims
- **Service-to-Service**: Client credentials flow for internal APIs

## 📈 Performance & Monitoring

### Performance Metrics
- **GraphQL Query Performance**: <200ms for simple queries, <1s for complex
- **AI Processing**: <5s per product for categorization, <2min for batch
- **Recommendation Generation**: <100ms for collaborative filtering
- **Database Queries**: <50ms for simple lookups, <200ms for aggregations

### Monitoring Stack
- **Prometheus**: Metrics collection and alerting
- **Grafana**: Visualization and dashboards
- **Jaeger**: Distributed tracing
- **Structured Logging**: JSON logs with correlation IDs

### Key Metrics Tracked
```go
// Custom metrics
var (
    aiCategorizationDuration = prometheus.NewHistogramVec(...)
    recommendationAccuracy = prometheus.NewGaugeVec(...)
    inventoryOptimizationSavings = prometheus.NewCounterVec(...)
    graphqlQueryComplexity = prometheus.NewHistogramVec(...)
)
```

## 🔧 Configuration

### Environment Variables

```bash
# Server Configuration
PRODUCTS_API_SERVER_HOST=0.0.0.0
PRODUCTS_API_SERVER_PORT=8080
PRODUCTS_API_SERVER_ENVIRONMENT=production

# Database
PRODUCTS_API_DATABASE_HOST=localhost
PRODUCTS_API_DATABASE_PORT=5432
PRODUCTS_API_DATABASE_NAME=products_db
PRODUCTS_API_DATABASE_USERNAME=products_user
PRODUCTS_API_DATABASE_PASSWORD=secure_password

# AI Configuration
PRODUCTS_API_AI_ENABLED=true
PRODUCTS_API_AI_MODEL_PROVIDER=openai
PRODUCTS_API_AI_MODEL_API_KEY=sk-...
PRODUCTS_API_AI_BATCH_SIZE=10
PRODUCTS_API_AI_CONFIDENCE_THRESHOLD=0.8

# GraphQL
PRODUCTS_API_GRAPHQL_PLAYGROUND=false  # Disable in production
PRODUCTS_API_GRAPHQL_COMPLEXITY_LIMIT=1000
PRODUCTS_API_GRAPHQL_ENABLE_DATALOADER=true

# Authentication
PRODUCTS_API_AUTH_ENABLED=true
PRODUCTS_API_AUTH_OAUTH2_ISSUER=https://auth.example.com
PRODUCTS_API_AUTH_JWT_SECRET=your-jwt-secret
```

### Configuration File (YAML)
```yaml
server:
  port: 8080
  environment: production
  
database:
  host: localhost
  port: 5432
  name: products_db
  max_open_conns: 25
  
ai:
  enabled: true
  categorization_enabled: true
  recommendation_enabled: true
  model_provider: openai
  batch_size: 10
  confidence_threshold: 0.8
  
graphql:
  playground: false
  complexity_limit: 1000
  enable_dataloader: true
```

## 📝 Development Guide

### Code Organization
```
products-endpoint-graphql/
├── cmd/server/              # Application entry point
├── internal/               # Private application code
│   ├── ai/                 # AI services
│   ├── config/             # Configuration management
│   ├── database/           # Database operations
│   └── service/            # Business logic
├── pkg/                    # Public packages
│   ├── models/             # Domain models
│   ├── resolvers/          # GraphQL resolvers
│   ├── directives/         # GraphQL directives
│   └── middleware/         # HTTP middleware
├── graph/                  # GraphQL schema and generated code
├── test/                   # Test utilities
└── migrations/             # Database migrations
```

### Adding New AI Features

1. **Define the AI service interface**
   ```go
   type NewAIService interface {
       ProcessProduct(ctx context.Context, product *models.Product) (*Result, error)
       BatchProcess(ctx context.Context, products []models.Product) (map[uuid.UUID]*Result, error)
   }
   ```

2. **Implement the service**
   ```go
   func (s *NewAIService) ProcessProduct(ctx context.Context, product *models.Product) (*Result, error) {
       // AI processing logic
       return result, nil
   }
   ```

3. **Add GraphQL schema**
   ```graphql
   extend type Product {
       newAIData: NewAIResult @complexity(multiplier: 3, maximum: 1)
   }
   
   extend type Mutation {
       processWithNewAI(productId: UUID!): NewAIResult! 
         @hasRole(roles: ["NEW_AI_FEATURE", "ADMIN"])
   }
   ```

4. **Add comprehensive tests**
   ```go
   func TestNewAIService(t *testing.T) {
       // Unit tests with 85%+ coverage
       // Integration tests with real data
       // Performance benchmarks
   }
   ```

### Database Migrations
```bash
# Create new migration
make db-migrate-create name=add_new_ai_feature

# Apply migrations
make db-migrate-up

# Rollback migrations  
make db-migrate-down
```

## 🚀 Deployment

### Production Checklist
- [ ] Environment variables configured
- [ ] Database migrations applied
- [ ] SSL/TLS certificates installed
- [ ] Monitoring and alerting configured
- [ ] Log aggregation setup
- [ ] Backup strategy implemented
- [ ] Security scan passed
- [ ] Load testing completed
- [ ] API key rotation configured

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: products-graphql-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: products-graphql-api
  template:
    spec:
      containers:
      - name: api
        image: erp-microservices/products-graphql-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: PRODUCTS_API_DATABASE_HOST
          value: "postgres-service"
        - name: PRODUCTS_API_AI_MODEL_API_KEY
          valueFrom:
            secretKeyRef:
              name: ai-secrets
              key: openai-api-key
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## 🤝 Contributing

### Development Setup
1. Fork the repository
2. Create a feature branch
3. Follow the established patterns
4. Write comprehensive tests (85%+ coverage required)
5. Update documentation
6. Submit a pull request

### Code Standards
- **Go Best Practices**: Follow effective Go patterns
- **GraphQL Guidelines**: Schema-first development
- **Testing Requirements**: BDD scenarios + comprehensive unit tests
- **Documentation**: Godoc for all public functions
- **Security**: No secrets in code, proper input validation

### Commit Convention
```
feat: add AI-powered inventory optimization
fix: resolve race condition in recommendation engine  
docs: update GraphQL schema documentation
test: add integration tests for categorization service
perf: optimize database queries for product search
```

## 🆘 Troubleshooting

### Common Issues

**AI Services Not Working**
```bash
# Check AI service configuration
make config-check

# Verify API keys
docker-compose exec products-api env | grep AI

# Check service logs
docker-compose logs products-api | grep -i "ai"
```

**Database Connection Issues**
```bash
# Test database connectivity
docker-compose exec postgres psql -U products_user -d products_db -c "SELECT 1;"

# Check connection pool
docker-compose exec products-api wget -O- http://localhost:8080/metrics | grep db_
```

**GraphQL Errors**
```bash
# Enable GraphQL query logging
export PRODUCTS_API_GRAPHQL_ENABLE_QUERY_LOG=true

# Check schema validation
make generate

# Test queries in playground
open http://localhost:8080/playground
```

### Performance Issues
- Monitor GraphQL query complexity
- Check Redis cache hit rates
- Review database query performance
- Analyze AI processing times

## 📚 Additional Resources

- [GraphQL Best Practices](https://graphql.org/learn/best-practices/)
- [Go Performance Optimization](https://golang.org/doc/effective_go.html)
- [PostgreSQL Performance Tuning](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [AI/ML Model Deployment](https://ml-ops.org/)
- [Kubernetes Production Readiness](https://kubernetes.io/docs/concepts/cluster-administration/)

## 📄 License

Apache License 2.0 - see LICENSE file for details.

## 🏷️ Version

Current version: **1.0.0**

Built with ❤️ using Go, GraphQL, and cutting-edge AI technology.