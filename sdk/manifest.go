package sdk

import pb "example.com/mud/plugin/proto"

// Command is a type alias for a command name string.
type Command = string

type Manifest struct {
	StartingRoom string
	Rooms        []*RoomDef
	Entities     []*EntityDef
	Commands     []*CommandDef
}

type RoomDef struct {
	ID          string
	Name        string
	Description string
	Icon        string
	Color       string
	Exits       map[string]string
	ChildIDs    []string
}

type EntityDef struct {
	ID                 string
	Name               string
	Description        string
	Aliases            []string
	Tags               []string
	Fields             map[string]string // field_name → encoded value ("42", "true", "hello")
	ContainerID        string
	ContainerComponent string
	HasInventory       bool
	HasContainer       bool
	ContainerPrefix    string
	ContainerRevealed  bool
	Reactions          map[Command][]Action
}

type CommandDef struct {
	Name     string
	Aliases  []string
	Patterns []CommandPattern
}

type CommandPattern struct {
	Syntax  string
	NoMatch string
	Help    string
}

func (m *Manifest) toProto() *pb.GameManifest {
	rooms := make([]*pb.RoomDef, 0, len(m.Rooms))
	for _, r := range m.Rooms {
		rooms = append(rooms, r.toProto())
	}

	ents := make([]*pb.EntityDef, 0, len(m.Entities))
	for _, e := range m.Entities {
		ents = append(ents, e.toProto())
	}

	cmds := make([]*pb.CommandDef, 0, len(m.Commands))
	for _, c := range m.Commands {
		cmds = append(cmds, c.toProto())
	}

	return &pb.GameManifest{
		StartingRoom: m.StartingRoom,
		Rooms:        rooms,
		Entities:     ents,
		Commands:     cmds,
	}
}

func (r *RoomDef) toProto() *pb.RoomDef {
	return &pb.RoomDef{
		Id:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Icon:        r.Icon,
		Color:       r.Color,
		Exits:       r.Exits,
		ChildIds:    r.ChildIDs,
	}
}

func (e *EntityDef) toProto() *pb.EntityDef {
	return &pb.EntityDef{
		Id:                 e.ID,
		Name:               e.Name,
		Description:        e.Description,
		Aliases:            e.Aliases,
		Tags:               e.Tags,
		Fields:             e.Fields,
		ContainerId:        e.ContainerID,
		ContainerComponent: e.ContainerComponent,
		HasInventory:       e.HasInventory,
		HasContainer:       e.HasContainer,
		ContainerPrefix:    e.ContainerPrefix,
		ContainerRevealed:  e.ContainerRevealed,
	}
}

func (c *CommandDef) toProto() *pb.CommandDef {
	pats := make([]*pb.CommandPattern, 0, len(c.Patterns))
	for _, p := range c.Patterns {
		pats = append(pats, &pb.CommandPattern{
			Syntax:  p.Syntax,
			NoMatch: p.NoMatch,
			Help:    p.Help,
		})
	}
	return &pb.CommandDef{
		Name:     c.Name,
		Aliases:  c.Aliases,
		Patterns: pats,
	}
}
