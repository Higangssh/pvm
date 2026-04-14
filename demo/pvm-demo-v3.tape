Output demo/pvm-demo-v3.mp4
Output demo/pvm-demo-v3.gif

Require python3
Require ./pvm

Set Shell "bash"
Set Theme "tokyonight"
Set FontSize 28
Set Width 1600
Set Height 900
Set Padding 24
Set Framerate 30
Set TypingSpeed 42ms
Set WindowBar Colorful
Set BorderRadius 12

Hide
Type "cd /Users/sanghee/.openclaw/workspace/projects/pvm" Sleep 300ms Enter
Type "export PATH=/Users/sanghee/.openclaw/workspace/projects/pvm:$PATH" Sleep 300ms Enter
Type "python3 -m venv /tmp/pvm-vhs-demo-3" Sleep 300ms Enter
Sleep 2s
Show
Type "pvm add /tmp/pvm-vhs-demo-3 --alias vhsdemo3" Sleep 300ms Enter
Sleep 2s
Type "pvm list" Sleep 300ms Enter
Sleep 2s
Type "pvm run vhsdemo3 -- --version" Sleep 300ms Enter
Sleep 2s
Type "pvm exec vhsdemo3 -- pwd" Sleep 300ms Enter
Sleep 2s
