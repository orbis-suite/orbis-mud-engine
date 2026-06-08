command Take {
    aliases is ["take", "grab", "pickup"]
    
    pattern {
        syntax is "take {target}"
        noMatch is "you can't pick that up."
    }
}

command Drop {
    aliases is ["drop"]

    pattern {
        syntax is "drop {target}"
        noMatch is "you can't drop that."
    }
}

command Give {
    aliases is ["give", "hand"]

    pattern {
        syntax is "give {instrument} to {target}"
        noMatch is "You can't give that to that."
    }

    pattern {
        syntax is "give {target} {instrument}"
        noMatch is "You can't give that to that."
    }
}