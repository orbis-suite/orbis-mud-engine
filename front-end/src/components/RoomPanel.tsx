import { Box, Button, Paper, Typography } from '@mui/material'
import PanelLabel from './PanelLabel'
import type { Direction, RoomContent } from '../types'

interface Props {
  room: RoomContent | null
  onMove: (direction: Direction) => void
}

export default function RoomPanel({ room, onMove }: Props) {
  return (
    <Paper sx={{ gridArea: 'room', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <PanelLabel>Room</PanelLabel>
      <Box sx={{ p: '10px 12px', overflowY: 'auto', flex: 1 }}>
        {room ? (
          <>
            <Typography sx={{ fontSize: 15, color: '#d4a8ff', letterSpacing: '0.03em', mb: 0.75 }}>
              {room.name}
            </Typography>
            {room.exits.length > 0 && (
              <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5, mb: 0.75 }}>
                {room.exits.map((exit) => (
                  <Button
                    key={exit}
                    size="small"
                    variant="outlined"
                    onClick={() => onMove(exit as Direction)}
                    sx={{ minWidth: 0, px: 1, py: 0.25, fontSize: 11 }}
                  >
                    {exit}
                  </Button>
                ))}
              </Box>
            )}
            <Typography variant="body2" sx={{ color: '#a89db8', whiteSpace: 'pre-wrap' }}>
              {room.description}
            </Typography>
          </>
        ) : <Typography color="text.secondary">—</Typography>}
      </Box>
    </Paper>
  )
}
