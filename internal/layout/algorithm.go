package layout

import (
	"math"
	"math/rand"

	"github.com/StortM/Structura/internal/schema"
)

type LayoutAlgorithm interface {
	ApplyLayout(schema *schema.Schema) error
}

type ForceDirectedLayout struct {
	CanvasWidth  float64
	CanvasHeight float64
	Iterations   int
}

func NewForceDirectedLayout() *ForceDirectedLayout {
	return &ForceDirectedLayout{
		CanvasWidth:  1200,
		CanvasHeight: 800,
		Iterations:   100,
	}
}

func (l *ForceDirectedLayout) ApplyLayout(s *schema.Schema) error {
	for i := range s.Tables {
		s.Tables[i].Position.X = rand.Float64() * l.CanvasWidth
		s.Tables[i].Position.Y = rand.Float64() * l.CanvasHeight
	}

	k := math.Sqrt(l.CanvasWidth * l.CanvasHeight / float64(len(s.Tables))) // Optimal distance
	temperature := l.CanvasWidth / 10

	for range l.Iterations {
		// Calculate repulsive forces between all pairs of tables
		forces := make([]schema.Position, len(s.Tables))

		// Apply repulsive forces (tables repel each other)
		for i := range s.Tables {
			for j := range s.Tables {
				if i == j {
					continue
				}

				// Calculate distance vector between tables i and j
				dx := s.Tables[i].Position.X - s.Tables[j].Position.X
				dy := s.Tables[i].Position.Y - s.Tables[j].Position.Y

				// Avoid division by zero
				distance := math.Max(1.0, math.Sqrt(dx*dx+dy*dy))

				// Calculate repulsive force (inversely proportional to distance)
				force := k * k / distance

				// Apply force along the distance vector
				unitDx := dx / distance
				unitDy := dy / distance

				forces[i].X += unitDx * force
				forces[i].Y += unitDy * force
			}
		}

		// Apply attractive forces along relationships
		for _, rel := range s.Relationships {
			// Find source and target table indices
			var sourceIdx, targetIdx int = -1, -1
			for i, table := range s.Tables {
				if table.ID == rel.SourceTable {
					sourceIdx = i
				}
				if table.ID == rel.TargetTable {
					targetIdx = i
				}
			}

			if sourceIdx == -1 || targetIdx == -1 {
				continue
			}

			// Calculate distance vector between related tables
			dx := s.Tables[sourceIdx].Position.X - s.Tables[targetIdx].Position.X
			dy := s.Tables[sourceIdx].Position.Y - s.Tables[targetIdx].Position.Y

			distance := math.Max(1.0, math.Sqrt(dx*dx+dy*dy))

			// Calculate attractive force (proportional to distance)
			force := distance * distance / k

			// Apply force along the distance vector
			unitDx := dx / distance
			unitDy := dy / distance

			forces[sourceIdx].X -= unitDx * force
			forces[sourceIdx].Y -= unitDy * force
			forces[targetIdx].X += unitDx * force
			forces[targetIdx].Y += unitDy * force
		}

		// Apply the forces, limited by current temperature
		for i := range s.Tables {
			// Calculate the magnitude of the force
			forceMagnitude := math.Sqrt(forces[i].X*forces[i].X + forces[i].Y*forces[i].Y)

			// Limit the force by temperature
			limitedMagnitude := math.Min(forceMagnitude, temperature)

			if forceMagnitude > 0 {
				// Apply the limited force
				s.Tables[i].Position.X += forces[i].X / forceMagnitude * limitedMagnitude
				s.Tables[i].Position.Y += forces[i].Y / forceMagnitude * limitedMagnitude

				// Keep tables within canvas boundaries
				s.Tables[i].Position.X = math.Max(0, math.Min(l.CanvasWidth-s.Tables[i].Size.Width, s.Tables[i].Position.X))
				s.Tables[i].Position.Y = math.Max(0, math.Min(l.CanvasHeight-s.Tables[i].Size.Height, s.Tables[i].Position.Y))
			}
		}

		// Cool down the temperature
		temperature *= 0.95
	}

	// Calculate control points for relationship lines
	l.calculateRelationshipPoints(s)

	return nil
}

func (l *ForceDirectedLayout) calculateRelationshipPoints(s *schema.Schema) {
	for i, rel := range s.Relationships {
		// Find source and target tables
		var sourceTable, targetTable schema.Table

		for _, table := range s.Tables {
			if table.ID == rel.SourceTable {
				sourceTable = table
			}
			if table.ID == rel.TargetTable {
				targetTable = table
			}
		}

		startX := sourceTable.Position.X + sourceTable.Size.Width/2
		startY := sourceTable.Position.Y + sourceTable.Size.Height/2
		endX := targetTable.Position.X + targetTable.Size.Width/2
		endY := targetTable.Position.Y + targetTable.Size.Height/2

		// Create a simple straight line with start and end points
		s.Relationships[i].Points = []schema.Position{
			{X: startX, Y: startY},
			{X: endX, Y: endY},
		}
	}
}

type GridLayout struct {
	CanvasWidth  float64
	CanvasHeight float64
	Padding      float64
}

func NewGridLayout() *GridLayout {
	return &GridLayout{
		CanvasWidth:  1200,
		CanvasHeight: 800,
		Padding:      50,
	}
}

func (l *GridLayout) ApplyLayout(s *schema.Schema) error {
	if len(s.Tables) == 0 {
		return nil
	}

	// Determine grid dimensions based on number of tables
	numTables := len(s.Tables)
	cols := int(math.Ceil(math.Sqrt(float64(numTables))))
	rows := int(math.Ceil(float64(numTables) / float64(cols)))

	// Calculate cell size
	cellWidth := (l.CanvasWidth - l.Padding*2) / float64(cols)
	cellHeight := (l.CanvasHeight - l.Padding*2) / float64(rows)

	// Position tables in a grid
	for i, table := range s.Tables {
		col := i % cols
		row := i / cols

		// Center the table in its cell
		s.Tables[i].Position.X = l.Padding + float64(col)*cellWidth + (cellWidth-table.Size.Width)/2
		s.Tables[i].Position.Y = l.Padding + float64(row)*cellHeight + (cellHeight-table.Size.Height)/2
	}

	// Calculate control points for relationship lines
	for i, rel := range s.Relationships {
		// Find source and target tables
		var sourceTable, targetTable schema.Table

		for _, table := range s.Tables {
			if table.ID == rel.SourceTable {
				sourceTable = table
			}
			if table.ID == rel.TargetTable {
				targetTable = table
			}
		}

		// Calculate start and end points (center of tables)
		startX := sourceTable.Position.X + sourceTable.Size.Width/2
		startY := sourceTable.Position.Y + sourceTable.Size.Height/2
		endX := targetTable.Position.X + targetTable.Size.Width/2
		endY := targetTable.Position.Y + targetTable.Size.Height/2

		// Create a simple straight line with start and end points
		s.Relationships[i].Points = []schema.Position{
			{X: startX, Y: startY},
			{X: endX, Y: endY},
		}
	}

	return nil
}
