# Flyway Migration Documentation

## Summary
All database projects in the ERP microservices system have been successfully converted from Liquibase to Flyway.

## Changes Made

### 1. SQL File Naming Convention
All SQL files have been renamed to follow Flyway's naming convention:
- Pattern: `V_{prefix}{sequence}__description.sql`
- Prefix: First 4 letters of the project name
- Sequence: 3-digit number starting at 001

#### Naming Prefixes by Database:
- `acco` - accounting_and_budgeting-database
- `ecom` - e_commerce-database  
- `huma` - human_resources-database
- `invo` - invoice-database
- `orde` - order-database
- `peop` - people_and_organizations-database
- `prod` - products-database
- `ship` - shipments-database
- `work` - work_effort-database

### 2. Configuration Files Added

#### `.env.example`
Each database now has an `.env.example` file with database connection settings:
```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME={database_name}
DB_USER={database_name}
DB_PASSWORD={database_name}
```

#### `flyway.conf`
Standard Flyway configuration file added to each database project:
- Migration location: `filesystem:sql`
- Baseline on migrate enabled
- Schema: public
- Connection retries: 10

### 3. Package.json Updates

All package.json files have been updated with:
- New Flyway migration scripts:
  - `migrate` - Run database migrations
  - `migrate:info` - Show migration status
  - `migrate:validate` - Validate migrations
- Updated build scripts to work with Flyway
- Added dependencies:
  - `dotenv`: ^16.0.3
  - `node-flywaydb`: ^3.0.7

### 4. Dockerfile Updates

All Dockerfiles have been updated to:
- Use `postgres:latest` instead of older versions
- Copy migration files directly to `/docker-entrypoint-initdb.d/`
- PostgreSQL will execute files in alphabetical order on first run
- Added UTF-8 encoding configuration

### 5. Removed Files

The following Liquibase-related files have been removed from all projects:
- `database_change_log.yml`
- `databasechangelog.csv`
- `liquibase-3.5.3-bin/` directory

## Usage

### Local Development

1. Copy `.env.example` to `.env` and adjust settings as needed:
   ```bash
   cp .env.example .env
   ```

2. Run migrations:
   ```bash
   npm run migrate
   ```

3. Check migration status:
   ```bash
   npm run migrate:info
   ```

### Docker Deployment

1. Build the Docker image:
   ```bash
   npm run build:docker
   ```

2. Start the container:
   ```bash
   npm run start
   ```

The PostgreSQL container will automatically execute all migration files in order on first run.

### Creating New Migrations

When adding new migrations, follow the naming convention:
```
V_{prefix}{sequence}__description.sql
```

For example, for the accounting database:
- `V_acco004__add_audit_tables.sql`
- `V_acco005__add_indexes.sql`

## Benefits of Flyway over Liquibase

1. **Simpler configuration** - No XML/YAML changesets required
2. **SQL-first approach** - Direct SQL files without wrapper format
3. **Better Docker integration** - PostgreSQL can execute migrations directly
4. **Cleaner version control** - SQL files are easier to review and diff
5. **Environment variable support** - Better security for credentials

## Testing

All existing BDD tests remain unchanged and should continue to work with the new Flyway-based structure. The test commands remain the same:

```bash
npm test
```

## Notes

- All databases now use consistent naming and structure
- Migration files are executed in alphabetical order
- The `build:database` script concatenates all migrations for offline SQL generation
- Docker containers load migrations automatically on first run
- No changes required to application code or GraphQL endpoints