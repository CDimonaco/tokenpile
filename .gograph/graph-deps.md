# Dependencies & Imports

**Root:** `/Users/cdimonaco/code/github.com/cdimonaco/tokenpile`  
**Generated:** 2026-07-05 15:04:58 UTC  

---

## External Dependencies (Tech Stack)

| Module | Version |
|--------|---------|
| `github.com/aymanbagabas/go-osc52/v2` | `v2.0.1` |
| `github.com/aymanbagabas/go-udiff` | `v0.3.1` |
| `github.com/charmbracelet/bubbletea` | `v1.3.10` |
| `github.com/charmbracelet/colorprofile` | `v0.3.2` |
| `github.com/charmbracelet/lipgloss` | `v1.1.0` |
| `github.com/charmbracelet/x/ansi` | `v0.10.1` |
| `github.com/charmbracelet/x/cellbuf` | `v0.0.13-0.20250311204145-2c3ea96c31dd` |
| `github.com/charmbracelet/x/exp/golden` | `v0.0.0-20240806155701-69247e0abc2a` |
| `github.com/charmbracelet/x/exp/teatest` | `v0.0.0-20260705004817-2cc9a8fe1146` |
| `github.com/charmbracelet/x/term` | `v0.2.1` |
| `github.com/cpuguy83/go-md2man/v2` | `v2.0.7` |
| `github.com/danieljoos/wincred` | `v1.2.3` |
| `github.com/davecgh/go-spew` | `v1.1.2-0.20180830191138-d8f796af33cc` |
| `github.com/dustin/go-humanize` | `v1.0.1` |
| `github.com/erikgeiser/coninput` | `v0.0.0-20211004153227-1c3628e74d0f` |
| `github.com/godbus/dbus/v5` | `v5.2.2` |
| `github.com/google/go-github/v68` | `v68.0.0` |
| `github.com/google/go-querystring` | `v1.1.0` |
| `github.com/google/uuid` | `v1.6.0` |
| `github.com/kr/text` | `v0.2.0` |
| `github.com/lucasb-eyer/go-colorful` | `v1.2.0` |
| `github.com/mattn/go-isatty` | `v0.0.20` |
| `github.com/mattn/go-localereader` | `v0.0.1` |
| `github.com/mattn/go-runewidth` | `v0.0.16` |
| `github.com/muesli/ansi` | `v0.0.0-20230316100256-276c6243b2f6` |
| `github.com/muesli/cancelreader` | `v0.2.2` |
| `github.com/muesli/termenv` | `v0.16.0` |
| `github.com/ncruces/go-strftime` | `v1.0.0` |
| `github.com/niemeyer/pretty` | `v0.0.0-20200227124842-a10e7caefd8e` |
| `github.com/pmezard/go-difflib` | `v1.0.1-0.20181226105442-5d4384ee4fb2` |
| `github.com/remyoudompheng/bigfft` | `v0.0.0-20230129092748-24d4a6f8daec` |
| `github.com/rivo/uniseg` | `v0.4.7` |
| `github.com/russross/blackfriday/v2` | `v2.1.0` |
| `github.com/stretchr/objx` | `v0.5.2` |
| `github.com/stretchr/testify` | `v1.11.1` |
| `github.com/urfave/cli/v2` | `v2.27.7` |
| `github.com/xo/terminfo` | `v0.0.0-20220910002029-abceb7e1c41e` |
| `github.com/xrash/smetrics` | `v0.0.0-20240521201337-686a1a2994c1` |
| `github.com/zalando/go-keyring` | `v0.2.8` |
| `golang.org/x/oauth2` | `v0.36.0` |
| `golang.org/x/sys` | `v0.44.0` |
| `golang.org/x/text` | `v0.28.0` |
| `gopkg.in/check.v1` | `v1.0.0-20200227125254-8fa46927fb4f` |
| `gopkg.in/yaml.v3` | `v3.0.1` |
| `modernc.org/libc` | `v1.73.4` |
| `modernc.org/mathutil` | `v1.7.1` |
| `modernc.org/memory` | `v1.11.0` |
| `modernc.org/sqlite` | `v1.53.0` |

## Package Imports

| Package | Imports |
|---------|--------|
| `config` | `crypto/ed25519`, `crypto/rand`, `encoding/pem`, `errors`, `fmt`, `log/slog`, `os`, `path/filepath` |
| `config_test` | `github.com/cdimonaco/tokenpile/internal/config`, `github.com/stretchr/testify/assert`, `github.com/stretchr/testify/require`, `os`, `path/filepath`, `testing` |
| `export` | `bytes`, `crypto/ed25519`, `crypto/sha256`, `encoding/base64`, `encoding/json`, `errors`, `fmt`, `github.com/cdimonaco/tokenpile/internal/schema`, `github.com/cdimonaco/tokenpile/internal/usage`, `github.com/stretchr/testify/assert`, `github.com/stretchr/testify/require`, `os`, `sort`, `testing`, `time` |
| `export_test` | `crypto/ed25519`, `crypto/rand`, `encoding/base64`, `encoding/json`, `github.com/cdimonaco/tokenpile/internal/export`, `github.com/cdimonaco/tokenpile/internal/usage`, `github.com/stretchr/testify/assert`, `github.com/stretchr/testify/require`, `os`, `testing`, `time` |
| `main` | `bytes`, `context`, `crypto/ed25519`, `crypto/subtle`, `encoding/base64`, `encoding/json`, `encoding/pem`, `errors`, `fmt`, `github.com/cdimonaco/tokenpile/internal/config`, `github.com/cdimonaco/tokenpile/internal/export`, `github.com/cdimonaco/tokenpile/internal/mocks`, `github.com/cdimonaco/tokenpile/internal/pricing`, `github.com/cdimonaco/tokenpile/internal/provider`, `github.com/cdimonaco/tokenpile/internal/skill`, `github.com/cdimonaco/tokenpile/internal/store`, `github.com/cdimonaco/tokenpile/internal/tui`, `github.com/cdimonaco/tokenpile/internal/usage`, `github.com/charmbracelet/bubbletea`, `github.com/google/uuid`, `github.com/stretchr/testify/assert`, `github.com/stretchr/testify/mock`, `github.com/stretchr/testify/require`, `github.com/urfave/cli/v2`, `log/slog`, `os`, `os/exec`, `path/filepath`, `sort`, `strings`, `testing`, `time`, `unicode/utf8` |
| `pricing` | `embed`, `fmt`, `gopkg.in/yaml.v3`, `maps`, `os` |
| `pricing_test` | `github.com/cdimonaco/tokenpile/internal/pricing`, `github.com/stretchr/testify/assert`, `github.com/stretchr/testify/require`, `os`, `path/filepath`, `testing` |
| `provider` | `context`, `crypto/aes`, `crypto/cipher`, `crypto/rand`, `crypto/sha256`, `encoding/hex`, `errors`, `fmt`, `github.com/cdimonaco/tokenpile/internal/usage`, `github.com/google/go-github/v68/github`, `github.com/stretchr/testify/assert`, `github.com/stretchr/testify/require`, `github.com/zalando/go-keyring`, `golang.org/x/oauth2`, `golang.org/x/oauth2/github`, `io`, `log/slog`, `net`, `net/http`, `os`, `os/exec`, `path/filepath`, `regexp`, `runtime`, `strconv`, `strings`, `testing`, `time` |
| `provider_test` | `context`, `encoding/json`, `github.com/cdimonaco/tokenpile/internal/mocks`, `github.com/cdimonaco/tokenpile/internal/provider`, `github.com/cdimonaco/tokenpile/internal/usage`, `github.com/stretchr/testify/assert`, `github.com/stretchr/testify/require`, `net/http`, `net/http/httptest`, `testing` |
| `schema` | `embed` |
| `skill` | `embed`, `errors`, `fmt`, `os`, `path/filepath`, `strings` |
| `skill_test` | `github.com/cdimonaco/tokenpile/internal/skill`, `github.com/stretchr/testify/assert`, `github.com/stretchr/testify/require`, `os`, `path/filepath`, `strings`, `testing` |
| `store` | `context`, `database/sql`, `encoding/json`, `errors`, `fmt`, `github.com/cdimonaco/tokenpile/internal/usage`, `github.com/google/uuid`, `log/slog`, `modernc.org/sqlite`, `strings`, `time` |
| `store_test` | `context`, `github.com/cdimonaco/tokenpile/internal/store`, `github.com/cdimonaco/tokenpile/internal/usage`, `github.com/stretchr/testify/assert`, `github.com/stretchr/testify/require`, `path/filepath`, `testing`, `time` |
| `tui` | `context`, `fmt`, `github.com/cdimonaco/tokenpile/internal/pricing`, `github.com/cdimonaco/tokenpile/internal/provider`, `github.com/cdimonaco/tokenpile/internal/store`, `github.com/cdimonaco/tokenpile/internal/usage`, `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, `github.com/charmbracelet/x/exp/teatest`, `github.com/stretchr/testify/assert`, `github.com/stretchr/testify/require`, `log/slog`, `os/exec`, `path/filepath`, `regexp`, `runtime`, `strings`, `testing`, `time` |
| `usage` | `time` |

