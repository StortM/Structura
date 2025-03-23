package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/StortM/Structura/internal/layout"
	"github.com/StortM/Structura/internal/schema"
	"github.com/gorilla/mux"
)

var SchemaStore = make(map[string]*schema.Schema)

func GetHomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/templates/index.html")
}

func ImportSchema(w http.ResponseWriter, r *http.Request) {
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

	SchemaStore[newSchema.ID] = newSchema

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newSchema)
}

func ApplyLayout(w http.ResponseWriter, r *http.Request) {
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

	schema, exists := SchemaStore[schemaID]
	if !exists {
		http.Error(w, "Schema not found", http.StatusNotFound)
		return
	}

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

	schema.UpdatedAt = time.Now().Unix()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

func UpdateTablePosition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	schemaID := vars["id"]
	tableID := vars["tableId"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var request struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

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

	// TODO - recalculate relationship lines

	schema.UpdatedAt = time.Now().Unix()

	w.WriteHeader(http.StatusOK)
}

func GetSchema(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	schemaID := vars["id"]

	schema, exists := SchemaStore[schemaID]
	if !exists {
		http.Error(w, "Schema not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

func ListSchemas(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schemas)
}

func DeleteSchema(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	schemaID := vars["id"]

	if _, exists := SchemaStore[schemaID]; !exists {
		http.Error(w, "Schema not found", http.StatusNotFound)
		return
	}

	delete(SchemaStore, schemaID)

	w.WriteHeader(http.StatusOK)
}