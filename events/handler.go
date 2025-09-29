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

// HandleKeyEvent обрабатывает события клавиш и возвращает флаг, если событие было "съедено"
func (h *EventHandler) HandleKeyEvent(model *models.AppModel, msg tea.KeyMsg) (*models.AppModel, tea.Cmd, bool) {
	// Глобальные обработчики (сохранение, удаление)
	if model.IsSaving() {
		return h.handleSaveAsPrompt(model, msg)
	}
	if model.IsDeleting() {
		return h.handleDeleteConfirmation(model, msg)
	}

	if !model.GetInputMode() {
		return h.handleNavigationMode(model, msg)
	}
	return h.handleInputMode(model, msg)
}

// handleNavigationMode обрабатывает события клавиш в режиме навигации
func (h *EventHandler) handleNavigationMode(model *models.AppModel, msg tea.KeyMsg) (*models.AppModel, tea.Cmd, bool) {
	switch msg.String() {
	case "ctrl+c", "q":
		return model, tea.Quit, true
	case "i", "a":
		model.SetInputMode(true)
		h.updateFocus(model)
		return model, nil, true
	case "ctrl+s":
		if model.GetActiveTab() == models.TabRequest {
			model.SetIsSaving(true)
			model.GetSaveNameInput().Focus()
		}
		return model, nil, true

	// Переключение вкладок и методов
	case "left", "h":
		if model.GetActiveTab() == models.TabRequest && model.GetActiveSection() == models.SectionMethod {
			model.SetSelectedMethod(models.HTTPMethod((int(model.GetSelectedMethod()) - 1 + len(models.MethodNames)) % len(models.MethodNames)))
		} else {
			currentTab := (int(model.GetActiveTab()) - 1 + 3) % 3
			model.SetActiveTab(models.Tab(currentTab))
		}
		return model, nil, true
	case "right", "l":
		if model.GetActiveTab() == models.TabRequest && model.GetActiveSection() == models.SectionMethod {
			model.SetSelectedMethod(models.HTTPMethod((int(model.GetSelectedMethod()) + 1) % len(models.MethodNames)))
		} else {
			currentTab := (int(model.GetActiveTab()) + 1) % 3
			model.SetActiveTab(models.Tab(currentTab))
		}
		return model, nil, true

	// Навигация по секциям на вкладке "Запрос"
	case "k", "up":
		if model.GetActiveTab() == models.TabRequest {
			model.SetActiveSection(models.Section((int(model.GetActiveSection()) + 4) % 5))
			h.updateFocus(model)
		} else if model.GetActiveTab() == models.TabSaved {
			*model.GetSavedList(), _ = model.GetSavedList().Update(msg)
		}
		return model, nil, true
	case "j", "down":
		if model.GetActiveTab() == models.TabRequest {
			model.SetActiveSection(models.Section((int(model.GetActiveSection()) + 1) % 5))
			h.updateFocus(model)
		} else if model.GetActiveTab() == models.TabSaved {
			*model.GetSavedList(), _ = model.GetSavedList().Update(msg)
		}
		return model, nil, true
	case "tab":
		if model.GetActiveTab() == models.TabRequest {
			model.SetActiveSection(models.Section((int(model.GetActiveSection()) + 1) % 5))
			h.updateFocus(model)
		}
		return model, nil, true
	case "shift+tab":
		if model.GetActiveTab() == models.TabRequest {
			model.SetActiveSection(models.Section((int(model.GetActiveSection()) + 4) % 5))
			h.updateFocus(model)
		}
		return model, nil, true
	case "1", "2", "3", "4", "5":
		if model.GetActiveTab() == models.TabRequest {
			section := 0
			switch msg.String() {
			case "1":
				section = 0
			case "2":
				section = 1
			case "3":
				section = 2
			case "4":
				section = 3
			case "5":
				section = 4
			}
			model.SetActiveSection(models.Section(section))
			h.updateFocus(model)
		}
		return model, nil, true

	case "d":
		if model.GetActiveTab() == models.TabSaved {
			model.SetIsDeleting(true)
		}
		return model, nil, true

	case "enter":
		model, cmd := h.handleEnterKey(model)
		return model, cmd, true
	case "backspace":
		model, cmd := h.handleBackspaceKey(model)
		return model, cmd, true
	}
	return model, nil, false
}

// handleInputMode обрабатывает события клавиш в режиме ввода
func (h *EventHandler) handleInputMode(model *models.AppModel, msg tea.KeyMsg) (*models.AppModel, tea.Cmd, bool) {
	switch msg.String() {
	case "esc":
		model.SetInputMode(false)
		h.updateFocus(model)
		return model, nil, true // Ключ обработан
	case "ctrl+c", "q":
		if model.GetActiveSection() != models.SectionBody {
			return model, tea.Quit, true
		}
	case "enter":
		// Позволяем добавлять заголовки/параметры по Enter в режиме ввода
		if model.GetActiveSection() == models.SectionHeaders || model.GetActiveSection() == models.SectionParams {
			model, cmd := h.handleEnterOnRequestTab(model)
			return model, cmd, true // "Съедаем" Enter
		}
	}
	return model, nil, false // Ключ не обработан, передать компоненту
}

// --- Обработчики специальных режимов ---

func (h *EventHandler) handleSaveAsPrompt(model *models.AppModel, msg tea.KeyMsg) (*models.AppModel, tea.Cmd, bool) {
	switch msg.String() {
	case "enter":
		name := model.GetSaveNameInput().Value()
		if name != "" {
			model.AddNewSavedRequest(name)
		}
		model.GetSaveNameInput().SetValue("")
		model.GetSaveNameInput().Blur()
		model.SetIsSaving(false)
	case "esc":
		model.GetSaveNameInput().SetValue("")
		model.GetSaveNameInput().Blur()
		model.SetIsSaving(false)
	}
	// Передаем событие в поле ввода имени
	*model.GetSaveNameInput(), _ = model.GetSaveNameInput().Update(msg)
	return model, nil, true // "Съедаем" событие в любом случае
}

func (h *EventHandler) handleDeleteConfirmation(model *models.AppModel, msg tea.KeyMsg) (*models.AppModel, tea.Cmd, bool) {
	switch strings.ToLower(msg.String()) {
	case "y":
		model.DeleteSelectedRequest()
		model.SetIsDeleting(false)
	case "n", "esc":
		model.SetIsDeleting(false)
	}
	return model, nil, true // "Съедаем" событие в любом случае
}

// --- Основные действия ---

func (h *EventHandler) handleEnterKey(model *models.AppModel) (*models.AppModel, tea.Cmd) {
	switch model.GetActiveTab() {
	case models.TabRequest:
		return h.handleEnterOnRequestTab(model)
	case models.TabSaved:
		model.LoadRequestFromSaved()
		return model, nil
	}
	return model, nil
}

func (h *EventHandler) handleEnterOnRequestTab(model *models.AppModel) (*models.AppModel, tea.Cmd) {
	switch model.GetActiveSection() {
	case models.SectionHeaders:
		if model.GetHeaderInput().Value() != "" {
			parts := strings.SplitN(model.GetHeaderInput().Value(), "=", 2)
			if len(parts) == 2 {
				model.AddHeader(parts[0], parts[1])
				model.GetHeaderInput().SetValue("")
			}
		}
	case models.SectionParams:
		if model.GetParamInput().Value() != "" {
			parts := strings.SplitN(model.GetParamInput().Value(), "=", 2)
			if len(parts) == 2 {
				model.AddParam(parts[0], parts[1])
				model.GetParamInput().SetValue("")
			}
		}
	default:
		if model.URLInputValue() != "" {
			model.SetLoading(true)
			return model, h.sendRequest(model)
		}
	}
	return model, nil
}

func (h *EventHandler) handleBackspaceKey(model *models.AppModel) (*models.AppModel, tea.Cmd) {
	if model.GetActiveTab() != models.TabRequest {
		return model, nil
	}
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

func (h *EventHandler) updateFocus(model *models.AppModel) {
	model.GetURLInput().Blur()
	model.GetHeaderInput().Blur()
	model.GetBodyInput().Blur()
	model.GetParamInput().Blur()

	if model.GetInputMode() && model.GetActiveTab() == models.TabRequest {
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

func (h *EventHandler) UpdateComponents(model *models.AppModel, msg tea.Msg) (*models.AppModel, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch model.GetActiveTab() {
	case models.TabRequest:
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
			}
		}
	case models.TabResponse:
		*model.GetResponseVP(), cmd = model.GetResponseVP().Update(msg)
		cmds = append(cmds, cmd)
	case models.TabSaved:
		*model.GetSavedList(), cmd = model.GetSavedList().Update(msg)
		cmds = append(cmds, cmd)
	}

	return model, tea.Batch(cmds...)
}
