module github.com/erpmicroservices/graphql-federation-gateway

go 1.21

require (
	github.com/cucumber/godog v0.14.0
	github.com/go-chi/chi/v5 v5.0.12
	github.com/rs/cors v1.10.1
	github.com/rs/zerolog v1.32.0
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/cucumber/gherkin/go/v26 v26.2.0 // indirect
	github.com/cucumber/messages/go/v21 v21.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gofrs/uuid v4.3.1+incompatible // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-memdb v1.3.4 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.18.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Replace directives for local development
replace github.com/erpmicroservices/auth-go => ../auth-go

replace github.com/erpmicroservices/common-go => ../common-go
