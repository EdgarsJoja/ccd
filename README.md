# Interactive CD command

## Usage
Use keys `up` and `down` to choose directory.

Use `right` or `enter` to select chosen directory.

Use `left` or `esc` or `backspace` to go back.

Use `h` to toggle hidden directories.

Use `q` to exit and cd into currently selected directory.

Use `ctrl+c` to exit without cd operation.

## Shell Integration

Integration with various shells.

### Bash/Zsh
```bash
ccd() {
  local start="${1:-$PWD}" dir
  dir="$("/path/to/ccd/binary" "$start")" || return
  [ -n "$dir" ] && builtin cd -- "$dir"
}
```

### Todo

[ ] Directory filter
[ ] Help output
[ ] Persistent memory between runs
