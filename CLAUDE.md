# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this module.

## Module Overview

The `application` module serves as the parent Maven POM for the entire ERP Micro Services project. It defines common dependencies, dependency management, properties, and build configuration that are inherited by all Java-based child modules in the ERP system.

## Technology Stack

- **Build Tool**: Maven 3.x
- **Java Version**: Java 8 (as defined in properties)
- **Parent**: Spring Boot Starter Parent 2.0.1.RELEASE
- **Packaging**: POM (parent project)
- **License**: Apache 2.0

## Project Structure

```
application/
├── LICENSE               # Apache 2.0 license
├── README.md            # Module documentation
└── pom.xml              # Parent Maven POM configuration
```

## Maven Configuration

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

## Dependency Management

### Core Dependencies Managed
- **Spring Boot**: 2.0.1.RELEASE (parent)
- **Hibernate**: 4.3.10.Final
- **Jersey**: 2.19 for RESTful services
- **JUnit**: 4.12 for testing
- **PostgreSQL**: 9.3-1100-jdbc41
- **Jackson**: 2.4.0+ for JSON processing
- **Functional Java**: 4.7 for functional programming

### Testing Framework Stack
- **JUnit**: Core testing framework
- **JBehave**: 4.1.3 for BDD testing
- **Selenium**: 2.46.0 for web testing
- **Mockito**: 2.15.0 for mocking

## Repository Configuration

### Maven Repositories
- **Spring Releases**: https://repo.spring.io/libs-release
- **Maven Central**: Default repository

## Build and Development

### Maven Commands
```bash
# Install all modules
mvn clean install

# Compile all modules
mvn compile

# Run tests across all modules
mvn test

# Generate project reports
mvn site
```

### Dependency Analysis
```bash
# View dependency tree
mvn dependency:tree

# Analyze dependencies
mvn dependency:analyze

# Check for updates
mvn versions:display-dependency-updates
```

## Child Module Integration

### For New Java Modules
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

### Common Dependencies Available
- **JPA/Hibernate**: For database access
- **Jersey**: For REST API development  
- **Jackson**: For JSON serialization
- **Commons**: Utilities and collections
- **Testing Stack**: JUnit, JBehave, Selenium, Mockito

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

## Development Workflow

### Adding Dependencies
1. **Check Managed Versions**: See if dependency is already managed in parent
2. **Add to dependencyManagement**: For new dependencies used by multiple modules
3. **Update Properties**: For version properties shared across modules
4. **Document Changes**: Update relevant documentation

### Version Management
- **SNAPSHOT Versions**: Used during development
- **Version Properties**: Centralized in parent POM
- **Dependency Alignment**: Ensure compatible versions across modules

## Important Notes

### Java Version Strategy
- **Current**: Java 8 (legacy compatibility)
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

## Module Dependencies

This parent POM is referenced by:
- **common**: Shared utilities and models
- **test_utils**: Testing utilities
- Java-based endpoint modules
- Java-based service modules

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