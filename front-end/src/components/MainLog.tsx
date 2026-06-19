import { forwardRef } from 'react'
import { Box, Paper } from '@mui/material'
import PanelLabel from './PanelLabel'

interface Props {
  lines: string[]
}

const MainLog = forwardRef<HTMLDivElement, Props>(({ lines }, ref) => (
  <Paper ref={ref} sx={{ gridArea: 'main', display: 'flex', flexDirection: 'column', overflowY: 'auto' }}>
    <PanelLabel>Main</PanelLabel>
    <Box sx={{ p: '8px 12px', display: 'flex', flexDirection: 'column', gap: '3px' }}>
      {lines.map((line, i) => (
        <Box key={i} sx={{
          lineHeight: 1.6,
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-word',
          color: '#e1ddeb',
          bgcolor: i % 2 === 0 ? 'transparent' : '#1c1a24',
        }}>
          {line}
        </Box>
      ))}
    </Box>
  </Paper>
))

MainLog.displayName = 'MainLog'

export default MainLog
