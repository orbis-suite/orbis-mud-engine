package player

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"example.com/mud/models"
	"example.com/mud/utils"
	"example.com/mud/world/entities"
	"example.com/mud/world/entities/components"
	"example.com/mud/world/scheduler"
)

var safeNameRegex = regexp.MustCompile(`[^a-zA-Z]+`)

type Player struct {
	Name        string
	Entity      *entities.Entity
	CurrentRoom *entities.Entity
	Pending     *entities.PendingAction

	mu            sync.Mutex
	nextActionAt  time.Time
	trackingAlias string
	world         World
}

type World interface {
	EntitiesById() map[string]*entities.Entity
	GetEntityById(id string) (*entities.Entity, bool)
	MovePlayer(p *Player, direction string) (string, error)

	Publish(room *entities.Entity, text string, exclude []*entities.Entity)
	PublishTo(room *entities.Entity, recipient *entities.Entity, text string)

	GetScheduler() *scheduler.Scheduler
}

func NewPlayer(name string, world World, currentRoom *entities.Entity) (*Player, error) {
	playerTemplate, ok := world.GetEntityById("Player")
	if !ok {
		return nil, fmt.Errorf("entity with ID 'Player' does not exist in world")
	}

	playerEntity := playerTemplate.Copy(nil)
	playerEntity.Name = name
	playerEntity.Description = fmt.Sprintf("%s the brave hero is here.", name)
	playerEntity.Aliases = []string{strings.ToLower(name)}

	return &Player{
		Name:        name,
		Entity:      playerEntity,
		CurrentRoom: currentRoom,
		world:       world,
	}, nil
}

func (p *Player) OpeningMessage() (string, error) {
	message, err := p.GetRoomDescription()
	if err != nil {
		return "", fmt.Errorf("opening message for player '%s': %w", p.Name, err)
	}

	return message, nil
}

func NameValidation(name string) string {
	if len(name) == 0 {
		return "Please, speak up! I didn't hear a name.\n"
	} else if len(name) > 20 {
		return "That's much too long to remember!\n"
	}

	testName := safeNameRegex.ReplaceAllString(name, "")

	if testName != name {
		return "I'm no good with numbers or spaces, and I only speak English!\n"
	}

	return ""
}

func (p *Player) CooldownRemaining() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	if now.Before(p.nextActionAt) {
		return time.Until(p.nextActionAt)
	}
	return 0
}

func (p *Player) StartCooldown(d time.Duration) {
	p.mu.Lock()
	p.nextActionAt = time.Now().Add(d)
	p.mu.Unlock()
}

func (p *Player) GetRoomDescription() (string, error) {
	var b strings.Builder

	room, err := entities.RequireComponent[*components.Room](p.CurrentRoom)
	if err != nil {
		return "", err
	}

	formattedTitle, err := utils.FormatText(fmt.Sprintf("{'%s' | bold | red}", p.CurrentRoom.Name), map[string]string{})
	if err != nil {
		return "", fmt.Errorf("could not format room '%s' name: %w", p.CurrentRoom.Name, err)
	}

	b.WriteString(formattedTitle)
	b.WriteString("\n")

	roomDescription := strings.TrimSpace(p.CurrentRoom.Description)
	b.WriteString(roomDescription)
	b.WriteString("\n")

	for _, e := range room.GetChildren().GetChildren() {
		if e == p.Entity {
			continue
		}

		description, err := e.GetDescription()
		if err != nil {
			return "", err
		}

		b.WriteString(fmt.Sprintf("%s%s", models.Tab, description))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(room.GetExitText())
	return b.String(), nil
}

func (p *Player) Move(direction string) (string, error) {
	message, err := p.world.MovePlayer(p, direction)
	return message, err
}

func (p *Player) Look(alias string) (string, error) {
	if alias == "" {
		message, err := p.GetRoomDescription()
		if err != nil {
			return "", fmt.Errorf("look room for player '%s': %w", p.Name, err)
		}
		return message, nil
	}

	matches, err := p.getEntitiesByAlias(alias)
	if err != nil {
		return "", fmt.Errorf("get look target for player '%s': %w", p.Name, err)
	}

	if len(matches) == 0 {
		return fmt.Sprintf("There is no %s for you to look upon.", alias), nil
	} else if len(matches) == 1 {
		description, err := matches[0].Entity.GetDescription()
		return description, err
	}

	slots := []entities.AmbiguitySlot{
		{
			Role:    entities.EventRoleTarget.String(),
			Prompt:  "Which target?",
			Matches: matches,
		},
	}

	return "", &entities.AmbiguityError{
		Slots: slots,
		Execute: func(inputMap map[string]*entities.Entity) (string, error) {
			t := inputMap[entities.EventRoleTarget.String()]

			description, err := t.GetDescription()
			return description, err
		},
	}
}

func (p *Player) Inventory() (string, error) {
	if inventory, ok := entities.GetComponent[*components.Inventory](p.Entity); ok {
		message, err := inventory.Print()
		if err != nil {
			return "", fmt.Errorf("inventory print for player '%s': %w", p.Name, err)
		}
		return message, nil
	}
	return "You couldn't possibly carry anything at all.", nil
}

func (p *Player) ActMessage(action, message, noMatchMessage string) (string, error) {
	return p.sendEventToEntity(p.Entity, &entities.Event{
		Type:         action,
		Publisher:    p.world,
		Scheduler:    p.world.GetScheduler(),
		EntitiesById: p.world.EntitiesById(),
		Room:         p.CurrentRoom,
		Source:       p.Entity,
		Message:      message,
	}, noMatchMessage)
}

func (p *Player) ActUponAlias(action, targetAlias, noMatchMessage string) (string, error) {
	matches, err := p.getEntitiesByAlias(targetAlias)
	if err != nil {
		return "", fmt.Errorf("act upon get target for player '%s': %w", p.Name, err)
	}

	if len(matches) == 0 {
		return fmt.Sprintf("You wish to %s %s, but that's not here.", action, targetAlias), nil
	} else if len(matches) == 1 {
		return p.actUponEntity(action, matches[0].Entity, noMatchMessage)
	}

	slots := []entities.AmbiguitySlot{
		{
			Role:    entities.EventRoleTarget.String(),
			Prompt:  "Which target?",
			Matches: matches,
		},
	}

	return "", &entities.AmbiguityError{
		Slots: slots,
		Execute: func(inputMap map[string]*entities.Entity) (string, error) {
			t := inputMap[entities.EventRoleTarget.String()]
			return p.actUponEntity(action, t, noMatchMessage)
		},
	}
}

func (p *Player) actUponEntity(action string, target *entities.Entity, noMatchMessage string) (string, error) {
	return p.sendEventToEntity(target, &entities.Event{
		Type:         action,
		Publisher:    p.world,
		Scheduler:    p.world.GetScheduler(),
		EntitiesById: p.world.EntitiesById(),
		Room:         p.CurrentRoom,
		Source:       p.Entity,
		Target:       target,
	}, noMatchMessage)
}

func (p *Player) ActUponMessageAlias(action, targetAlias, message, noMatchMessage string) (string, error) {
	matches, err := p.getEntitiesByAlias(targetAlias)
	if err != nil {
		return "", fmt.Errorf("act upon message get target for player '%s': %w", p.Name, err)
	}

	if len(matches) == 0 {
		return fmt.Sprintf("You can't %s without %s here", action, targetAlias), nil
	} else if len(matches) == 1 {
		return p.actUponMessageEntity(action, matches[0].Entity, message, noMatchMessage)
	}

	slots := []entities.AmbiguitySlot{
		{
			Role:    entities.EventRoleTarget.String(),
			Prompt:  "Which target?",
			Matches: matches,
		},
	}

	return "", &entities.AmbiguityError{
		Slots: slots,
		Execute: func(inputMap map[string]*entities.Entity) (string, error) {
			t := inputMap[entities.EventRoleTarget.String()]
			return p.actUponMessageEntity(action, t, message, noMatchMessage)
		},
	}
}

func (p *Player) actUponMessageEntity(action string, target *entities.Entity, message, noMatchMessage string) (string, error) {
	return p.sendEventToEntity(target, &entities.Event{
		Type:         action,
		Publisher:    p.world,
		Scheduler:    p.world.GetScheduler(),
		EntitiesById: p.world.EntitiesById(),
		Room:         p.CurrentRoom,
		Source:       p.Entity,
		Target:       target,
		Message:      message,
	}, noMatchMessage)
}

func (p *Player) ActUponWithAlias(action, targetAlias, instrumentAlias, noMatchMessage string) (string, error) {
	// Build slots for any ambiguous pieces
	var slots []entities.AmbiguitySlot
	var target, instrument *entities.Entity

	targetMatches, err := p.getEntitiesByAlias(targetAlias)
	if err != nil {
		return "", fmt.Errorf("act upon with get target for player '%s': %w", p.Name, err)
	}
	if len(targetMatches) == 0 {
		return fmt.Sprintf("There is no %s here.", targetAlias), nil
	} else if len(targetMatches) == 1 {
		target = targetMatches[0].Entity
	} else {
		slots = append(slots, entities.AmbiguitySlot{
			Role:    entities.EventRoleTarget.String(),
			Prompt:  fmt.Sprintf("Which target to %s?", action),
			Matches: targetMatches,
		})
	}

	instrumentMatches, err := p.getEntitiesByAlias(instrumentAlias)
	if err != nil {
		return "", fmt.Errorf("act upon with get instrument for player '%s': %w", p.Name, err)
	}
	if len(instrumentMatches) == 0 {
		return fmt.Sprintf("You don't have %s available.", instrumentAlias), nil
	} else if len(instrumentMatches) == 1 {
		instrument = instrumentMatches[0].Entity
	} else {
		slots = append(slots, entities.AmbiguitySlot{
			Role:    entities.EventRoleInstrument.String(),
			Prompt:  fmt.Sprintf("Use what to %s?", action),
			Matches: instrumentMatches,
		})
	}

	if len(slots) == 0 {
		return p.actUponWithEntities(action, target, instrument, noMatchMessage)
	}

	return "", &entities.AmbiguityError{
		Slots: slots,
		Execute: func(inputMap map[string]*entities.Entity) (string, error) {
			t := target
			if t == nil {
				t = inputMap[entities.EventRoleTarget.String()]
			}

			i := instrument
			if i == nil {
				i = inputMap[entities.EventRoleInstrument.String()]
			}

			return p.actUponWithEntities(action, t, i, noMatchMessage)
		},
	}
}

func (p *Player) actUponWithEntities(action string, target, instrument *entities.Entity, noMatchMessage string) (string, error) {
	return p.sendEventToEntity(target, &entities.Event{
		Type:         action,
		Publisher:    p.world,
		Scheduler:    p.world.GetScheduler(),
		EntitiesById: p.world.EntitiesById(),
		Room:         p.CurrentRoom,
		Source:       p.Entity,
		Instrument:   instrument,
		Target:       target,
	}, noMatchMessage)
}

func (p *Player) sendEventToEntity(entity *entities.Entity, event *entities.Event, noMatchMessage string) (string, error) {
	if entity == nil {
		return "", fmt.Errorf("player '%s' send event nil entity", p.Name)
	}

	if eventful, ok := entities.GetComponent[*components.Eventful](entity); ok {

		match, err := eventful.OnEvent(event)
		if err != nil {
			return "", fmt.Errorf("player '%s' send event to '%s' on event error: %w", p.Name, entity.Name, err)
		}

		if match {
			return "", nil
		}
	}

	message, err := entities.FormatEventMessage(noMatchMessage, event)
	if err != nil {
		return "", fmt.Errorf("player '%s' send event to '%s' no match format: %w", p.Name, entity.Name, err)
	}

	return message, nil
}

func (p *Player) getEntitiesByAlias(alias string) ([]entities.AmbiguityOption, error) {
	eMatches := make([]entities.AmbiguityOption, 0, 10)

	// check if the room itself has a matching alias
	if slices.Contains(p.CurrentRoom.Aliases, alias) {
		eMatches = append(eMatches, entities.AmbiguityOption{
			Text:   fmt.Sprintf("The room: %s", p.CurrentRoom.Name),
			Entity: p.CurrentRoom,
		})
	}

	// look for matches in the room
	room, err := entities.RequireComponent[*components.Room](p.CurrentRoom)
	if err != nil {
		return nil, fmt.Errorf("getEntityByAlias for player '%s': %w", p.Name, err)
	} else {
		if cMatches := room.GetChildren().GetChildrenByAlias(alias); len(cMatches) > 0 {
			eMatches = append(eMatches, cMatches...)
		}
	}

	// look for matches in the player's inventory
	if inventory, ok := entities.GetComponent[*components.Inventory](p.Entity); ok {
		if iMatches := inventory.GetChildren().GetChildrenByAlias(alias); len(iMatches) > 0 {
			eMatches = append(eMatches, iMatches...)
		}
	}

	return eMatches, nil
}
