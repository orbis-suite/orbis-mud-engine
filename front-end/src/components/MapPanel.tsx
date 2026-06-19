import { useEffect, useRef } from 'react'
import { Box, Paper, Typography } from '@mui/material'
import PanelLabel from './PanelLabel'
import type { MapData } from '../types'
import playerImgSrc from '../assets/player.png'

const CONN_SIZE = 5 // pixels for connector cells (odd grid indices)

const playerImg = new Image()
playerImg.src = playerImgSrc

// Collect all PNG assets: { 'hut': '/assets/hut.png', ... }
const assetUrls: Record<string, string> = Object.fromEntries(
  Object.entries(import.meta.glob('../assets/*.png', { eager: true, import: 'default' }) as Record<string, string>)
    .map(([path, url]) => [path.replace(/^.*\//, '').replace(/\.png$/, ''), url])
)

const iconCache: Record<string, HTMLImageElement> = {}
const SAFE_ICON = /^[a-zA-Z0-9_-]+$/

function getOrLoadIcon(name: string, onLoad: () => void): HTMLImageElement | null {
  if (!SAFE_ICON.test(name)) return null
  if (iconCache[name]) return iconCache[name]
  const url = assetUrls[name]
  if (!url) return null
  const img = new Image()
  img.onload = onLoad
  img.src = url
  iconCache[name] = img
  return img
}

interface Props {
  map: MapData | null
}

// Maps a single grid axis coordinate to canvas { pos, size }.
// Even indices are rooms, odd indices are connectors.
function cellLayout(g: number, roomSize: number): { pos: number; size: number } {
  const idx = Math.floor(g / 2)
  const isRoom = g % 2 === 0
  return {
    pos: idx * (roomSize + CONN_SIZE) + (isRoom ? 0 : roomSize),
    size: isRoom ? roomSize : CONN_SIZE,
  }
}

export default function MapPanel({ map }: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef(map)
  mapRef.current = map

  useEffect(() => {
    const canvas = canvasRef.current
    const container = containerRef.current
    if (!canvas || !container) return

    function draw() {
      const m = mapRef.current
      if (!canvas || !container || !m || m.grid.length === 0) return

      const gridCols = m.grid[0]?.length ?? 0
      const gridRows = m.grid.length
      if (gridCols === 0) return

      // Count how many room vs connector columns/rows exist
      const roomCols = Math.ceil(gridCols / 2)
      const connCols = Math.floor(gridCols / 2)
      const roomRows = Math.ceil(gridRows / 2)
      const connRows = Math.floor(gridRows / 2)

      // Compute room size so the map fills the container, keeping rooms square
      const roomSizeX = Math.floor((container.clientWidth  - connCols * CONN_SIZE) / roomCols)
      const roomSizeY = Math.floor((container.clientHeight - connRows * CONN_SIZE) / roomRows)
      const roomSize  = Math.max(1, Math.min(roomSizeX, roomSizeY))

      const canvasW = roomCols * roomSize + connCols * CONN_SIZE
      const canvasH = roomRows * roomSize + connRows * CONN_SIZE
      canvas.width  = canvasW
      canvas.height = canvasH

      const ctx = canvas.getContext('2d')
      if (!ctx) return
      ctx.clearRect(0, 0, canvasW, canvasH)

      for (let gy = 0; gy < gridRows; gy++) {
        const ly = cellLayout(gy, roomSize)
        for (let gx = 0; gx < gridCols; gx++) {
          const { color, icon } = m.grid[gy][gx]
          const lx = cellLayout(gx, roomSize)

          if (color) {
            ctx.fillStyle = color
            ctx.fillRect(lx.pos, ly.pos, lx.size, ly.size)
          }

          ctx.imageSmoothingEnabled = false

          if (icon) {
            const img = getOrLoadIcon(icon, draw)
            if (img?.complete) ctx.drawImage(img, lx.pos, ly.pos, lx.size, ly.size)
          }

          if (map.playerX == gx && map.playerY == gy) {
            ctx.drawImage(playerImg, lx.pos, ly.pos, lx.size, ly.size)
          }
        }
      }
    }

    if (playerImg.complete) {
      draw()
    } else {
      playerImg.onload = draw
    }

    const ro = new ResizeObserver(draw)
    ro.observe(container)
    return () => ro.disconnect()
  }, [map])

  return (
    <Paper sx={{ gridArea: 'map', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <PanelLabel>Map</PanelLabel>
      <Box
        ref={containerRef}
        sx={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', overflow: 'hidden' }}
      >
        {map
          ? <canvas ref={canvasRef} style={{ imageRendering: 'pixelated', display: 'block' }} />
          : <Typography color="text.secondary" sx={{ p: 1 }}>—</Typography>
        }
      </Box>
    </Paper>
  )
}
