package events

import (
	"strings"

	"github.com/KharpukhaevV/postui/httpclient"
	"github.com/KharpukhaevV/postui/models"
	tea "github.com/charmbracelet/bubbletea"
)

// EventHandler обрабатывает события пользовательского ввода
type EventHandler struct {
	httpClient *httpclient.HTTPClient
}

// NewEventHandler создает новый обработчик событий
func NewEventHandler() *EventHandler {
	return &EventHandler{
		httpClient: httpclient.NewHTTPClient(),
	}
}

// HandleKeyEvent обрабатывает события клавиш и возвращает соответствующие команды
func (h *EventHandler) HandleKeyEvent(model *models.AppModel, msg tea.KeyMsg) (*models.AppModel, tea.Cmd) {
	if !model.GetInputMode() {
		return h.handleNavigationMode(model, msg)
	}
	return h.handleInputMode(model, msg)
}

// handleNavigationMode обрабатывает события клавиш в режиме навигации
func (h *EventHandler) handleNavigationMode(model *models.AppModel, msg tea.KeyMsg) (*models.AppModel, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return model, tea.Quit
	case "i", "a":
		model.SetInputMode(true)
		h.updateFocus(model)
		return model, nil
	case "tab":
		model.SetActiveSection(models.Section((int(model.GetActiveSection()) + 1) % 6))
		h.updateFocus(model)
		return model, nil
	case "shift+tab":
		model.SetActiveSection(models.Section((int(model.GetActiveSection()) + 5) % 6))
		h.updateFocus(model)
		return model, nil
	case "1":
		model.SetActiveSection(models.SectionMethod)
		h.updateFocus(model)
		return model, nil
	case "2":
		model.SetActiveSection(models.SectionURL)
		h.updateFocus(model)
		return model, nil
	case "3":
		model.SetActiveSection(models.SectionHeaders)
		h.updateFocus(model)
		return model, nil
	case "4":
		model.SetActiveSection(models.SectionBody)
		h.updateFocus(model)
		return model, nil
	case "5":
		model.SetActiveSection(models.SectionParams)
		h.updateFocus(model)
		return model, nil
	case "6":
		model.SetActiveSection(models.SectionResponse)
		h.updateFocus(model)
		return model, nil
	case "up", "k":
		if model.GetActiveSection() == models.SectionResponse {
			model.GetResponseVP().LineUp(1)
		}
		return model, nil
	case "down", "j":
		if model.GetActiveSection() == models.SectionResponse {
			model.GetResponseVP().LineDown(1)
		}
		return model, nil
	case "pageup":
		if model.GetActiveSection() == models.SectionResponse {
			model.GetResponseVP().PageUp()
		}
		return model, nil
	case "pagedown":
		if model.GetActiveSection() == models.SectionResponse {
			model.GetResponseVP().PageDown()
		}
		return model, nil
	case "left", "h":
		if model.GetActiveSection() == models.SectionMethod {
			model.SetSelectedMethod(models.HTTPMethod((int(model.GetSelectedMethod()) - 1 + len(models.MethodNames)) % len(models.MethodNames)))
		}
		return model, nil
	case "right", "l":
		if model.GetActiveSection() == models.SectionMethod {
			model.SetSelectedMethod(models.HTTPMethod((int(model.GetSelectedMethod()) + 1) % len(models.MethodNames)))
		}
		return model, nil
	case "enter":
		return h.handleEnterKey(model)
	case "backspace":
		return h.handleBackspaceKey(model)
	}
	return model, nil
}

// handleInputMode обрабатывает события клавиш в режиме ввода
func (h *EventHandler) handleInputMode(model *models.AppModel, msg tea.KeyMsg) (*models.AppModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		model.SetInputMode(false)
		h.updateFocus(model)
		return model, nil
	case "ctrl+c", "q":
		if model.GetActiveSection() != models.SectionBody {
			return model, tea.Quit
		}
		return model, nil
	case "up", "k":
		if model.GetActiveSection() == models.SectionResponse {
			model.GetResponseVP().LineUp(1)
			return model, nil
		}
	case "down", "j":
		if model.GetActiveSection() == models.SectionResponse {
			model.GetResponseVP().LineDown(1)
			return model, nil
		}
	case "pageup":
		if model.GetActiveSection() == models.SectionResponse {
			model.GetResponseVP().PageUp()
			return model, nil
		}
	case "pagedown":
		if model.GetActiveSection() == models.SectionResponse {
			model.GetResponseVP().PageDown()
			return model, nil
		}
	}
	return model, nil
}

// handleEnterKey обрабатывает клавишу Enter в зависимости от активной секции
func (h *EventHandler) handleEnterKey(model *models.AppModel) (*models.AppModel, tea.Cmd) {
	switch model.GetActiveSection() {
	case models.SectionHeaders:
		if model.GetHeaderInput().Value() != "" {
			parts := strings.SplitN(model.GetHeaderInput().Value(), "=", 2)
			if len(parts) == 2 {
				model.AddHeader(parts[0], parts[1])
				model.GetHeaderInput().SetValue("")
			}
		}
		return model, nil
	case models.SectionParams:
		if model.GetParamInput().Value() != "" {
			parts := strings.SplitN(model.GetParamInput().Value(), "=", 2)
			if len(parts) == 2 {
				model.AddParam(parts[0], parts[1])
				model.GetParamInput().SetValue("")
			}
		}
		return model, nil
	default:
		if model.URLInputValue() != "" {
			model.SetLoading(true)
			return model, h.sendRequest(model)
		}
		return model, nil
	}
}

// handleBackspaceKey обрабатывает клавишу Backspace в зависимости от активной секции
func (h *EventHandler) handleBackspaceKey(model *models.AppModel) (*models.AppModel, tea.Cmd) {
	switch model.GetActiveSection() {
	case models.SectionHeaders:
		if len(model.GetHeaders()) > 0 && model.GetHeaderInput().Value() == "" {
			model.RemoveLastHeader()
		}
	case models.SectionParams:
		if len(model.GetParams()) > 0 && model.GetParamInput().Value() == "" {
			model.RemoveLastParam()
		}
	}
	return model, nil
}

// sendRequest отправляет HTTP запрос и возвращает соответствующую команду
func (h *EventHandler) sendRequest(model *models.AppModel) tea.Cmd {
	return func() tea.Msg {
		req := httpclient.NewHTTPRequest(model)
		response, err := h.httpClient.SendRequest(&req)
		if err != nil {
			return models.ErrorData{Message: err.Error()}
		}
		return response
	}
}

// updateFocus обновляет фокус для компонентов ввода в зависимости от активной секции
func (h *EventHandler) updateFocus(model *models.AppModel) {
	model.GetURLInput().Blur()
	model.GetHeaderInput().Blur()
	model.GetBodyInput().Blur()
	model.GetParamInput().Blur()

	if model.GetInputMode() {
		switch model.GetActiveSection() {
		case models.SectionURL:
			model.GetURLInput().Focus()
		case models.SectionHeaders:
			model.GetHeaderInput().Focus()
		case models.SectionBody:
			model.GetBodyInput().Focus()
		case models.SectionParams:
			model.GetParamInput().Focus()
		}
	}
}

// UpdateComponents обновляет компоненты ввода в зависимости от активной секции
func (h *EventHandler) UpdateComponents(model *models.AppModel, msg tea.Msg) (*models.AppModel, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if model.GetInputMode() {
		switch model.GetActiveSection() {
		case models.SectionURL:
			*model.GetURLInput(), cmd = model.GetURLInput().Update(msg)
			cmds = append(cmds, cmd)
		case models.SectionHeaders:
			*model.GetHeaderInput(), cmd = model.GetHeaderInput().Update(msg)
			cmds = append(cmds, cmd)
		case models.SectionBody:
			*model.GetBodyInput(), cmd = model.GetBodyInput().Update(msg)
			cmds = append(cmds, cmd)
		case models.SectionParams:
			*model.GetParamInput(), cmd = model.GetParamInput().Update(msg)
			cmds = append(cmds, cmd)
		case models.SectionResponse:
			*model.GetResponseVP(), cmd = model.GetResponseVP().Update(msg)
			cmds = append(cmds, cmd)
		}
	} else {
		if model.GetActiveSection() == models.SectionResponse {
			*model.GetResponseVP(), cmd = model.GetResponseVP().Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return model, tea.Batch(cmds...)
}
