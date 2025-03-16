// internal/api/handlers.go
package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/StortM/Structura/internal/layout"
	"github.com/StortM/Structura/internal/schema"
)

// SchemaStore is a simple in-memory store for schemas (in a real app, you'd use a database)
var SchemaStore = make(map[string]*schema.Schema)

// GetHomePage handles the root route and serves the main HTML page
func GetHomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/templates/index.html")
}

// ImportSchema handles SQL schema import
func ImportSchema(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse request as JSON
	var request struct {
		SQL string `json:"sql"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	// Parse the SQL
	parser := &schema.SQLParser{}
	newSchema, err := parser.Parse(request.SQL)
	if err != nil {
		http.Error(w, "Error parsing SQL: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Apply layout algorithm
	layoutAlgo := layout.NewForceDirectedLayout()
	if err := layoutAlgo.ApplyLayout(newSchema); err != nil {
		http.Error(w, "Error applying layout: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Store the schema
	SchemaStore[newSchema.ID] = newSchema

	// Return the schema
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newSchema)
}

// ApplyLayout applies a layout algorithm to a schema
func ApplyLayout(w http.ResponseWriter, r *http.Request) {
	// Get schema ID from URL parameters
	vars := mux.Vars(r)
	schemaID := vars["id"]

	// Get layout type from request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var request struct {
		LayoutType string `json:"layoutType"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	// Find the schema
	schema, exists := SchemaStore[schemaID]
	if !exists {
		http.Error(w, "Schema not found", http.StatusNotFound)
		return
	}

	// Apply the selected layout algorithm
	var layoutAlgo layout.LayoutAlgorithm
	switch request.LayoutType {
	case "force":
		layoutAlgo = layout.NewForceDirectedLayout()
	case "grid":
		layoutAlgo = layout.NewGridLayout()
	default:
		layoutAlgo = layout.NewForceDirectedLayout()
	}

	if err := layoutAlgo.ApplyLayout(schema); err != nil {
		http.Error(w, "Error applying layout: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update timestamp
	schema.UpdatedAt = time.Now().Unix()

	// Return the updated schema
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

// UpdateTablePosition updates the position of a table
func UpdateTablePosition(w http.ResponseWriter, r *http.Request) {
	// Get schema ID and table ID from URL parameters
	vars := mux.Vars(r)
	schemaID := vars["id"]
	tableID := vars["tableId"]

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse request as JSON
	var request struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	// Find the schema
	schema, exists := SchemaStore[schemaID]
	if !exists {
		http.Error(w, "Schema not found", http.StatusNotFound)
		return
	}

	// Find and update the table
	found := false
	for i := range schema.Tables {
		if schema.Tables[i].ID == tableID {
			schema.Tables[i].Position.X = request.X
			schema.Tables[i].Position.Y = request.Y
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Table not found", http.StatusNotFound)
		return
	}

	// Update relationship control points if necessary
	// (In a real implementation, you would recalculate the relationship lines)

	// Update timestamp
	schema.UpdatedAt = time.Now().Unix()

	// Return success
	w.WriteHeader(http.StatusOK)
}

// GetSchema returns a specific schema
func GetSchema(w http.ResponseWriter, r *http.Request) {
	// Get schema ID from URL parameters
	vars := mux.Vars(r)
	schemaID := vars["id"]

	// Find the schema
	schema, exists := SchemaStore[schemaID]
	if !exists {
		http.Error(w, "Schema not found", http.StatusNotFound)
		return
	}

	// Return the schema
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

// ListSchemas returns all schemas
func ListSchemas(w http.ResponseWriter, r *http.Request) {
	// Extract basic info for each schema
	schemas := make([]struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		CreatedAt int64  `json:"createdAt"`
		UpdatedAt int64  `json:"updatedAt"`
	}, 0, len(SchemaStore))

	for _, s := range SchemaStore {
		schemas = append(schemas, struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			CreatedAt int64  `json:"createdAt"`
			UpdatedAt int64  `json:"updatedAt"`
		}{
			ID:        s.ID,
			Name:      s.Name,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		})
	}

	// Return the list
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schemas)
}

// DeleteSchema deletes a schema
func DeleteSchema(w http.ResponseWriter, r *http.Request) {
	// Get schema ID from URL parameters
	vars := mux.Vars(r)
	schemaID := vars["id"]

	// Check if schema exists
	if _, exists := SchemaStore[schemaID]; !exists {
		http.Error(w, "Schema not found", http.StatusNotFound)
		return
	}

	// Delete the schema
	delete(SchemaStore, schemaID)

	// Return success
	w.WriteHeader(http.StatusOK)
}