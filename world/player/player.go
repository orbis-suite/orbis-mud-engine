package player

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"example.com/mud/world/entities"
	"example.com/mud/world/scheduler"
)

var safeNameRegex = regexp.MustCompile(`[^a-zA-Z]+`)

type Player struct {
	Name    string
	Entity  *entities.Entity
	Pending *entities.PendingAction

	mu            sync.Mutex
	nextActionAt  time.Time
	trackingAlias string
	world         World
}

type World interface {
	EntitiesById() map[string]*entities.Entity
	GetEntityById(id string) (*entities.Entity, bool)

	Publish(room *entities.Entity, text string, exclude []*entities.Entity)
	PublishTo(recipient *entities.Entity, text string)
	Move(toRoom *entities.Entity, player *entities.Entity)

	GetScheduler() *scheduler.Scheduler
}

func NewPlayer(name string, world World, parentRoom *entities.Entity) (*Player, error) {
	playerTemplate, ok := world.GetEntityById("player")
	if !ok {
		return nil, fmt.Errorf("entity with ID 'Player' does not exist in world")
	}

	playerEntity := playerTemplate.Copy(parentRoom)
	playerEntity.Name = name
	playerEntity.Description = fmt.Sprintf("%s the brave hero is here.", name)
	playerEntity.Aliases = []string{strings.ToLower(name)}

	return &Player{
		Name:   name,
		Entity: playerEntity,
		world:  world,
	}, nil
}

func (p *Player) Init() {
	p.Entity.InitFunc(&entities.Event{
		Type:         "init",
		Publisher:    p.world,
		Scheduler:    p.world.GetScheduler(),
		EntitiesById: p.world.EntitiesById(),
		Room:         p.Entity.Parent,
		Source:       p.Entity,
	})
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
	// var b strings.Builder

	// room, err := entities.RequireComponent[*components.Room](p.CurrentRoom)
	// if err != nil {
	// 	return "", err
	// }

	// formattedTitle, err := utils.FormatText(fmt.Sprintf("{'%s' | bold | red}", p.CurrentRoom.Name), map[string]string{})
	// if err != nil {
	// 	return "", fmt.Errorf("could not format room '%s' name: %w", p.CurrentRoom.Name, err)
	// }

	// b.WriteString(formattedTitle)
	// b.WriteString("\n")

	// roomDescription := strings.TrimSpace(p.CurrentRoom.Description)
	// b.WriteString(roomDescription)
	// b.WriteString("\n")

	// for _, e := range room.GetChildren().GetChildren() {
	// 	if e == p.Entity {
	// 		continue
	// 	}

	// 	description, err := e.GetDescription()
	// 	if err != nil {
	// 		return "", err
	// 	}

	// 	b.WriteString(fmt.Sprintf("%s%s", models.Tab, description))
	// 	b.WriteString("\n")
	// }

	// b.WriteString("\n")
	// b.WriteString(room.GetExitText())
	// return b.String(), nil

	return p.Entity.Parent.Description, nil
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
	// if inventory, ok := entities.GetComponent[*components.Inventory](p.Entity); ok {
	// 	message, err := inventory.Print()
	// 	if err != nil {
	// 		return "", fmt.Errorf("inventory print for player '%s': %w", p.Name, err)
	// 	}
	// 	return message, nil
	// }
	// return "You couldn't possibly carry anything at all.", nil
	return "", nil
}

func (p *Player) Act(action string, parameters map[string]string, noMatchMessage string) (string, error) {
	return p.sendEventToEntities(&entities.Event{
		Type:              action,
		Publisher:         p.world,
		Scheduler:         p.world.GetScheduler(),
		EntitiesById:      p.world.EntitiesById(),
		CommandParameters: parameters,
		Room:              p.Entity.Parent,
		Source:            p.Entity,
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
	return p.sendEventToEntities(&entities.Event{
		Type:         action,
		Publisher:    p.world,
		Scheduler:    p.world.GetScheduler(),
		EntitiesById: p.world.EntitiesById(),
		Room:         p.Entity.Parent,
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
	return p.sendEventToEntities(&entities.Event{
		Type:         action,
		Publisher:    p.world,
		Scheduler:    p.world.GetScheduler(),
		EntitiesById: p.world.EntitiesById(),
		// TODO add command parameters
		Room:   p.Entity.Parent,
		Source: p.Entity,
		Target: target,
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
	return p.sendEventToEntities(&entities.Event{
		Type:         action,
		Publisher:    p.world,
		Scheduler:    p.world.GetScheduler(),
		EntitiesById: p.world.EntitiesById(),
		Room:         p.Entity.Parent,
		Source:       p.Entity,
		Instrument:   instrument,
		Target:       target,
	}, noMatchMessage)
}

func (p *Player) sendEventToEntities(event *entities.Event, noMatchMessage string) (string, error) {
	match := false

	if source := event.Source; source != nil {
		sourceMatch := p.sendEventToEntity(source, entities.EventRoleSource, event)
		match = match || sourceMatch
	}

	if instrument := event.Instrument; instrument != nil {
		sourceMatch := p.sendEventToEntity(instrument, entities.EventRoleInstrument, event)
		match = match || sourceMatch
	}

	if target := event.Target; target != nil {
		sourceMatch := p.sendEventToEntity(target, entities.EventRoleTarget, event)
		match = match || sourceMatch
	}

	if match {
		return "", nil
	}

	message, err := entities.FormatEventMessage(noMatchMessage, event)
	if err != nil {
		return "", fmt.Errorf("player '%s' send event to '%s' no match format: %w", p.Name, p.Entity.Name, err)
	}

	return message, nil
}

func (p *Player) sendEventToEntity(entity *entities.Entity, role entities.EventRole, event *entities.Event) bool {
	reaction, ok := entity.GetReaction(event.Type, role)
	if !ok {
		return false
	}

	reaction(event)
	return true
}

func (p *Player) getEntitiesByAlias(alias string) ([]entities.AmbiguityOption, error) {
	eMatches := make([]entities.AmbiguityOption, 0, 10)
	room := p.Entity.Parent

	// check if the room itself has a matching alias
	if slices.Contains(room.Aliases, alias) {
		eMatches = append(eMatches, entities.AmbiguityOption{
			Text:   fmt.Sprintf("The room: %s", p.Entity.Parent.Name),
			Entity: p.Entity.Parent,
		})
	}

	eMatches = append(
		eMatches,
		room.GetChildrenByAlias(alias)...,
	)

	return eMatches, nil
}
