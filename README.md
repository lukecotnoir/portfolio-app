# A terminal-ish portfolio

two ways to access it: an Angular web app and a Go SSH server,
both rendering the terminal look over the same content model.

```
                    content/content.json
                   (single source of truth)
                    /                  \
            web/ (Angular)         ssh/ (Go + wish + bubbletea)
            ng build → Pages       binary → :22 on a $5 VPS
            browser renders        ssh guest@lcotnoir.dev
            styled DOM lines       renders styled TUI lines
```

## layout

| path | what |
| --- | --- |
| `content/` | resume + copy as data. edit this, run `make sync` |
| `web/` | Angular app. `ng serve` / `ng build` |
| `ssh/` | Go SSH server (wish + bubbletea). `go run .` → port 2222 |
| `deploy/` | systemd unit + go-live notes |
| `.github/workflows/` | pushes web → GitHub Pages on every commit to main |

## dev

```sh
make sync       # copy content.json into web/public + ssh/
make dev-web    # ng serve → http://localhost:4200
make dev-ssh    # go run . → ssh -p 2222 guest@localhost
make build      # production build of both
```

## try the ssh side locally

```sh
cd ssh && go run .
# in another terminal:
ssh -p 2222 guest@localhost    # type 'help', then 'exit'
```

## go live

See `deploy/DEPLOY.md`. Short version: domain + $5/mo VPS,
`scp` the binary, run it under systemd on port 22, point DNS.
