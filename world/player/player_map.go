package player

import (
	"fmt"

	"example.com/mud/world/entities"
	"example.com/mud/world/entities/components"
	"example.com/mud/world/response"
)

// sgrToHex maps the SGR color names used by room entities to web hex codes.
var sgrToHex = map[string]string{
	"black":   "#1c1a24",
	"red":     "#ff5555",
	"green":   "#55ff55",
	"yellow":  "#ffff55",
	"blue":    "#5555ff",
	"magenta": "#ff55ff",
	"cyan":    "#55ffff",
	"white":   "#ffffff",
}

func (p *Player) MapCommand() (response.Response, error) {
	// We don't support the "map" command directly anymore, but may want to eventually.
	return response.Text{Value: "You gaze upon your surroundings and see what's in the map panel."}, nil
}

func (p *Player) Map() (response.MapView, error) {
	coordByRoom, err := assignCoordinates(p.CurrentRoom, p.world, 5)
	if err != nil {
		return response.MapView{}, fmt.Errorf("map: assign coordinates: %w", err)
	}

	currentRoom, err := entities.RequireComponent[*components.Room](p.CurrentRoom)
	if err != nil {
		return response.MapView{}, fmt.Errorf("cannot map non-room area: %w", err)
	}

	grid, playerX, playerY, err := p.renderMap(coordByRoom, currentRoom, p.world)
	if err != nil {
		return response.MapView{}, fmt.Errorf("map: render map: %w", err)
	}

	return response.MapView{Grid: grid, PlayerX: playerX, PlayerY: playerY}, nil
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

func (p *Player) renderMap(coordByRoom map[*components.Room]coord, currentRoom *components.Room, world World) ([][]response.MapCell, int, int, error) {
	if len(coordByRoom) == 0 {
		return nil, 0, 0, nil
	}

	minX, maxX, minY, maxY := 0, 0, 0, 0
	for _, c := range coordByRoom {
		if c.X < minX {
			minX = c.X
		}
		if c.X > maxX {
			maxX = c.X
		}
		if c.Y < minY {
			minY = c.Y
		}
		if c.Y > maxY {
			maxY = c.Y
		}
	}

	width := (maxX-minX)*2 + 1
	height := (maxY-minY)*2 + 1
	grid := make([][]response.MapCell, height)
	for i := range grid {
		grid[i] = make([]response.MapCell, width)
		for j := range grid[i] {
			grid[i][j] = response.MapCell{Color: "", Icon: " "}
		}
	}

	var playerX, playerY int

	for r, c := range coordByRoom {
		gx := (c.X - minX) * 2
		gy := (c.Y - minY) * 2

		if r == currentRoom {
			grid[gy][gx] = response.MapCell{Color: "#ff5555", Icon: "@"}
			playerX, playerY = gx, gy
		} else if p.trackingAlias != "" && len(r.GetChildren().GetChildrenByAlias(p.trackingAlias)) > 0 {
			grid[gy][gx] = response.MapCell{Color: "#ffff55", Icon: "!"}
		} else {
			color := sgrToHex[r.MapColor]
			if color == "" {
				color = "#c8c4d0"
			}
			grid[gy][gx] = response.MapCell{Color: color, Icon: r.MapIcon}
		}

		for _, roomId := range r.Exits {
			roomEntity, ok := world.GetEntityById(roomId)
			if !ok {
				return nil, 0, 0, fmt.Errorf("entity with id '%s' does not exist", roomId)
			}

			room, err := entities.RequireComponent[*components.Room](roomEntity)
			if err != nil {
				return nil, 0, 0, fmt.Errorf("render map: %w", err)
			}

			npos, ok := coordByRoom[room]
			if !ok {
				continue
			}
			dx := npos.X - c.X
			dy := npos.Y - c.Y
			if dx == 1 {
				grid[gy][gx+1] = response.MapCell{Color: "#6b5f80", Icon: "-"}
			} else if dx == -1 {
				grid[gy][gx-1] = response.MapCell{Color: "#6b5f80", Icon: "-"}
			} else if dy == 1 {
				grid[gy+1][gx] = response.MapCell{Color: "#6b5f80", Icon: "|"}
			} else if dy == -1 {
				grid[gy-1][gx] = response.MapCell{Color: "#6b5f80", Icon: "|"}
			}
		}
	}

	return grid, playerX, playerY, nil
}

