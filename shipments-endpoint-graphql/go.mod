module github.com/erpmicroservices/shipments-endpoint-graphql

go 1.23.0

toolchain go1.23.5

require (
	github.com/99designs/gqlgen v0.17.78
	github.com/go-chi/chi/v5 v5.0.12
	github.com/go-chi/cors v1.2.1
	github.com/golang-migrate/migrate/v4 v4.17.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/mux v1.8.1
	github.com/jmoiron/sqlx v1.3.5
	github.com/lib/pq v1.10.9
	github.com/rs/zerolog v1.32.0
	github.com/shopspring/decimal v1.4.0
	github.com/spf13/viper v1.18.2
	github.com/stretchr/testify v1.10.0
	github.com/vektah/gqlparser/v2 v2.5.30
	golang.org/x/sync v0.16.0
)

require (
	github.com/erpmicroservices/auth-go v0.0.0
	github.com/erpmicroservices/common-go v0.0.0
)

require (
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	github.com/urfave/cli/v2 v2.27.7 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	golang.org/x/mod v0.26.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/tools v0.35.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Local development - replace with actual module paths in production
replace github.com/erpmicroservices/common-go => ../common-go

replace github.com/erpmicroservices/auth-go => ../auth-go