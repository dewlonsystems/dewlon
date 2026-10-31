# dewlon

[![CI](https://github.com/dewlonsystems/dewlon/actions/workflows/ci.yml/badge.svg)](https://github.com/dewlonsystems/dewlon/actions/workflows/ci.yml)
[![Latest release](https://img.shields.io/github/v/release/dewlonsystems/dewlon)](https://github.com/dewlonsystems/dewlon/releases/latest)

`dewlon` is a command line tool that creates new projects from Dewlon's maintained templates. One command gives you a working project with your name filled in, its dependencies installed, and the template's checksum verified.

```
dewlon new hospital-api
```

## At a glance

| | |
|---|---|
| **What it does** | Creates new projects from templates, adds ready-made modules (Docker, CI, database) to existing Go projects, and checks that your machine is set up |
| **Command name** | `dewlon` |
| **Templates** | `fiber` (Go REST API on Fiber), `flutter` (Flutter app) |
| **Works on** | Windows, macOS, Linux (Intel/AMD and ARM) |
| **Install on Windows** | `irm https://raw.githubusercontent.com/dewlonsystems/dewlon/main/scripts/install.ps1 \| iex` (in PowerShell) |
| **Install on macOS/Linux** | `curl -fsSL https://raw.githubusercontent.com/dewlonsystems/dewlon/main/scripts/install.sh \| sh` |
| **Check it works** | `dewlon version` then `dewlon doctor` |
| **Create a project** | `dewlon new my-project` |
| **Source code** | https://github.com/dewlonsystems/dewlon |
| **Downloads** | https://github.com/dewlonsystems/dewlon/releases/latest |
| **License** | See the `LICENSE` file in this repository |

## Contents

- [Install](#install)
  - [Windows](#install-on-windows)
  - [macOS and Linux](#install-on-macos-and-linux)
  - [Download by hand](#download-by-hand-no-commands)
  - [Other ways](#other-ways-to-install)
  - [Check that it worked](#check-that-it-worked)
- [Create your first project](#create-your-first-project)
- [Commands](#commands)
- [Templates](#templates)
- [Add modules to an existing project](#add-modules-to-an-existing-project)
- [Update dewlon](#update-dewlon)
- [Uninstall](#uninstall)
- [Security and verification](#security-and-verification)
- [Private templates](#private-templates)
- [Settings and environment variables](#settings-and-environment-variables)
- [Troubleshooting](#troubleshooting)
- [Questions and answers](#questions-and-answers)
- [For contributors](#for-contributors)

---

## Install

Pick **one** way to install and stick with it on each computer. Mixing methods (for example the installer and `go install`) can leave two copies of `dewlon` on one machine, and the older one may be the one that runs.

### Install on Windows

1. Open **PowerShell**. Press the Windows key, type `PowerShell`, and press Enter. (Not "Command Prompt": the command below only works in PowerShell.)
2. Paste this and press Enter:

   ```powershell
   irm https://raw.githubusercontent.com/dewlonsystems/dewlon/main/scripts/install.ps1 | iex
   ```

3. **Close the window and open a new one.** Terminals only pick up PATH changes when they start.
4. Run `dewlon version` to confirm.

What the installer does:

- finds the latest release (or the one named in `DEWLON_VERSION`)
- downloads the Windows build for your processor (Intel/AMD or ARM)
- checks the download against the release's published SHA-256 checksum and stops if it doesn't match
- installs `dewlon.exe` to `%LOCALAPPDATA%\Programs\dewlon` (usually `C:\Users\<you>\AppData\Local\Programs\dewlon`)
- adds that folder to your user PATH, so `dewlon` works from any folder

It does not need administrator rights.

### Install on macOS and Linux

1. Open a terminal.
2. Run:

   ```bash
   curl -fsSL https://raw.githubusercontent.com/dewlonsystems/dewlon/main/scripts/install.sh | sh
   ```

3. Run `dewlon version` to confirm.

What the installer does:

- finds the latest release (or the one named in `DEWLON_VERSION`)
- downloads the build for your system (`linux` or `darwin`, `amd64` or `arm64`)
- checks the download against the published SHA-256 checksum and stops if it doesn't match
- installs `dewlon` to `/usr/local/bin` if you can write there, otherwise to `~/.local/bin`
- tells you if the install folder is not on your PATH

To choose the folder yourself:

```bash
curl -fsSL https://raw.githubusercontent.com/dewlonsystems/dewlon/main/scripts/install.sh | DEWLON_INSTALL_DIR="$HOME/bin" sh
```

To install a specific version instead of the latest:

```bash
curl -fsSL https://raw.githubusercontent.com/dewlonsystems/dewlon/main/scripts/install.sh | DEWLON_VERSION=v0.4.0 sh
```

On Windows, set the version first: `$env:DEWLON_VERSION = "v0.4.0"`, then run the install line.

### Download by hand (no commands)

Use this if you prefer not to run an installer script, or your network blocks it.

#### Windows, by hand

1. Go to https://github.com/dewlonsystems/dewlon/releases/latest.
2. Under **Assets**, download `dewlon_<version>_windows_amd64.zip`. Choose `windows_arm64` only if your PC has an ARM processor (see *Settings, System, About, System type*).
3. Right-click the downloaded zip, choose **Properties**, and tick **Unblock** if the box is there. Click OK.
4. Right-click the zip and choose **Extract All**. Inside is `dewlon.exe`.
5. Create a permanent folder, for example `C:\Tools\dewlon`, and move `dewlon.exe` into it. Don't leave it in Downloads: if that folder is cleaned up, dewlon disappears.
6. Add that folder to PATH:
   1. Press the Windows key, type `Edit environment variables for your account`, and open it.
   2. Under **User variables**, select **Path** and click **Edit**.
   3. Click **New** and type `C:\Tools\dewlon` (the folder, not the `.exe`).
   4. Click **OK** on every window.
7. Open a **new** terminal and run `dewlon version`.

If Windows shows "Windows protected your PC" (SmartScreen), click **More info**, then **Run anyway**. dewlon is not code-signed yet, so Windows doesn't recognise it.

#### macOS and Linux, by hand

1. Go to https://github.com/dewlonsystems/dewlon/releases/latest.
2. Download the `.tar.gz` that matches your system:
   - `darwin_arm64`: Mac with Apple Silicon (M1 or newer)
   - `darwin_amd64`: Mac with an Intel processor
   - `linux_amd64`: most Linux PCs and servers
   - `linux_arm64`: ARM Linux (for example Raspberry Pi 64-bit)
3. Extract it. You get a file named `dewlon`.
4. Move `dewlon` into a folder on your PATH, for example `/usr/local/bin`, and make sure it is executable (`chmod +x dewlon`).
5. On a Mac, the first run may be blocked as "from an unidentified developer". Open **System Settings, Privacy & Security** and click **Open Anyway**.

#### Archive names

Release files follow one pattern:

```
dewlon_<version>_<os>_<arch>.<ext>
```

- `<version>` has no leading `v` (for example `0.4.0`)
- `<os>` is `windows`, `darwin` (macOS) or `linux`
- `<arch>` is `amd64` (Intel/AMD) or `arm64` (ARM)
- `<ext>` is `zip` on Windows and `tar.gz` everywhere else

Every release also has a `checksums.txt` with the SHA-256 of each file.

#### Checking a manual download (optional)

Compare the file's SHA-256 with the matching line in `checksums.txt` from the same release:

```powershell
# Windows PowerShell
Get-FileHash .\dewlon_0.4.0_windows_amd64.zip -Algorithm SHA256
```

```bash
# macOS
shasum -a 256 dewlon_0.4.0_darwin_arm64.tar.gz

# Linux
sha256sum dewlon_0.4.0_linux_amd64.tar.gz
```

If the two values differ, don't use the file. Download it again.

### Other ways to install

| Method | Command | Notes |
|---|---|---|
| **Homebrew** (macOS) | `brew install dewlonsystems/tap/dewlon` | Update with `brew upgrade dewlon` |
| **Scoop** (Windows) | `scoop bucket add dewlon https://github.com/dewlonsystems/scoop-bucket` then `scoop install dewlon` | Update with `scoop update dewlon` |
| **Go** (any system) | `go install github.com/dewlonsystems/dewlon@latest` | Needs Go installed. Builds from source. Installs to `<GOPATH>/bin` (on Windows, `C:\Users\<you>\go\bin`), which must be on your PATH |

### Check that it worked

Open a **new** terminal and run:

```
dewlon version
```

You should see something like:

```
dewlon v0.4.0
  commit:   d522a40
  built:    2026-10-30T18:39:16Z
  go:       go1.27.1
  platform: windows/amd64
```

Then run:

```
dewlon doctor
```

`doctor` checks for the tools your projects use (Go, Flutter, Git) and whether github.com is reachable, and tells you where to download anything that's missing.

If `dewlon` says "not recognized" or "command not found", see [Troubleshooting](#troubleshooting).

---

## Create your first project

```
dewlon new hospital-api
```

This creates a folder called `hospital-api` in the current folder, using the default template (`fiber`).

When it finishes, dewlon prints the next steps, for example:

```
✓ Created hospital-api in 3.2s
✓ Template checksum verified

Next steps:
  cd hospital-api
  go run .
```

More examples:

```
dewlon new school-api --template=fiber       # choose a template by name
dewlon new mobile-app -t flutter             # short form of --template
dewlon new billing --module github.com/acme/billing    # your own Go module path
dewlon new clinic-api --dir ~/work --ref v1.2.0        # create elsewhere, pin a template version
dewlon new                                   # no name: dewlon asks for the name and template
```

What happens when you run `dewlon new`:

1. The name, template, module path and options are checked first, so mistakes fail fast.
2. dewlon works out where to download the template from and which checksum to expect.
3. The template is downloaded, checked, unpacked into a hidden staging folder, and your project name is filled into the files.
4. Only when everything succeeded is the folder moved to its final name. If anything fails, or you press Ctrl+C, nothing is left behind under the project name.
5. Dependencies are installed (`go mod tidy` for fiber, `flutter pub get` for flutter). If this step fails, for example because you're offline, your project is **kept** and you get a warning with the command to run yourself.

### Project name rules

- letters, numbers, hyphens (`-`) and underscores (`_)` only
- must start with a letter or number
- at most 100 characters
- names Windows reserves for devices (`con`, `nul`, `aux`, `com1` and similar) are rejected
- dewlon never overwrites: if the folder already exists, it stops and tells you

---

## Commands

Run `dewlon --help` or `dewlon <command> --help` at any time.

| Command | What it does |
|---|---|
| `dewlon new [name]` | Create a project from a template |
| `dewlon add [module]` | Add a module to an existing Go project |
| `dewlon templates` | List the templates (aliases: `list`, `ls`) |
| `dewlon doctor` | Check your machine is ready |
| `dewlon update` | Check whether a newer dewlon release exists |
| `dewlon version` | Show version and build details |
| `dewlon completion <shell>` | Print a shell completion script (`bash`, `zsh`, `fish`, `powershell`) |

### Options for every command

| Option | Meaning |
|---|---|
| `-v`, `--verbose` | Show each step as it happens |
| `-q`, `--quiet` | Print only errors (and warnings) |
| `--no-color` | Turn off colors. The `NO_COLOR` environment variable does the same |
| `--version` | Print the version and exit |
| `-h`, `--help` | Show help |

### `dewlon new`

| Option | Meaning |
|---|---|
| `-t`, `--template <name>` | Template to use: `fiber` (default) or `flutter` |
| `-d`, `--dir <folder>` | Create the project inside this folder (default: current folder) |
| `-m`, `--module <path>` | Go module path (default: `github.com/dewlonsystems/<name>`) |
| `--ref <ref>` | Template branch, tag or commit to build from (default: the template's main branch) |
| `--sha256 <hash>` | Require the downloaded template to have this SHA-256 checksum |
| `--no-verify` | Skip checksum verification for tagged releases |
| `--no-install` | Don't install dependencies after creating the project |

With no project name:

- in a terminal, dewlon asks for the name and the template
- in a script or CI (no terminal), it stops with an error asking you to pass the name

### `dewlon doctor`

Checks for `go`, `flutter` and `git`, and whether `github.com` can be reached. Missing tools are shown as warnings, since you may only need some of them. Add `--strict` to exit with an error if any check fails (useful in CI).

### `dewlon update`

Looks up the latest release and tells you if yours is behind. It never changes anything on your machine. See [Update dewlon](#update-dewlon).

### `dewlon version`

Shows the version, commit, build date, Go version and platform. Use `dewlon version --short` to print only the version number.

### Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | An error occurred (a message is printed to stderr) |
| `130` | Cancelled with Ctrl+C |

---

## Templates

List them any time with `dewlon templates`.

| Template | Creates | Source repository | Dependency step |
|---|---|---|---|
| `fiber` (default) | A Go REST API built on Fiber | `dewlonsystems/fiber-template` | `go mod tidy` |
| `flutter` | A Flutter application | `dewlonsystems/flutter-template` | `flutter pub get` |

You need **Go** installed to build the fiber template and **Flutter** for the flutter template. Run `dewlon doctor` to check.

### Names filled in automatically

Templates contain placeholders that dewlon replaces, both inside files and in file and folder names:

| Placeholder | Replaced with | Example for `my-app` |
|---|---|---|
| `{{PROJECT_NAME}}` | the project name | `my-app` |
| `{{APP_NAME}}` | the project name | `my-app` |
| `{{MODULE_NAME}}` | the Go module path | `github.com/dewlonsystems/my-app` |
| `{{PACKAGE_NAME}}` | the name with hyphens turned into underscores, lowercase (needed by Dart) | `my_app` |
| `{{YEAR}}` | the current year | `2026` |

Binary files are left alone. If a template uses a placeholder like `{{SOMETHING}}` that dewlon doesn't know, you get a warning listing it.

---

## Add modules to an existing project

`dewlon add` drops ready-made pieces into a Go project you already have. Run it from the folder that holds `go.mod`, or pass `--dir`.

```
dewlon add                  # list the modules
dewlon add docker           # add the Docker files
dewlon add database --dry-run   # preview, change nothing
```

| Module | What it adds |
|---|---|
| `docker` | `Dockerfile` (multi-stage, small final image) and `.dockerignore` |
| `ci` | `.github/workflows/ci.yml`: runs `go vet` and `go test -race` on every push and pull request |
| `database` | `internal/database/database.go`: a PostgreSQL connection pool (pgx) configured from the `DATABASE_URL` environment variable. Also runs `go get` and `go mod tidy` for pgx |

The module reads your `go.mod` to fill in the Go version, the module path and where your `main` package is (the project root, or the only folder under `cmd/` that contains one).

Safety rules:

- existing files are **never overwritten** unless you pass `--force`
- `--dry-run` shows exactly which files would be created or overwritten
- if writing fails halfway, files already written are put back
- `--no-install` skips the automatic `go get`

| Option | Meaning |
|---|---|
| `-d`, `--dir <folder>` | Project folder (the one containing `go.mod`). Default: current folder |
| `--force` | Overwrite files that already exist |
| `--dry-run` | Show what would be written without changing anything |
| `--no-install` | Don't fetch Go dependencies afterwards |

---

## Update dewlon

```
dewlon update
```

This asks GitHub for the latest release and compares it with the version you're running. If yours is older, it shows the new version, a link to the release notes, and the command that fits the way you installed:

| Installed with | Upgrade command shown |
|---|---|
| the Windows installer | `irm https://raw.githubusercontent.com/dewlonsystems/dewlon/main/scripts/install.ps1 \| iex` |
| the macOS/Linux installer | `curl -fsSL https://raw.githubusercontent.com/dewlonsystems/dewlon/main/scripts/install.sh \| sh` |
| Homebrew | `brew upgrade dewlon` |
| Scoop | `scoop update dewlon` |
| `go install` | `go install github.com/dewlonsystems/dewlon@latest` |

For scripts and CI, `dewlon update --fail-if-outdated` exits with an error when a newer release exists.

If you built dewlon from source it reports a "development build" and just names the latest release.

Versions before the one that introduced `update` don't have this command. Upgrade them once with the installer command for your system, as in [Install](#install).

---

## Uninstall

| Installed with | How to remove it |
|---|---|
| Windows installer | Delete the folder `%LOCALAPPDATA%\Programs\dewlon`, then remove that folder from your PATH (*Edit environment variables for your account*, then **Path**) |
| macOS/Linux installer | Delete the file `dewlon` from where it was installed (`which dewlon` shows where) |
| Homebrew | `brew uninstall dewlon` |
| Scoop | `scoop uninstall dewlon` |
| `go install` | Delete `dewlon` (or `dewlon.exe`) from your Go `bin` folder |
| By hand | Delete the folder you put it in and remove it from PATH |

dewlon keeps no configuration or data on your computer, so there is nothing else to clean up. Projects you created are ordinary folders and stay where they are.

---

## Security and verification

- **Installers verify downloads.** Both install scripts compare the file's SHA-256 with the release's `checksums.txt` and refuse to install on a mismatch.
- **Templates are verified.** When you build from a release tag (`--ref v1.2.0`), dewlon downloads the template's release files and checks them against the checksum its maintainers published. See [Template checksums](#template-checksums).
- **Archives are unpacked safely.** dewlon refuses files that would land outside the project folder, refuses symbolic links, and enforces limits on the number of files (20,000), total unpacked size (500 MB) and download size (200 MB).
- **Tokens stay with GitHub.** A GitHub token from your environment is only ever sent to github.com and api.github.com.
- **Nothing is executed from templates.** dewlon only downloads, unpacks and fills in names. The one command it runs afterwards is the dependency step (`go mod tidy` or `flutter pub get`), which you can skip with `--no-install`.

### Template checksums

| You pass | What dewlon does |
|---|---|
| nothing | Uses the template's default branch. **Not verified**, because a branch changes over time |
| `--ref v1.2.0` (a release tag) | Downloads `template.zip` from that release and checks it against the release's `SHA256SUMS`. Fails if the release has neither file |
| `--ref v1.2.0 --no-verify` | Uses the plain source archive for that tag, unverified |
| `--ref main --sha256 <hash>` | Downloads the branch and requires it to match the hash you give |
| `--ref v1.2.0 --sha256 <hash>` | Requires the published checksum **and** your hash to agree |

With `--verbose`, dewlon prints `Checksum verified` or says the download wasn't verified.

**For template authors:** to make a tag verifiable, the release must contain two files:

- `template.zip`: the template, with a single top-level folder (`git archive --format=zip --prefix=<name>/ ...` does this)
- `SHA256SUMS`: checksums in the standard `sha256sum` format, listing `template.zip`

The workflow in the template repositories builds and attaches both when you push a tag such as `v1.0.0`.

---

## Private templates

If a template repository is private, give dewlon a GitHub token that can read it, in an environment variable:

```powershell
# Windows PowerShell (this window only)
$env:DEWLON_GITHUB_TOKEN = "your-token"
```

```bash
# macOS / Linux (this terminal only)
export DEWLON_GITHUB_TOKEN="your-token"
```

dewlon looks for a token in this order: `DEWLON_GITHUB_TOKEN`, `GITHUB_TOKEN`, `GH_TOKEN`.

A fine-grained personal access token with **Contents: Read** on the template repositories is enough. A token also raises GitHub's rate limits for `dewlon update` and for release lookups.

---

## Settings and environment variables

dewlon has no configuration file. These environment variables change its behavior:

| Variable | Used by | Meaning |
|---|---|---|
| `DEWLON_GITHUB_TOKEN`, `GITHUB_TOKEN`, `GH_TOKEN` | dewlon | GitHub token for private templates and higher rate limits (checked in this order) |
| `NO_COLOR` | dewlon | If set, turns colors off |
| `DEWLON_VERSION` | install scripts | Version to install, such as `v0.4.0`. Default: the latest |
| `DEWLON_INSTALL_DIR` | install scripts | Folder to install into |

---

## Troubleshooting

### "dewlon is not recognized" or "command not found"

The folder holding dewlon isn't on your PATH, or your terminal was opened before it was added.

1. **Close all terminals and open a new one.** This fixes it most of the time.
2. Check where it is: `where.exe dewlon` on Windows, `which dewlon` on macOS/Linux.
3. If nothing is found, add the install folder to PATH (see [Install](#install)).

### `dewlon version` shows an old version after installing

You probably have two copies, and the old one comes first on PATH.

- Windows PowerShell: run `where.exe dewlon` (with `.exe`). In PowerShell, plain `where` is a different command and prints nothing.
- macOS/Linux: run `which -a dewlon`.
- Delete the old copy and open a new terminal.

### `where dewlon` prints nothing in PowerShell

Use `where.exe dewlon`. `where` alone is a PowerShell alias for something else.

### The Windows install command fails or does nothing in cmd

The one-line installer needs **PowerShell**. In Command Prompt (cmd) it won't work. Open PowerShell and run it there.

### Windows says "Windows protected your PC"

dewlon is not code-signed yet. Click **More info**, then **Run anyway**.

### macOS says the app is from an unidentified developer

Open **System Settings, Privacy & Security** and click **Open Anyway**. This only happens for files downloaded in a browser. Installing with Homebrew or the install script avoids it.

### "checksum mismatch"

The downloaded file doesn't match the published checksum, so dewlon discarded it. This can happen if the download was interrupted or tampered with. Try again. If it keeps happening, report it at https://github.com/dewlonsystems/dewlon/issues.

### "not found (404)" when creating a project

- Check the template name (`dewlon templates`) and the `--ref` value (tags need the exact spelling, including the `v`).
- If the template repository is private, set `DEWLON_GITHUB_TOKEN` (see [Private templates](#private-templates)).

### "GitHub refused the request (403)" or rate limit

GitHub limits unauthenticated requests. Set `DEWLON_GITHUB_TOKEN` to raise the limit.

### "has no template.zip and SHA256SUMS, so it can't be verified"

The release you asked for was published without the verification files. Either pick a release that has them, or pass `--no-verify` to use the plain source archive.

### "already exists; choose a different name"

dewlon never writes into an existing folder. Pick another name or remove the folder first.

### Warning: "dependencies were not installed"

The project was created, but the dependency step failed or the tool isn't installed. Run the command from the warning inside the project folder (`go mod tidy` or `flutter pub get`). Run `dewlon doctor` to see what's missing.

### "no project name given and stdin is not a terminal"

You ran `dewlon new` without a name in a place with no interactive terminal (a script or CI). Pass the name: `dewlon new my-project`.

---

## Questions and answers

**Do I need Go installed to use dewlon?**
No, the dewlon program itself is a single file with nothing else required. You need Go only to build and run projects from the `fiber` template (and Flutter for the `flutter` template).

**Does dewlon work offline?**
Creating a project needs internet access to download the template. `dewlon add` works offline except for fetching dependencies, which you can skip with `--no-install`.

**Where do I put the project?**
Wherever you run the command, or use `--dir <folder>`.

**Can I use my own Go module path?**
Yes: `dewlon new my-api --module github.com/acme/my-api`.

**Can I pin a template version?**
Yes: `--ref v1.2.0`. Release tags are verified against published checksums.

**Does dewlon collect any data?**
No. It talks only to GitHub (to download templates and check for releases) and stores nothing on your machine.

**How do I get autocomplete?**
Run `dewlon completion --help` for your shell (`bash`, `zsh`, `fish` or `powershell`). It prints a script to load in your shell profile.

**How do I report a bug or ask for a feature?**
Open an issue at https://github.com/dewlonsystems/dewlon/issues. Include the output of `dewlon version` and `dewlon doctor`, and the exact command you ran.

---

## For contributors

### Build and test

You need Go (the version in `go.mod`).

```
git clone https://github.com/dewlonsystems/dewlon.git
cd dewlon
go build -o dewlon .
go vet ./...
go test ./...
```

Run the linter with [golangci-lint](https://golangci-lint.run) (`golangci-lint run`). The same checks run in CI on Linux, macOS and Windows for every push and pull request.

### Project layout

```
main.go                     entry point
cmd/                        the commands (new, add, templates, doctor, update, version)
internal/buildinfo/         version, commit and build date
internal/templates/         the list of templates and where they live
internal/source/            decides where a template is downloaded from and verifies checksums
internal/downloader/        HTTP download with retries, size limits and hashing
internal/github/            small GitHub API client (releases, tokens)
internal/unzip/             safe zip extraction
internal/filesystem/        folder helpers
internal/placeholder/       fills in {{PLACEHOLDER}} values
internal/runner/            runs go, flutter and other tools
internal/project/           builds a new project from a template
internal/modules/           modules for `dewlon add` (files are embedded in the binary)
internal/update/            release lookup and version comparison
scripts/                    install.sh (macOS/Linux) and install.ps1 (Windows)
```

### Adding a template

Add an entry to the list in `internal/templates/templates.go` with its key, description, GitHub repository, default branch, dependency command and next steps. The `new` command, the interactive picker, `dewlon templates` and shell completion all read from that one list.

### Adding a module

1. Create a folder `internal/modules/files/<key>/` containing the files to add. Give every file a `.tmpl` suffix (it is removed when written). Use the placeholders `{{PROJECT_NAME}}`, `{{MODULE_NAME}}`, `{{GO_VERSION}}`, `{{MAIN_PACKAGE}}` and `{{YEAR}}`.
2. Add an entry for it in `internal/modules/modules.go`.

### Commit messages and releases

Versions are chosen automatically from commit messages (the [Conventional Commits](https://www.conventionalcommits.org) style) by [release-please](https://github.com/googleapis/release-please). You never type a version number or create a tag by hand.

| Commit starts with | Effect on the version (while below 1.0) |
|---|---|
| `fix:` | patch bump: 0.4.0 to 0.4.1 |
| `feat:` | minor bump: 0.4.0 to 0.5.0 |
| `feat!:` or a `BREAKING CHANGE:` footer | minor bump (until version 1.0) |
| `docs:`, `chore:`, `ci:`, `test:`, `refactor:` | no release on their own |

How a release happens:

1. You push commits to `main`.
2. release-please keeps a pull request open titled like **"chore(main): release 0.5.0"**, with the new version and changelog.
3. When you merge that pull request, it creates the tag and GitHub release. The same workflow then runs [GoReleaser](https://goreleaser.com), which builds the binaries for Windows, macOS and Linux, writes `checksums.txt`, uploads everything to the release, and updates the Homebrew tap and Scoop bucket.

To force a specific next version, add an empty commit with a footer: `git commit --allow-empty -m "chore: release 0.5.0" -m "Release-As: 0.5.0"`.

If a release build fails after the tag exists, run the **Release** workflow manually from the Actions tab and enter the tag.

Release setup requirements (already in place for this repository):

- a secret named `PACKAGES_GITHUB_TOKEN` (a token with *Contents: read and write* on the `homebrew-tap` and `scoop-bucket` repositories)
- the repositories `dewlonsystems/homebrew-tap` and `dewlonsystems/scoop-bucket`
- *Settings, Actions, General*: "Allow GitHub Actions to create and approve pull requests" turned on

---

Copyright © 2026 Dewlon Systems.