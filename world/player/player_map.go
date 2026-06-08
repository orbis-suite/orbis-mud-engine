package player

import (
	"fmt"
	"strings"

	"example.com/mud/models"
	"example.com/mud/world/entities"
	"example.com/mud/world/entities/components"
)

func (p *Player) Map() (string, error) {
	coordByRoom, err := assignCoordinates(p.CurrentRoom, p.world, 5)
	if err != nil {
		return "", fmt.Errorf("map: assign coordinates: %w", err)
	}

	currentRoom, err := entities.RequireComponent[*components.Room](p.CurrentRoom)
	if err != nil {
		return "", fmt.Errorf("cannot map non-room area: %w", err)
	}

	ascii, err := p.renderMap(coordByRoom, currentRoom, p.world)
	if err != nil {
		return "", fmt.Errorf("map: render map: %w", err)
	}

	return ascii, nil
}

type coord struct{ X, Y int }

var cardinalDelta = map[string]coord{
	"north": {0, -1},
	"south": {0, 1},
	"east":  {1, 0},
	"west":  {-1, 0},
}

// fixed-order so random ranging over map doesn't influence mapping
var dirOrder = []string{"north", "east", "south", "west"}

func assignCoordinates(start *entities.Entity, world World, maxDepth int) (map[*components.Room]coord, error) {
	coordByRoom := make(map[*components.Room]coord)
	roomAtCoord := make(map[coord]*components.Room)
	seen := make(map[*components.Room]bool)

	type item struct {
		e     *entities.Entity
		x, y  int
		depth int
	}

	// queue with index-based pop (no O(n) slice shifting)
	queue := []item{{e: start, x: 0, y: 0, depth: 0}}
	head := 0

	for head < len(queue) {
		it := queue[head]
		head++

		if it.depth > maxDepth {
			continue
		}

		r, err := entities.RequireComponent[*components.Room](it.e)
		if err != nil {
			return nil, fmt.Errorf("cannot map non-room: %w", err)
		}
		if seen[r] {
			continue
		}

		c := coord{it.x, it.y}
		if existing, ok := roomAtCoord[c]; ok && existing != r {
			return nil, fmt.Errorf("mapping non-euclidean space at %v", c)
		}

		seen[r] = true
		coordByRoom[r] = c
		roomAtCoord[c] = r

		// Expand neighbors in deterministic NESW order
		exits := r.Exits
		for _, dir := range dirOrder {
			roomID, ok := exits[dir]
			if !ok {
				continue
			}
			delta := cardinalDelta[dir]

			nextEntity, ok := world.GetEntityById(roomID)
			if !ok {
				return nil, fmt.Errorf("entity with id %q does not exist", roomID)
			}

			// Optional early type check (helps avoid enqueuing non-rooms)
			nr, err := entities.RequireComponent[*components.Room](nextEntity)
			if err != nil {
				return nil, fmt.Errorf("cannot map non-room: %w", err)
			}
			if seen[nr] {
				continue
			}

			queue = append(queue, item{
				e:     nextEntity,
				x:     it.x + delta.X,
				y:     it.y + delta.Y,
				depth: it.depth + 1,
			})
		}
	}

	return coordByRoom, nil
}

func (p *Player) renderMap(coordByRoom map[*components.Room]coord, currentRoom *components.Room, world World) (string, error) {
	if len(coordByRoom) == 0 {
		return "", nil
	}

	minX, maxX, minY, maxY := 0, 0, 0, 0
	for _, p := range coordByRoom {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	width := (maxX-minX)*2 + 1
	height := (maxY-minY)*2 + 1
	grid := make([][]string, height)
	for i := range grid {
		grid[i] = make([]string, width)
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}

	for r, c := range coordByRoom {
		gx := (c.X - minX) * 2
		gy := (c.Y - minY) * 2

		if r == currentRoom {
			grid[gy][gx] = fmt.Sprintf("%s%s%s", models.SGR["red"], "@", models.SGR["reset"])
		} else if len(r.GetChildren().GetChildrenByAlias(p.trackingAlias)) > 0 {
			grid[gy][gx] = fmt.Sprintf("%s%s%s", models.SGR["yellow"], "!", models.SGR["reset"])
		} else {
			color := models.SGR[r.MapColor]
			icon := r.MapIcon
			grid[gy][gx] = fmt.Sprintf("%s%s%s", color, icon, models.SGR["reset"])
		}

		for _, roomId := range r.Exits {
			roomEntity, ok := world.GetEntityById(roomId)
			if !ok {
				return "", fmt.Errorf("entity with id '%s' does not exist", roomId)
			}

			room, err := entities.RequireComponent[*components.Room](roomEntity)
			if err != nil {
				return "", fmt.Errorf("render map: %w", err)
			}

			npos, ok := coordByRoom[room]
			if !ok {
				continue
			}
			dx := npos.X - c.X
			dy := npos.Y - c.Y
			if dx == 1 {
				grid[gy][gx+1] = "-"
			} else if dx == -1 {
				grid[gy][gx-1] = "-"
			} else if dy == 1 {
				grid[gy+1][gx] = "|"
			} else if dy == -1 {
				grid[gy-1][gx] = "|"
			}
		}
	}

	var b strings.Builder
	b.Grow(height * (width + 1))
	for _, row := range grid {
		b.WriteString(strings.Join(row, ""))
		b.WriteByte('\n')
	}
	return b.String(), nil
}

func (p *Player) Track(alias string) (string, error) {
	p.trackingAlias = alias

	return fmt.Sprintf(`Rooms with "%s" will now appear as "!" on your map.`, alias), nil
}
