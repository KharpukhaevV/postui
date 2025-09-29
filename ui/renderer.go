package ui

import (
	"fmt"
	"strings"

	"github.com/KharpukhaevV/postui/models"
	"github.com/charmbracelet/lipgloss"
)

// UIRenderer обрабатывает рендеринг интерфейса
type UIRenderer struct {
	styles *UIStyles
}

// UIStyles содержит все определения стилей интерфейса
type UIStyles struct {
	docStyle            lipgloss.Style
	titleStyle          lipgloss.Style
	labelStyle          lipgloss.Style
	inputStyle          lipgloss.Style
	activeInputStyle    lipgloss.Style
	errorStyle          lipgloss.Style
	successStyle        lipgloss.Style
	sectionStyle        lipgloss.Style
	activeSectionStyle  lipgloss.Style
	methodStyle         lipgloss.Style
	selectedMethodStyle lipgloss.Style
	tabStyle            lipgloss.Style
	activeTabStyle      lipgloss.Style
	helpTextStyle       lipgloss.Style
	promptStyle         lipgloss.Style
}

// NewUIRenderer создает новый рендерер интерфейса со стилями по умолчанию
func NewUIRenderer() *UIRenderer {
	activeColor := lipgloss.Color("205")

	return &UIRenderer{
		styles: &UIStyles{
			docStyle:            lipgloss.NewStyle().Padding(1, 2),
			titleStyle:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")),
			labelStyle:          lipgloss.NewStyle().Bold(true).Width(12).MarginRight(1),
			inputStyle:          lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1),
			activeInputStyle:    lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1).BorderForeground(activeColor),
			errorStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
			successStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true),
			sectionStyle:        lipgloss.NewStyle().Margin(0, 0, 1, 0),
			activeSectionStyle:  lipgloss.NewStyle().Foreground(activeColor).Bold(true),
			methodStyle:         lipgloss.NewStyle().Padding(0, 1).Margin(0, 1),
			selectedMethodStyle: lipgloss.NewStyle().Padding(0, 1).Margin(0, 1).Foreground(activeColor).Bold(true).Underline(true),
			tabStyle:            lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.Color("240")),
			activeTabStyle:      lipgloss.NewStyle().Padding(0, 2).Bold(true).Underline(true).Foreground(activeColor),
			helpTextStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
			promptStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true),
		},
	}
}

// Render рендерит полный интерфейс
func (r *UIRenderer) Render(model *models.AppModel) string {
	width, _ := model.GetDimensions()
	if width == 0 {
		return "Инициализация..."
	}

	var currentView string
	switch model.GetActiveTab() {
	case models.TabRequest:
		currentView = r.renderRequestView(model)
	case models.TabResponse:
		currentView = r.renderResponseView(model)
	case models.TabSaved:
		currentView = r.renderSavedView(model)
	}

	header := r.renderHeader(model)
	footer := r.renderFooter(model)

	finalView := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"", // Отступ
		currentView,
		footer,
	)

	return r.styles.docStyle.Render(finalView)
}

// renderHeader рендерит заголовок и вкладки на одной строке
func (r *UIRenderer) renderHeader(model *models.AppModel) string {
	width, _ := model.GetDimensions()
	title := r.styles.titleStyle.Render("REST Client TUI")
	tabs := r.renderTabs(model)

	spacerWidth := width - lipgloss.Width(title) - lipgloss.Width(tabs) - r.styles.docStyle.GetHorizontalFrameSize()
	if spacerWidth < 0 {
		spacerWidth = 0
	}
	spacer := lipgloss.NewStyle().Width(spacerWidth).Render("")

	return lipgloss.JoinHorizontal(lipgloss.Left, title, spacer, tabs)
}

// renderFooter рендерит нижнюю часть интерфейса
func (r *UIRenderer) renderFooter(model *models.AppModel) string {
	// Приоритетные сообщения (сохранение, удаление)
	if model.IsSaving() {
		return r.styles.promptStyle.Render("Сохранить как: ") + model.GetSaveNameInput().View()
	}
	if model.IsDeleting() {
		return r.styles.errorStyle.Render(fmt.Sprintf("Удалить '%s'? (y/n)", model.GetSavedList().SelectedItem().(models.SavedRequest).Title()))
	}

	// Статус выполнения запроса
	if model.GetLoading() {
		return "⏳ Отправка запроса..."
	}
	if model.GetErrorMsg() != "" {
		return r.styles.errorStyle.Render("Ошибка: " + model.GetErrorMsg())
	}
	if model.GetStatus() != "" {
		successMsg := r.styles.successStyle.Render("✓ Запрос выполнен")
		statusInfo := lipgloss.JoinHorizontal(lipgloss.Top,
			r.styles.labelStyle.Render("Статус:"),
			lipgloss.NewStyle().Width(20).Render(model.GetStatus()),
			r.styles.labelStyle.Render("Время:"),
			lipgloss.NewStyle().Width(15).Render(model.GetResponseTime()),
		)
		return lipgloss.JoinHorizontal(lipgloss.Left, successMsg, "  ", statusInfo)
	}

	// Подсказка по умолчанию
	return r.styles.helpTextStyle.Render("←/h/l/→: вкладки | j/k: навигация | i: ввод | enter: выбрать/отправить | q: выход")
}

// renderTabs рендерит панель вкладок
func (r *UIRenderer) renderTabs(model *models.AppModel) string {
	requestTab := r.styles.tabStyle.Render("Запрос")
	responseTab := r.styles.tabStyle.Render("Ответ")
	savedTab := r.styles.tabStyle.Render("Сохраненные")

	switch model.GetActiveTab() {
	case models.TabRequest:
		requestTab = r.styles.activeTabStyle.Render("Запрос")
	case models.TabResponse:
		responseTab = r.styles.activeTabStyle.Render("Ответ")
	case models.TabSaved:
		savedTab = r.styles.activeTabStyle.Render("Сохраненные")
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, requestTab, responseTab, savedTab)
}

// --- Рендеринг содержимого вкладок ---

func (r *UIRenderer) renderRequestView(model *models.AppModel) string {
	return lipgloss.JoinVertical(lipgloss.Left,
		r.styles.sectionStyle.Render(r.renderMethodSection(model)),
		r.styles.sectionStyle.Render(r.renderURLSection(model)),
		r.styles.sectionStyle.Render(r.renderHeadersSection(model)),
		r.styles.sectionStyle.Render(r.renderBodySection(model)),
		r.styles.sectionStyle.Render(r.renderParamsSection(model)),
	)
}

func (r *UIRenderer) renderResponseView(model *models.AppModel) string {
	return model.GetResponseVP().View()
}

func (r *UIRenderer) renderSavedView(model *models.AppModel) string {
	return model.GetSavedList().View()
}

// --- Рендеринг секций для вкладки "Запрос" ---

func (r *UIRenderer) renderMethodSection(model *models.AppModel) string {
	label := "[1] Метод:"
	if model.GetActiveSection() == models.SectionMethod {
		label = r.styles.activeSectionStyle.Render("[1] Метод:")
	}
	methodItems := make([]string, len(models.MethodNames))
	for i, method := range models.MethodNames {
		if models.HTTPMethod(i) == model.GetSelectedMethod() {
			methodItems[i] = r.styles.selectedMethodStyle.Render(method)
		} else {
			methodItems[i] = r.styles.methodStyle.Render(method)
		}
	}
	methodsRow := lipgloss.JoinHorizontal(lipgloss.Left, methodItems...)
	return lipgloss.JoinHorizontal(lipgloss.Left, r.styles.labelStyle.Render(label), methodsRow)
}

func (r *UIRenderer) renderURLSection(model *models.AppModel) string {
	label := "[2] URL:"
	if model.GetActiveSection() == models.SectionURL {
		label = r.styles.activeSectionStyle.Render("[2] URL:")
	}
	input := model.GetURLInput()
	view := input.View()
	style := r.styles.inputStyle.Width(input.Width)
	if model.GetActiveSection() == models.SectionURL && model.GetInputMode() {
		style = r.styles.activeInputStyle.Width(input.Width)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, r.styles.labelStyle.Render(label), style.Render(view))
}

func (r *UIRenderer) renderHeadersSection(model *models.AppModel) string {
	label := "[3] Заголовки:"
	if model.GetActiveSection() == models.SectionHeaders {
		label = r.styles.activeSectionStyle.Render("[3] Заголовки:")
	}
	var sb strings.Builder
	for _, h := range model.GetHeaders() {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", h.Key, h.Value))
	}
	sb.WriteString(model.GetHeaderInput().View())

	input := model.GetHeaderInput()
	style := r.styles.inputStyle.Width(input.Width).Height(len(model.GetHeaders()) + 1)
	if model.GetActiveSection() == models.SectionHeaders && model.GetInputMode() {
		style = r.styles.activeInputStyle.Width(input.Width).Height(len(model.GetHeaders()) + 1)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, r.styles.labelStyle.Render(label), style.Render(sb.String()))
}

func (r *UIRenderer) renderBodySection(model *models.AppModel) string {
	label := "[4] Тело:"
	if model.GetActiveSection() == models.SectionBody {
		label = r.styles.activeSectionStyle.Render("[4] Тело:")
	}
	input := model.GetBodyInput()
	view := input.View()
	style := r.styles.inputStyle.Width(input.Width()).Height(input.Height())
	if model.GetActiveSection() == models.SectionBody && model.GetInputMode() {
		style = r.styles.activeInputStyle.Width(input.Width()).Height(input.Height())
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, r.styles.labelStyle.Render(label), style.Render(view))
}

func (r *UIRenderer) renderParamsSection(model *models.AppModel) string {
	label := "[5] Параметры:"
	if model.GetActiveSection() == models.SectionParams {
		label = r.styles.activeSectionStyle.Render("[5] Параметры:")
	}
	var sb strings.Builder
	for _, p := range model.GetParams() {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", p.Key, p.Value))
	}
	sb.WriteString(model.GetParamInput().View())

	input := model.GetParamInput()
	style := r.styles.inputStyle.Width(input.Width).Height(len(model.GetParams()) + 1)
	if model.GetActiveSection() == models.SectionParams && model.GetInputMode() {
		style = r.styles.activeInputStyle.Width(input.Width).Height(len(model.GetParams()) + 1)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, r.styles.labelStyle.Render(label), style.Render(sb.String()))
}
