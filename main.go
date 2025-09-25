package main

import (
	"fmt"

	"github.com/KharpukhaevV/postui/events"
	"github.com/KharpukhaevV/postui/models"
	"github.com/KharpukhaevV/postui/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// App представляет основное приложение
type App struct {
	model        *models.AppModel
	eventHandler *events.EventHandler
	uiRenderer   *ui.UIRenderer
}

// NewApp создает новый экземпляр приложения
func NewApp() *App {
	return &App{
		model:        models.NewAppModel(),
		eventHandler: events.NewEventHandler(),
		uiRenderer:   ui.NewUIRenderer(),
	}
}

// Init инициализирует приложение
func (a *App) Init() tea.Cmd {
	return nil
}

// Update обрабатывает сообщения и обновляет состояние приложения
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.model.UpdateDimensions(msg.Width, msg.Height)

	case tea.KeyMsg:
		a.model, cmd = a.eventHandler.HandleKeyEvent(a.model, msg)
		cmds = append(cmds, cmd)

	case models.ResponseData:
		a.model.SetResponseData(msg)

	case models.ErrorData:
		a.model.SetError(msg)
	}

	// Обновляем компоненты ввода
	a.model, cmd = a.eventHandler.UpdateComponents(a.model, msg)
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

// View рендерит интерфейс приложения
func (a *App) View() string {
	return a.uiRenderer.Render(a.model)
}

func main() {
	app := NewApp()
	program := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Printf("Ошибка: %v", err)
	}
}
