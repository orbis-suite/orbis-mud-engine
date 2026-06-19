import { useEffect, useRef, useReducer, useState } from 'react'
import { ThemeProvider, createTheme, CssBaseline, GlobalStyles, Box } from '@mui/material'

import NameDialog from './components/NameDialog'
import InventoryDialog from './components/InventoryDialog'
import RoomPanel from './components/RoomPanel'
import MapPanel from './components/MapPanel'
import MainLog from './components/MainLog'
import ItemsPanel from './components/ItemsPanel'
import InputBar from './components/InputBar'

import type { State, Action, WSMessage, ClientMessage, Direction, RoomContent, InventoryContent, TextContent, EntityContent, MapData } from './types'

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

// ── Reducer ───────────────────────────────────────────────────────────────────

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

// ── App ───────────────────────────────────────────────────────────────────────

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

  useEffect(() => {
    const dirMap: Partial<Record<string, Direction>> = {
      w: 'north', s: 'south', a: 'west', d: 'east',
    }
    const handleKeyDown = (e: KeyboardEvent) => {
      if (document.activeElement === cmdRef.current) return
      const direction = dirMap[e.key.toLowerCase()]
      if (direction) sendMessage({ type: 'move', direction })
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [])

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

  function sendMessage(msg: ClientMessage) {
    if (!ws.current || ws.current.readyState !== WebSocket.OPEN) return
    ws.current.send(JSON.stringify(msg))
  }

  function send() {
    const line = cmdInput.trim()
    if (!line) return
    sendMessage({ type: 'text', text: line })
    cmdRef.current?.select()
  }

  const { room, map, lines, inventory, inventoryOpen } = state

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <GlobalStyles styles={{ 'html, body, #root': { height: '100%' } }} />

      <NameDialog
        phase={state.phase}
        nameInput={nameInput}
        setNameInput={setNameInput}
        nameError={state.nameError}
        onConnect={connect}
      />

      <InventoryDialog
        open={inventoryOpen}
        inventory={inventory}
        onClose={() => dispatch({ type: 'close_inventory' })}
        onExited={() => dispatch({ type: 'clear_inventory' })}
      />

      {state.phase === 'playing' && (
        <Box
          onKeyDown={(e) => { if (e.key === 'Escape') dispatch({ type: 'close_inventory' }) }}
          sx={{
            height: '100vh',
            display: 'grid',
            gridTemplateColumns: '3fr 1fr',
            gridTemplateRows: '1fr 2fr 42px',
            gridTemplateAreas: '"room map" "main items" "input input"',
            gap: '4px',
            p: '4px',
            bgcolor: 'background.default',
          }}
        >
          <RoomPanel room={room} onMove={(dir) => sendMessage({ type: 'move', direction: dir })} />
          <MapPanel map={map} />
          <MainLog ref={logRef} lines={lines} />
          <ItemsPanel room={room} />
          <InputBar cmdInput={cmdInput} setCmdInput={setCmdInput} cmdRef={cmdRef} onSend={send} />
        </Box>
      )}
    </ThemeProvider>
  )
}
