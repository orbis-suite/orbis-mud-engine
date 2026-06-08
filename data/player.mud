entity Player {
    name is "Player"
    description is "Player Template"
    aliases is ["player"]
    tags is ["player"]

    component Inventory {
        children is [
            "Egg"
        ]
    }

    react attack {
        when {
            instrument has tag "player"
        } then {
            print source "You attempt to beat {target} with {instrument}, but they are too heavy to lift."
        }

        when {
            source is target
        } then {
            print source "You hit yourself upon the head, hard enough to hurt."
            publish "{source} hits themselves upon the head, their rage directed inwward."
        }
        
        when {
            instrument is target
        } then {
            print source "You wonder if you might be able to beat {target} with themselves, but disregard the idea."
        }

        then {
            print source "You beat a great big indent into {target}'s head"
            print target "{source} caves your head in."
            publish "{source} violently whacks {target} upon their head."
        }
    }
}

entity Egg {
    name is "Egg"
    description is "A bulbous, green-speckled egg."
    aliases is ["egg"]
    tags is ["egg"]

    angry is false

    react attack {
        when {
            target.angry is false
        } then {
            set target.angry to true
            print source "The egg is now angry that you hit it."
        }
        
        then {
            set target.angry to false
            print source "The egg is calmed after you strike it again"
        }
    }
}