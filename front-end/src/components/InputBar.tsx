import { type RefObject } from 'react'
import { Box, Button, InputBase, Paper, Typography } from '@mui/material'

interface Props {
  cmdInput: string
  setCmdInput: (v: string) => void
  cmdRef: RefObject<HTMLInputElement | null>
  onSend: () => void
}

export default function InputBar({ cmdInput, setCmdInput, cmdRef, onSend }: Props) {
  return (
    <Paper sx={{ gridArea: 'input', display: 'flex', alignItems: 'center', gap: 1, px: 1 }}>
      <Typography sx={{ color: 'primary.main', flexShrink: 0 }}>{'>'}</Typography>
      <InputBase
        inputRef={cmdRef}
        value={cmdInput}
        onChange={(e) => setCmdInput(e.target.value)}
        onKeyDown={(e) => { if (e.key === 'Enter') onSend() }}
        autoFocus
        sx={{ flex: 1, color: 'text.primary', fontFamily: 'inherit', fontSize: 'inherit' }}
        inputProps={{ spellCheck: false, autoComplete: 'off' }}
      />
      <Button onClick={onSend} variant="outlined" size="small" sx={{ flexShrink: 0, fontSize: 12 }}>
        Send
      </Button>
    </Paper>
  )
}
