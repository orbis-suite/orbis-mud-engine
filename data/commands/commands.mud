command Attack {
    aliases is ["attack", "hit", "beat"]

    pattern {
        syntax is "attack {target}"
        noMatch is "You don't want to attack that."
    }

    pattern {
        syntax is "attack {target} with {instrument}"
        noMatch is "You don't want to attack that with that."
    }
}

command Kiss {
    aliases is ["smooch"]

    pattern {
        syntax is "kiss {target}"
        noMatch is "you don't want to kiss that."
    }
}