# Interactive CD command


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
