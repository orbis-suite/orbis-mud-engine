package entities

import (
	"example.com/mud/sdk"
)

type entity struct {
	def   *sdk.EntityDef
	react func(*sdk.Event) []sdk.Action
}

// all is the registry of every entity in the game.
var all = []entity{
	{playerDef, playerReact},
	{eggDef, eggReact},
	{couchDef, couchReact},
	{nickelDef, nil},
	{lampDef, standardReact},
	{boxDef, boxReact},
	{bookDef, itemReact},
	{shoeDef, itemReact},
	{bedDef, standardReact},
	{bedroomLampDef, standardReact},
	{goblinDef, goblinReact},
	{toiletDef, nil},
	{medicineDef, nil},
}

func All() []*sdk.EntityDef {
	out := make([]*sdk.EntityDef, 0, len(all))
	for _, e := range all {
		out = append(out, e.def)
	}
	return out
}

func Dispatch(e *sdk.Event) []sdk.Action {
	// TODO: this shouldn't be an O(N) operation
	for _, h := range all {
		if h.def.ID != e.TargetID {
			continue
		}
		if acts, ok := h.def.Reactions[e.Command]; ok {
			return acts
		}
		if h.react != nil {
			return h.react(e)
		}
		return nil
	}
	return nil
}
