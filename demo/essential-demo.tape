# Essential mdd workflow: the everyday CRUD + find loop, nothing else.
#
# Run from inside demo/:
#   cd demo && vhs essential-demo.tape
#
# Isolates $HOME so this always starts from mdd's real default settings,
# and prepends $HOME/go/bin onto PATH to use the freshly built binary.

Require mdd

Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 550
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 35ms

Output essential-demo.gif
Output essential-demo.mp4

Hide
Type `export PATH="$HOME/go/bin:$PATH"`
Enter
Type `export HOME="$(mktemp -d)"`
Enter
Type `export PS1="$ "`
Enter
Type "clear"
Enter
Show

Type "mdd list"
Sleep 800ms
Enter
Sleep 2s

Type "mdd Ship weekly digest email"
Sleep 800ms
Enter
Sleep 500ms
Type `mdd "Follow up with design on empty states @bb"`
Sleep 800ms
Enter
Sleep 500ms
Type "mdd list"
Sleep 800ms
Enter
Sleep 2200ms

Type "mdd toggle 2"
Sleep 800ms
Enter
Sleep 1200ms

Type "mdd complete 1"
Sleep 800ms
Enter
Sleep 1200ms

Type `mdd edit 4 "Dark mode toggle for settings screen"`
Sleep 800ms
Enter
Sleep 1200ms

Type "mdd find login"
Sleep 800ms
Enter
Sleep 1200ms

Type "mdd remove 8"
Sleep 800ms
Enter
Sleep 1200ms

Type "mdd list"
Sleep 800ms
Enter
Sleep 2500ms
