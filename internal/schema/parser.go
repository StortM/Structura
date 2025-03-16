// internal/schema/parser.go
package schema

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Parser interface for different database schema formats
type Parser interface {
	Parse(input string) (*Schema, error)
}

// SQLParser implements the Parser interface for SQL
type SQLParser struct{}

// Parse parses an SQL schema definition and returns a Schema
func (p *SQLParser) Parse(input string) (*Schema, error) {
	// Create a new schema with a unique ID
	schema := &Schema{
		ID:        uuid.New().String(),
		Name:      "Imported Schema",
		Tables:    make([]Table, 0),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	// Extract CREATE TABLE statements
	tableRegex := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+[` + "`" + `]?(\w+)[` + "`" + `]?\s*\(([\s\S]*?)\);`)
	tableMatches := tableRegex.FindAllStringSubmatch(input, -1)

	// Process each table
	for _, match := range tableMatches {
		if len(match) < 3 {
			continue
		}

		tableName := match[1]
		columnsText := match[2]

		table := Table{
			ID:       uuid.New().String(),
			Name:     tableName,
			Columns:  make([]Column, 0),
			Position: Position{X: 0, Y: 0},        // Initial position, will be updated by layout algorithm
			Size:     Size{Width: 200, Height: 0}, // Width fixed, height will depend on number of columns
		}

		// Extract column definitions
		columnLines := strings.Split(columnsText, ",")
		primaryKeys := make([]string, 0)

		// First pass - extract PRIMARY KEY constraint if defined separately
		for _, line := range columnLines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(strings.ToUpper(line), "PRIMARY KEY") {
				pkRegex := regexp.MustCompile(`PRIMARY\s+KEY\s*\(([^)]+)\)`)
				pkMatch := pkRegex.FindStringSubmatch(line)
				if len(pkMatch) > 1 {
					keys := strings.Split(pkMatch[1], ",")
					for _, key := range keys {
						primaryKeys = append(primaryKeys, strings.Trim(strings.TrimSpace(key), "`"))
					}
				}
			}
		}

		// Second pass - extract columns
		for _, line := range columnLines {
			line = strings.TrimSpace(line)

			// Skip constraints and keys for now
			if strings.HasPrefix(strings.ToUpper(line), "PRIMARY KEY") ||
				strings.HasPrefix(strings.ToUpper(line), "FOREIGN KEY") ||
				strings.HasPrefix(strings.ToUpper(line), "CONSTRAINT") ||
				strings.HasPrefix(strings.ToUpper(line), "INDEX") ||
				strings.HasPrefix(strings.ToUpper(line), "UNIQUE") {
				continue
			}

			// Extract column name and type
			columnRegex := regexp.MustCompile(`[` + "`" + `]?(\w+)[` + "`" + `]?\s+(\w+(\(\d+\))?)`)
			columnMatch := columnRegex.FindStringSubmatch(line)
			if len(columnMatch) < 3 {
				continue
			}

			columnName := columnMatch[1]
			dataType := columnMatch[2]

			// Check if column is part of the primary key
			isPrimaryKey := false
			for _, pk := range primaryKeys {
				if pk == columnName {
					isPrimaryKey = true
					break
				}
			}

			// Check if PRIMARY KEY is defined inline
			isPrimaryKey = isPrimaryKey || strings.Contains(strings.ToUpper(line), "PRIMARY KEY")

			// Check if column is nullable
			isNullable := !strings.Contains(strings.ToUpper(line), "NOT NULL")

			// Check if auto increment
			autoIncrement := strings.Contains(strings.ToUpper(line), "AUTO_INCREMENT")

			// Create the column
			column := Column{
				Name:          columnName,
				DataType:      dataType,
				IsPrimaryKey:  isPrimaryKey,
				IsNullable:    isNullable,
				AutoIncrement: autoIncrement,
			}

			table.Columns = append(table.Columns, column)
		}

		// Calculate table height based on number of columns
		table.Size.Height = float64(30 + len(table.Columns)*25) // Header + rows

		schema.Tables = append(schema.Tables, table)
	}

	// Extract foreign key relationships
	fkRegex := regexp.MustCompile(`(?i)FOREIGN\s+KEY\s*\(([^)]+)\)\s+REFERENCES\s+[` + "`" + `]?(\w+)[` + "`" + `]?\s*\(([^)]+)\)`)
	for _, match := range tableRegex.FindAllStringSubmatch(input, -1) {
		if len(match) < 3 {
			continue
		}

		tableName := match[1]
		tableBody := match[2]

		fkMatches := fkRegex.FindAllStringSubmatch(tableBody, -1)
		for _, fkMatch := range fkMatches {
			if len(fkMatch) < 4 {
				continue
			}

			sourceColumn := strings.Trim(strings.TrimSpace(fkMatch[1]), "`")
			targetTable := fkMatch[2]
			targetColumn := strings.Trim(strings.TrimSpace(fkMatch[3]), "`")

			// Find the source table ID
			var sourceTableID string
			for _, table := range schema.Tables {
				if table.Name == tableName {
					sourceTableID = table.ID
					break
				}
			}

			// Find the target table ID
			var targetTableID string
			for _, table := range schema.Tables {
				if table.Name == targetTable {
					targetTableID = table.ID
					break
				}
			}

			// Mark the column as a foreign key
			for i, table := range schema.Tables {
				if table.Name == tableName {
					for j, col := range table.Columns {
						if col.Name == sourceColumn {
							schema.Tables[i].Columns[j].IsForeignKey = true
						}
					}
				}
			}

			// Add the relationship
			relationship := Relationship{
				ID:           uuid.New().String(),
				SourceTable:  sourceTableID,
				TargetTable:  targetTableID,
				SourceColumn: sourceColumn,
				TargetColumn: targetColumn,
				RelType:      "one-to-many", // Default relationship type
			}

			schema.Relationships = append(schema.Relationships, relationship)
		}
	}

	return schema, nil
}
