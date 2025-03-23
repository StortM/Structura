package schema

import (
	"regexp"
	"strings"
	"time"

	"slices"

	"github.com/google/uuid"
)

type Parser interface {
	Parse(input string) (*Schema, error)
}

type SQLParser struct{}

func (p *SQLParser) Parse(input string) (*Schema, error) {
	schema := &Schema{
		ID:        uuid.New().String(),
		Name:      "Imported Schema",
		Tables:    make([]Table, 0),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	tableRegex := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+[` + "`" + `]?(\w+)[` + "`" + `]?\s*\(([\s\S]*?)\);`)
	tableMatches := tableRegex.FindAllStringSubmatch(input, -1)

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
			Position: Position{X: 0, Y: 0},        
			Size:     Size{Width: 200, Height: 0}, 
		}

		columnLines := strings.Split(columnsText, ",")
		primaryKeys := make([]string, 0)

		// First pass - extract primary keys
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

			// TODO - handle constraints and indexes
			if strings.HasPrefix(strings.ToUpper(line), "PRIMARY KEY") ||
				strings.HasPrefix(strings.ToUpper(line), "FOREIGN KEY") ||
				strings.HasPrefix(strings.ToUpper(line), "CONSTRAINT") ||
				strings.HasPrefix(strings.ToUpper(line), "INDEX") ||
				strings.HasPrefix(strings.ToUpper(line), "UNIQUE") {
				continue
			}

			columnRegex := regexp.MustCompile(`[` + "`" + `]?(\w+)[` + "`" + `]?\s+(\w+(\(\d+\))?)`)
			columnMatch := columnRegex.FindStringSubmatch(line)
			if len(columnMatch) < 3 {
				continue
			}

			columnName := columnMatch[1]
			dataType := columnMatch[2]

			isPrimaryKey := slices.Contains(primaryKeys, columnName)

			isPrimaryKey = isPrimaryKey || strings.Contains(strings.ToUpper(line), "PRIMARY KEY")

			isNullable := !strings.Contains(strings.ToUpper(line), "NOT NULL")

			autoIncrement := strings.Contains(strings.ToUpper(line), "AUTO_INCREMENT")

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
		table.Size.Height = float64(30 + len(table.Columns)*25) // 30px for table name, 25px per column

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
				RelType:      "one-to-many", // TODO - infer relationship type
			}

			schema.Relationships = append(schema.Relationships, relationship)
		}
	}

	return schema, nil
}
