package main

const bashCompletion = `# bash completion for tod — source this file or add to ~/.bashrc:
#   source <(tod completion bash)
_tod() {
    local cur cmds
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    cmds="add ls list done undone reopen rm del edit pri due search find stats undo redo clear export path completion version help"

    if [ "$COMP_CWORD" -eq 1 ]; then
        COMPREPLY=( $(compgen -W "$cmds" -- "$cur") )
        return 0
    fi

    case "${COMP_WORDS[1]}" in
        ls|list|search|find)
            COMPREPLY=( $(compgen -W "--all --done --tag --project --pri --plain --no-color" -- "$cur") )
            ;;
        pri|priority)
            COMPREPLY=( $(compgen -W "high medium low none" -- "$cur") )
            ;;
        due)
            COMPREPLY=( $(compgen -W "today tomorrow mon tue wed thu fri sat sun none" -- "$cur") )
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") )
            ;;
    esac
}
complete -F _tod tod
`

const zshCompletion = `#compdef tod
# zsh completion for tod — save as _tod in your fpath, or:
#   tod completion zsh > "${fpath[1]}/_tod"
_tod() {
    local -a cmds
    cmds=(
        'add:Add a task'
        'ls:List tasks'
        'done:Complete tasks'
        'undone:Reopen tasks'
        'rm:Delete tasks'
        'edit:Edit a task'
        'pri:Set priority'
        'due:Set due date'
        'search:Find tasks'
        'stats:Productivity dashboard'
        'undo:Undo last change'
        'redo:Redo last undo'
        'clear:Remove completed tasks'
        'export:Print tasks as JSON'
        'path:Show data file locations'
        'completion:Shell completion script'
        'version:Print version'
        'help:Show help'
    )
    if (( CURRENT == 2 )); then
        _describe 'command' cmds
        return
    fi
    case "$words[2]" in
        ls|list|search|find)
            _arguments '--all[include completed]' '--done[completed only]' \
                '--tag[filter by tag]' '--project[filter by project]' \
                '--pri[filter by priority]' '--plain[ASCII output]' '--no-color[no color]'
            ;;
        pri|priority)
            _values 'priority' high medium low none
            ;;
        completion)
            _values 'shell' bash zsh fish
            ;;
    esac
}
compdef _tod tod
`

const fishCompletion = `# fish completion for tod — save to ~/.config/fish/completions/tod.fish:
#   tod completion fish > ~/.config/fish/completions/tod.fish
complete -c tod -n '__fish_use_subcommand' -a add -d 'Add a task'
complete -c tod -n '__fish_use_subcommand' -a 'ls list' -d 'List tasks'
complete -c tod -n '__fish_use_subcommand' -a done -d 'Complete tasks'
complete -c tod -n '__fish_use_subcommand' -a 'undone reopen' -d 'Reopen tasks'
complete -c tod -n '__fish_use_subcommand' -a 'rm del' -d 'Delete tasks'
complete -c tod -n '__fish_use_subcommand' -a edit -d 'Edit a task'
complete -c tod -n '__fish_use_subcommand' -a pri -d 'Set priority'
complete -c tod -n '__fish_use_subcommand' -a due -d 'Set due date'
complete -c tod -n '__fish_use_subcommand' -a 'search find' -d 'Find tasks'
complete -c tod -n '__fish_use_subcommand' -a stats -d 'Productivity dashboard'
complete -c tod -n '__fish_use_subcommand' -a undo -d 'Undo last change'
complete -c tod -n '__fish_use_subcommand' -a redo -d 'Redo last undo'
complete -c tod -n '__fish_use_subcommand' -a clear -d 'Remove completed tasks'
complete -c tod -n '__fish_use_subcommand' -a export -d 'Print tasks as JSON'
complete -c tod -n '__fish_use_subcommand' -a path -d 'Show data locations'
complete -c tod -n '__fish_use_subcommand' -a completion -d 'Shell completions'
complete -c tod -n '__fish_use_subcommand' -a version -d 'Print version'
complete -c tod -n '__fish_use_subcommand' -a help -d 'Show help'
complete -c tod -n '__fish_seen_subcommand_from ls list search find' -l all -d 'Include completed'
complete -c tod -n '__fish_seen_subcommand_from ls list search find' -l done -d 'Completed only'
complete -c tod -n '__fish_seen_subcommand_from ls list search find' -l plain -d 'ASCII output'
complete -c tod -n '__fish_seen_subcommand_from ls list search find' -l no-color -d 'No color'
complete -c tod -n '__fish_seen_subcommand_from ls list search find' -l tag -d 'Filter by tag'
complete -c tod -n '__fish_seen_subcommand_from ls list search find' -l project -d 'Filter by project'
complete -c tod -n '__fish_seen_subcommand_from ls list search find' -l pri -d 'Filter by priority'
complete -c tod -n '__fish_seen_subcommand_from pri priority' -a 'high medium low none'
complete -c tod -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'
`
