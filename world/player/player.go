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
	"example.com/mud/world/response"
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
	MovePlayer(p *Player, direction string) (response.Response, error)

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

func (p *Player) OpeningMessage() (response.RoomDescription, error) {
	room, err := p.GetRoomDescription()
	if err != nil {
		return response.RoomDescription{}, fmt.Errorf("opening message for player '%s': %w", p.Name, err)
	}
	return room, nil
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

func (p *Player) GetRoomDescription() (response.RoomDescription, error) {
	room, err := entities.RequireComponent[*components.Room](p.CurrentRoom)
	if err != nil {
		return response.RoomDescription{}, err
	}

	// Collect child summaries (skip the player entity itself)
	children := make([]response.ChildSummary, 0)
	for _, e := range room.GetChildren().GetChildren() {
		if e == p.Entity {
			continue
		}
		children = append(children, response.ChildSummary{
			Name:        e.Name,
			Description: e.Description,
		})
	}
	slices.SortFunc(children, func(a, b response.ChildSummary) int {
		return strings.Compare(a.Name, b.Name)
	})

	// Collect exit directions
	exits := make([]string, 0, len(room.Exits))
	for dir := range room.Exits {
		exits = append(exits, dir)
	}
	slices.Sort(exits)

	return response.RoomDescription{
		Name:        p.CurrentRoom.Name,
		Description: strings.TrimSpace(p.CurrentRoom.Description),
		Exits:       exits,
		Children:    children,
	}, nil
}

func (p *Player) Move(direction string) (response.Response, error) {
	return p.world.MovePlayer(p, direction)
}

func (p *Player) Look(alias string) (response.Response, error) {
	if alias == "" {
		room, err := p.GetRoomDescription()
		if err != nil {
			return nil, fmt.Errorf("look room for player '%s': %w", p.Name, err)
		}
		return response.Text{Value: room.String()}, nil
	}

	matches, err := p.getEntitiesByAlias(alias)
	if err != nil {
		return nil, fmt.Errorf("get look target for player '%s': %w", p.Name, err)
	}

	if len(matches) == 0 {
		return response.Text{Value: fmt.Sprintf("There is no %s for you to look upon.", alias)}, nil
	} else if len(matches) == 1 {
		return entityDescription(matches[0].Entity)
	}

	slots := []entities.AmbiguitySlot{
		{
			Role:    entities.EventRoleTarget.String(),
			Prompt:  "Which target?",
			Matches: matches,
		},
	}

	return nil, &entities.AmbiguityError{
		Slots: slots,
		Execute: func(inputMap map[string]*entities.Entity) (string, error) {
			t := inputMap[entities.EventRoleTarget.String()]
			desc, err := t.GetDescription()
			return desc, err
		},
	}
}

// entityDescription builds an EntityDescription response from an entity.
func entityDescription(e *entities.Entity) (response.EntityDescription, error) {
	var children []response.ChildSummary
	for _, cwc := range e.GetComponentsWithChildren() {
		if !cwc.GetChildren().GetRevealed() {
			continue
		}
		for _, child := range cwc.GetChildren().GetChildren() {
			children = append(children, response.ChildSummary{
				Name:        child.Name,
				Description: child.Description,
			})
		}
	}
	return response.EntityDescription{
		Name:        e.Name,
		Description: e.Description,
		Children:    children,
	}, nil
}

func (p *Player) Inventory() (response.InventoryList, error) {
	if inventory, ok := entities.GetComponent[*components.Inventory](p.Entity); ok {
		var items []string
		for _, child := range inventory.GetChildren().GetChildren() {
			if child.Name != "" {
				items = append(items, child.Name)
			}
		}
		if items == nil {
			items = []string{}
		}
		return response.InventoryList{Items: items}, nil
	}
	return response.InventoryList{Items: []string{}}, nil
}

func (p *Player) ActMessage(action, message, noMatchMessage string) (response.Text, error) {
	str, err := p.sendEventToEntity(p.Entity, &entities.Event{
		Type:         action,
		Publisher:    p.world,
		Scheduler:    p.world.GetScheduler(),
		EntitiesById: p.world.EntitiesById(),
		Room:         p.CurrentRoom,
		Source:       p.Entity,
		Message:      message,
	}, noMatchMessage)
	return response.Text{Value: str}, err
}

func (p *Player) ActUponAlias(action, targetAlias, noMatchMessage string) (response.Text, error) {
	matches, err := p.getEntitiesByAlias(targetAlias)
	if err != nil {
		return response.Text{}, fmt.Errorf("act upon get target for player '%s': %w", p.Name, err)
	}

	if len(matches) == 0 {
		return response.Text{Value: fmt.Sprintf("You wish to %s %s, but that's not here.", action, targetAlias)}, nil
	} else if len(matches) == 1 {
		str, err := p.actUponEntity(action, matches[0].Entity, noMatchMessage)
		return response.Text{Value: str}, err
	}

	slots := []entities.AmbiguitySlot{
		{
			Role:    entities.EventRoleTarget.String(),
			Prompt:  "Which target?",
			Matches: matches,
		},
	}

	return response.Text{}, &entities.AmbiguityError{
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

func (p *Player) ActUponMessageAlias(action, targetAlias, message, noMatchMessage string) (response.Text, error) {
	matches, err := p.getEntitiesByAlias(targetAlias)
	if err != nil {
		return response.Text{}, fmt.Errorf("act upon message get target for player '%s': %w", p.Name, err)
	}

	if len(matches) == 0 {
		return response.Text{Value: fmt.Sprintf("You can't %s without %s here", action, targetAlias)}, nil
	} else if len(matches) == 1 {
		str, err := p.actUponMessageEntity(action, matches[0].Entity, message, noMatchMessage)
		return response.Text{Value: str}, err
	}

	slots := []entities.AmbiguitySlot{
		{
			Role:    entities.EventRoleTarget.String(),
			Prompt:  "Which target?",
			Matches: matches,
		},
	}

	return response.Text{}, &entities.AmbiguityError{
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

func (p *Player) ActUponWithAlias(action, targetAlias, instrumentAlias, noMatchMessage string) (response.Text, error) {
	// Build slots for any ambiguous pieces
	var slots []entities.AmbiguitySlot
	var target, instrument *entities.Entity

	targetMatches, err := p.getEntitiesByAlias(targetAlias)
	if err != nil {
		return response.Text{}, fmt.Errorf("act upon with get target for player '%s': %w", p.Name, err)
	}
	if len(targetMatches) == 0 {
		return response.Text{Value: fmt.Sprintf("There is no %s here.", targetAlias)}, nil
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
		return response.Text{}, fmt.Errorf("act upon with get instrument for player '%s': %w", p.Name, err)
	}
	if len(instrumentMatches) == 0 {
		return response.Text{Value: fmt.Sprintf("You don't have %s available.", instrumentAlias)}, nil
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
		str, err := p.actUponWithEntities(action, target, instrument, noMatchMessage)
		return response.Text{Value: str}, err
	}

	return response.Text{}, &entities.AmbiguityError{
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

	if reactor, ok := entity.GetReactor(); ok {
		match, err := reactor.OnEvent(event)
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

func (p *Player) Track(alias string) (response.Text, error) {
	p.trackingAlias = alias
	return response.Text{Value: fmt.Sprintf(`Rooms with "%s" will now appear as "!" on your map.`, alias)}, nil
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

// renderForTelnet converts a Response to a plain-text string for the telnet client.
// ANSI formatting is applied here so the wire format stays clean.
func RenderForTelnet(r response.Response) (string, error) {
	switch v := r.(type) {
	case response.RoomDescription:
		var b strings.Builder
		title, err := utils.FormatText(fmt.Sprintf("{'%s' | bold | red}", v.Name), map[string]string{})
		if err != nil {
			return "", fmt.Errorf("format room title: %w", err)
		}
		b.WriteString(title)
		b.WriteByte('\n')
		b.WriteString(v.Description)
		b.WriteByte('\n')
		for _, child := range v.Children {
			formatted, err := utils.FormatText(child.Description, map[string]string{})
			if err != nil {
				return "", fmt.Errorf("format child description: %w", err)
			}
			b.WriteString(fmt.Sprintf("%s%s\n", models.Tab, formatted))
		}
		b.WriteByte('\n')
		if len(v.Exits) > 0 {
			b.WriteString("Exits: ")
			b.WriteString(strings.Join(v.Exits, ", "))
		}
		return b.String(), nil

	case response.EntityDescription:
		var b strings.Builder
		formatted, err := utils.FormatText(v.Description, map[string]string{})
		if err != nil {
			return "", fmt.Errorf("format entity description: %w", err)
		}
		b.WriteString(fmt.Sprintf("- %s", formatted))
		for _, child := range v.Children {
			cf, err := utils.FormatText(child.Description, map[string]string{})
			if err != nil {
				return "", fmt.Errorf("format child description: %w", err)
			}
			b.WriteString(fmt.Sprintf("\n  - %s", cf))
		}
		return b.String(), nil

	case response.InventoryList:
		if len(v.Items) == 0 {
			return "You are carrying: []", nil
		}
		return fmt.Sprintf("You are carrying: [%s]", strings.Join(v.Items, ", ")), nil

	case response.MapView:
		var b strings.Builder
		for _, row := range v.Grid {
			for _, cell := range row {
				b.WriteString(cell.Icon)
			}
			b.WriteByte('\n')
		}
		return b.String(), nil

	case response.Text:
		return v.Value, nil

	default:
		return fmt.Sprintf("%+v", v), nil
	}
}
