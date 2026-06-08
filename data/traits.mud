trait Standard {
    trait Kissable
    trait Hittable
}

trait Kissable {
    react kiss {
        then {
            print source "You kiss the {target}"
            publish "{source} kisses the {target}."
        }
    }
}

trait Hittable {
    react attack {
        when {
            instrument is target
        } then {
            print source "You can't hit something with itself."
        }

        when {
            instrument exists
        } then {
            print source "You hit the {target} with {instrument}"
            publish "{source} hits the {target} with {instrument}."
        }

        then {
            print source "You hit the {target}"
            publish "{source} hits the {target}."
        }
    }
}

trait Item {
    react take {
        when {
            not target in source.Inventory
        } then {
            print source "You pocket {target}"
            publish "{source} pockets {target}"
            move target to source.Inventory
        }

        then {
            print source "You're already carrying {target}"
        }
    }

    react drop {
        when {
            target in source.Inventory
        } then {
            print source "You drop {target} onto the ground."
            publish "{source} drops {target} onto the ground."
            move target to room.Room 
        }

        then {
            print source "You aren't carrying {target}"
        }
    }
}