command Open {
    aliases is ["open"]

    pattern {
        syntax is "open {target}"
        noMatch is "You can't open that."
    }
}

command Close {
    aliases is ["close"]

    pattern {
        syntax is "close {target}"
        noMatch is "You can't close that."
    }
}