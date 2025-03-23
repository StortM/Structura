package schema

type Table struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Columns   []Column    `json:"columns"`
	Position  Position    `json:"position"`
	Size      Size        `json:"size"`
	TableType string      `json:"tableType"`
}

type Column struct {
	Name          string `json:"name"`
	DataType      string `json:"dataType"`
	IsPrimaryKey  bool   `json:"isPrimaryKey"`
	IsForeignKey  bool   `json:"isForeignKey"`
	IsNullable    bool   `json:"isNullable"`
	DefaultValue  string `json:"defaultValue,omitempty"`
	AutoIncrement bool   `json:"autoIncrement,omitempty"`
}

type Relationship struct {
	ID           string     `json:"id"`
	SourceTable  string     `json:"sourceTable"` 
	TargetTable  string     `json:"targetTable"` 
	SourceColumn string     `json:"sourceColumn"`
	TargetColumn string     `json:"targetColumn"`
	RelType      string     `json:"relType"`     
	Points       []Position `json:"points,omitempty"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type Schema struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Tables        []Table        `json:"tables"`
	Relationships []Relationship `json:"relationships"`
	CreatedAt     int64          `json:"createdAt"` 
	UpdatedAt     int64          `json:"updatedAt"` 
}