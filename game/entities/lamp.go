package entities

import "example.com/mud/sdk"

var lampDef = &sdk.EntityDef{
	ID:          "Lamp",
	Name:        "Lamp",
	Description: "A dimly lit {'lamp' | bold | yellow} stands quietly in the corner, its weak glow casting just enough light to soften the edges of the room.",
	Aliases:     []string{"lamp"},
	Tags:        []string{"furniture"},
}

var bedroomLampDef = &sdk.EntityDef{
	ID:          "BedroomLamp",
	Name:        "Lamp",
	Description: "A dimly lit {'lamp' | bold | yellow} stands quietly in the corner, its weak glow casting just enough light to soften the edges of the room.",
	Aliases:     []string{"lamp"},
	Tags:        []string{"furniture"},
}
