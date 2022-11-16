# VHS documentation
#
# Output:
#   Output <path>.gif               Create a GIF output at the given <path>
#   Output <path>.mp4               Create an MP4 output at the given <path>
#   Output <path>.webm              Create a WebM output at the given <path>
#
# Settings:
#   Set FontSize <number>           Set the font size of the terminal
#   Set FontFamily <string>         Set the font family of the terminal
#   Set Height <number>             Set the height of the terminal
#   Set Width <number>              Set the width of the terminal
#   Set LetterSpacing <float>       Set the font letter spacing (tracking)
#   Set LineHeight <float>          Set the font line height
#   Set Theme <string>              Set the theme of the terminal (JSON)
#   Set Padding <number>            Set the padding of the terminal
#   Set Framerate <number>          Set the framerate of the recording
#   Set PlaybackSpeed <float>       Set the playback speed of the recording
#
# Sleep:
#   Sleep <time>                    Sleep for a set amount of <time> in seconds
#
# Type:
#   Type[@<time>] "<characters>"    Type <characters> into the terminal with a
#                                   <time> delay between each character
#
# Keys:
#   Backspace[@<time>] [number]     Press the Backspace key
#   Down[@<time>] [number]          Press the Down key
#   Enter[@<time>] [number]         Press the Enter key
#   Space[@<time>] [number]         Press the Space key
#   Tab[@<time>] [number]           Press the Tab key
#   Left[@<time>] [number]          Press the Left Arrow key
#   Right[@<time>] [number]         Press the Right Arrow key
#   Up[@<time>] [number]            Press the Up Arrow key
#   Down[@<time>] [number]          Press the Down Arrow key
#   Ctrl+<key>                      Press the Control key + <key> (e.g. Ctrl+C)
#
# Display:
#   Hide                            Hide the subsequent commands from the output
#   Show                            Show the subsequent commands in the output

Output docs/examples/arsenic-config.gif

Set FontSize 15
Set FontFamily "DroidSansMono Nerd Font"
Set Width 1200
Set Height 650
Set Padding 10
Set Framerate 30

Hide
Type 'export PATH="/usr/local/go/bin:$PATH"' Enter Sleep 500ms
Type "export GOPATH=$HOME/go" Enter Sleep 500ms
Type "source arsenic.rc" Enter Sleep 500ms
Type "cd /tmp" Enter Sleep 500ms
Type "clear" Enter Sleep 500ms

Show
Type "arsenic config" Sleep 500ms Enter
Sleep 3s

Type "arsenic config example_key example_value" Sleep 500ms Enter
Sleep 2s

Type "arsenic config example_key" Sleep 500ms Enter
Sleep 2s

Type "arsenic config --remove-reset example_key" Sleep 500ms Enter
Sleep 2s

Type "arsenic config example_key" Sleep 500ms Enter
Sleep 2s

Type "arsenic config --sub-keys" Sleep 500ms Enter
Sleep 2s

Type "arsenic config scripts.discover --sub-keys" Sleep 500ms Enter
Sleep 2s

Type "arsenic config scripts.discover.as-subdomain-discovery.enabled false" Sleep 500ms Enter
Sleep 1s

Type "arsenic config scripts.discover.as-subdomain-discovery" Sleep 500ms Enter
Sleep 3s

Type "arsenic config scripts.discover.example-script.count 1" Sleep 500ms Enter
Sleep 1s
Type "arsenic config scripts.discover.example-script.enabled true" Sleep 500ms Enter
Sleep 1s
Type "arsenic config scripts.discover.example-script.order 500" Sleep 500ms Enter
Sleep 1s
Type "arsenic config scripts.discover.example-script.script path/to/example.sh" Sleep 500ms Enter
Sleep 1s
Type "arsenic config scripts.discover.example-script" Sleep 500ms Enter
Sleep 3s

Hide
Type "rm arsenic.yaml" Enter Sleep 500ms
