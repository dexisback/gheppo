_gheppo_once() {
    if declare -p PROMPT_COMMAND 2>/dev/null | grep -q 'declare -a'; then
        PROMPT_COMMAND=("${_GHEPPO_ORIGINAL_PROMPT_COMMAND[@]}")
    else
        PROMPT_COMMAND="${_GHEPPO_ORIGINAL_PROMPT_COMMAND-}"
    fi

    unset _GHEPPO_ORIGINAL_PROMPT_COMMAND
    unset -f _gheppo_once

    command gheppo
}

if declare -p PROMPT_COMMAND 2>/dev/null | grep -q 'declare -a'; then
    _GHEPPO_ORIGINAL_PROMPT_COMMAND=("${PROMPT_COMMAND[@]}")
    PROMPT_COMMAND=("_gheppo_once" "${PROMPT_COMMAND[@]}")
else
    _GHEPPO_ORIGINAL_PROMPT_COMMAND="${PROMPT_COMMAND-}"
    PROMPT_COMMAND="_gheppo_once${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
fi


