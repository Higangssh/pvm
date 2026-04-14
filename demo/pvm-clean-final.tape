Output demo/pvm-clean-final.mp4
Output demo/pvm-clean-final.gif

Require python3

Set Shell "bash"
Set Theme "tokyonight"
Set FontSize 28
Set Width 1600
Set Height 900
Set Padding 24
Set Framerate 30
Set TypingSpeed 40ms
Set WindowBar Colorful
Set BorderRadius 12

Hide
Type "ln -sf /tmp/pvm-demo-bin /tmp/pvm" Sleep 100ms Enter
Type "export PS1='❯ '" Sleep 100ms Enter
Type "export PATH=/tmp:$PATH" Sleep 100ms Enter
Type "export HOME=/tmp/pvm-demo-home-clean" Sleep 100ms Enter
Type "export XDG_CONFIG_HOME=/tmp/pvm-demo-home-clean/.config" Sleep 100ms Enter
Type "mkdir -p /tmp/pvm-demo-home-clean/.config /tmp/pvm-demo-project-clean /tmp/pvm-demo-venvs" Sleep 100ms Enter
Type "cd /tmp/pvm-demo-project-clean" Sleep 100ms Enter
Type "clear" Sleep 100ms Enter
Show
Type "pvm list" Sleep 250ms Enter
Sleep 2800ms
Type "pvm ui" Sleep 250ms Enter
Sleep 2200ms
Enter
Sleep 2000ms
Type "python --version" Sleep 250ms Enter
Sleep 1500ms
Type "pwd" Sleep 250ms Enter
Sleep 1500ms
Type "exit" Sleep 250ms Enter
Sleep 1200ms
