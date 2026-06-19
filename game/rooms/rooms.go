package rooms

import "example.com/mud/sdk"

func All() []*sdk.RoomDef {
	return []*sdk.RoomDef{
		Hut(),
		LivingRoom(),
		BedRoom(),
		Bathroom(),
		MedicineCabinet(),
	}
}

func Hut() *sdk.RoomDef {
	return &sdk.RoomDef{
		ID:          "Hut",
		Name:        "Hut",
		Description: "A cozy little hut.",
		Icon:        "H",
		Color:       "yellow",
		Exits:       map[string]string{"east": "LivingRoom"},
		ChildIDs:    []string{"Goblin", "Goblin", "Couch"},
	}
}

func LivingRoom() *sdk.RoomDef {
	return &sdk.RoomDef{
		ID:          "LivingRoom",
		Name:        "Living Room",
		Description: "A welcoming and warm living room, clean and orderly with a quiet sense of comfort.",
		Icon:        "L",
		Color:       "magenta",
		Exits:       map[string]string{"north": "BedRoom", "east": "Bathroom", "west": "Hut"},
		ChildIDs:    []string{"Couch", "Lamp", "Box"},
	}
}

func BedRoom() *sdk.RoomDef {
	return &sdk.RoomDef{
		ID:          "BedRoom",
		Name:        "Bedroom",
		Description: "A fun little bedroom.",
		Exits:       map[string]string{"south": "LivingRoom"},
		ChildIDs:    []string{"Bed", "BedroomLamp"},
	}
}

func Bathroom() *sdk.RoomDef {
	return &sdk.RoomDef{
		ID:          "Bathroom",
		Name:        "Bathroom",
		Description: "A bathroom, a perfect place to relax and excrete.",
		Exits:       map[string]string{"west": "LivingRoom", "east": "MedicineCabinet"},
		ChildIDs:    []string{"Toilet", "Goblin"},
	}
}

func MedicineCabinet() *sdk.RoomDef {
	return &sdk.RoomDef{
		ID:          "MedicineCabinet",
		Name:        "Medicine Cabinet",
		Description: "A medicine cabinet. God only knows how you managed to fit in here.",
		Exits:       map[string]string{"west": "Bathroom"},
		ChildIDs:    []string{"Medicine"},
	}
}
