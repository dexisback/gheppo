autoload -Uz add-zsh-hook

_gheppo_once() {
    add-zsh-hook -d precmd _gheppo_once
    command gheppo
}

add-zsh-hook precmd _gheppo_once
