package ui

import (
	"fmt"

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
}

// NewUIRenderer создает новый рендерер интерфейса со стилями по умолчанию
func NewUIRenderer() *UIRenderer {
	return &UIRenderer{
		styles: &UIStyles{
			docStyle:            lipgloss.NewStyle().Margin(1, 2),
			titleStyle:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")),
			labelStyle:          lipgloss.NewStyle().Bold(true).Width(12).MarginRight(1),
			inputStyle:          lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1),
			activeInputStyle:    lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1).BorderForeground(lipgloss.Color("205")),
			errorStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
			successStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true),
			sectionStyle:        lipgloss.NewStyle().Margin(0, 0, 1, 0),
			activeSectionStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true),
			methodStyle:         lipgloss.NewStyle().Padding(0, 1).Margin(0, 1),
			selectedMethodStyle: lipgloss.NewStyle().Padding(0, 1).Margin(0, 1).Foreground(lipgloss.Color("205")).Bold(true).Underline(true),
		},
	}
}

// Render рендерит полный интерфейс
func (r *UIRenderer) Render(model *models.AppModel) string {
	// Получаем размеры из модели (нужно добавить геттеры)
	width, height := model.GetDimensions()
	if width == 0 || height == 0 {
		return "Инициализация..."
	}

	leftWidth := (width - 4) * 2 / 3
	leftColumn := lipgloss.NewStyle().Width(leftWidth).Height(height - 4)
	rightColumn := lipgloss.NewStyle().Width(width - leftWidth - 6).Height(height - 4)

	leftContent := r.renderLeftColumn(model, leftWidth)
	rightContent := r.renderRightColumn(model, rightColumn.GetWidth())

	mainLayout := lipgloss.JoinHorizontal(lipgloss.Top,
		leftColumn.Render(leftContent),
		lipgloss.NewStyle().Width(2).Render(""),
		rightColumn.Render(rightContent),
	)

	return r.styles.docStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			r.styles.titleStyle.Render("REST Client TUI"),
			"",
			mainLayout,
			"\nНажмите 1-6 для переключения секций, TAB/SHIFT+TAB для навигации, ENTER для отправки запроса, ↑↓/PageUp/PageDown для прокрутки ответа, q для выхода",
		),
	)
}

// renderLeftColumn рендерит левую колонку с конфигурацией запроса
func (r *UIRenderer) renderLeftColumn(model *models.AppModel, leftWidth int) string {
	methodSection := r.renderMethodSection(model)
	urlSection := r.renderURLSection(model, leftWidth)
	headersSection := r.renderHeadersSection(model, leftWidth)
	bodySection := r.renderBodySection(model, leftWidth)
	paramsSection := r.renderParamsSection(model, leftWidth)
	statusMessage := r.renderStatusMessage(model)

	return lipgloss.JoinVertical(lipgloss.Left,
		r.styles.sectionStyle.Render(methodSection),
		r.styles.sectionStyle.Render(urlSection),
		r.styles.sectionStyle.Render(headersSection),
		r.styles.sectionStyle.Render(bodySection),
		r.styles.sectionStyle.Render(paramsSection),
		statusMessage,
	)
}

// renderRightColumn рендерит правую колонку с ответом
func (r *UIRenderer) renderRightColumn(model *models.AppModel, rightWidth int) string {
	responseSection := r.renderResponseSection(model, rightWidth)
	statusInfo := r.renderStatusInfo(model)

	return lipgloss.JoinVertical(lipgloss.Left,
		responseSection,
		statusInfo,
	)
}

// renderMethodSection рендерит выбор HTTP метода
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

	return lipgloss.JoinHorizontal(lipgloss.Left,
		r.styles.labelStyle.Render(label),
		methodsRow,
	)
}

// renderURLSection рендерит секцию ввода URL
func (r *UIRenderer) renderURLSection(model *models.AppModel, leftWidth int) string {
	label := "[2] URL:"
	if model.GetActiveSection() == models.SectionURL {
		label = r.styles.activeSectionStyle.Render("[2] URL:")
	}

	input := model.GetURLInput().View()
	if model.GetActiveSection() == models.SectionURL && model.GetInputMode() {
		input = r.styles.activeInputStyle.Width(leftWidth - 4).Render(input)
	} else {
		input = r.styles.inputStyle.Width(leftWidth - 4).Render(input)
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		r.styles.labelStyle.Render(label),
		input,
	)
}

// renderHeadersSection рендерит секцию заголовков
func (r *UIRenderer) renderHeadersSection(model *models.AppModel, leftWidth int) string {
	label := "[3] Заголовки:"
	if model.GetActiveSection() == models.SectionHeaders {
		label = r.styles.activeSectionStyle.Render("[3] Заголовки:")
	}

	headersText := ""
	for _, h := range model.GetHeaders() {
		headersText += fmt.Sprintf("  %s: %s\n", h.Key, h.Value)
	}

	input := headersText + model.GetHeaderInput().View()
	if model.GetActiveSection() == models.SectionHeaders && model.GetInputMode() {
		input = r.styles.activeInputStyle.Width(leftWidth - 4).Height(len(model.GetHeaders()) + 1).Render(input)
	} else {
		input = r.styles.inputStyle.Width(leftWidth - 4).Height(len(model.GetHeaders()) + 1).Render(input)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		r.styles.labelStyle.Render(label),
		input,
	)
}

// renderBodySection рендерит секцию тела запроса
func (r *UIRenderer) renderBodySection(model *models.AppModel, leftWidth int) string {
	label := "[4] Тело:"
	if model.GetActiveSection() == models.SectionBody {
		label = r.styles.activeSectionStyle.Render("[4] Тело:")
	}

	input := model.GetBodyInput().View()
	if model.GetActiveSection() == models.SectionBody && model.GetInputMode() {
		input = r.styles.activeInputStyle.Width(leftWidth - 4).Height(12).Render(input)
	} else {
		input = r.styles.inputStyle.Width(leftWidth - 4).Height(12).Render(input)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		r.styles.labelStyle.Render(label),
		input,
	)
}

// renderParamsSection рендерит секцию параметров запроса
func (r *UIRenderer) renderParamsSection(model *models.AppModel, leftWidth int) string {
	label := "[5] Параметры:"
	if model.GetActiveSection() == models.SectionParams {
		label = r.styles.activeSectionStyle.Render("[5] Параметры:")
	}

	paramsText := ""
	for _, p := range model.GetParams() {
		paramsText += fmt.Sprintf("  %s: %s\n", p.Key, p.Value)
	}

	input := paramsText + model.GetParamInput().View()
	if model.GetActiveSection() == models.SectionParams {
		input = r.styles.activeInputStyle.Width(leftWidth - 4).Height(len(model.GetParams()) + 1).Render(input)
	} else {
		input = r.styles.inputStyle.Width(leftWidth - 4).Height(len(model.GetParams()) + 1).Render(input)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		r.styles.labelStyle.Render(label),
		input,
	)
}

// renderResponseSection рендерит секцию ответа
func (r *UIRenderer) renderResponseSection(model *models.AppModel, rightWidth int) string {
	label := "[6] Ответ:"
	if model.GetActiveSection() == models.SectionResponse {
		label = r.styles.activeSectionStyle.Render("[6] Ответ:")
	}

	viewportHeight := model.GetResponseVP().Height
	input := model.GetResponseVP().View()
	if model.GetActiveSection() == models.SectionResponse {
		input = r.styles.activeInputStyle.Width(rightWidth - 4).Height(viewportHeight).Render(input)
	} else {
		input = r.styles.inputStyle.Width(rightWidth - 4).Height(viewportHeight).Render(input)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		r.styles.labelStyle.Render(label),
		input,
	)
}

// renderStatusInfo рендерит информацию о статусе
func (r *UIRenderer) renderStatusInfo(model *models.AppModel) string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		r.styles.labelStyle.Render("Статус:"),
		lipgloss.NewStyle().Width(20).Render(model.GetStatus()),
		r.styles.labelStyle.Render("Время:"),
		lipgloss.NewStyle().Width(15).Render(model.GetResponseTime()),
	)
}

// renderStatusMessage рендерит сообщение о статусе
func (r *UIRenderer) renderStatusMessage(model *models.AppModel) string {
	if model.GetLoading() {
		return "⏳ Отправка запроса..."
	} else if model.GetErrorMsg() != "" {
		return r.styles.errorStyle.Render("Ошибка: " + model.GetErrorMsg())
	} else if model.GetStatus() != "" {
		return r.styles.successStyle.Render("✓ Запрос выполнен")
	}
	return ""
}
