# countdown

Countdown is a terminal-based multi-event countdown timer with a graphical timeline. It uses the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework from [Charm\_](https://charm.sh/).

## Features

- **Multiple events**: Track as many countdowns as you need
- **Visual timeline**: See upcoming events on a proportional timeline
- **Urgency colors**: Events change color as they get closer (green → yellow → orange → red)
- **Past events**: Track events that have already passed
- **Live updates**: Countdowns update every second
- **Detailed statistics**: View total seconds, minutes, hours, days, and years
- **Responsive layout**: Adapts to your terminal size

## Installation

Install using Golang's tools:

```bash
go install github.com/rom41572/countdown@latest
```

Or clone and build:

```bash
git clone https://github.com/rom41572/countdown.git
cd countdown
go build -o countdown main.go
```

## Configuration

When you launch it for the first time, an `events.json` file will be created in the user's system-defined config directory:

- **Linux**: `~/.config/countdown/`
- **macOS**: `~/Library/Application Support/countdown/`
- **Windows**: `%APPDATA%\countdown\`

On the first startup, one prepopulated event (Golang's next anniversary) will be shown.

## Usage

### Keyboard Controls

| Key         | Action                    |
| ----------- | ------------------------- |
| `+`         | Add a new event           |
| `-`         | Remove selected event     |
| `e`         | Edit selected event       |
| `↑`/`↓`     | Navigate events           |
| `/`         | Filter events             |
| `Tab`       | Next field (in forms)     |
| `Shift+Tab` | Previous field (in forms) |
| `Enter`     | Select/confirm            |
| `Esc`       | Cancel/go back            |
| `q`         | Quit                      |

### Date Formats

When adding or editing events, use one of these formats:

- **Date only**: `2025-12-31` (time defaults to 00:00:00)
- **Date and time**: `2025-12-31 18:30:00`

### Interface

The interface has three panels:

1. **Events List** (left): All your events sorted by date
2. **Event Details** (center): Detailed countdown for the selected event
3. **Timeline** (right): Visual timeline of upcoming events with proportional bars

### Timeline

The timeline shows upcoming events with visual bars representing time distance:

```
▼ NOW
│
├─■■■■■·························
│ ● Meeting
│   Dec 12 • 3 days
│
├─■■■■■■■■■■■■■■■···············
│ ◆ Birthday Party
│   Dec 25 • 16 days
│
▽ future
```

- `◆` = Selected event
- `●` = Future event
- Bar length = relative time until event

### Urgency Colors

Events are color-coded based on how soon they occur:

| Time Remaining | Color       |
| -------------- | ----------- |
| > 30 days      | Green       |
| 14-30 days     | Light green |
| 7-14 days      | Yellow      |
| 3-7 days       | Orange      |
| 1-3 days       | Red         |
| < 1 day        | Dark red    |
| Past           | Purple      |

## License

MIT
