# ArvanCloud Status CLI

A terminal UI that polls the [ArvanCloud status page](https://www.arvancloudstatus.ir)
and shows live service health and recent incidents. Built with
[Bubble Tea v2](https://charm.land/bubbletea/v2).

```
╭──────────────────────────────────────────────────────────────────────╮
│ ArvanCloud  all systems operational                                  │
│  ● Cloud Server                                                      │
│      ● Forough (Iran-Tehran/Central 2)                               │
│      ● Simin (Iran-Tehran/Central 1)                                 │
│  ● CDN                                                               │
│      ● DNS                                                           │
│      ● CDN                                                           │
╰──────────────────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────────╮
│ Incidents (50 shown)                                                 │
│ ▸ ● [MINOR]     Foreign Connectivity Issue   resolved · 2d ago       │
│   ● [MINOR]     DNS Resolution Issue         resolved · 3d ago       │
│   ● [SCHEDULED] LCT Network Maintenance      resolved · 1w ago       │
╰──────────────────────────────────────────────────────────────────────╯
↑/k up • ↓/j down • tab switch pane • enter detail • r refresh • q quit
```

## Install

```sh
go install github.com/parham-alvani/arvanstatus-cli@latest
```

…or build from source:

```sh
git clone https://github.com/parham-alvani/arvanstatus-cli
cd arvanstatus-cli
go build -o arvanstatus-cli .
./arvanstatus-cli
```

## Usage

Just run the binary — no flags, no config file:

```sh
arvanstatus-cli
```

The top pane shows the full ArvanCloud service tree (region groups → services)
colored by status. The bottom pane lists the most recent incidents from the
public Cachet API. Both panes refresh automatically every 60 seconds.

### Keybindings

| Key       | Action                          |
|-----------|---------------------------------|
| `↑` / `k` | move cursor up                  |
| `↓` / `j` | move cursor down                |
| `tab`     | switch focused pane             |
| `enter`   | open incident detail            |
| `esc`     | close detail                    |
| `r`       | refresh now                     |
| `?`       | toggle full help                |
| `q`, `^C` | quit                            |

### Status colors

| Color  | Component status    | Incident status |
|--------|---------------------|-----------------|
| green  | operational         | fixed           |
| yellow | performance issues  | watching        |
| orange | partial outage      | identified      |
| red    | major outage        | investigating   |

## Data source

ArvanCloud's status page is powered by [Cachet](https://cachethq.io) and served
at [`www.arvancloudstatus.ir`](https://www.arvancloudstatus.ir). This tool calls
its unauthenticated Cachet v3 JSON:API:

- `GET https://www.arvancloudstatus.ir/api/component-groups`
- `GET https://www.arvancloudstatus.ir/api/components?include=group`
- `GET https://www.arvancloudstatus.ir/api/incidents?include=updates,components`

Components and their groups are assembled into a region → service tree.
Timestamps are returned in `Asia/Tehran` local time without offset and are
parsed accordingly.
