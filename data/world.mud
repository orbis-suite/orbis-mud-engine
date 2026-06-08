entity LivingRoom {
    name is "Living Room"
    description is "A welcoming and warm living room, clean and orderly with a quiet sense of comfort."
    aliases is ["room"]
    tags is ["room"]

    component Room {
        icon is "L"
        color is "magenta"

        exits is {
            "north": "BedRoom",
            "east": "Bathroom"
        }

        children is [
            "Couch",
            "Lamp",
            "Box"
        ]
    }
}

entity Couch {
    name is "Couch"
    description is "A soft, inviting {'couch' | bold | yellow} rests here, its cushions sagged just enough to suggest long use. It seems comfortable, with plenty of room for something to be hidden within."
    aliases is ["couch"]
    tags is ["furniture"]

    trait Kissable

    react attack {
        when {
            instrument has tag "egg"
        } then {
            print source "You hit the egg upon the couch, gently, as to not disturb the egg."
            publish "{source} hits their egg upon the couch, smiling vacantly to themselves."
        }

        then {
            print source "As you beat upon the couch, a {'nickel' | bold | yellow} falls out."
            publish "{source} beats upon the couch, and a shining {'nickel'| bold | yellow} falls out from under a cushion."
            copy "Nickel" to room.Room
        }
    }
}

entity Nickel {
    name is "Nickel"
    description is "A shining {'nickel' | bold | yellow} lies here, Thomas Jefferson’s handsome side profile glinting faintly as though pleased with its escape."
    aliases is ["nickel"]
    tags is ["item"]

   trait Item 
}

entity Lamp {
    name is "Lamp"
    description is "A dimly lit {'lamp' | bold | yellow} stands quietly in the corner, its weak glow casting just enough light to soften the edges of the room."
    aliases is ["lamp"]
    tags is ["furniture"]
}

entity Box {
    name is "Box"
    description is "A cardboard {'box' | bold | yellow} is here, too."
    aliases is ["box"]
    tags is ["furniture"]
    
    component Container {
        prefix is "Inside the box:"
        revealed is true
        children is [
            "Book",
            "Shoe"
        ]
    }

    react open {
       then {
            reveal target.Container
            print source "You open the box."
            publish "{source} opens the box"
        } 
    }

    react close {
        then {
            hide target.Container
            print source "You close the box."
            publish "{source} closes the box."
        }
    }
}

entity Book {
    name is "Book"
    description is "A {'book' | bold | yellow} with a leather cover, a bold adaptation of VeggieTales with human characters."
    aliases is ["book"]
    tags is ["item"]
    
    trait Item 
}

entity Shoe {
    name is "Shoe"
    description is "A battered left shoe, the sole hangs unattached at the toe."
    aliases is ["shoe"]
    tags is ["item"]
    
    trait Item 
}

entity BedRoom {
    name is "Bedroom"
    description is "A fun little bedroom."
    aliases is ["room"]
    tags is ["room"]

    component Room {
        exits is {
            "south": "LivingRoom"
        }

        children is [
            "Bed",
            "Lamp"
        ]
    }
}

entity Bed {
    name is "Bed"
    description is "A {'bed' | bold | yellow} is well-made and looks inviting."
    aliases is ["bed"]
    tags is ["furniture"]

    trait Standard
}

entity Bathroom {
    name is "Bathroom"
    description is "A bathroom, a perfect place to relax and excrete."
    aliases is ["room"]
    tags is ["room"]

    component Room {
        exits is {
            "west": "LivingRoom",
            "east": "MedicineCabinet"
        }

        children is [
            "Toilet",
            "Goblin"
        ]
    }
}

entity Goblin {
    name is "Goblin"
    description is "A funny {'goblin' | bold | yellow} man no bigger than your fist smiles warmly."
    aliases is ["goblin", "man"]
    tags is ["npc"]

    component Inventory {}

    react attack {
        then {
            print source "As you throw a {'punch' | yellow} at the goblin, he jumps around you, {'kissing' | red} your forehead."
            publish "{source} tries and fails to attack the goblin, yet they're rewarded with a gentle {'kiss' | red } from the creature."
        }
    }

    react kiss {
        when {
            not target in source.Inventory
        } then {
            print source "You give the goblin a kiss upon his {'sweaty' | blue} brow, and he {'hops' | italic} into your pocket."
            publish "{source} gives the goblin a {'kiss' | bold | red}, before the goblin {'jumps' | italic} into {source}'s pocket."
            move target to source.Inventory
        }

        then {
            print source "You look into your pocket and plant another kiss upon the goblin's cheek."
            publish "{source} gives the goblin in their pocket a big wet {'kiss' | bold | red}."
        }
    }

    react give {
        when {
            instrument has tag "item"
        } then {
            print source "You give the goblin your {instrument}, and he accepts it happily. 'You win!' He says, 'You win the game for giving the goblin an item!'"
            publish "{source} gives the goblin {instrument}. The goblin is overjoyed."
            move instrument to target.Inventory
        }

        then {
            print source "The goblin gives you a smile, shaking his head softly. 'I don't want that stupid smelly thing,' he says."
            publish "{source} tries to give the goblin {instrument}, but he refuses to take it."
        }
    }
}

entity Toilet {
    name is "Toilet"
    description is "A {'toilet' | bold | yellow}, it's shiny and porcelain."
    aliases is ["toilet"]
    tags is ["furniture"]

    react kiss {
        then {
            print source "Maybe... no. You reconsider. {'Do not kiss the toilet.' | bold | underline }"
        }
    }
}