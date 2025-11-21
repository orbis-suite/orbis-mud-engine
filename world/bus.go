package world

import (
	"sync"

	"example.com/mud/world/entities"
)

type Bus struct {
	mu              sync.RWMutex
	roomSubscribers map[*entities.Entity]map[*entities.Entity]chan string // room -> (player -> inbox channel)
	playerRooms     map[*entities.Entity]*entities.Entity                 // player -> room
}

func NewBus() *Bus {
	return &Bus{
		roomSubscribers: make(map[*entities.Entity]map[*entities.Entity]chan string),
		playerRooms:     make(map[*entities.Entity]*entities.Entity),
	}
}

func (b *Bus) Subscribe(newRoom *entities.Entity, player *entities.Entity, inbox chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// unsubscribe from current room
	if oldRoom, ok := b.playerRooms[player]; ok {
		if subscribers := b.roomSubscribers[oldRoom]; subscribers != nil {
			delete(subscribers, player)
			if len(subscribers) == 0 {
				delete(b.roomSubscribers, oldRoom)
			}
		}
	}

	// subscribe to new room
	subscribers := b.roomSubscribers[newRoom]
	if subscribers == nil {
		subscribers = make(map[*entities.Entity]chan string)
		b.roomSubscribers[newRoom] = subscribers
	}
	subscribers[player] = inbox

	// update index
	b.playerRooms[player] = newRoom
}

func (b *Bus) Unsubscribe(room *entities.Entity, player *entities.Entity) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subscribers := b.roomSubscribers[room]; subscribers != nil {
		delete(subscribers, player)
		if len(subscribers) == 0 {
			delete(b.roomSubscribers, room)
		}
	}

	delete(b.playerRooms, player)
}

func (b *Bus) Move(toRoom *entities.Entity, player *entities.Entity) {
	b.mu.RLock()
	oldRoom := b.playerRooms[player]
	var inbox chan string
	if oldRoom != nil {
		if subs := b.roomSubscribers[oldRoom]; subs != nil {
			inbox = subs[player]
		}
	}
	b.mu.RUnlock()

	if inbox != nil {
		b.Subscribe(toRoom, player, inbox)
	}
}

func (b *Bus) Publish(room *entities.Entity, text string, exclude []*entities.Entity) {
	excludeSet := make(map[*entities.Entity]struct{}, len(exclude))
	for _, ex := range exclude {
		excludeSet[ex] = struct{}{}
	}

	b.mu.RLock()
	subscribers := b.roomSubscribers[room]
	var targets []chan string
	for p, inbox := range subscribers {
		if _, excluded := excludeSet[p]; excluded {
			continue
		}
		targets = append(targets, inbox)
	}
	b.mu.RUnlock()

	for _, inbox := range targets {
		select {
		case inbox <- text:
		default:
			// drop if receiver is slow
		}
	}
}

func (b *Bus) PublishTo(recipient *entities.Entity, text string) {
	b.mu.RLock()

	// get room that player is subscribed to, and send the message
	room := b.playerRooms[recipient]
	subscribers := b.roomSubscribers[room]
	inbox := subscribers[recipient]

	b.mu.RUnlock()

	select {
	case inbox <- text:
	default:
		// drop if receiver is slow
	}
}
