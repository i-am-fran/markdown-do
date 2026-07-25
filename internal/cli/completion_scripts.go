package cli

import "fmt"

const bashCompletionScript = `_mdd_complete() {
    local cur words
    cur="${COMP_WORDS[COMP_CWORD]}"
    words=("${COMP_WORDS[@]:1:COMP_CWORD}")
    COMPREPLY=($(compgen -W "$(mdd __complete "${words[@]}")" -- "$cur"))
}
complete -F _mdd_complete mdd
`

const zshCompletionScript = `#compdef mdd
_mdd() {
    local -a completions
    completions=(${(f)"$(mdd __complete "${words[@]:1}")"})
    _describe 'command' completions
}
compdef _mdd mdd
`

// CompletionScript returns the shell completion script for shell ("bash" or
// "zsh"), or an error for anything else.
func CompletionScript(shell string) (string, error) {
	switch shell {
	case "bash":
		return bashCompletionScript, nil
	case "zsh":
		return zshCompletionScript, nil
	default:
		return "", fmt.Errorf("unsupported shell %q — use bash or zsh", shell)
	}
}
