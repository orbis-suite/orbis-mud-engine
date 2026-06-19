package rooms

import (
	"math"
	"math/rand"
)

// perlin generates 2D Perlin noise using a seeded permutation table.
type perlin struct {
	perm [512]int
}

func newPerlin(seed int64) *perlin {
	p := &perlin{}
	r := rand.New(rand.NewSource(seed))
	src := r.Perm(256)
	for i := 0; i < 256; i++ {
		p.perm[i] = src[i]
		p.perm[i+256] = src[i]
	}
	return p
}

func (p *perlin) fade(t float64) float64 {
	return t * t * t * (t*(t*6-15) + 10)
}

func (p *perlin) lerp(t, a, b float64) float64 {
	return a + t*(b-a)
}

func (p *perlin) grad(hash int, x, y float64) float64 {
	switch hash & 3 {
	case 0:
		return x + y
	case 1:
		return -x + y
	case 2:
		return x - y
	default:
		return -x - y
	}
}

func (p *perlin) noise(x, y float64) float64 {
	xi := int(math.Floor(x)) & 255
	yi := int(math.Floor(y)) & 255
	xf := x - math.Floor(x)
	yf := y - math.Floor(y)
	u := p.fade(xf)
	v := p.fade(yf)

	aa := p.perm[p.perm[xi]+yi]
	ab := p.perm[p.perm[xi]+yi+1]
	ba := p.perm[p.perm[xi+1]+yi]
	bb := p.perm[p.perm[xi+1]+yi+1]

	x1 := p.lerp(u, p.grad(aa, xf, yf), p.grad(ba, xf-1, yf))
	x2 := p.lerp(u, p.grad(ab, xf, yf-1), p.grad(bb, xf-1, yf-1))
	return p.lerp(v, x1, x2)
}

// octaveNoise layers multiple noise passes for more natural-looking terrain.
func (p *perlin) octaveNoise(x, y float64, octaves int, persistence float64) float64 {
	total := 0.0
	frequency := 1.0
	amplitude := 1.0
	maxValue := 0.0
	for i := 0; i < octaves; i++ {
		total += p.noise(x*frequency, y*frequency) * amplitude
		maxValue += amplitude
		amplitude *= persistence
		frequency *= 2
	}
	return total / maxValue
}
