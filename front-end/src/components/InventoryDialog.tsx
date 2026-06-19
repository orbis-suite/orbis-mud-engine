import { Button, Dialog, DialogActions, DialogContent, DialogTitle, List, ListItem, ListItemText, Typography } from '@mui/material'

interface Props {
  open: boolean
  inventory: string[]
  onClose: () => void
  onExited: () => void
}

export default function InventoryDialog({ open, inventory, onClose, onExited }: Props) {
  return (
    <Dialog
      open={open}
      onClose={onClose}
      slotProps={{ transition: { onExited } }}
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
        <Button onClick={onClose} variant="outlined">Close</Button>
      </DialogActions>
    </Dialog>
  )
}
