# gotmojify

A fast, single-binary Go clone of [folke/devmoji](https://github.com/folke/devmoji) that adds color and [Gitmoji](https://gitmoji.dev/)-inspired emojis to [Conventional Commits](https://www.conventionalcommits.org/).

## Features

- **Emojify / Demojify / Devmojify / Strip**: Convert text between `unicode` emojis, GitHub `shortcode`s (`:sparkles:`), `devmoji` aliases (`:feat:`, `:chore-release:`, `:security:`), or `strip` all emojis.
- **Conventional Commits (`--commit`)**: Automatically inserts matching type, scope, and breaking-change emojis into commit headers (`feat(cli): ✨ ...`, `fix(security): 🐛 🔒️ ...`, `feat!: 💥 ✨ ...`).
- **Git Hook Integration (`--edit`, `--lint`)**: Formats and validates `.git/COMMIT_EDITMSG` in place (with git worktree support).
- **Pretty `git log` Formatting (`--log`)**: Colorizes and emojifies commit messages across streamed log lines.
- **Layered Configuration**: Supports YAML, JSON, and TOML config files (`.gotmojify.{yaml,yml,json,toml}`, `gotmojify.{yaml,yml,json,toml}`, `devmoji.config.{yaml,yml,json,toml}`), `GOTMOJIFY_*` environment variables, and CLI flags.
- **Shell Completions**: Built-in Cobra completions (`gotmojify completion bash|zsh|fish|powershell`).

## Installation

### Using `go install`

```sh
go install github.com/DO3B/gotmojify/cmd/gotmojify@latest
```

### Pre-built Binaries

Download pre-compiled archives for Linux, macOS, and Windows (`amd64` / `arm64`) from the [GitHub Releases](https://github.com/DO3B/gotmojify/releases) page.

### Build from Source

```sh
git clone https://github.com/DO3B/gotmojify.git
cd gotmojify
go build -o bin/gotmojify ./cmd/gotmojify
```

## Usage

```text
gotmojify [flags]
gotmojify [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print version information

Flags:
  -c, --config string   location of config file (.gotmojify.{yaml,json,toml} or devmoji.config.*)
  -l, --list            list all known devmojis
  -t, --text string     text to format. reads from stdin when omitted
      --lint            lint the conventional commit. disabled for --log
  -f, --format string   output format: unicode, shortcode, devmoji, strip (default "unicode")
      --commit          automatically add a devmoji to the conventional commit header (default true)
      --no-commit       do not process conventional commit headers
  -e, --edit            read last commit message from .git/COMMIT_EDITMSG in the git root
      --log             format conventional commits in text similar to git log
      --color           use colors for formatting (enabled by default for TTY output)
      --no-color        don't use colors
  -h, --help            help for gotmojify
  -v, --version         version for gotmojify
```

### Emojify & Format Conversion

| Format | Description |
| --- | --- |
| `unicode` | Outputs Unicode emojis such as `✨`, `🚀` (default) |
| `shortcode` | Outputs GitHub Markdown shortcodes such as `:sparkles:`, `:rocket:` |
| `devmoji` | Outputs devmoji aliases such as `:feat:`, `:chore-release:` |
| `strip` | Removes all emojis and shortcodes from the input |

```sh
echo "This is a :test: of the first :release: :boom: ✨" | gotmojify --no-commit --format shortcode
# This is a :rotating_light: of the first :rocket: :boom: :sparkles:

echo "This is a :test: of the first :release: :boom: :sparkles:" | gotmojify --no-commit --format unicode
# This is a 🚨 of the first 🚀 💥 ✨

echo "🚀 :boom: :sparkles:" | gotmojify --no-commit --format devmoji
# :chore-release: :breaking: :feat:

echo "test 🚀 :boom: :sparkles: :security:" | gotmojify --no-commit --format strip
# test
```

### Conventional Commits (`--commit`)

```sh
gotmojify --text "feat: added a new feature :smile:"
# feat: ✨ added a new feature 😄

gotmojify --text "chore(release): 1.1.1"
# chore(release): 🚀 1.1.1

gotmojify --text "fix(security): upgraded dependency"
# fix(security): 🐛 🔒️ upgraded dependency

gotmojify --text "feat(cli)!: breaking change"
# feat(cli)!: 💥 ✨ breaking change
```

### Git Hooks (`--edit`, `--lint`)

Format and lint `.git/COMMIT_EDITMSG` in `.git/hooks/prepare-commit-msg` or `.git/hooks/commit-msg`:

```sh
#!/bin/sh
gotmojify --edit --lint
```

Or configure with [Lefthook](https://lefthook.dev/) in `lefthook.yml`:

```yaml
commit-msg:
  jobs:
    - name: gotmojify
      run: gotmojify --edit --lint
```

### Pretty `git log` (`--log`)

```sh
git log --oneline --color=always | gotmojify --log --color
```

### List Available Devmojis (`--list`)

```sh
gotmojify --list
```

## Configuration

`gotmojify` loads the file passed via `--config`, or automatically searches in order:

1. Current working directory
2. Nearest parent directory containing `go.mod`
3. Nearest parent directory containing `package.json`
4. Nearest parent directory containing `.git`
5. User home directory (`$HOME`)

Supported filenames: `.gotmojify.{yaml,yml,json,toml}`, `gotmojify.{yaml,yml,json,toml}`, and `devmoji.config.{yaml,yml,json,toml}`.

### Example `.gotmojify.yaml`

```yaml
format: unicode
types:
  - lint
devmoji:
  # Override an existing type's emoji
  - code: feat
    emoji: boom
  # Add a custom devmoji
  - code: fail
    emoji: poop
    description: something got messed up pretty bad
  # Inherit description from built-in gitmoji
  - code: css
    gitmoji: art
  # Inherit description from gitmoji while overriding the emoji
  - code: config
    gitmoji: wrench
    emoji: gear
```

### Environment Variables

All core options can also be overridden via `GOTMOJIFY_*` environment variables (precedence: CLI flag > environment variable > config file > default):

- `GOTMOJIFY_FORMAT` (`unicode`, `shortcode`, `devmoji`, `strip`)
- `GOTMOJIFY_COMMIT` (`true` / `false`)
- `GOTMOJIFY_LINT` (`true` / `false`)
- `GOTMOJIFY_LOG` (`true` / `false`)
- `GOTMOJIFY_COLOR` (`true` / `false`, or standard `NO_COLOR`)

## Default Devmoji Reference

| Emoji | Devmoji Code | Description |
| --- | --- | --- |
| ✨ | `:feat:` | **feat:** a new feature |
| 🐛 | `:fix:` | **fix:** a bug fix |
| 📚️ | `:docs:` | **docs:** documentation only changes |
| 🎨 | `:style:` | **style:** changes that do not affect the meaning of the code |
| ♻️ | `:refactor:` | **refactor:** a code change that neither fixes a bug nor adds a feature |
| ⚡️ | `:perf:` | **perf:** a code change that improves performance |
| 🚨 | `:test:` | **test:** adding missing or correcting existing tests |
| 🔧 | `:chore:` | **chore:** changes to the build process or auxiliary tools |
| 🚀 | `:chore-release:` | **chore(release):** code deployment or publishing |
| 🔗 | `:chore-deps:` | **chore(deps):** add or delete dependencies |
| 📦️ | `:build:` | **build:** changes related to build processes |
| 👷 | `:ci:` | **ci:** updates to the continuous integration system |
| 🚀 | `:release:` | code deployment or publishing to external repositories |
| 🔒️ | `:security:` | Fix security or privacy issues. |
| 🌐 | `:i18n:` | Internationalization and localization. |
| 💥 | `:breaking:` | Introduce breaking changes. |
| ⚙️ | `:config:` | Add or update configuration files. |
| ➕ | `:add:` | add something |
| ➖ | `:remove:` | remove something |

## Development

This repository uses [Aqua](https://aquaproj.github.io/) for declarative binary management (`aqua.yaml`), [Task](https://taskfile.dev/) as the task runner (`Taskfile.yml`), and [Lefthook](https://lefthook.dev/) for git hooks (`lefthook.yml`).

```sh
# Install pinned CLI tools (task, golangci-lint, goreleaser, lefthook, govulncheck)
aqua i

# Install git hooks (pre-commit, commit-msg, pre-push)
task hooks:install

# List all available tasks
task --list

# Run vet, lint, tests, and vulnerability scan
task check
```

## Automated Releases & CI

- **CI**: GitHub Actions workflows run unit tests across Go `1.26`, `1.27`, and `stable` with `-race`, `golangci-lint`, `govulncheck`, `gosec`, and CodeQL.
- **Releases**: [Release Please](https://github.com/googleapis/release-please) (`.github/workflows/release-please.yml`) runs automatically once `Tests`, `Lint`, and `Security` all succeed on `main`, managing release PRs and changelogs via `release-please-config.json` and `.release-please-manifest.json`. Merging a Release Please PR creates the version tag and GitHub Release and invokes `.github/workflows/release.yml` (GoReleaser) to publish binaries and checksums; pushing a manual `v*` tag (`git tag v1.0.0 && git push origin v1.0.0`) also continues to work.
- **Dependency Updates**: Configured via `.github/dependabot.yml` for Go modules and GitHub Actions.

## License

Apache License 2.0 — see [LICENSE](LICENSE).
