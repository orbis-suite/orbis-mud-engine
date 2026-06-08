package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"example.com/mud/config"
	"example.com/mud/world"
	"example.com/mud/world/entities"
	"example.com/mud/world/player"
	"example.com/mud/world/response"
)

var upgrader = websocket.Upgrader{
	// TODO: restrict allowed origins in production
	CheckOrigin: func(r *http.Request) bool { return true },
}

// wsConn serialises writes to a *websocket.Conn.
// gorilla/websocket does not allow concurrent writers.
type wsConn struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

func (w *wsConn) writeResp(r response.Response) error {
	content, err := json.Marshal(r)
	if err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteJSON(response.WSMessage{Panel: r.Panel(), Content: content})
}

func (w *wsConn) writeText(text string) error {
	return w.writeResp(response.Text{Value: text})
}

func (w *wsConn) readLine() (string, error) {
	_, b, err := w.conn.ReadMessage()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func (w *wsConn) closeWithError(msg string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	_ = w.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, msg),
	)
}

// HandleWS upgrades the HTTP connection to WebSocket and runs the full session.
// The client must supply the player name as the "name" query parameter.
func HandleWS(w http.ResponseWriter, r *http.Request, gameWorld *world.World, cfg *config.Config) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))

	if vdn := player.NameValidation(name); vdn != "" {
		http.Error(w, strings.TrimSpace(vdn), http.StatusBadRequest)
		return
	}

	raw, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	conn := &wsConn{conn: raw}
	defer raw.Close()

	inbox := make(chan string, 64)
	p, err := gameWorld.AddPlayer(name, inbox)
	if err != nil {
		msg := fmt.Sprintf("error adding player: %v", err)
		fmt.Println(msg)
		conn.closeWithError(msg)
		return
	}

	opening, err := p.OpeningMessage()
	if err != nil {
		msg := fmt.Sprintf("error printing opening message: %v", err)
		fmt.Println(msg)
		conn.closeWithError(msg)
		return
	}
	if err := conn.writeResp(opening); err != nil {
		return
	}

	pushRoom := func() {
		if room, err := p.GetRoomDescription(); err == nil {
			_ = conn.writeResp(room)
		}
		if mapView, err := p.Map(); err == nil {
			_ = conn.writeResp(mapView)
		}
	}

	pushRoom()

	// inbox goroutine: async room-broadcast messages always go to main panel.
	// After each narrative message, push a fresh room description so the
	// Items in Room panel stays in sync when other players pick up / drop items.
	go func() {
		for msg := range inbox {
			_ = conn.writeText(msg)
			pushRoom()
		}
	}()

	handleWSOutgoing(conn, gameWorld, p, cfg, pushRoom)

	gameWorld.DisconnectPlayer(p)
	fmt.Println("WebSocket connection closed")
}

func handleWSOutgoing(conn *wsConn, gameWorld *world.World, p *player.Player, cfg *config.Config, pushRoom func()) {
	for {
		line, err := conn.readLine()
		if err != nil {
			break
		}
		if line == "" {
			continue
		}
		if strings.ToLower(line) == "quit" {
			break
		}

		// pending ambiguity resolution
		if pending := p.Pending; pending != nil {
			n, nerr := strconv.Atoi(line)
			if nerr == nil {
				slot := pending.Ambiguity.Slots[pending.StepIndex]
				if n >= 1 && n <= len(slot.Matches) {
					pending.Selected[slot.Role] = n - 1
					pending.StepIndex++
					if pending.StepIndex < len(pending.Ambiguity.Slots) {
						promptWSSlot(conn, pending)
					} else {
						chosen := make(map[string]*entities.Entity, len(pending.Selected))
						for _, s := range pending.Ambiguity.Slots {
							idx := pending.Selected[s.Role]
							chosen[s.Role] = s.Matches[idx].Entity
						}
						out, execErr := pending.Ambiguity.Execute(chosen)
						p.Pending = nil
						if execErr != nil {
							_ = conn.writeText(execErr.Error())
						} else if out != "" {
							_ = conn.writeText(out)
						}
					}
				}
				continue
			}
			p.Pending = nil
		}

		if cooldown := p.CooldownRemaining(); cooldown > 0 {
			_ = conn.writeText(fmt.Sprintf("You need to catch your breath. Try again in %.1fs", cooldown.Seconds()))
			continue
		}

		resp, parseErr := gameWorld.Parse(p, line)
		if parseErr != nil {
			var amb *entities.AmbiguityError
			if errors.As(parseErr, &amb) {
				p.Pending = &entities.PendingAction{
					Ambiguity: amb,
					StepIndex: 0,
					Selected:  map[string]int{},
				}
				promptWSSlot(conn, p.Pending)
				continue
			}
			wrapped := fmt.Sprintf("error received: %v", parseErr)
			fmt.Println(wrapped)
			_ = conn.writeText(wrapped)
		} else if resp != nil {
			// Skip empty text responses — they indicate a handled event with no
			// direct player feedback (e.g. say/tell broadcasts to the room).
			txt, isText := resp.(response.Text)
			if !isText || txt.Value != "" {
				_ = conn.writeResp(resp)
				// After any response, push a fresh room description and map so
				// the Items in Room and Map panels stay in sync.
				pushRoom()
			}
		}

		p.StartCooldown(time.Duration(cfg.PlayerRateLimit) * time.Millisecond)
	}
}

func promptWSSlot(conn *wsConn, p *entities.PendingAction) {
	slot := p.Ambiguity.Slots[p.StepIndex]
	var sb strings.Builder
	sb.WriteString(slot.Prompt)
	for i, opt := range slot.Matches {
		sb.WriteString(fmt.Sprintf("\n  %d) %s", i+1, opt.Text))
	}
	_ = conn.writeText(sb.String())
}
