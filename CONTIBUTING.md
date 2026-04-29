# Contributing

Wanna learn Go? This is a great project for it — it's small enough to understand fully but has real moving parts: terminal UI, process I/O, crontab parsing, and packaging for multiple distros.

Before making a change, please open an issue or start a discussion so we can talk through the approach first. This avoids duplicate work and keeps the codebase consistent.

## Pull Requests

All code changes go through pull requests:

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. If your change affects user-facing behavior, update the documentation.
4. Make sure your code follows [Effective Go](https://golang.org/doc/effective_go.html) as much as possible.
5. Test your changes manually — run `go run cmd/lazy-cron/main.go` and make sure nothing is broken.
6. Write a [good commit message](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).
7. Open the pull request.

## Project structure

```
cmd/lazy-cron/    entry point
internal/
  cron/             cron job parsing, reading, writing (manager.go, describe.go)
  styles/           all lipgloss styles and colors
  ui/               root bubbletea model (app.go)
  views/            individual screens (list.go, form.go, confirm.go)
packaging/
  arch/PKGBUILD     Arch Linux package
  nfpm.yaml         Debian/RPM package config
  lazy-cron.1     man page
```

## Running locally

```bash
go run cmd/lazy-cron/main.go
```

## Dependencies

```bash
go mod tidy
```

Key dependencies: `charmbracelet/bubbletea`, `charmbracelet/bubbles`, `charmbracelet/lipgloss`, `robfig/cron`.

## Reporting bugs

Use [GitHub Issues](https://github.com/domenez-dev/lazy-cron/issues). Include your distro, terminal emulator, and what you expected vs what happened.

## Code of conduct

Be respectful. This is a small project and a welcoming space for people learning Go and systems programming.

## License

By contributing, you agree that your submissions are under the same [MIT License](LICENSE) that covers the project.
