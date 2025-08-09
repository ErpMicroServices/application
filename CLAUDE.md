# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an ERP (Enterprise Resource Planning) microservices system built with a polyglot architecture. Each business domain is implemented as a set of modular services following consistent patterns. The `application` module serves as the parent Maven POM for all Java-based child modules in the ERP system.

## Technology Stack

### Backend Services
- **Java**: 8-21 (varies by service, parent POM uses Java 8)
- **Spring Boot**: 2.x-3.x with Spring Cloud (parent uses 2.0.1.RELEASE)
- **Build Tools**: Maven (primary for parent POM) and Gradle (for individual services)
- **API Layer**: GraphQL (Spring GraphQL)
- **Authentication**: OAuth2 with Spring Authorization Server

### Database Layer
- **PostgreSQL**: 15.x (9.3 JDBC driver in parent POM)
- **Schema Management**: Liquibase
- **Node.js**: For database tooling and BDD tests

### Frontend
- **React**: 18.x
- **Build Tool**: npm/yarn
- **Testing**: Cucumber BDD

### Infrastructure
- **Docker**: All services containerized
- **Spring Cloud Config**: Centralized configuration
- **Service Discovery**: Spring Cloud patterns

## Project Structure

### Parent POM Structure
```
application/
├── LICENSE               # Apache 2.0 license
├── README.md            # Module documentation
├── pom.xml              # Parent Maven POM configuration
└── [submodules]/        # Git submodules for each service
```

### Service Module Pattern
Each business domain follows this modular pattern:
```
{domain}-database/        # PostgreSQL + Liquibase migrations
{domain}-features/        # Gherkin BDD specifications  
{domain}-endpoint-graphql/# Spring Boot GraphQL API
{domain}-ui-web/         # React frontend
```

### Core Business Domains
- **people_and_organizations**: Party management, relationships, communication events
- **accounting_and_budgeting**: Financial data and budgets
- **e_commerce**: Web content, subscriptions, user preferences
- **products**: Product catalog management
- **order/invoice/shipments**: Order fulfillment pipeline
- **human_resources**: Employee management
- **work_effort**: Project and task management

### Shared Components
- **common**: Shared Java utilities and base classes
- **config_server**: Spring Cloud Config Server
- **authorization_server**: OAuth2 authorization server
- **test_utils**: Testing utilities

## Maven Parent POM Configuration

### Project Coordinates
- **GroupId**: `erp_microservices`
- **ArtifactId**: `application`
- **Version**: `1.0.0-SNAPSHOT`
- **Packaging**: `pom`

### Key Properties
```xml
<java.version>8</java.version>
<hibernate.core.version>4.3.10.Final</hibernate.core.version>
<jersey.version>2.19</jersey.version>
<junit.version>4.12</junit.version>
```

### Dependency Management

#### Core Dependencies Managed
- **Spring Boot**: 2.0.1.RELEASE (parent)
- **Hibernate**: 4.3.10.Final
- **Jersey**: 2.19 for RESTful services
- **JUnit**: 4.12 for testing
- **PostgreSQL**: 9.3-1100-jdbc41
- **Jackson**: 2.4.0+ for JSON processing
- **Functional Java**: 4.7 for functional programming

#### Testing Framework Stack
- **JUnit**: Core testing framework
- **JBehave**: 4.1.3 for BDD testing
- **Selenium**: 2.46.0 for web testing
- **Mockito**: 2.15.0 for mocking

## Common Development Commands

### Parent POM / All Modules (Maven)
```bash
# Install all modules
mvn clean install

# Compile all modules
mvn compile

# Run tests across all modules
mvn test

# Generate project reports
mvn site

# View dependency tree
mvn dependency:tree

# Analyze dependencies
mvn dependency:analyze

# Check for updates
mvn versions:display-dependency-updates
```

### Database Services (*-database modules)
```bash
# Install dependencies
npm install

# Run Liquibase migrations
npm run update_database

# Run BDD tests
npm test

# Build Docker image
docker build -t {service}-db .
```

### Java Services (Gradle-based)
```bash
# Build and run tests
./gradlew build

# Run BDD/integration tests only
./gradlew behaviorTest

# Run application locally
./gradlew bootRun

# Build Docker image using buildpacks
./gradlew bootBuildImage

# Clean build
./gradlew clean build
```

### Java Services (Maven-based)
```bash
# Build and run tests
mvn clean install

# Run application
mvn spring-boot:run

# Skip tests during build
mvn clean install -DskipTests
```

### React UI Services (*-ui-web modules)
```bash
# Install dependencies
npm install

# Start development server
npm start

# Build for production
npm run build

# Run BDD tests
npm run bdd_test

# Run unit tests
npm test
```

### Docker Operations
```bash
# Start service with dependencies
docker-compose up -d

# View logs
docker-compose logs -f {service-name}

# Rebuild and restart
docker-compose down && docker-compose build && docker-compose up -d
```

## Testing Strategy

### BDD Testing Workflow
1. Features are defined in Gherkin files (`*.feature`)
2. Database modules test schema and data integrity
3. Endpoint modules test GraphQL APIs with full integration
4. UI modules test user interactions

### Running Tests
```bash
# Database BDD tests (from *-database directory)
npm test

# Java BDD tests (from *-endpoint-graphql directory)  
./gradlew behaviorTest

# React BDD tests (from *-ui-web directory)
npm run bdd_test

# Run specific Cucumber scenario
./gradlew behaviorTest -Dcucumber.filter.tags="@specific-tag"
```

## GraphQL Development

### Schema Files
- Schema definitions: `src/main/resources/graphql/schema.graphqls`
- Test queries: `src/test/resources/graphql-test/*.graphql`

### Testing GraphQL Endpoints
```bash
# Use GraphQL test files in src/test/resources/graphql-test/
# Run specific query tests via behaviorTest
```

## Database Development

### Liquibase Migrations
- Change logs: `database_change_log.yml`
- SQL scripts: `sql/` directory
- Rollback: `sql/rollback_database.sql`

### Database Connection
```yaml
# Default local PostgreSQL configuration
host: localhost
port: 5432
database: {service_name}_db
username: {service_name}_user
password: {service_name}_password
```

## Spring Profiles

Services use Spring profiles for environment-specific configuration:
- `local`: Local development
- `test`: Testing environment
- `cicd`: CI/CD pipeline
- `prod`: Production

Activate profile:
```bash
./gradlew bootRun --args='--spring.profiles.active=local'
```

## Development Workflow

### Child Module Integration
Child modules should:
1. **Inherit Parent**: Reference this POM as parent
```xml
<parent>
    <groupId>erp_microservices</groupId>
    <artifactId>application</artifactId>
    <version>1.0.0-SNAPSHOT</version>
</parent>
```

2. **Use Managed Dependencies**: Leverage dependency management without version specifications
3. **Follow Conventions**: Use established patterns for packaging and naming

### Starting a New Feature
1. Create feature branch following naming conventions
2. Write BDD features in `*-features` module
3. Implement database changes in `*-database` module
4. Run database BDD tests to verify schema
5. Implement GraphQL API in `*-endpoint-graphql` module
6. Run integration tests to verify API
7. Implement UI changes in `*-ui-web` module if needed
8. Run full test suite before PR

### Service Dependencies
Most services depend on:
1. PostgreSQL database (start with Docker)
2. Config Server (for centralized configuration)
3. Authorization Server (for OAuth2 authentication)

Start infrastructure:
```bash
# From project root (application directory)
docker-compose -f infrastructure-docker-compose.yml up -d
```

### Adding Dependencies
1. **Check Managed Versions**: See if dependency is already managed in parent POM
2. **Add to dependencyManagement**: For new dependencies used by multiple modules
3. **Update Properties**: For version properties shared across modules
4. **Document Changes**: Update relevant documentation

### Version Management
- **SNAPSHOT Versions**: Used during development
- **Version Properties**: Centralized in parent POM
- **Dependency Alignment**: Ensure compatible versions across modules

## Architecture Principles

### Domain-Driven Design
- Each service represents a bounded context
- Entities follow DDD patterns (AggregateEntity, Type hierarchies)
- Repository pattern for data access

### GraphQL Best Practices
- Schema-first development
- Comprehensive integration testing
- Consistent error handling with GraphQL error types

### Database Design
- UUID primary keys for distributed systems
- Audit fields on all tables (created_at, updated_at)
- Type tables for extensible enumerations
- Soft deletes with status fields

## Quality and Reporting

### Reporting Plugins Configured
- **Maven Project Info Reports**: Standard project information
- **PMD**: Code quality analysis
- **Checkstyle**: Code style enforcement

### Quality Standards
- **PMD Analysis**: Enabled for code quality checks
- **Checkstyle**: Configured for consistent formatting
- **Target JDK**: Java 1.8
- **Source Encoding**: UTF-8

## Repository Standards

### Artifact Naming
- **GroupId**: `erp_microservices`
- **Version**: `1.0.0-SNAPSHOT` for development
- **Packaging**: Appropriate for module type (jar, war, pom)

### Build Standards
- **Source Encoding**: UTF-8
- **Report Encoding**: UTF-8  
- **Compiler Target**: Java 8
- **License**: Apache 2.0

### Maven Repositories
- **Spring Releases**: https://repo.spring.io/libs-release
- **Maven Central**: Default repository

## Module Dependencies

This parent POM is referenced by:
- **common**: Shared utilities and models
- **test_utils**: Testing utilities
- Java-based endpoint modules
- Java-based service modules

## Troubleshooting

### Common Issues

1. **Database connection failures**
   - Ensure PostgreSQL is running on port 5432
   - Check credentials in application.yml
   - Verify database exists and migrations have run

2. **Test failures**
   - Run `npm install` or `./gradlew clean` to refresh dependencies
   - Check test database is clean (rollback if needed)
   - Verify Spring profiles are correctly set

3. **Build failures**
   - Java version compatibility (check gradle.properties or pom.xml)
   - Clear Gradle/Maven cache: `./gradlew clean` or `mvn clean`
   - Update dependencies: `./gradlew --refresh-dependencies`

## Important Notes

### Java Version Strategy
- **Current**: Java 8 (legacy compatibility in parent POM)
- **Migration Planning**: Consider upgrading to newer Java versions
- **Compatibility**: Ensure all child modules support target Java version

### Spring Boot Integration
- **Parent Chain**: Spring Boot → Application → Child Modules
- **Version Management**: Spring Boot manages many dependency versions
- **Override Capability**: Can override Spring Boot versions in dependencyManagement

### Legacy Considerations
- **Older Dependencies**: Some dependencies may need updates
- **Security**: Regular dependency security scans recommended
- **Compatibility**: Test dependency updates across all child modules

### General Guidelines
- Always run BDD tests before committing changes
- Each service can be developed and deployed independently
- Use Spring profiles to manage environment-specific configuration
- Follow existing patterns for consistency across services
- GraphQL schemas should be backwards compatible

## Task Master AI Instructions
**Import Task Master's development workflow commands and guidelines, treat as if import is in the main CLAUDE.md file.**
@../.taskmaster/CLAUDE.md

## Task Master AI Instructions
**Import Task Master's development workflow commands and guidelines, treat as if import is in the main CLAUDE.md file.**
@./.taskmaster/CLAUDE.md
