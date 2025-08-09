package fixtures

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/rs/zerolog/log"
)

// MockService represents a mock GraphQL service for testing
type MockService struct {
	Name     string
	Schema   string
	Server   *httptest.Server
	Handler  *MockServiceHandler
}

// MockServiceHandler handles GraphQL requests for mock services
type MockServiceHandler struct {
	name      string
	schema    string
	responses map[string]interface{}
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   interface{} `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string `json:"message"`
	Path    []interface{} `json:"path,omitempty"`
}

// NewMockService creates a new mock GraphQL service
func NewMockService(name, schema string) *MockService {
	handler := &MockServiceHandler{
		name:      name,
		schema:    schema,
		responses: make(map[string]interface{}),
	}

	server := httptest.NewServer(http.HandlerFunc(handler.ServeHTTP))

	return &MockService{
		Name:    name,
		Schema:  schema,
		Server:  server,
		Handler: handler,
	}
}

// Close closes the mock service
func (ms *MockService) Close() {
	ms.Server.Close()
}

// URL returns the mock service URL
func (ms *MockService) URL() string {
	return ms.Server.URL
}

// AddResponse adds a mock response for a specific query
func (ms *MockService) AddResponse(query string, response interface{}) {
	ms.Handler.responses[query] = response
}

// ServeHTTP handles HTTP requests to the mock service
func (h *MockServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/graphql" && r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Handle introspection query
	if r.URL.Query().Get("query") != "" {
		h.handleIntrospection(w, r)
		return
	}

	// Parse GraphQL request
	var req GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Failed to decode GraphQL request")
		h.sendError(w, "Invalid request format")
		return
	}

	// Handle schema request
	if req.Query == "{ __schema { queryType { name } } }" {
		h.handleSchemaRequest(w, r)
		return
	}

	// Look for mock response
	if response, exists := h.responses[req.Query]; exists {
		h.sendResponse(w, GraphQLResponse{Data: response})
		return
	}

	// Default response
	h.sendResponse(w, GraphQLResponse{
		Data: map[string]interface{}{
			"service": h.name,
			"status":  "mock_response",
		},
	})
}

// handleIntrospection handles GraphQL introspection queries
func (h *MockServiceHandler) handleIntrospection(w http.ResponseWriter, r *http.Request) {
	response := GraphQLResponse{
		Data: map[string]interface{}{
			"__schema": map[string]interface{}{
				"queryType": map[string]interface{}{
					"name": "Query",
				},
			},
		},
	}
	h.sendResponse(w, response)
}

// handleSchemaRequest handles schema requests
func (h *MockServiceHandler) handleSchemaRequest(w http.ResponseWriter, r *http.Request) {
	response := GraphQLResponse{
		Data: map[string]interface{}{
			"__schema": map[string]interface{}{
				"queryType": map[string]interface{}{
					"name": "Query",
				},
			},
		},
	}
	h.sendResponse(w, response)
}

// sendResponse sends a JSON response
func (h *MockServiceHandler) sendResponse(w http.ResponseWriter, response GraphQLResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// sendError sends an error response
func (h *MockServiceHandler) sendError(w http.ResponseWriter, message string) {
	response := GraphQLResponse{
		Errors: []GraphQLError{
			{Message: message},
		},
	}
	h.sendResponse(w, response)
}

// CreateMockServices creates a set of mock services for testing
func CreateMockServices() map[string]*MockService {
	services := make(map[string]*MockService)

	// People & Organizations service
	peopleSchema := `
		type Query {
			person(id: ID!): Person
			people: [Person!]!
		}
		
		type Person @key(fields: "id") {
			id: ID!
			name: String!
			email: String!
		}
	`
	peopleService := NewMockService("people-organizations", peopleSchema)
	peopleService.AddResponse(`query { people { id name email } }`, map[string]interface{}{
		"people": []map[string]interface{}{
			{"id": "1", "name": "John Doe", "email": "john@example.com"},
			{"id": "2", "name": "Jane Smith", "email": "jane@example.com"},
		},
	})
	services["people-organizations"] = peopleService

	// E-Commerce service
	ecommerceSchema := `
		type Query {
			orders(personId: ID!): [Order!]!
		}
		
		type Order @key(fields: "id") {
			id: ID!
			personId: ID!
			status: String!
			total: Float!
		}
		
		extend type Person @key(fields: "id") {
			id: ID! @external
			orders: [Order!]!
		}
	`
	ecommerceService := NewMockService("e-commerce", ecommerceSchema)
	ecommerceService.AddResponse(`query { orders(personId: "1") { id status total } }`, map[string]interface{}{
		"orders": []map[string]interface{}{
			{"id": "order-1", "status": "completed", "total": 99.99},
			{"id": "order-2", "status": "pending", "total": 149.99},
		},
	})
	services["e-commerce"] = ecommerceService

	// Products service
	productsSchema := `
		type Query {
			product(id: ID!): Product
		}
		
		type Product @key(fields: "id") {
			id: ID!
			name: String!
			price: Float!
		}
	`
	productsService := NewMockService("products", productsSchema)
	productsService.AddResponse(`query { product(id: "1") { id name price } }`, map[string]interface{}{
		"product": map[string]interface{}{
			"id": "1", "name": "Example Product", "price": 29.99,
		},
	})
	services["products"] = productsService

	return services
}

// CleanupMockServices closes all mock services
func CleanupMockServices(services map[string]*MockService) {
	for _, service := range services {
		service.Close()
	}
}