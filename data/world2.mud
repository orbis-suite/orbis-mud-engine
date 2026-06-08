entity MedicineCabinet {
    name is "Medicine Cabinet"
    description is "A medicine cabinet. God only knows how you managed to fit in here."
    aliases is ["room"]
    tags is ["room"]

    component Room {
        exits is {
            "west": "Bathroom"
        }

        children is [
            "Medicine"
        ]
    }
}

entity Medicine {
    name is "Medicine"
    description is "Bottles of pills line the shelves -- the eponymous medicine for which the cabinet is named."
    aliases is ["medicine", "pills", "bottles"]
    tags is ["consumable"]
}