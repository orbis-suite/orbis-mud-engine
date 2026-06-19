package sdk

import (
	"fmt"
	"strconv"

	pb "example.com/mud/plugin/proto"
)

// Action is a game action to be executed by the engine.
type Action interface {
	toProto() *pb.Action
}

// ── Print ────────────────────────────────────────────────────────────────────

type printAction struct{ role, message string }

func Print(role, message string) Action { return &printAction{role, message} }

func (a *printAction) toProto() *pb.Action {
	return &pb.Action{Kind: &pb.Action_Print{Print: &pb.PrintAction{Role: a.role, Message: a.message}}}
}

// ── Publish ──────────────────────────────────────────────────────────────────

type publishAction struct {
	message      string
	excludeRoles []string
}

func Publish(message string, excludeRoles ...string) Action {
	return &publishAction{message, excludeRoles}
}

func (a *publishAction) toProto() *pb.Action {
	return &pb.Action{Kind: &pb.Action_Publish{Publish: &pb.PublishAction{
		Message:      a.message,
		ExcludeRoles: a.excludeRoles,
	}}}
}

// ── Move ─────────────────────────────────────────────────────────────────────

type moveAction struct {
	entityRole, destRole, destComponent string
}

func Move(entityRole, destRole, destComponent string) Action {
	return &moveAction{entityRole, destRole, destComponent}
}

func (a *moveAction) toProto() *pb.Action {
	return &pb.Action{Kind: &pb.Action_Move{Move: &pb.MoveAction{
		EntityRole:    a.entityRole,
		DestRole:      a.destRole,
		DestComponent: a.destComponent,
	}}}
}

// ── SetField ─────────────────────────────────────────────────────────────────

type setFieldAction struct {
	role, field string
	value       interface{}
}

func SetField(role, field string, value interface{}) Action {
	return &setFieldAction{role, field, value}
}

func (a *setFieldAction) toProto() *pb.Action {
	var val, valType string
	switch v := a.value.(type) {
	case int:
		val = strconv.Itoa(v)
		valType = "int"
	case bool:
		val = strconv.FormatBool(v)
		valType = "bool"
	case string:
		val = v
		valType = "string"
	default:
		val = fmt.Sprintf("%v", v)
		valType = "string"
	}
	return &pb.Action{Kind: &pb.Action_SetField{SetField: &pb.SetFieldAction{
		Role: a.role, Field: a.field, Value: val, ValueType: valType,
	}}}
}

// ── Destroy ──────────────────────────────────────────────────────────────────

type destroyAction struct{ role string }

func Destroy(role string) Action { return &destroyAction{role} }

func (a *destroyAction) toProto() *pb.Action {
	return &pb.Action{Kind: &pb.Action_Destroy{Destroy: &pb.DestroyAction{Role: a.role}}}
}

// ── Spawn ────────────────────────────────────────────────────────────────────

type spawnAction struct{ templateID, destRole, destComponent string }

func Spawn(templateID, destRole, destComponent string) Action {
	return &spawnAction{templateID, destRole, destComponent}
}

func (a *spawnAction) toProto() *pb.Action {
	return &pb.Action{Kind: &pb.Action_Spawn{Spawn: &pb.SpawnAction{
		TemplateId: a.templateID, DestRole: a.destRole, DestComponent: a.destComponent,
	}}}
}

// ── After ────────────────────────────────────────────────────────────────────

type afterAction struct {
	delayMs int64
	actions []Action
}

func After(delayMs int64, acts ...Action) Action {
	return &afterAction{delayMs, acts}
}

func (a *afterAction) toProto() *pb.Action {
	protoActs := make([]*pb.Action, 0, len(a.actions))
	for _, act := range a.actions {
		protoActs = append(protoActs, act.toProto())
	}
	return &pb.Action{Kind: &pb.Action_After{After: &pb.AfterAction{
		DelayMs: a.delayMs, Actions: protoActs,
	}}}
}

// ── Reveal / Hide ────────────────────────────────────────────────────────────

type revealAction struct{ role, component string }
type hideAction struct{ role, component string }

func Reveal(role, component string) Action { return &revealAction{role, component} }
func Hide(role, component string) Action   { return &hideAction{role, component} }

func (a *revealAction) toProto() *pb.Action {
	return &pb.Action{Kind: &pb.Action_Reveal{Reveal: &pb.RevealAction{Role: a.role, Component: a.component}}}
}

func (a *hideAction) toProto() *pb.Action {
	return &pb.Action{Kind: &pb.Action_Hide{Hide: &pb.HideAction{Role: a.role, Component: a.component}}}
}
