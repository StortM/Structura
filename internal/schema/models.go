// internal/schema/models.go
package schema

// Table represents a database table in the schema
type Table struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Columns   []Column    `json:"columns"`
	Position  Position    `json:"position"`
	Size      Size        `json:"size"`
	TableType string      `json:"tableType"` // e.g., "entity", "junction", etc.
}

// Column represents a column in a database table
type Column struct {
	Name          string `json:"name"`
	DataType      string `json:"dataType"`
	IsPrimaryKey  bool   `json:"isPrimaryKey"`
	IsForeignKey  bool   `json:"isForeignKey"`
	IsNullable    bool   `json:"isNullable"`
	DefaultValue  string `json:"defaultValue,omitempty"`
	AutoIncrement bool   `json:"autoIncrement,omitempty"`
}

// Relationship represents a relationship between two tables
type Relationship struct {
	ID           string     `json:"id"`
	SourceTable  string     `json:"sourceTable"`  // ID of source table
	TargetTable  string     `json:"targetTable"`  // ID of target table
	SourceColumn string     `json:"sourceColumn"` // Name of source column
	TargetColumn string     `json:"targetColumn"` // Name of target column
	RelType      string     `json:"relType"`      // one-to-one, one-to-many, many-to-many
	Points       []Position `json:"points,omitempty"` // Control points for rendering the relationship line
}

// Position represents x, y coordinates for positioning elements
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Size represents width and height dimensions
type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Schema represents the entire database schema
type Schema struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Tables        []Table        `json:"tables"`
	Relationships []Relationship `json:"relationships"`
	CreatedAt     int64          `json:"createdAt"` // Unix timestamp
	UpdatedAt     int64          `json:"updatedAt"` // Unix timestamp
}