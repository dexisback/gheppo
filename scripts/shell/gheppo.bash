_gheppo_once() {
    local gheppo_prompt_command=""
    local command_part

    for command_part in ${PROMPT_COMMAND-}; do
        if [[ "$command_part" != "_gheppo_once" ]]; then
            if [[ -n "$gheppo_prompt_command" ]]; then
                gheppo_prompt_command+=";"
            fi
            gheppo_prompt_command+="$command_part"
        fi
    done

    PROMPT_COMMAND="$gheppo_prompt_command"
    command gheppo
}

PROMPT_COMMAND="_gheppo_once${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
