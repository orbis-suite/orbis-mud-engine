import { Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, TextField, Typography } from '@mui/material'
import type { Phase } from '../types'

interface Props {
  phase: Phase
  nameInput: string
  setNameInput: (v: string) => void
  nameError: string
  onConnect: (name: string) => void
}

export default function NameDialog({ phase, nameInput, setNameInput, nameError, onConnect }: Props) {
  return (
    <Dialog open={phase !== 'playing'}>
      <Box component="form" onSubmit={(e) => { e.preventDefault(); onConnect(nameInput.trim()) }}>
        <DialogTitle>Enter the World</DialogTitle>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: '8px !important', width: 300 }}>
          <Typography variant="body2" color="text.secondary">
            What is your name, weary adventurer?
          </Typography>
          <TextField
            value={nameInput}
            onChange={(e) => setNameInput(e.target.value)}
            disabled={phase === 'connecting'}
            autoFocus
            size="small"
            slotProps={{ htmlInput: { spellCheck: false, autoComplete: 'off', maxLength: 20 } }}
          />
          {nameError && (
            <Typography variant="caption" color="error">{nameError}</Typography>
          )}
        </DialogContent>
        <DialogActions>
          <Button type="submit" disabled={phase === 'connecting'} fullWidth variant="outlined">
            {phase === 'connecting' ? 'Connecting…' : 'Enter'}
          </Button>
        </DialogActions>
      </Box>
    </Dialog>
  )
}
