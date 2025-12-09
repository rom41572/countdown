package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	secondsPerYear     = 31557600
	secondsPerDay      = 86400
	secondsPerHour     = 3600
	secondsPerMinute   = 60
	timeout            = 365 * 24 * time.Hour
	minListWidth       = 28
	minDetailWidth     = 50
	minTimelineWidth   = 45
	appName            = "countdown"
	eventsFileName     = "events.json"
	inputTimeFormShort = "2006-01-02"
	inputTimeFormLong  = "2006-01-02 15:04:05"
	cError             = "#CF002E"
	cItemTitleDark     = "#F5EB6D"
	cItemTitleLight    = "#F3B512"
	cItemDescDark      = "#9E9742"
	cItemDescLight     = "#FFD975"
	cTitle             = "#2389D3"
	cDetailTitle       = "#D32389"
	cPromptBorder      = "#D32389"
	cDimmedTitleDark   = "#DDDDDD"
	cDimmedTitleLight  = "#222222"
	cDimmedDescDark    = "#999999"
	cDimmedDescLight   = "#555555"
	cTextLightGray     = "#000000ff"
	cSuccess           = "#146034ff"
	cWarning           = "#F39C12"
	cHint              = "#7F8C8D"
	cUrgency1          = "#347a51ff" // > 30 days (green)
	cUrgency2          = "#58D68D"   // 14-30 days (light green)
	cUrgency3          = "#F4D03F"   // 7-14 days (yellow)
	cUrgency4          = "#F39C12"   // 3-7 days (orange)
	cUrgency5          = "#E74C3C"   // 1-3 days (red)
	cUrgency6          = "#C0392B"   // < 1 day (dark red)
	cPast              = "#9B59B6"   // past events (purple)
	cBarEmpty          = "#2C3E50"
	cTimelineTrack     = "#34495E"
	cTimelineNow       = "#E74C3C"
	cTimelineFuture    = "#3498DB"
	cTimelineSelected  = "#F39C12"
)

func getEventsFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}

	appConfigDir := filepath.Join(configDir, appName)
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(appConfigDir, eventsFileName), nil
}

var AppStyle = lipgloss.NewStyle().Margin(0, 1)
var TitleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(cTextLightGray)).
	Background(lipgloss.Color(cTitle)).
	Padding(0, 1)
var SelectedTitle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder(), false, false, false, true).
	BorderForeground(lipgloss.AdaptiveColor{Light: cItemTitleLight, Dark: cItemTitleDark}).
	Foreground(lipgloss.AdaptiveColor{Light: cItemTitleLight, Dark: cItemTitleDark}).
	Padding(0, 0, 0, 1)
var SelectedDesc = SelectedTitle.Copy().
	Foreground(lipgloss.AdaptiveColor{Light: cItemDescLight, Dark: cItemDescDark})
var DimmedTitle = lipgloss.NewStyle().
	Foreground(lipgloss.AdaptiveColor{Light: cDimmedTitleLight, Dark: cDimmedTitleDark}).
	Padding(0, 0, 0, 2)
var DimmedDesc = DimmedTitle.Copy().
	Foreground(lipgloss.AdaptiveColor{Light: cDimmedDescDark, Dark: cDimmedDescLight})
var ErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(cError)).Render
var SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(cSuccess)).Render
var WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(cWarning)).Render
var HintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(cHint)).Render
var NoStyle = lipgloss.NewStyle()
var FocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(cPromptBorder))
var BlurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
var InputLabelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.AdaptiveColor{Light: cDimmedTitleLight, Dark: cDimmedTitleDark}).
	Bold(true).
	MarginTop(1)
var DatePreviewStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(cHint)).
	Italic(true).
	MarginLeft(2)
var ButtonStyle = lipgloss.NewStyle().
	Padding(0, 2).
	Border(lipgloss.RoundedBorder(), true).
	BorderForeground(lipgloss.Color("240"))
var ButtonFocusedStyle = ButtonStyle.Copy().
	BorderForeground(lipgloss.Color(cPromptBorder)).
	Foreground(lipgloss.Color(cPromptBorder)).
	Bold(true)

var BrightTextStyle = lipgloss.NewStyle().
	Foreground(lipgloss.AdaptiveColor{Light: cDimmedTitleLight, Dark: cDimmedTitleDark}).Render
var NormalTextStyle = lipgloss.NewStyle().
	Foreground(lipgloss.AdaptiveColor{Light: cDimmedDescLight, Dark: cDimmedDescDark}).Render

var TimelineTitleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(cTextLightGray)).
	Background(lipgloss.Color(cTitle)).
	Padding(0, 1).
	MarginBottom(1)
var TimelineTrackStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(cTimelineTrack))
var TimelineNowStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(cTimelineNow)).
	Bold(true)
var TimelineSelectedStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(cTimelineSelected)).
	Bold(true)

type keymap struct {
	Add    key.Binding
	Remove key.Binding
	Edit   key.Binding
	Next   key.Binding
	Prev   key.Binding
	Enter  key.Binding
	Back   key.Binding
	Quit   key.Binding
}

var Keymap = keymap{
	Add: key.NewBinding(
		key.WithKeys("+"),
		key.WithHelp("+", "add"),
	),
	Remove: key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "remove"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Next: key.NewBinding(
		key.WithKeys("tab"),
	),
	Prev: key.NewBinding(
		key.WithKeys("shift+tab"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctlr+c", "q"),
		key.WithHelp("q", "quit"),
	),
}

type sessionState int

const (
	showEvents sessionState = iota
	showInput
	showEdit
	noEvents
)

type inputFields int

const (
	inputNameField inputFields = iota
	inputTimeField
	inputCancelButton
	inputSubmitButton
)

type Event struct {
	Name string `json:"name"`
	Time int64  `json:"ts"`
}

func (e Event) ToBasicString() string {
	return time.Unix(e.Time, 0).String()
}

func (e Event) Title() string       { return e.Name }
func (e Event) Description() string { return countdownParser(e.Time) }
func (e Event) FilterValue() string { return e.Name }

type MainModel struct {
	state         sessionState
	focus         int
	events        list.Model
	inputs        []textinput.Model
	timer         timer.Model
	inputStatus   string
	datePreview   string
	dateValid     bool
	editIndex     int
	windowWidth   int
	windowHeight  int
	listWidth     int
	detailWidth   int
	timelineWidth int
}

func (m *MainModel) calculateWidths() {
	availableWidth := m.windowWidth - 6

	if availableWidth < minListWidth+minDetailWidth+minTimelineWidth {
		m.listWidth = minListWidth
		m.detailWidth = minDetailWidth
		m.timelineWidth = minTimelineWidth
	} else {
		m.listWidth = max(minListWidth, availableWidth*25/100)
		m.detailWidth = max(minDetailWidth, availableWidth*40/100)
		m.timelineWidth = max(minTimelineWidth, availableWidth*35/100)
	}
}

func NewMainModel() MainModel {
	m := MainModel{
		state:         showEvents,
		timer:         timer.NewWithInterval(timeout, time.Second),
		editIndex:     -1,
		windowWidth:   120,
		windowHeight:  40,
		listWidth:     minListWidth,
		detailWidth:   minDetailWidth,
		timelineWidth: minTimelineWidth,
	}
	events, err := readEventsFile()
	if err != nil {
		panic(err)
	}
	items := make([]list.Item, len(events))
	for i := range events {
		items[i] = events[i]
	}
	m.inputs = make([]textinput.Model, 2)
	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CharLimit = 50
		switch i {
		case 0:
			t.Placeholder = "e.g., Birthday Party"
			t.Focus()
			t.PromptStyle = FocusedStyle
			t.TextStyle = FocusedStyle
		case 1:
			t.Placeholder = "2025-12-31 or 2025-12-31 18:00:00"
			t.CharLimit = 19
		}
		m.inputs[i] = t
	}
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = SelectedTitle
	delegate.Styles.SelectedDesc = SelectedDesc
	delegate.Styles.DimmedTitle = DimmedTitle
	delegate.Styles.DimmedDesc = DimmedDesc
	delegate.ShortHelpFunc = func() []key.Binding { return []key.Binding{Keymap.Add, Keymap.Remove, Keymap.Edit} }
	delegate.FullHelpFunc = func() [][]key.Binding { return [][]key.Binding{{Keymap.Add, Keymap.Remove, Keymap.Edit}} }
	m.events = list.New(items, delegate, m.listWidth, 40)
	m.events.Title = "Events"
	m.events.Styles.Title = TitleStyle
	m.events.Styles.HelpStyle = lipgloss.NewStyle().Width(m.listWidth).Height(5)
	m.events.SetShowPagination(true)
	if len(m.events.Items()) == 0 {
		m.state = noEvents
	}
	return m
}

func (m MainModel) Init() tea.Cmd {
	return m.timer.Init()
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch m.state {
	case noEvents:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.windowWidth = msg.Width
			m.windowHeight = msg.Height
			m.calculateWidths()
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, Keymap.Add):
				m.state = showInput
			case key.Matches(msg, Keymap.Quit):
				return m, tea.Quit
			}
		}
	case showEvents:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.windowWidth = msg.Width
			m.windowHeight = msg.Height
			m.calculateWidths()
			_, v := AppStyle.GetFrameSize()
			m.events.SetSize(m.listWidth, msg.Height-v)
			m.events.Styles.HelpStyle = lipgloss.NewStyle().Width(m.listWidth).Height(5)
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, Keymap.Quit):
				return m, tea.Quit
			case key.Matches(msg, Keymap.Add):
				m.state = showInput
			case key.Matches(msg, Keymap.Edit):
				if len(m.events.Items()) > 0 {
					m.editIndex = m.events.Index()
					event := m.events.SelectedItem().(Event)
					m.inputs[0].SetValue(event.Name)
					ts := time.Unix(event.Time, 0)
					m.inputs[1].SetValue(ts.Format(inputTimeFormLong))
					m.updateDatePreview()
					m.state = showEdit
				}
			case key.Matches(msg, Keymap.Remove):
				m.events.RemoveItem(m.events.Index())
				if err := m.saveEventsToFile(); err != nil {
					panic(err)
				}
				if len(m.events.Items()) == 0 {
					m.state = noEvents
				}
			}
		}
		newEvents, newCmd := m.events.Update(msg)
		m.events = newEvents
		cmd = newCmd
	case showInput, showEdit:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.windowWidth = msg.Width
			m.windowHeight = msg.Height
			m.calculateWidths()
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, Keymap.Back):
				m.resetInputs()
				m.state = showEvents
				if len(m.events.Items()) == 0 {
					m.state = noEvents
				}
			case key.Matches(msg, Keymap.Next):
				m.focus++
				if m.focus > int(inputSubmitButton) {
					m.focus = int(inputNameField)
				}
			case key.Matches(msg, Keymap.Prev):
				m.focus--
				if m.focus < int(inputNameField) {
					m.focus = int(inputSubmitButton)
				}
			case key.Matches(msg, Keymap.Enter):
				switch inputFields(m.focus) {
				case inputNameField, inputTimeField:
					m.focus++
				case inputCancelButton:
					m.resetInputs()
					m.state = showEvents
					if len(m.events.Items()) == 0 {
						m.state = noEvents
					}
				case inputSubmitButton:
					e, err := m.validateInputs()
					if err != nil {
						m.inputs[inputNameField].Reset()
						m.inputs[inputTimeField].Reset()
						m.focus = 0
						m.inputStatus = fmt.Sprintf("Error: %v", err)
						m.datePreview = ""
						m.dateValid = false
						break
					}

					if m.state == showEdit {
						m.events.RemoveItem(m.editIndex)
					}

					if len(m.events.Items()) == 0 {
						m.events.InsertItem(0, e)
					} else {
						index := 0
						for _, item := range m.events.Items() {
							if e.Time >= item.(Event).Time {
								index++
							}
						}
						m.events.InsertItem(index, e)
					}

					if err := m.saveEventsToFile(); err != nil {
						panic(err)
					}

					newEvents, newCmd := m.events.Update(msg)
					m.events = newEvents
					cmd = newCmd
					m.resetInputs()
					m.state = showEvents
				}
			}
		}
		cmds = append(cmds, m.updateInputs()...)
		for i := 0; i < len(m.inputs); i++ {
			newModel, cmd := m.inputs[i].Update(msg)
			m.inputs[i] = newModel
			cmds = append(cmds, cmd)
		}
		m.updateDatePreview()
	}
	timerModel, timerCmd := m.timer.Update(msg)
	m.timer = timerModel
	cmds = append(cmds, timerCmd)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m MainModel) View() string {
	switch m.state {
	case noEvents:
		content := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(cPromptBorder)).
			Padding(2, 4).
			Render("No events, add one with '+'\n\nPress 'q' to quit")
		return lipgloss.Place(m.windowWidth, m.windowHeight, lipgloss.Center, lipgloss.Center, content)
	case showInput:
		return m.inputView("✨ New Event")
	case showEdit:
		return m.inputView("✏️  Edit Event")
	default:
		listStr := AppStyle.Render(m.events.View())
		detailStr := m.detailsString()
		timelineStr := m.renderTimeline()
		return lipgloss.JoinHorizontal(lipgloss.Top, listStr, detailStr, timelineStr)
	}
}

func main() {
	p := tea.NewProgram(NewMainModel(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("There was an error: %v", err)
		os.Exit(1)
	}
}

func getUrgencyColor(ts int64) string {
	t := time.Unix(ts, 0)
	diff := time.Until(t)

	if diff < 0 {
		return cPast
	}

	days := diff.Hours() / 24

	switch {
	case days < 1:
		return cUrgency6 // < 1 day - dark red
	case days < 3:
		return cUrgency5 // 1-3 days - red
	case days < 7:
		return cUrgency4 // 3-7 days - orange
	case days < 14:
		return cUrgency3 // 7-14 days - yellow
	case days < 30:
		return cUrgency2 // 14-30 days - light green
	default:
		return cUrgency1 // > 30 days - green
	}
}

func formatLargeNumber(n int64) string {
	if n < 0 {
		return "-" + formatLargeNumber(-n)
	}

	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	remainder := len(str) % 3
	if remainder > 0 {
		result.WriteString(str[:remainder])
		if len(str) > remainder {
			result.WriteString(",")
		}
	}

	for i := remainder; i < len(str); i += 3 {
		result.WriteString(str[i : i+3])
		if i+3 < len(str) {
			result.WriteString(",")
		}
	}

	return result.String()
}

func formatLargeFloat(f float64, precision int) string {
	negative := f < 0
	if negative {
		f = -f
	}

	intPart := int64(f)
	fracPart := f - float64(intPart)

	intStr := formatLargeNumber(intPart)
	fracStr := fmt.Sprintf("%.*f", precision, fracPart)[1:] // Remove leading "0"

	result := intStr + fracStr
	if negative {
		return "-" + result
	}
	return result
}

func renderProgressBar(value, max float64, width int, color string) string {
	if max <= 0 {
		max = 1
	}
	if value < 0 {
		value = 0
	}
	if value > max {
		value = max
	}

	filled := int((value / max) * float64(width))
	if filled > width {
		filled = width
	}

	filledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cBarEmpty))

	bar := filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", width-filled))

	return bar
}

func renderTimeBlocks(years, days, hours, minutes, seconds int, color string, width int) string {
	var b strings.Builder
	blockStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cDimmedDescDark)).Width(10)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cDimmedTitleDark)).Width(4).Align(lipgloss.Right)

	// Calculate max bar width
	barWidth := width - 20
	if barWidth < 10 {
		barWidth = 10
	}
	if barWidth > 30 {
		barWidth = 30
	}

	type timeUnit struct {
		label    string
		value    int
		maxValue int
	}

	units := []timeUnit{
		{"Years", years, 10},
		{"Days", days, 365},
		{"Hours", hours, 24},
		{"Minutes", minutes, 60},
		{"Seconds", seconds, 60},
	}

	for _, unit := range units {
		if unit.value == 0 && unit.label == "Years" {
			continue
		}

		blocks := (unit.value * barWidth) / unit.maxValue
		if unit.value > 0 && blocks == 0 {
			blocks = 1
		}
		if blocks > barWidth {
			blocks = barWidth
		}

		b.WriteString(labelStyle.Render(unit.label))
		b.WriteString(valueStyle.Render(fmt.Sprintf("%d", unit.value)))
		b.WriteString(" [")
		b.WriteString(blockStyle.Render(strings.Repeat("■", blocks)))
		b.WriteString(emptyStyle.Render(strings.Repeat("·", barWidth-blocks)))
		b.WriteString("]\n")
	}

	return strings.TrimSuffix(b.String(), "\n")
}

func (m MainModel) renderTimeline() string {
	var b strings.Builder

	items := m.events.Items()
	if len(items) == 0 {
		return ""
	}

	titleStyle := TimelineTitleStyle.Copy().Width(m.timelineWidth - 4)
	b.WriteString("\n" + titleStyle.Render("📅 Upcoming Events") + "\n\n")

	type timelineEvent struct {
		event     Event
		index     int
		selected  bool
		daysAway  int
		hoursAway int
	}

	var events []timelineEvent
	selectedIdx := m.events.Index()
	now := time.Now()

	for i, item := range items {
		e := item.(Event)
		eventTime := time.Unix(e.Time, 0)
		if eventTime.After(now) {
			diff := eventTime.Sub(now)
			events = append(events, timelineEvent{
				event:     e,
				index:     i,
				selected:  i == selectedIdx,
				daysAway:  int(diff.Hours() / 24),
				hoursAway: int(diff.Hours()),
			})
		}
	}

	if len(events) == 0 {
		b.WriteString(HintStyle("  No upcoming events\n"))
		return m.timelineStyle().Render(b.String())
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].event.Time < events[j].event.Time
	})

	barWidth := m.timelineWidth - 25
	if barWidth < 10 {
		barWidth = 10
	}
	if barWidth > 40 {
		barWidth = 40
	}

	b.WriteString(TimelineNowStyle.Render("  ▼ NOW") + "\n")
	b.WriteString(TimelineNowStyle.Render("  │") + "\n")

	maxEvents := (m.windowHeight - 15) / 5
	if maxEvents < 3 {
		maxEvents = 3
	}
	if maxEvents > 8 {
		maxEvents = 8
	}
	displayed := 0

	for _, te := range events {
		if displayed >= maxEvents {
			b.WriteString(TimelineTrackStyle.Render("  │\n"))
			b.WriteString(TimelineTrackStyle.Render("  ⋮") + HintStyle(fmt.Sprintf(" +%d more events\n", len(events)-displayed)))
			break
		}

		var timeBar string
		var timeLabel string
		color := getUrgencyColor(te.event.Time)
		barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cBarEmpty))
		trackStyle := TimelineTrackStyle

		if te.daysAway == 0 {
			hours := te.hoursAway
			if hours == 0 {
				timeLabel = "< 1 hour"
				blocks := 1
				timeBar = barStyle.Render(strings.Repeat("▓", blocks)) + emptyStyle.Render(strings.Repeat("░", barWidth-blocks))
			} else {
				timeLabel = fmt.Sprintf("%d hours", hours)
				blocks := (hours * barWidth) / 24
				if blocks == 0 {
					blocks = 1
				}
				if blocks > barWidth {
					blocks = barWidth
				}
				timeBar = barStyle.Render(strings.Repeat("▓", blocks)) + emptyStyle.Render(strings.Repeat("░", barWidth-blocks))
			}
		} else if te.daysAway <= 7 {
			timeLabel = fmt.Sprintf("%d day", te.daysAway)
			if te.daysAway > 1 {
				timeLabel += "s"
			}
			blocks := (te.daysAway * barWidth) / 7
			if blocks == 0 {
				blocks = 1
			}
			timeBar = barStyle.Render(strings.Repeat("█", blocks)) + emptyStyle.Render(strings.Repeat("░", barWidth-blocks))
		} else if te.daysAway <= 30 {
			timeLabel = fmt.Sprintf("%d days", te.daysAway)
			blocks := (te.daysAway * barWidth) / 30
			if blocks == 0 {
				blocks = 1
			}
			timeBar = barStyle.Render(strings.Repeat("█", blocks)) + emptyStyle.Render(strings.Repeat("░", barWidth-blocks))
		} else if te.daysAway <= 365 {
			months := te.daysAway / 30
			timeLabel = fmt.Sprintf("~%d month", months)
			if months > 1 {
				timeLabel += "s"
			}
			blocks := (te.daysAway * barWidth) / 365
			if blocks == 0 {
				blocks = 1
			}
			if blocks > barWidth {
				blocks = barWidth
			}
			timeBar = barStyle.Render(strings.Repeat("█", blocks)) + emptyStyle.Render(strings.Repeat("░", barWidth-blocks))
		} else {
			years := te.daysAway / 365
			timeLabel = fmt.Sprintf("~%d year", years)
			if years > 1 {
				timeLabel += "s"
			}
			timeBar = barStyle.Render(strings.Repeat("█", barWidth))
		}

		var marker string
		var eventStyle lipgloss.Style
		if te.selected {
			marker = "◆"
			eventStyle = TimelineSelectedStyle
		} else {
			marker = "●"
			eventStyle = barStyle
		}

		b.WriteString(trackStyle.Render("  │") + "\n")
		b.WriteString(trackStyle.Render("  ├─") + timeBar + "\n")

		eventTime := time.Unix(te.event.Time, 0)
		dateStr := eventTime.Format("Jan 02")
		if eventTime.Year() != now.Year() {
			dateStr = eventTime.Format("Jan 02 '06")
		}

		name := te.event.Name
		maxNameLen := m.timelineWidth - 10
		if maxNameLen < 15 {
			maxNameLen = 15
		}
		if len(name) > maxNameLen {
			name = name[:maxNameLen-3] + "..."
		}

		b.WriteString(trackStyle.Render("  │ "))
		b.WriteString(eventStyle.Render(marker+" "+name) + "\n")
		b.WriteString(trackStyle.Render("  │   "))
		b.WriteString(HintStyle(dateStr+" • "+timeLabel) + "\n")

		displayed++
	}

	b.WriteString(TimelineTrackStyle.Render("  │\n"))
	b.WriteString(TimelineTrackStyle.Render("  ▽ future\n"))
	b.WriteString("\n" + HintStyle("Bar length = time distance"))

	return m.timelineStyle().Render(b.String())
}

func (m MainModel) timelineStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(m.timelineWidth).
		Padding(1, 2).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color(cTimelineFuture))
}

func (m MainModel) detailsString() string {
	var b strings.Builder
	event := m.events.SelectedItem().(Event)
	urgencyColor := getUrgencyColor(event.Time)

	titleStyle := lipgloss.NewStyle().
		Width(m.detailWidth-6).
		Foreground(lipgloss.Color(cTextLightGray)).
		Background(lipgloss.Color(urgencyColor)).
		Padding(0, 1).
		Align(lipgloss.Center)

	b.WriteString(titleStyle.Render(event.Name) + "\n\n")

	ts := time.Unix(event.Time, 0)

	b.WriteString(NormalTextStyle("📅 "))
	b.WriteString(BrightTextStyle(ts.Format("Monday, January 2, 2006")) + "\n")
	b.WriteString(NormalTextStyle("🕐 "))
	b.WriteString(BrightTextStyle(ts.Format("3:04:05 PM MST")) + "\n\n")

	countdownTitleStyle := lipgloss.NewStyle().
		Width(m.detailWidth-6).
		Foreground(lipgloss.Color(cTextLightGray)).
		Background(lipgloss.Color(urgencyColor)).
		Padding(0, 1).
		Align(lipgloss.Center)

	diff := time.Until(ts).Seconds()
	isPast := diff < 0
	if isPast {
		b.WriteString(countdownTitleStyle.Render("⏪ Time Since") + "\n\n")
		diff = -diff
	} else {
		b.WriteString(countdownTitleStyle.Render("⏳ Time Until") + "\n\n")
	}

	totalSeconds := int(diff)
	years := totalSeconds / secondsPerYear
	days := (totalSeconds - years*secondsPerYear) / secondsPerDay
	hours := (totalSeconds - years*secondsPerYear - days*secondsPerDay) / secondsPerHour
	minutes := (totalSeconds - years*secondsPerYear - days*secondsPerDay - hours*secondsPerHour) / secondsPerMinute
	seconds := totalSeconds - years*secondsPerYear - days*secondsPerDay - hours*secondsPerHour - minutes*secondsPerMinute

	b.WriteString(renderTimeBlocks(years, days, hours, minutes, seconds, urgencyColor, m.detailWidth))
	b.WriteString("\n\n")

	compactStyle := lipgloss.NewStyle().
		Width(m.detailWidth - 6).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color(urgencyColor)).
		Bold(true)

	var countdownStr string
	if years > 0 {
		countdownStr = fmt.Sprintf("%dy %dd %dh %dm %ds", years, days, hours, minutes, seconds)
	} else if days > 0 {
		countdownStr = fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		countdownStr = fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		countdownStr = fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		countdownStr = fmt.Sprintf("%ds", seconds)
	}
	if isPast {
		countdownStr += " ago"
	}
	b.WriteString(compactStyle.Render(countdownStr) + "\n\n")

	progressWidth := m.detailWidth - 30
	if progressWidth < 10 {
		progressWidth = 10
	}
	if progressWidth > 30 {
		progressWidth = 30
	}
	b.WriteString(NormalTextStyle("Day progress: "))
	dayProgress := float64(hours*3600+minutes*60+seconds) / float64(secondsPerDay)
	b.WriteString(renderProgressBar(dayProgress, 1.0, progressWidth, urgencyColor))
	b.WriteString(fmt.Sprintf(" %.1f%%\n\n", dayProgress*100))

	statsTitleStyle := lipgloss.NewStyle().
		Width(m.detailWidth-6).
		Foreground(lipgloss.Color(cTextLightGray)).
		Background(lipgloss.Color(cTitle)).
		Padding(0, 1).
		Align(lipgloss.Center)
	b.WriteString(statsTitleStyle.Render("📊 Statistics") + "\n\n")

	totalSecondsFloat := diff
	totalMinutes := totalSecondsFloat / float64(secondsPerMinute)
	totalHours := totalSecondsFloat / float64(secondsPerHour)
	totalDays := totalSecondsFloat / float64(secondsPerDay)
	totalYears := totalSecondsFloat / float64(secondsPerYear)

	statsLabelStyle := lipgloss.NewStyle().
		Width(16).
		Foreground(lipgloss.AdaptiveColor{Light: cDimmedDescLight, Dark: cDimmedDescDark})
	statsValueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: cDimmedTitleLight, Dark: cDimmedTitleDark})

	b.WriteString(statsLabelStyle.Render("Total seconds:"))
	b.WriteString(statsValueStyle.Render(formatLargeNumber(int64(totalSecondsFloat))) + "\n")
	b.WriteString(statsLabelStyle.Render("Total minutes:"))
	b.WriteString(statsValueStyle.Render(formatLargeFloat(totalMinutes, 2)) + "\n")
	b.WriteString(statsLabelStyle.Render("Total hours:"))
	b.WriteString(statsValueStyle.Render(formatLargeFloat(totalHours, 2)) + "\n")
	b.WriteString(statsLabelStyle.Render("Total days:"))
	b.WriteString(statsValueStyle.Render(formatLargeFloat(totalDays, 2)) + "\n")
	b.WriteString(statsLabelStyle.Render("Total years:"))
	b.WriteString(statsValueStyle.Render(formatLargeFloat(totalYears, 4)) + "\n")

	detailStyle := lipgloss.NewStyle().
		Width(m.detailWidth).
		Padding(1, 2).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: cItemTitleLight, Dark: cItemTitleDark})

	return detailStyle.Render(b.String())
}

func countdownParser(ts int64) string {
	t := time.Unix(ts, 0)
	diff := int(time.Until(t).Seconds())
	isPast := diff < 0
	if isPast {
		diff = -diff
	}
	years := diff / secondsPerYear
	days := (diff - years*secondsPerYear) / secondsPerDay
	hours := (diff - years*secondsPerYear - days*secondsPerDay) / secondsPerHour
	minutes := (diff - years*secondsPerYear - days*secondsPerDay - hours*secondsPerHour) / secondsPerMinute
	seconds := diff - years*secondsPerYear - days*secondsPerDay - hours*secondsPerHour - minutes*secondsPerMinute
	var result string
	if years > 0 {
		result = fmt.Sprintf("%dy %dd %dh %dm %ds", years, days, hours, minutes, seconds)
	} else if days > 0 {
		result = fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		result = fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		result = fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		result = fmt.Sprintf("%ds", seconds)
	}

	color := getUrgencyColor(ts)
	coloredStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))

	if isPast {
		result = coloredStyle.Render(result + " ago")
	} else {
		result = coloredStyle.Render(result)
	}
	return result
}

func readEventsFile() ([]Event, error) {
	eventsFile, err := getEventsFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get events file path: %w", err)
	}

	var events []Event
	if _, err := os.Stat(eventsFile); errors.Is(err, os.ErrNotExist) {
		_, err := os.Create(eventsFile)
		if err != nil {
			return events, err
		}
		event := nextGolangAnniversary()
		events = append(events, event)
		bytes, err := json.MarshalIndent(events, "", "  ")
		if err != nil {
			return events, err
		}
		err = os.WriteFile(eventsFile, bytes, 0644)
		return events, err
	}
	bytes, err := os.ReadFile(eventsFile)
	if err != nil {
		return events, err
	}
	err = json.Unmarshal(bytes, &events)
	if err != nil {
		return events, err
	}
	return events, nil
}

func (m MainModel) saveEventsToFile() error {
	eventsFile, err := getEventsFilePath()
	if err != nil {
		return fmt.Errorf("failed to get events file path: %w", err)
	}

	items := m.events.Items()
	events := make([]Event, len(items))
	for i := range items {
		events[i] = items[i].(Event)
	}
	bytes, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(eventsFile, bytes, 0644)
	return err
}

func (m MainModel) inputView(title string) string {
	var b strings.Builder

	inputWidth := m.windowWidth / 2
	if inputWidth < 50 {
		inputWidth = 50
	}
	if inputWidth > 80 {
		inputWidth = 80
	}

	titleStyle := lipgloss.NewStyle().
		Width(inputWidth-6).
		Foreground(lipgloss.Color(cTextLightGray)).
		Background(lipgloss.Color(cDetailTitle)).
		Padding(0, 1).
		Align(lipgloss.Center)

	b.WriteString(titleStyle.Render(title) + "\n\n")

	fieldStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(inputWidth - 10)
	fieldFocusedStyle := fieldStyle.Copy().
		BorderForeground(lipgloss.Color(cPromptBorder))

	b.WriteString(InputLabelStyle.Render("📝 Event Name") + "\n")
	nameFieldStyle := fieldStyle
	if m.focus == int(inputNameField) {
		nameFieldStyle = fieldFocusedStyle
	}
	b.WriteString(nameFieldStyle.Render(m.inputs[0].View()) + "\n")

	b.WriteString(InputLabelStyle.Render("📅 Date & Time") + "\n")
	timeFieldStyle := fieldStyle
	if m.focus == int(inputTimeField) {
		timeFieldStyle = fieldFocusedStyle
	}
	b.WriteString(timeFieldStyle.Render(m.inputs[1].View()) + "\n")

	b.WriteString(HintStyle("   Format: YYYY-MM-DD or YYYY-MM-DD HH:MM:SS") + "\n")
	b.WriteString(HintStyle("   Example: 2025-12-31 or 2025-12-31 18:30:00") + "\n")

	if m.datePreview != "" {
		if m.dateValid {
			b.WriteString(DatePreviewStyle.Render("→ "+m.datePreview) + "\n")
		} else {
			b.WriteString(ErrStyle("   ✗ "+m.datePreview) + "\n")
		}
	} else {
		b.WriteString("\n")
	}

	cancelButton := ButtonStyle
	if m.focus == int(inputCancelButton) {
		cancelButton = ButtonFocusedStyle
	}
	submitButton := ButtonStyle
	if m.focus == int(inputSubmitButton) {
		submitButton = ButtonFocusedStyle
	}

	submitLabel := "✓ Create"
	if m.state == showEdit {
		submitLabel = "✓ Update"
	}

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		cancelButton.Render("✗ Cancel"),
		"  ",
		submitButton.Render(submitLabel),
	)
	b.WriteString("\n" + buttons + "\n")

	if m.inputStatus != "" {
		b.WriteString("\n" + ErrStyle(m.inputStatus))
	}

	b.WriteString("\n\n" + HintStyle("Tab: next field • Shift+Tab: previous • Enter: select • Esc: cancel"))

	inputStyle := lipgloss.NewStyle().
		Width(inputWidth).
		Margin(1, 1).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color(cPromptBorder))

	// Center the input form
	return lipgloss.Place(m.windowWidth, m.windowHeight, lipgloss.Center, lipgloss.Center, inputStyle.Render(b.String()))
}

func (m *MainModel) updateDatePreview() {
	dateStr := m.inputs[inputTimeField].Value()
	if dateStr == "" {
		m.datePreview = ""
		m.dateValid = false
		return
	}

	timeFormat := inputTimeFormLong
	if len(dateStr) <= len(inputTimeFormShort) {
		timeFormat = inputTimeFormShort
	}

	ts, err := time.ParseInLocation(timeFormat, dateStr, time.Local)
	if err != nil {
		m.datePreview = "Invalid date format"
		m.dateValid = false
		return
	}

	m.dateValid = true
	if ts.Before(time.Now()) {
		m.datePreview = ts.Format("Mon, Jan 2, 2006 at 3:04 PM") + " (past event)"
	} else {
		m.datePreview = ts.Format("Mon, Jan 2, 2006 at 3:04 PM")
	}
}

func (m *MainModel) updateInputs() []tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focus {
			cmds[i] = m.inputs[i].Focus()
			m.inputs[i].PromptStyle = FocusedStyle
			m.inputs[i].TextStyle = FocusedStyle
			continue
		}
		m.inputs[i].Blur()
		m.inputs[i].PromptStyle = NoStyle
		m.inputs[i].TextStyle = NoStyle
	}
	return cmds
}

func (m *MainModel) resetInputs() {
	m.inputs[inputNameField].Reset()
	m.inputs[inputTimeField].Reset()
	m.focus = 0
	m.inputStatus = ""
	m.datePreview = ""
	m.dateValid = false
	m.editIndex = -1
}

func (m MainModel) validateInputs() (Event, error) {
	var event Event
	name := m.inputs[0].Value()
	t := m.inputs[1].Value()
	if name == "" {
		return event, fmt.Errorf("event name is required")
	}
	if t == "" {
		return event, fmt.Errorf("date/time is required")
	}
	timeFormat := inputTimeFormLong
	if len(t) < len(inputTimeFormLong) {
		timeFormat = inputTimeFormShort
	}
	ts, err := time.ParseInLocation(timeFormat, t, time.Local)
	if err != nil {
		return event, fmt.Errorf("invalid date format")
	}
	event = Event{Name: name, Time: ts.Unix()}
	return event, nil
}

func nextGolangAnniversary() Event {
	nameStr := "Golang's Birthday"
	now := time.Now()
	year := now.Year()
	thisYear := time.Date(year, 11, 10, 0, 0, 0, 0, time.Local)
	nextYear := time.Date(year+1, 11, 10, 0, 0, 0, 0, time.Local)
	if now.Before(thisYear) {
		return Event{nameStr, thisYear.Unix()}
	}
	return Event{nameStr, nextYear.Unix()}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
