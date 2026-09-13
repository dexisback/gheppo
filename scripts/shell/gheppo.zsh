[[ -o interactive ]] || return 0

if [[ -o zle ]]; then
    autoload -Uz add-zle-hook-widget

    _gheppo_once() {
        add-zle-hook-widget -d line-init _gheppo_once
        zle && zle -I
        command gheppo
    }

    add-zle-hook-widget line-init _gheppo_once
else
    command gheppo
fi
