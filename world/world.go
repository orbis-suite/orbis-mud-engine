package world

import (
	"fmt"
	"log"
	"strings"

	"example.com/mud/parser"
	"example.com/mud/parser/commands"
	"example.com/mud/world/entities"
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

func (w *World) Init() error {
	for _, e := range w.entityMap {
		err := w.initEntityAndChildren(e)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *World) initEntityAndChildren(e *entities.Entity) error {
	if e.InitFunc != nil {
		e.InitFunc(&entities.Event{
			Type:         "init",
			Publisher:    w,
			Scheduler:    w.GetScheduler(),
			EntitiesById: w.entityMap,
			Room:         e.Parent,
			Source:       e,
		})
	}

	for _, children := range e.GetChildren() {
		for _, child := range children {
			w.initEntityAndChildren(child)
		}
	}

	return nil
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

	// if room, ok := entities.GetComponent[*components.Room](newPlayer.CurrentRoom); ok {
	// 	room.AddChild(newPlayer.Entity)
	// }

	w.bus.Subscribe(startingRoom, newPlayer.Entity, inbox)
	w.Publish(startingRoom, fmt.Sprintf("%s enters the room.", newPlayer.Name), []*entities.Entity{newPlayer.Entity})

	return newPlayer, nil
}

func (w *World) DisconnectPlayer(p *player.Player) {
	// TODO
	// if room, ok := entities.GetComponent[*components.Room](p.Entity.Parent); ok {
	// 	room.RemoveChild(p.Entity)
	// }

	w.bus.Unsubscribe(p.Entity.Parent, p.Entity)
	w.Publish(p.Entity.Parent, fmt.Sprintf("%s leaves the room.", p.Name), []*entities.Entity{p.Entity})
}

func (w *World) GetEntityById(id string) (*entities.Entity, bool) {
	entity, ok := w.entityMap[id]
	return entity, ok
}

func (w *World) Publish(room *entities.Entity, text string, exclude []*entities.Entity) {
	w.bus.Publish(room, text, exclude)
}

func (w *World) PublishTo(recipient *entities.Entity, text string) {
	w.bus.PublishTo(recipient, text)
}

func (w *World) Move(toRoom *entities.Entity, player *entities.Entity) {
	w.bus.Move(toRoom, player)
}

func (w *World) GetScheduler() *scheduler.Scheduler {
	return w.Scheduler
}

func (w *World) Parse(p *player.Player, line string) (string, error) {
	// TODO append to errors rather than just returning them
	cmd := parser.Parse(line)
	if cmd == nil {
		return "What in the nine hells?", nil
	}

	switch cmd.Kind {
	case "help":
		return w.HelpMessage(cmd.Params["command"]), nil
	case "look":
		return p.Look(cmd.Params["target"])
	case "inventory":
		return p.Inventory()
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

	response, err := p.Act(cmd.Kind, cmd.Params, cmd.NoMatchMessage)
	return response, err
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
