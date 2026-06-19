package plugin

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/hashicorp/go-plugin"

	"example.com/mud/models"
	"example.com/mud/world/entities"
	"example.com/mud/world/entities/components"
	pb "example.com/mud/plugin/proto"
)

// Launch starts the game binary as a plugin and returns a client + cleanup func.
func Launch(binaryPath string) (GameClient, func(), error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  Handshake,
		Plugins:          PluginMap,
		Cmd:              exec.Command(binaryPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("launch game plugin: %w", err)
	}

	raw, err := rpcClient.Dispense("game")
	if err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("dispense game plugin: %w", err)
	}

	gameClient, ok := raw.(GameClient)
	if !ok {
		client.Kill()
		return nil, nil, fmt.Errorf("dispensed plugin does not implement GameClient")
	}

	return gameClient, client.Kill, nil
}

// ManifestToWorld converts a GameManifest into the engine's entity map and command list.
func ManifestToWorld(manifest *pb.GameManifest, client GameClient) (map[string]*entities.Entity, []*models.CommandDefinition, error) {
	entityMap := make(map[string]*entities.Entity)

	// First pass: create all entity stubs (without children resolved)
	for _, ed := range manifest.GetEntities() {
		e := buildEntity(ed, client)
		entityMap[ed.Id] = e
	}

	// Create all room entities
	for _, rd := range manifest.GetRooms() {
		e := buildRoomEntity(rd, client)
		entityMap[rd.Id] = e
	}

	// Second pass: resolve children
	// Room children
	for _, rd := range manifest.GetRooms() {
		roomEntity := entityMap[rd.Id]
		roomComp, err := entities.RequireComponent[*components.Room](roomEntity)
		if err != nil {
			return nil, nil, fmt.Errorf("room %q missing Room component: %w", rd.Id, err)
		}
		for _, childID := range rd.ChildIds {
			childEntity, ok := entityMap[childID]
			if !ok {
				return nil, nil, fmt.Errorf("room %q references unknown child %q", rd.Id, childID)
			}
			if err := roomComp.AddChild(childEntity); err != nil {
				return nil, nil, fmt.Errorf("add child %q to room %q: %w", childID, rd.Id, err)
			}
		}
	}

	// Entity container/inventory children
	for _, ed := range manifest.GetEntities() {
		if ed.ContainerId == "" {
			continue
		}
		parentEntity, ok := entityMap[ed.ContainerId]
		if !ok {
			return nil, nil, fmt.Errorf("entity %q references unknown parent %q", ed.Id, ed.ContainerId)
		}
		childEntity := entityMap[ed.Id]

		compType, err := entities.ParseComponentType(ed.ContainerComponent)
		if err != nil {
			return nil, nil, fmt.Errorf("entity %q: %w", ed.Id, err)
		}

		parent, err := parentEntity.RequireComponentWithChildren(compType)
		if err != nil {
			return nil, nil, fmt.Errorf("entity %q parent %q missing component %q: %w", ed.Id, ed.ContainerId, ed.ContainerComponent, err)
		}

		if err := parent.AddChild(childEntity); err != nil {
			return nil, nil, fmt.Errorf("add %q to %q: %w", ed.Id, ed.ContainerId, err)
		}
	}

	// Convert commands
	cmds := make([]*models.CommandDefinition, 0, len(manifest.GetCommands()))
	for _, cd := range manifest.GetCommands() {
		cmd := &models.CommandDefinition{
			Name:    cd.Name,
			Aliases: cd.Aliases,
		}
		for _, pat := range cd.Patterns {
			tokens := models.ParseSyntax(pat.Syntax)
			cmd.Patterns = append(cmd.Patterns, models.CommandPattern{
				Tokens:         tokens,
				NoMatchMessage: pat.NoMatch,
				HelpMessage:    pat.Help,
			})
		}
		cmds = append(cmds, cmd)
	}

	return entityMap, cmds, nil
}

func buildEntity(ed *pb.EntityDef, client GameClient) *entities.Entity {
	fields := decodeEntityFields(ed.Fields)

	e := entities.NewEntity(ed.Name, ed.Description, ed.Aliases, ed.Tags, fields, nil)

	// Add plugin eventful for reaction handling
	e.Add(&PluginEventful{TemplateID: ed.Id, Client: client})

	// Add optional components
	if ed.HasInventory {
		e.Add(components.NewInventory())
	}
	if ed.HasContainer {
		c := components.NewContainer()
		c.GetChildren().SetPrefix(ed.ContainerPrefix)
		c.GetChildren().SetRevealed(ed.ContainerRevealed)
		e.Add(c)
	}

	return e
}

func buildRoomEntity(rd *pb.RoomDef, client GameClient) *entities.Entity {
	e := entities.NewEntity(rd.Name, rd.Description, []string{"room"}, []string{"room"}, map[string]models.Value{}, nil)

	room := components.NewRoom()
	room.MapIcon = rd.Icon
	if room.MapIcon == "" {
		room.MapIcon = "O"
	}
	room.MapColor = rd.Color
	room.Exits = rd.Exits

	e.Add(room)

	// Rooms can also have plugin eventful if needed
	e.Add(&PluginEventful{TemplateID: rd.Id, Client: client})

	return e
}

func decodeEntityFields(raw map[string]string) map[string]models.Value {
	out := make(map[string]models.Value, len(raw))
	for k, v := range raw {
		// Try bool first (before int, since "true"/"false" could technically parse as int in some libs)
		if b, err := strconv.ParseBool(v); err == nil {
			out[k] = models.VBool(b)
			continue
		}
		// Try int
		if i, err := strconv.Atoi(v); err == nil {
			out[k] = models.VInt(i)
			continue
		}
		// Default to string
		out[k] = models.VStr(v)
	}
	return out
}

// StartEventStream starts a goroutine that consumes proactive actions from the game binary.
func StartEventStream(ctx context.Context, client GameClient, execAction func([]*pb.Action) error) {
	// The bidirectional stream is optional for Phase 1.
	// Game binaries that don't need proactive events simply don't use it.
	_ = ctx
	_ = client
	_ = execAction
}
