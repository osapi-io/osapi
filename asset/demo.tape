# OSAPI Demo Recording
# Run from the repo root with: vhs asset/demo.tape
# Requires: vhs, tmux, osapi binary built in current directory

Output asset/demo.gif

Set Shell "zsh"
Set FontSize 14
Set Width 1400
Set Height 800
Set Theme "Catppuccin Mocha"
Set WindowBar Colorful
Set BorderRadius 8
Set Padding 10
Set TypingSpeed 40ms
Set CursorBlink false

# VHS passes --no-rcs to zsh, so source dotfiles and fix PATH
Hide
Type "source ~/.zshrc"
Enter
Sleep 1s
Type "export PATH=.:$PATH"
Enter
Sleep 500ms
Type "tmux kill-session -t demo 2>/dev/null; true"
Enter
Sleep 500ms
Type "clear"
Enter
Sleep 500ms
Show

# Start tmux
Type "tmux new-session -s demo"
Enter
Sleep 2s

# Top pane: start osapi
Type "osapi start"
Enter
Sleep 5s

# Split pane horizontally (top pane ~1/3, bottom ~2/3)
Ctrl+A
Sleep 500ms
Type@0 `"`
Sleep 1s
Ctrl+A
Sleep 300ms
Type@0 ":"
Sleep 300ms
Type@0 "resize-pane -U 15"
Enter
Sleep 1s

# Bottom pane: source env so prompt and PATH work here too
Hide
Type "source ~/.zshrc"
Enter
Sleep 1s
Type "export PATH=.:$PATH"
Enter
Sleep 500ms
Type "clear"
Enter
Sleep 500ms
Show

# Bottom pane: run client commands

# 1. Health check
Sleep 1s
Type "osapi client health status"
Sleep 500ms
Enter
Sleep 4s

# 2. Agent list
Type "osapi client agent list"
Sleep 500ms
Enter
Sleep 4s

# 3. Node status
Type "osapi client node status"
Sleep 500ms
Enter
Sleep 5s

# 4. Run a command
Type@50ms `osapi client node command exec --command "uname" --args "-a"`
Sleep 500ms
Enter
Sleep 5s

# 5. Audit trail
Type "osapi client audit list"
Sleep 500ms
Enter
Sleep 5s

# 6. JSON output
Type "osapi client node hostname --json"
Sleep 500ms
Enter
Sleep 4s

# Final pause
Sleep 2s

# Cleanup (hidden)
Hide
Ctrl+A
Sleep 300ms
Type@0 ":"
Sleep 300ms
Type@0 "kill-session"
Enter
