package rooms

import (
	"fmt"

	"example.com/mud/sdk"
)

const worldSize = 10

// --- Height levels ---

type heightLevel int

const (
	heightWater heightLevel = iota // no room generated
	heightLow
	heightMid
	heightHigh
)

// heightThresholds controls how Perlin values map to height levels.
// Perlin output roughly ranges [-0.7, 0.7] with 4 octaves.
var heightThresholds = struct{ water, low, mid float64 }{
	water: -0.10, // below → water
	low:   0.10,  // below → low
	mid:   0.35,  // below → mid, else → high
}

// --- Moisture levels ---

type moistureLevel int

const (
	moistureLow moistureLevel = iota
	moistureHigh
)

// moistureThreshold controls the low/high moisture split.
var moistureThreshold = 0.0

// --- Biomes ---

type biomeKey struct {
	height   heightLevel
	moisture moistureLevel
}

// BiomeDef holds display properties for a biome. Edit names/descriptions here.
type BiomeDef struct {
	Name        string
	Description string
	Icon        string
	Color       string
}

// biomes is the single source of truth for biome properties.
// Change Name/Description here to rename or retune any biome.
var biomes = map[biomeKey]BiomeDef{
	{heightLow, moistureLow}: {
		Name:        "Coast",
		Description: "Sandy shores where land meets the sea. Gulls cry overhead.",
		Icon:        "~",
		Color:       "yellow",
	},
	{heightLow, moistureHigh}: {
		Name:        "Swamp",
		Description: "Murky wetlands choked with moss and drifting fog.",
		Icon:        "S",
		Color:       "green",
	},
	{heightMid, moistureLow}: {
		Name:        "Plains",
		Description: "Open grasslands swept by dry winds under a wide sky.",
		Icon:        ".",
		Color:       "yellow",
	},
	{heightMid, moistureHigh}: {
		Name:        "Forest",
		Description: "Dense woodland rich with birdsong and deep shadow.",
		Icon:        "F",
		Color:       "green",
	},
	{heightHigh, moistureLow}: {
		Name:        "Highland",
		Description: "Rocky ridges and windswept moors, sparse and cold.",
		Icon:        "^",
		Color:       "white",
	},
	{heightHigh, moistureHigh}: {
		Name:        "Mountain",
		Description: "Mist-shrouded peaks where snow meets cloud and silence reigns.",
		Icon:        "M",
		Color:       "cyan",
	},
}

// --- Noise seeds ---

const (
	heightSeed   int64 = 45
	moistureSeed int64 = 140
)

// --- Classification helpers ---

func classifyHeight(v float64) heightLevel {
	switch {
	case v < heightThresholds.water:
		return heightWater
	case v < heightThresholds.low:
		return heightLow
	case v < heightThresholds.mid:
		return heightMid
	default:
		return heightHigh
	}
}

func classifyMoisture(v float64) moistureLevel {
	if v < moistureThreshold {
		return moistureLow
	}
	return moistureHigh
}

func worldID(x, y int) string {
	return fmt.Sprintf("world_%d_%d", x, y)
}

// --- World generation ---

type worldCell struct {
	h heightLevel
	m moistureLevel
}

var cardinalDirs = []struct {
	name   string
	dx, dy int
}{
	{"north", 0, -1},
	{"south", 0, 1},
	{"east", 1, 0},
	{"west", -1, 0},
}

// GenerateWorld builds a worldSize×worldSize grid of rooms using Perlin noise
// for height and moisture, returning one RoomDef per non-water cell.
func GenerateWorld() []*sdk.RoomDef {
	heightNoise := newPerlin(heightSeed)
	moistureNoise := newPerlin(moistureSeed)

	// Scale controls how zoomed-in the noise is (larger = smoother terrain).
	const scale = 20.0

	grid := [worldSize][worldSize]worldCell{}
	for y := 0; y < worldSize; y++ {
		for x := 0; x < worldSize; x++ {
			hv := heightNoise.octaveNoise(float64(x)/scale, float64(y)/scale, 4, 0.5)
			mv := moistureNoise.octaveNoise(float64(x)/scale, float64(y)/scale, 4, 0.5)
			grid[y][x] = worldCell{classifyHeight(hv), classifyMoisture(mv)}
		}
	}

	var rooms []*sdk.RoomDef
	for y := 0; y < worldSize; y++ {
		for x := 0; x < worldSize; x++ {
			c := grid[y][x]
			if c.h == heightWater {
				continue
			}
			biome := biomes[biomeKey{c.h, c.m}]
			exits := map[string]string{}
			for _, d := range cardinalDirs {
				nx, ny := x+d.dx, y+d.dy
				if nx < 0 || nx >= worldSize || ny < 0 || ny >= worldSize {
					continue
				}
				if grid[ny][nx].h != heightWater {
					exits[d.name] = worldID(nx, ny)
				}
			}
			rooms = append(rooms, &sdk.RoomDef{
				ID:          worldID(x, y),
				Name:        biome.Name,
				Description: biome.Description,
				Icon:        biome.Icon,
				Color:       biome.Color,
				Exits:       exits,
			})
		}
	}
	return rooms
}
