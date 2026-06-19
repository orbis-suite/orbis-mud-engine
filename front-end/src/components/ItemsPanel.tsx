import { Box, List, ListItem, Paper, Typography } from '@mui/material'
import PanelLabel from './PanelLabel'
import type { RoomContent } from '../types'

interface Props {
  room: RoomContent | null
}

export default function ItemsPanel({ room }: Props) {
  return (
    <Paper sx={{ gridArea: 'items', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <PanelLabel>Items in Room</PanelLabel>
      <Box sx={{ p: '10px 12px', overflowY: 'auto', flex: 1 }}>
        {room && room.children.length > 0 ? (
          <List dense disablePadding>
            {room.children.map((child, i) => (
              <ListItem key={i} disableGutters sx={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start', mb: 0.5 }}>
                <Typography variant="body2">{child.name}</Typography>
                {child.description && (
                  <Typography variant="caption" color="text.secondary">{child.description}</Typography>
                )}
              </ListItem>
            ))}
          </List>
        ) : <Typography color="text.secondary">Nothing here.</Typography>}
      </Box>
    </Paper>
  )
}
