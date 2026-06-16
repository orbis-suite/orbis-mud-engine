import { Typography } from '@mui/material'

export default function PanelLabel({ children }: { children: string }) {
  return (
    <Typography variant="caption" sx={{
      px: 1, py: 0.5, display: 'block', flexShrink: 0,
      color: 'text.secondary', textTransform: 'uppercase', letterSpacing: '0.08em',
      borderBottom: 1, borderColor: 'divider',
    }}>
      {children}
    </Typography>
  )
}
