package plugin

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"example.com/mud/models"
	pb "example.com/mud/plugin/proto"
	"example.com/mud/world/entities"
	"example.com/mud/world/entities/actions"
	"example.com/mud/world/entities/expressions"
)

// PluginEventful delegates all event handling to the game plugin via gRPC.
type PluginEventful struct {
	TemplateID string
	Client     GameClient
}

var _ entities.Component = &PluginEventful{}
var _ entities.Reactor = &PluginEventful{}

func (p *PluginEventful) Id() entities.ComponentType {
	return entities.ComponentEventful
}

func (p *PluginEventful) Copy() entities.Component {
	return &PluginEventful{
		TemplateID: p.TemplateID,
		Client:     p.Client,
	}
}

func (p *PluginEventful) OnEvent(ev *entities.Event) (bool, error) {
	req := buildEventRequest(p.TemplateID, ev)

	resp, err := p.Client.HandleEvent(context.Background(), req)
	if err != nil {
		return false, fmt.Errorf("plugin handle event: %w", err)
	}

	if len(resp.GetActions()) == 0 {
		return false, nil
	}

	if err := executeProtoActions(resp.GetActions(), ev); err != nil {
		return false, err
	}

	return true, nil
}

// buildEventRequest serializes an Event into a proto EventRequest.
func buildEventRequest(targetID string, ev *entities.Event) *pb.EventRequest {
	req := &pb.EventRequest{
		Command:  ev.Type,
		TargetId: targetID,
		Message:  ev.Message,
	}

	if ev.Source != nil {
		req.Source = snapshotEntity(ev.Source, ev)
	}
	if ev.Target != nil {
		req.Target = snapshotEntity(ev.Target, ev)
	}
	if ev.Instrument != nil {
		req.Instrument = snapshotEntity(ev.Instrument, ev)
	}
	if ev.Room != nil {
		req.Room = snapshotEntity(ev.Room, ev)
	}

	return req
}

func snapshotEntity(e *entities.Entity, ev *entities.Event) *pb.EntitySnapshot {
	snap := &pb.EntitySnapshot{
		TemplateId:  templateIDOf(e, ev),
		Name:        e.Name,
		Description: e.Description,
		Aliases:     e.Aliases,
		Tags:        e.Tags,
		Fields:      encodeFields(e.Fields),
	}

	for _, cwc := range e.GetComponentsWithChildren() {
		compName := cwc.(entities.Component).Id().String()
		for _, child := range cwc.GetChildren().GetChildren() {
			snap.Children = append(snap.Children, &pb.ChildRef{
				TemplateId: templateIDOf(child, ev),
				Name:       child.Name,
				Tags:       child.Tags,
				Component:  compName,
			})
		}
	}

	return snap
}

// templateIDOf finds the template ID for an entity by reverse lookup in EntitiesById.
// For player entities not in the map, falls back to the entity name.
func templateIDOf(e *entities.Entity, ev *entities.Event) string {
	for id, ent := range ev.EntitiesById {
		if ent == e {
			return id
		}
	}
	return e.Name
}

func encodeFields(fields map[string]models.Value) map[string]string {
	out := make(map[string]string, len(fields))
	for k, v := range fields {
		switch v.K {
		case models.KindInt:
			out[k] = strconv.Itoa(v.I)
		case models.KindString:
			out[k] = v.S
		case models.KindBool:
			out[k] = strconv.FormatBool(v.B)
		}
	}
	return out
}

// executeProtoActions converts proto actions to engine actions and executes them.
func executeProtoActions(protoActions []*pb.Action, ev *entities.Event) error {
	for _, pa := range protoActions {
		a, err := protoActionToEngineAction(pa, ev)
		if err != nil {
			return err
		}
		if a != nil {
			if err := a.Execute(ev); err != nil {
				return fmt.Errorf("execute action: %w", err)
			}
		}
	}
	return nil
}

func protoActionToEngineAction(pa *pb.Action, ev *entities.Event) (entities.Action, error) {
	switch kind := pa.Kind.(type) {
	case *pb.Action_Print:
		role, err := parseRole(kind.Print.Role)
		if err != nil {
			return nil, err
		}
		return &actions.Print{Text: kind.Print.Message, EventRole: role}, nil

	case *pb.Action_Publish:
		// The engine's Publish action always excludes source, instrument, and target.
		// We use it as-is; the exclude_roles field from the proto is advisory only.
		return &actions.Publish{Text: kind.Publish.Message}, nil

	case *pb.Action_Move:
		entityRole, err := parseRole(kind.Move.EntityRole)
		if err != nil {
			return nil, err
		}
		destRole, err := parseRole(kind.Move.DestRole)
		if err != nil {
			return nil, err
		}
		destComp, err := entities.ParseComponentType(kind.Move.DestComponent)
		if err != nil {
			return nil, err
		}
		return &actions.Move{
			RoleObject:      entityRole,
			RoleDestination: destRole,
			ComponentType:   destComp,
		}, nil

	case *pb.Action_SetField:
		role, err := parseRole(kind.SetField.Role)
		if err != nil {
			return nil, err
		}
		val, err := decodeFieldValue(kind.SetField.Value, kind.SetField.ValueType)
		if err != nil {
			return nil, err
		}
		return &actions.SetField{
			Role:       role,
			Field:      kind.SetField.Field,
			Expression: &expressions.ExpressionConst{V: val},
		}, nil

	case *pb.Action_Destroy:
		role, err := parseRole(kind.Destroy.Role)
		if err != nil {
			return nil, err
		}
		return &actions.Destroy{Role: role}, nil

	case *pb.Action_Spawn:
		destRole, err := parseRole(kind.Spawn.DestRole)
		if err != nil {
			return nil, err
		}
		destComp, err := entities.ParseComponentType(kind.Spawn.DestComponent)
		if err != nil {
			return nil, err
		}
		return &actions.Copy{
			EntityId:      kind.Spawn.TemplateId,
			EventRole:     destRole,
			ComponentType: destComp,
		}, nil

	case *pb.Action_After:
		childActions := make([]entities.Action, 0, len(kind.After.Actions))
		for _, ca := range kind.After.Actions {
			a, err := protoActionToEngineAction(ca, ev)
			if err != nil {
				return nil, err
			}
			if a != nil {
				childActions = append(childActions, a)
			}
		}
		return &actions.ScheduleOnce{
			Nanoseconds: time.Duration(kind.After.DelayMs) * time.Millisecond,
			Actions:     childActions,
		}, nil

	case *pb.Action_Reveal:
		role, err := parseRole(kind.Reveal.Role)
		if err != nil {
			return nil, err
		}
		comp, err := entities.ParseComponentType(kind.Reveal.Component)
		if err != nil {
			return nil, err
		}
		return &actions.RevealChildren{Role: role, ComponentType: comp, Reveal: true}, nil

	case *pb.Action_Hide:
		role, err := parseRole(kind.Hide.Role)
		if err != nil {
			return nil, err
		}
		comp, err := entities.ParseComponentType(kind.Hide.Component)
		if err != nil {
			return nil, err
		}
		return &actions.RevealChildren{Role: role, ComponentType: comp, Reveal: false}, nil
	}

	return nil, fmt.Errorf("unknown action kind: %T", pa.Kind)
}

func parseRole(s string) (entities.EventRole, error) {
	return entities.ParseEventRole(s)
}

func decodeFieldValue(value, valueType string) (models.Value, error) {
	switch valueType {
	case "int":
		i, err := strconv.Atoi(value)
		if err != nil {
			return models.Value{}, fmt.Errorf("decode int field: %w", err)
		}
		return models.VInt(i), nil
	case "string":
		return models.VStr(value), nil
	case "bool":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return models.Value{}, fmt.Errorf("decode bool field: %w", err)
		}
		return models.VBool(b), nil
	default:
		return models.Value{}, fmt.Errorf("unknown value type: %q", valueType)
	}
}
