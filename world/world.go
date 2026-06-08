package world

import (
	"fmt"
	"log"
	"strings"

	"example.com/mud/parser"
	"example.com/mud/parser/commands"
	"example.com/mud/world/entities"
	"example.com/mud/world/entities/components"
	"example.com/mud/world/player"
	"example.com/mud/world/scheduler"
)

type World struct {
	Scheduler *scheduler.Scheduler

	entityMap    map[string]*entities.Entity
	startingRoom string
	bus          *Bus
}

func NewWorld(entityMap map[string]*entities.Entity, startingRoom string) *World {
	return &World{
		entityMap:    entityMap,
		startingRoom: startingRoom,
		Scheduler:    scheduler.NewScheduler(),
		bus:          NewBus(),
	}
}

func (w *World) EntitiesById() map[string]*entities.Entity { return w.entityMap }

func (w *World) AddPlayer(name string, inbox chan string) (*player.Player, error) {
	startingRoom, ok := w.entityMap[w.startingRoom]
	if !ok {
		log.Fatalf("add player: room '%s' does not exist in world.", w.startingRoom)
	}

	newPlayer, err := player.NewPlayer(name, w, startingRoom)
	if err != nil {
		return nil, fmt.Errorf("could not create player '%s': %w", name, err)
	}

	if room, ok := entities.GetComponent[*components.Room](newPlayer.CurrentRoom); ok {
		room.AddChild(newPlayer.Entity)
	}

	w.bus.Subscribe(newPlayer.CurrentRoom, newPlayer.Entity, inbox)
	w.Publish(newPlayer.CurrentRoom, fmt.Sprintf("%s enters the room.", newPlayer.Name), []*entities.Entity{newPlayer.Entity})

	return newPlayer, nil
}

func (w *World) DisconnectPlayer(p *player.Player) {
	if room, ok := entities.GetComponent[*components.Room](p.CurrentRoom); ok {
		room.RemoveChild(p.Entity)
	}

	w.bus.Unsubscribe(p.CurrentRoom, p.Entity)
	w.Publish(p.CurrentRoom, fmt.Sprintf("%s leaves the room.", p.Name), []*entities.Entity{p.Entity})
}

func (w *World) GetEntityById(id string) (*entities.Entity, bool) {
	entity, ok := w.entityMap[id]
	return entity, ok
}

func (w *World) Publish(room *entities.Entity, text string, exclude []*entities.Entity) {
	w.bus.Publish(room, text, exclude)
}

func (w *World) PublishTo(room *entities.Entity, recipient *entities.Entity, text string) {
	w.bus.PublishTo(room, recipient, text)
}

func (w *World) GetScheduler() *scheduler.Scheduler {
	return w.Scheduler
}

func (w *World) Parse(p *player.Player, line string) (string, error) {
	cmd := parser.Parse(line)
	if cmd == nil {
		return "What in the nine hells?", nil
	}

	switch cmd.Kind {
	case "help":
		return w.HelpMessage(cmd.Params["command"]), nil
	case "move":
		return p.Move(cmd.Params["direction"])
	case "look":
		return p.Look(cmd.Params["target"])
	case "inventory":
		return p.Inventory()
	case "map":
		return p.Map()
	case "track":
		return p.Track(cmd.Params["target"])
	}

	// see if it has target
	if target := cmd.Params["target"]; target != "" {
		if instrument := cmd.Params["instrument"]; instrument != "" {
			response, err := p.ActUponWithAlias(cmd.Kind, target, instrument, cmd.NoMatchMessage)
			return response, err
		} else if message := cmd.Params["message"]; message != "" {
			response, err := p.ActUponMessageAlias(cmd.Kind, target, message, cmd.NoMatchMessage)
			return response, err
		} else {
			response, err := p.ActUponAlias(cmd.Kind, target, cmd.NoMatchMessage)
			return response, err
		}
	}

	// see if it has a message
	if message := cmd.Params["message"]; message != "" {
		response, err := p.ActMessage(cmd.Kind, message, cmd.NoMatchMessage)
		return response, err
	}

	return "What the hell are you talking about?", nil
}

func (w *World) HelpMessage(command string) string {
	if command == "" {
		return w.HelpGeneral()
	}

	canonical, ok := commands.VerbAliases[command]
	if !ok {
		return fmt.Sprintf("Unrecognized command: %s", command)
	}

	var b strings.Builder

	for _, p := range commands.Patterns {
		if strings.ToLower(p.Kind) == canonical {
			b.WriteString("- ")
			b.WriteString(p.String())

			if p.HelpMessage != "" {
				b.WriteString(": ")
				b.WriteString(p.HelpMessage)
			}

			b.WriteString("\n")
		}
	}

	return b.String()
}

func (w *World) HelpGeneral() string {
	var b strings.Builder

	for _, p := range commands.Patterns {
		b.WriteString("- ")
		b.WriteString(p.String())

		if p.HelpMessage != "" {
			b.WriteString(": ")
			b.WriteString(p.HelpMessage)
		}

		b.WriteString("\n")
	}

	return b.String()
}

func (w *World) MovePlayer(p *player.Player, direction string) (string, error) {
	playerRoom, err := entities.RequireComponent[*components.Room](p.CurrentRoom)
	if err != nil {
		return "", fmt.Errorf("move for player '%s': %w", p.Name, err)
	}

	newRoom := w.getNeighboringRoom(playerRoom, direction)
	if newRoom != nil {
		w.Publish(p.CurrentRoom, fmt.Sprintf("%s leaves the room.", p.Name), []*entities.Entity{p.Entity})

		playerRoom.RemoveChild(p.Entity)
		p.CurrentRoom = newRoom

		if room, ok := entities.GetComponent[*components.Room](p.CurrentRoom); ok {
			room.AddChild(p.Entity)
		}

		w.bus.Move(p.CurrentRoom, p.Entity)
		w.Publish(p.CurrentRoom, fmt.Sprintf("%s enters the room.", p.Name), []*entities.Entity{p.Entity})

		return p.GetRoomDescription()
	}

	return "You can't go there.", nil
}

func (w *World) getNeighboringRoom(r *components.Room, direction string) *entities.Entity {
	if roomId, ok := r.GetNeighboringRoomId(direction); ok {
		room := w.entityMap[roomId]
		return room
	}
	return nil
}
