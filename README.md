# li

A text editor

Work in progress

# features

* fully concurrent architecture
* vim-like modal editing
* flexible UI widget system
* 24-bit color terminal support
* fuzzy-matching auto completion
* time-based undo / redo
* command palette
* portable rendering backend
* no plugin scripting, hackable go codes only

# planning features

* multiple cursor / selection
* block selection
* language server protocol client
* context menu
* key stroke macro
* mouse operations

# screenshot

![Screenshot](misc/screenshot-2019-09-27.png?raw=true "Screenshot")

# install

`go get github.com/reusee/li`

Check li/config_default.go for key mappings and other configurations.

Create ~/.config/li-editor/config.toml to overwrite the default configurations.

