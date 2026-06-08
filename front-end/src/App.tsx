import { useEffect, useRef, useReducer, useState } from 'react'
import {
  ThemeProvider, createTheme, CssBaseline, GlobalStyles,
  Box, Paper, Typography, InputBase, Button,
  Dialog, DialogTitle, DialogContent, DialogActions,
  TextField, List, ListItem, ListItemText,
} from '@mui/material'

// ── Theme ─────────────────────────────────────────────────────────────────────

const theme = createTheme({
  palette: {
    mode: 'dark',
    background: { default: '#0d0d0f', paper: '#16141c' },
    primary: { main: '#7c5cbf' },
    text: { primary: '#c8c4d0', secondary: '#6b5f80' },
    divider: '#2a2733',
    error: { main: '#c06060' },
  },
  typography: {
    fontFamily: "ui-monospace, Consolas, 'Courier New', monospace",
    fontSize: 14,
  },
  components: {
    MuiPaper: {
      defaultProps: { elevation: 0 },
      styleOverrides: { root: { border: '1px solid #2a2733' } },
    },
    MuiDialog: {
      styleOverrides: { paper: { border: '1px solid #2a2733' } },
    },
    MuiButton: {
      styleOverrides: { root: { fontFamily: "ui-monospace, Consolas, 'Courier New', monospace", textTransform: 'none' } },
    },
  },
})

// ── Wire types ────────────────────────────────────────────────────────────────

interface RoomContent {
  name: string
  description: string
  exits: string[]
  children: { name: string; description: string }[]
}

interface TextContent { text: string }
interface EntityContent { name: string; description: string }
interface InventoryContent { items: string[] }
interface MapCell { color: string; icon: string }
interface MapData { grid: MapCell[][]; playerX: number; playerY: number }

interface WSMessage {
  panel: 'main' | 'map' | 'inventory' | 'room'
  content: unknown
}

// ── State ─────────────────────────────────────────────────────────────────────

type Phase = 'modal' | 'connecting' | 'playing'

interface State {
  phase: Phase
  nameError: string
  lines: string[]
  room: RoomContent | null
  map: MapData | null
  inventoryOpen: boolean
  inventory: string[]
}

type Action =
  | { type: 'connecting' }
  | { type: 'disconnected' }
  | { type: 'name_error'; error: string }
  | { type: 'message'; msg: WSMessage }
  | { type: 'close_inventory' }
  | { type: 'clear_inventory' }

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'close_inventory':
      return { ...state, inventoryOpen: false }
    case 'clear_inventory':
      return { ...state, inventory: [] }
    case 'connecting':
      return { ...state, phase: 'connecting', nameError: '' }
    case 'disconnected':
      if (state.phase === 'connecting') return { ...state, phase: 'modal' }
      return { ...state, phase: 'modal', lines: [...state.lines, '— disconnected —'] }
    case 'name_error':
      return { ...state, phase: 'modal', nameError: action.error }
    case 'message': {
      const { panel, content } = action.msg
      switch (panel) {
        case 'room':
          return { ...state, phase: 'playing', room: content as RoomContent }
        case 'map':
          return { ...state, map: content as MapData }
        case 'inventory':
          return { ...state, inventoryOpen: true, inventory: (content as InventoryContent).items }
        case 'main':
        default: {
          const c = content as TextContent | EntityContent
          const text = 'text' in c ? c.text : `${c.name}: ${c.description}`
          if (!text.trim()) return state
          return { ...state, lines: [...state.lines, text] }
        }
      }
    }
  }
}

const INITIAL: State = { phase: 'modal', nameError: '', lines: [], room: null, map: null, inventoryOpen: false, inventory: [] }

// ── Helpers ───────────────────────────────────────────────────────────────────

function validateName(name: string): string {
  if (name.length === 0) return 'Please enter a name.'
  if (name.length > 20) return 'Name must be 20 characters or fewer.'
  if (!/^[a-zA-Z]+$/.test(name)) return 'Name may only contain letters.'
  return ''
}

function PanelLabel({ children }: { children: string }) {
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

// ── Component ─────────────────────────────────────────────────────────────────

export default function App() {
  const [state, dispatch] = useReducer(reducer, INITIAL)
  const [nameInput, setNameInput] = useState('')
  const [cmdInput, setCmdInput] = useState('')
  const ws = useRef<WebSocket | null>(null)
  const logRef = useRef<HTMLDivElement>(null)
  const cmdRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (logRef.current) logRef.current.scrollTop = logRef.current.scrollHeight
  }, [state.lines])

  function connect(name: string) {
    const err = validateName(name)
    if (err) { dispatch({ type: 'name_error', error: err }); return }

    dispatch({ type: 'connecting' })
    const socket = new WebSocket(`${import.meta.env.VITE_WS_URL}?name=${encodeURIComponent(name)}`)
    ws.current = socket

    socket.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        dispatch({ type: 'message', msg })
      } catch { /* ignore malformed frames */ }
    }

    socket.onclose = (event) => {
      ws.current = null
      dispatch(event.reason ? { type: 'name_error', error: event.reason } : { type: 'disconnected' })
    }

    socket.onerror = () => {
      ws.current = null
      dispatch({ type: 'name_error', error: 'Could not connect to server.' })
    }
  }

  function send() {
    const line = cmdInput.trim()
    if (!line || !ws.current || ws.current.readyState !== WebSocket.OPEN) return
    ws.current.send(line)
    cmdRef.current?.select()
  }

  const { room, map, lines, inventory, inventoryOpen } = state

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <GlobalStyles styles={{ 'html, body, #root': { height: '100%' } }} />

      {/* ── Name dialog ────────────────────────────────────────────────── */}
      <Dialog open={state.phase !== 'playing'}>
        <Box component="form" onSubmit={(e) => { e.preventDefault(); connect(nameInput.trim()) }}>
          <DialogTitle>Enter the World</DialogTitle>
          <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: '8px !important', width: 300 }}>
            <Typography variant="body2" color="text.secondary">
              What is your name, weary adventurer?
            </Typography>
            <TextField
              value={nameInput}
              onChange={(e) => setNameInput(e.target.value)}
              disabled={state.phase === 'connecting'}
              autoFocus
              size="small"
              slotProps={{ htmlInput: { spellCheck: false, autoComplete: 'off', maxLength: 20 } }}
            />
            {state.nameError && (
              <Typography variant="caption" color="error">{state.nameError}</Typography>
            )}
          </DialogContent>
          <DialogActions>
            <Button type="submit" disabled={state.phase === 'connecting'} fullWidth variant="outlined">
              {state.phase === 'connecting' ? 'Connecting…' : 'Enter'}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>

      {/* ── Inventory dialog ────────────────────────────────────────────── */}
      <Dialog
        open={inventoryOpen}
        onClose={() => dispatch({ type: 'close_inventory' })}
        slotProps={{ transition: { onExited: () => dispatch({ type: 'clear_inventory' }) } }}
      >
        <DialogTitle>Inventory</DialogTitle>
        <DialogContent sx={{ minWidth: 260 }}>
          {inventory.length > 0 ? (
            <List dense disablePadding>
              {inventory.map((item, i) => (
                <ListItem key={i} disableGutters>
                  <ListItemText primary={item} />
                </ListItem>
              ))}
            </List>
          ) : (
            <Typography color="text.secondary">You are carrying nothing.</Typography>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => dispatch({ type: 'close_inventory' })} variant="outlined">Close</Button>
        </DialogActions>
      </Dialog>

      {/* ── Game layout ─────────────────────────────────────────────────── */}
      {state.phase === 'playing' && (
        <Box
          onKeyDown={(e) => { if (e.key === 'Escape') dispatch({ type: 'close_inventory' }) }}
          sx={{
            height: '100vh',
            display: 'grid',
            gridTemplateColumns: '1fr 1fr',
            gridTemplateRows: '1fr 2fr 42px',
            gridTemplateAreas: '"room map" "main items" "input input"',
            gap: '4px',
            p: '4px',
            bgcolor: 'background.default',
          }}
        >
          {/* Room */}
          <Paper sx={{ gridArea: 'room', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
            <PanelLabel>Room</PanelLabel>
            <Box sx={{ p: '10px 12px', overflowY: 'auto', flex: 1 }}>
              {room ? (
                <>
                  <Typography sx={{ fontSize: 15, color: '#d4a8ff', letterSpacing: '0.03em', mb: 0.75 }}>
                    {room.name}
                  </Typography>
                  <Typography variant="body2" sx={{ color: '#a89db8', whiteSpace: 'pre-wrap', mb: 0.75 }}>
                    {room.description}
                  </Typography>
                  {room.exits.length > 0 && (
                    <Typography variant="caption" color="text.secondary">
                      Exits: {room.exits.join(', ')}
                    </Typography>
                  )}
                </>
              ) : <Typography color="text.secondary">—</Typography>}
            </Box>
          </Paper>

          {/* Map */}
          <Paper sx={{ gridArea: 'map', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
            <PanelLabel>Map</PanelLabel>
            <Box sx={{ p: '10px 12px', overflowY: 'auto', flex: 1, fontFamily: 'monospace', lineHeight: 1.4, whiteSpace: 'pre' }}>
              {map ? map.grid.map((row, y) => (
                <Box key={y} component="div">
                  {row.map((cell, x) => (
                    <Box key={x} component="span" sx={{ color: cell.color || 'inherit' }}>{cell.icon}</Box>
                  ))}
                </Box>
              )) : <Typography color="text.secondary">—</Typography>}
            </Box>
          </Paper>

          {/* Main log */}
          <Paper ref={logRef} sx={{ gridArea: 'main', display: 'flex', flexDirection: 'column', overflowY: 'auto' }}>
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

          {/* Items in room */}
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

          {/* Input bar */}
          <Paper sx={{ gridArea: 'input', display: 'flex', alignItems: 'center', gap: 1, px: 1 }}>
            <Typography sx={{ color: 'primary.main', flexShrink: 0 }}>{'>'}</Typography>
            <InputBase
              inputRef={cmdRef}
              value={cmdInput}
              onChange={(e) => setCmdInput(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') send() }}
              autoFocus
              sx={{ flex: 1, color: 'text.primary', fontFamily: 'inherit', fontSize: 'inherit' }}
              inputProps={{ spellCheck: false, autoComplete: 'off' }}
            />
            <Button onClick={send} variant="outlined" size="small" sx={{ flexShrink: 0, fontSize: 12 }}>
              Send
            </Button>
          </Paper>
        </Box>
      )}
    </ThemeProvider>
  )
}
