package sdk

import (
	"strconv"

	pb "example.com/mud/plugin/proto"
)

// Event is passed to HandleEvent with a snapshot of the current game state.
type Event struct {
	Command    string
	TargetID   string
	Source     *EntitySnapshot
	Target     *EntitySnapshot
	Instrument *EntitySnapshot
	Room       *EntitySnapshot
	Message    string
}

// EntitySnapshot is a read-only view of an entity's current state.
type EntitySnapshot struct {
	TemplateID  string
	Name        string
	Description string
	Aliases     []string
	Tags        []string
	fields      map[string]string
	Children    []*ChildRef
}

type ChildRef struct {
	TemplateID string
	Name       string
	Tags       []string
	Component  string // "Room", "Inventory", "Container"
}

func (s *EntitySnapshot) HasTag(tag string) bool {
	for _, t := range s.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (s *EntitySnapshot) FieldInt(name string) int {
	if s == nil {
		return 0
	}
	i, _ := strconv.Atoi(s.fields[name])
	return i
}

func (s *EntitySnapshot) FieldString(name string) string {
	if s == nil {
		return ""
	}
	return s.fields[name]
}

func (s *EntitySnapshot) FieldBool(name string) bool {
	if s == nil {
		return false
	}
	b, _ := strconv.ParseBool(s.fields[name])
	return b
}

// HasChildInComponent returns true if any child with the given template ID exists in the given component.
func (s *EntitySnapshot) HasChildInComponent(templateID, component string) bool {
	if s == nil {
		return false
	}
	for _, c := range s.Children {
		if c.TemplateID == templateID && c.Component == component {
			return true
		}
	}
	return false
}

func eventFromProto(req *pb.EventRequest) *Event {
	return &Event{
		Command:    req.Command,
		TargetID:   req.TargetId,
		Source:     snapshotFromProto(req.Source),
		Target:     snapshotFromProto(req.Target),
		Instrument: snapshotFromProto(req.Instrument),
		Room:       snapshotFromProto(req.Room),
		Message:    req.Message,
	}
}

func snapshotFromProto(s *pb.EntitySnapshot) *EntitySnapshot {
	if s == nil {
		return nil
	}

	children := make([]*ChildRef, 0, len(s.Children))
	for _, c := range s.Children {
		children = append(children, &ChildRef{
			TemplateID: c.TemplateId,
			Name:       c.Name,
			Tags:       c.Tags,
			Component:  c.Component,
		})
	}

	return &EntitySnapshot{
		TemplateID:  s.TemplateId,
		Name:        s.Name,
		Description: s.Description,
		Aliases:     s.Aliases,
		Tags:        s.Tags,
		fields:      s.Fields,
		Children:    children,
	}
}
