# lazy-chrony

A fast, keyboard-driven **terminal UI cron job manager** for Linux.  
Built with Go + [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea).

```
 lazy-chrony  v0.1.0   [/] filter:all

 #    S  SRC      SCHEDULE               COMMAND                              NEXT RUN
 1    *  user     */5 * * * *            /usr/bin/backup.sh                   3m 12s
 2    *  user     @daily                 /home/user/cleanup.sh                18h 4m
 3    -  user     0 12 * * *             /usr/bin/notify.sh                   disabled
 4    *  system   17 *  * * *            root    cd / && run-parts ...        43m 0s
 5    *  cron.d   25 6  * * *            root    test -x /usr/sbin/ana...     5h 25m

 a add  e edit  d del  space toggle  ? help  q quit
```

## Features

- List all cron jobs: user crontab + `/etc/crontab` + `/etc/cron.d/*`
- Add, edit, delete user cron jobs
- Enable/disable (comment/uncomment) user jobs
- Next run time shown for every job
- Filter by source or status (user, system, enabled, disabled)
- Vim-style navigation (j/k, g/G)
- Schedule presets (ctrl+p in the form)
- No root required for user jobs

## Installation

### Arch Linux (recommended)

```bash
# From source with makepkg
git clone https://github.com/domenez-dev/lazy-chrony
cd lazy-chrony/packaging/arch
makepkg -si
```

Or with an AUR helper once published:
```bash
yay -S lazy-chrony
```

### Debian / Ubuntu

Download the latest `.deb` from the releases page:
```bash
sudo dpkg -i lazy-chrony_0.1.0_amd64.deb
```

### RHEL / CentOS / Fedora / openSUSE

Download the latest `.rpm` from the releases page:
```bash
sudo rpm -i lazy-chrony-0.1.0-1.x86_64.rpm
# or
sudo dnf install ./lazy-chrony-0.1.0-1.x86_64.rpm
```

### From source (any distro)

Requirements: Go 1.21+

```bash
git clone https://github.com/domenez-dev/lazy-chrony
cd lazy-chrony
make install        # builds and installs to /usr/local/bin
```

## Build

```bash
make build          # binary in dist/
make build-all      # linux amd64 + arm64
make deb            # .deb via nfpm
make rpm            # .rpm via nfpm
```

### Building packages with nfpm

Install [nfpm](https://nfpm.goreleaser.com/install/):

```bash
# Arch
yay -S nfpm
# or via Go
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest

# Build packages
make deb
make rpm
```

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `k` | Down / Up |
| `g` / `G` | Top / Bottom |
| `a` | Add new job |
| `e` | Edit selected job |
| `d` | Delete selected job |
| `space` / `t` | Toggle enable/disable |
| `r` | Refresh from disk |
| `/` | Cycle filter |
| `?` | Toggle help panel |
| `q` | Quit |

### In the Add/Edit form:

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | Next / previous field |
| `ctrl+p` | Cycle schedule presets |
| `ctrl+s` | Save |
| `esc` | Cancel |

## Schedule presets (ctrl+p)

`@reboot`, `@hourly`, `@daily`, `@weekly`, `@monthly`,  
`*/5 * * * *`, `*/15 * * * *`, `*/30 * * * *`, `0 0 * * *`, `0 12 * * *`

## Notes

- Only **user** jobs can be edited/deleted/toggled (your own crontab).
- System jobs in `/etc/crontab` and `/etc/cron.d/` are shown read-only.
  To edit them, run `sudo lazy-chrony` (or edit directly with `sudo vim /etc/cron.d/myjob`).
- The app calls `crontab -l` and `crontab -` internally to read/write,
  so your existing crontab comments and env vars are preserved.

## License

MIT
