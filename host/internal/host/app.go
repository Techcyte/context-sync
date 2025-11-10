package host

import (
	"fmt"
	"net/http"
	"strings"

	"tcs/internal/model"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type App struct {
	Manager   *Manager
	Spinner   spinner.Model
	Viewport  viewport.Model
	TextInput textinput.Model
	Messages  []string
	Quitting  bool
	Ready     bool
	Err       error
}

func NewApp(manager *Manager) App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("202"))

	input := textinput.New()
	input.Placeholder = "Case Number"
	input.CharLimit = 32
	input.Width = 32

	return App{
		Manager:   manager,
		Spinner:   s,
		TextInput: input,
		Messages:  []string{},
	}
}

func (app App) Init() tea.Cmd {
	go func() {
		err := http.ListenAndServe(app.Manager.Address, nil)
		if err != nil {
			panic(err)
		}
	}()

	return tea.Batch(app.Spinner.Tick, textinput.Blink)
}

func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	const heightOffset = 8
	inPutFocused := app.TextInput.Focused()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if !inPutFocused {
				app.Quitting = true
				return app, tea.Quit
			}
		case "c":
			if !inPutFocused {
				app.ClearLog()
				return app, nil
			}
		case "n":
			if !inPutFocused {
				app.TextInput.Focus()
				return app, nil
			}
		case "a":
			if app.Manager.Voting {
				app.Manager.Accept()
				return app, nil
			}
		case "r":
			if app.Manager.Voting {
				app.Manager.Reject()
				return app, nil
			}
		case "esc":
			app.TextInput.Blur()
			return app, nil
		case "enter":
			if app.TextInput.Focused() {
				if app.Manager.ClientCount() == 0 {
					app.Manager.CurrentCase = app.TextInput.Value()
					app.Manager.Context = []model.ContextItem{{Key: "case", Value: app.Manager.CurrentCase}}
					app.TextInput.SetValue("")
					app.TextInput.Blur()

					return app, nil
				}

				app.Manager.ContextChangeRequest(app.TextInput.Value())
				app.TextInput.SetValue("")
				app.TextInput.Blur()
			}

			return app, nil
		}
	case tea.WindowSizeMsg:
		if !app.Ready {
			app.Viewport = viewport.New(msg.Width, msg.Height-heightOffset)
			app.Viewport.YPosition = 5
			app.Viewport.SetContent("")
			app.Ready = true
		} else {
			app.Viewport.Width = msg.Width
			app.Viewport.Height = msg.Height - heightOffset
		}
	case error:
		app.Err = msg
		return app, nil
	}

	if len(app.Manager.MessagesToAdd) > 0 {
		for _, msg := range app.Manager.MessagesToAdd {
			msg = fmt.Sprintf("%v: %v", len(app.Messages)+1, msg)
			app.Messages = append(app.Messages, msg)
		}

		app.Viewport.SetContent(strings.Join(app.Messages, "\n"))
		app.Viewport.GotoBottom()
		app.Manager.MessagesToAdd = nil
	}

	cmds := []tea.Cmd{}
	var cmd tea.Cmd

	app.Spinner, cmd = app.Spinner.Update(msg)
	cmds = append(cmds, cmd)

	app.Viewport, cmd = app.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	app.TextInput, cmd = app.TextInput.Update(msg)
	cmds = append(cmds, cmd)

	return app, tea.Batch(cmds...)
}

func (app App) View() string {
	if app.Err != nil {
		return app.Err.Error()
	}

	str := fmt.Sprintf("\n\t⚡️ Context sync manager is running at %v %v", app.Manager.Address, app.Spinner.View())
	str = fmt.Sprintf("%v\t\tChange case %v", str, app.TextInput.View())

	if app.Manager.Voting {
		str = fmt.Sprintf("%v\tChange case to '%v'? accept <a> * reject <r>\n", str, app.Manager.VoteCase)
	} else if app.Manager.AutoAccept {
		str = fmt.Sprintf("%v\t\033[93mAuto accept enabled\033[0m\n", str)
	} else {
		str = fmt.Sprintf("%v\n", str)
	}

	str = fmt.Sprintf("%v\tConnected clients: %v", str, len(app.Manager.Clients))
	str = fmt.Sprintf("%v\t\t\t\t\tCurrent case: '%v'\n", str, app.Manager.CurrentCase)

	for i := 0; i < app.Viewport.Width; i++ {
		str = fmt.Sprintf("%v─", str)
	}

	str = fmt.Sprintf("\n%v\n%v\n", str, app.Viewport.View())

	controls := "clear <c> * change case <n> * quit <q>"
	lineLen := app.Viewport.Width - len(controls) - 2
	for range lineLen / 2 {
		str = fmt.Sprintf("%v─", str)
	}

	str = fmt.Sprintf("%v┤\033[32m%v\033[0m├", str, controls)

	for range lineLen / 2 {
		str = fmt.Sprintf("%v─", str)
	}

	if app.Quitting {
		return str + "\n"
	}

	return str
}

func (app *App) ClearLog() {
	app.Viewport.SetContent("")
	app.Messages = []string{}
}
