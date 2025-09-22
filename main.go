package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	mainView sessionState = iota
)

type model struct {
	// Основные поля
	urlInput   textinput.Model
	bodyInput  textarea.Model
	responseVP viewport.Model

	// Параметры и заголовки
	params      []param
	headers     []header
	paramInput  textinput.Model
	headerInput textinput.Model

	// Состояние
	activeSection  int // 0: Method, 1: URL, 2: Headers, 3: Body, 4: Params, 5: Response
	selectedMethod int // Индекс выбранного метода
	loading        bool
	response       string
	status         string
	responseTime   string
	errorMsg       string
	inputMode      bool // Режим ввода (true) или навигации (false)

	// Размеры
	width  int
	height int
}

type param struct {
	key   string
	value string
}

type header struct {
	key   string
	value string
}

type responseMsg struct {
	body       string
	status     string
	time       string
	statusCode int
}

type errorMsg struct {
	message string
}

// Доступные методы
var methods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

func initialModel() *model {
	// URL input
	urlInput := textinput.New()
	urlInput.Placeholder = "https://api.example.com/endpoint"
	urlInput.CharLimit = 500

	// Body input (textarea вместо textinput)
	bodyInput := textarea.New()
	bodyInput.Placeholder = "{\n  \"key\": \"value\"\n}"
	bodyInput.SetWidth(50)
	bodyInput.SetHeight(13)
	bodyInput.ShowLineNumbers = false
	bodyInput.FocusedStyle.CursorLine = lipgloss.NewStyle()
	
	// Response viewport
	responseVP := viewport.New(50, 20)

	// Param and header inputs
	paramInput := textinput.New()
	paramInput.Placeholder = "key=value"
	paramInput.CharLimit = 100

	headerInput := textinput.New()
	headerInput.Placeholder = "Content-Type=application/json"
	headerInput.CharLimit = 100

	m := &model{
		urlInput:      urlInput,
		bodyInput:     bodyInput,
		responseVP:    responseVP,
		params:        []param{},
		headers:       []header{},
		paramInput:    paramInput,
		headerInput:   headerInput,
		activeSection: 0, // Начинаем с URL
		selectedMethod: 0, // GET по умолчанию
		loading:       false,
		response:      "",
		status:        "",
		responseTime:  "",
		errorMsg:      "",
		inputMode:     false, // Начинаем в режиме навигации
	}

	// Устанавливаем начальный фокус
	m.updateFocus()
	return m
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Настраиваем размеры для двухколоночного layout
		leftWidth := (msg.Width - 4) * 2 / 3
		m.responseVP.Width = msg.Width - leftWidth - 6
		m.responseVP.Height = msg.Height - 10

		// Обновляем размеры полей ввода
		m.urlInput.Width = leftWidth - 4
		m.bodyInput.SetWidth(leftWidth - 4)
		m.paramInput.Width = leftWidth - 4
		m.headerInput.Width = leftWidth - 4

	case tea.KeyMsg:
		if !m.inputMode {
			// Режим навигации
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "i", "a": // Vim-подобные команды для входа в режим ввода
				m.inputMode = true
				m.updateFocus()
				return m, nil
			case "tab":
				m.activeSection = (m.activeSection + 1) % 6
				m.updateFocus()
			case "shift+tab":
				m.activeSection = (m.activeSection + 5) % 6
				m.updateFocus()
			case "1":
				m.activeSection = 0 // Method
				m.updateFocus()
			case "2":
				m.activeSection = 1 // URL
				m.updateFocus()
			case "3":
				m.activeSection = 2 // Headers
				m.updateFocus()
			case "4":
				m.activeSection = 3 // Body
				m.updateFocus()
			case "5":
				m.activeSection = 4 // Params
				m.updateFocus()
			case "6":
				m.activeSection = 5 // Response
				m.updateFocus()
			case "left", "h":
				if m.activeSection == 0 {
					m.selectedMethod = (m.selectedMethod - 1 + len(methods)) % len(methods)
				}
			case "right", "l":
				if m.activeSection == 0 {
					m.selectedMethod = (m.selectedMethod + 1) % len(methods)
				}
			case "enter":
				if m.activeSection == 2 && m.headerInput.Value() != "" {
					parts := strings.SplitN(m.headerInput.Value(), "=", 2)
					if len(parts) == 2 {
						m.headers = append(m.headers, header{key: parts[0], value: parts[1]})
						m.headerInput.SetValue("")
					}
				} else if m.activeSection == 4 && m.paramInput.Value() != "" {
					parts := strings.SplitN(m.paramInput.Value(), "=", 2)
					if len(parts) == 2 {
						m.params = append(m.params, param{key: parts[0], value: parts[1]})
						m.paramInput.SetValue("")
					}
				} else if m.urlInput.Value() != "" {
					m.loading = true
					return m, m.sendRequest
				}
			case "backspace":
				if m.activeSection == 2 && len(m.headers) > 0 && m.headerInput.Value() == "" {
					m.headers = m.headers[:len(m.headers)-1]
				} else if m.activeSection == 4 && len(m.params) > 0 && m.paramInput.Value() == "" {
					m.params = m.params[:len(m.params)-1]
				}
			}
		} else {
			// Режим ввода
			switch msg.String() {
			case "esc": // Выход из режима ввода
				m.inputMode = false
				m.updateFocus()
				return m, nil
			case "ctrl+c", "q":
				if m.activeSection != 3 { // Для body обрабатываем отдельно
					return m, tea.Quit
				}
			}
		}

	case responseMsg:
		m.loading = false
		m.response = m.formatJSON(msg.body)
		m.status = fmt.Sprintf("%s (%d)", msg.status, msg.statusCode)
		m.responseTime = msg.time
		m.responseVP.SetContent(m.response)
		m.errorMsg = ""

	case errorMsg:
		m.loading = false
		m.errorMsg = msg.message
		m.response = ""
		m.status = "Error"
		m.responseTime = ""
		m.responseVP.SetContent("")
	}

	// Обновляем компоненты в зависимости от активной секции
	if m.inputMode {
		switch m.activeSection {
		case 1: // URL
			m.urlInput, cmd = m.urlInput.Update(msg)
			cmds = append(cmds, cmd)
		case 2: // Headers
			m.headerInput, cmd = m.headerInput.Update(msg)
			cmds = append(cmds, cmd)
		case 3: // Body
			m.bodyInput, cmd = m.bodyInput.Update(msg)
			cmds = append(cmds, cmd)
		case 4: // Params
			m.paramInput, cmd = m.paramInput.Update(msg)
			cmds = append(cmds, cmd)
		case 5: // Response
			m.responseVP, cmd = m.responseVP.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) updateFocus() {
	// Снимаем фокус со всех полей
	m.urlInput.Blur()
	m.headerInput.Blur()
	m.bodyInput.Blur()
	m.paramInput.Blur()

	// Устанавливаем фокус на активное поле только в режиме ввода
	if m.inputMode {
		switch m.activeSection {
		case 1: // URL
			m.urlInput.Focus()
		case 2: // Headers
			m.headerInput.Focus()
		case 3: // Body
			m.bodyInput.Focus()
		case 4: // Params
			m.paramInput.Focus()
		}
	}
}

func (m model) sendRequest() tea.Msg {
	start := time.Now()

	// Парсим URL и добавляем параметры
	baseURL := m.urlInput.Value()
	if baseURL == "" {
		return errorMsg{message: "URL is required"}
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return errorMsg{message: fmt.Sprintf("Invalid URL: %v", err)}
	}

	// Добавляем query параметры
	query := parsedURL.Query()
	for _, p := range m.params {
		query.Add(p.key, p.value)
	}
	parsedURL.RawQuery = query.Encode()

	// Создаем тело запроса
	var bodyBytes []byte
	if m.bodyInput.Value() != "" {
		bodyBytes = []byte(m.bodyInput.Value())
	}

	// Получаем выбранный метод
	method := methods[m.selectedMethod]

	// Создаем запрос
	req, err := http.NewRequest(method, parsedURL.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return errorMsg{message: fmt.Sprintf("Failed to create request: %v", err)}
	}

	// Добавляем заголовки
	for _, h := range m.headers {
		req.Header.Add(h.key, h.value)
	}

	// Выполняем запрос
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return errorMsg{message: fmt.Sprintf("Request failed: %v", err)}
	}
	defer resp.Body.Close()

	// Читаем ответ
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	responseBody := buf.String()

	elapsed := time.Since(start).Round(time.Millisecond)

	return responseMsg{
		body:       responseBody,
		status:     resp.Status,
		statusCode: resp.StatusCode,
		time:       elapsed.String(),
	}
}

func (m model) formatJSON(input string) string {
	if strings.TrimSpace(input) == "" {
		return ""
	}

	var formatted bytes.Buffer
	err := json.Indent(&formatted, []byte(input), "", "  ")
	if err != nil {
		return input // Возвращаем как есть если не JSON
	}
	return formatted.String()
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Стили
	docStyle := lipgloss.NewStyle().Margin(1, 2)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	labelStyle := lipgloss.NewStyle().Bold(true).Width(12).MarginRight(1)
	inputStyle := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1)
	activeInputStyle := inputStyle.Copy().BorderForeground(lipgloss.Color("205"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	sectionStyle := lipgloss.NewStyle().Margin(0, 0, 1, 0)

	// Стиль для активной секции
	activeSectionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	
	// Стили для методов
	methodStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 1)
	
	selectedMethodStyle := methodStyle.Copy().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Underline(true)
	
	// Левая колонка (запрос)
	leftWidth := (m.width - 4) * 2 / 3
	leftColumn := lipgloss.NewStyle().Width(leftWidth).Height(m.height - 4)

	// Правая колонка (ответ)
	rightColumn := lipgloss.NewStyle().Width(m.width - leftWidth - 6).Height(m.height - 4)

	// Метод - горизонтальный список
	methodLabel := "[1] Method:"
	if m.activeSection == 0 {
		methodLabel = activeSectionStyle.Render("[1] Method:")
	}
	
	// Создаем горизонтальный список методов
	methodItems := make([]string, len(methods))
	for i, method := range methods {
		if i == m.selectedMethod {
			methodItems[i] = selectedMethodStyle.Render(method)
		} else {
			methodItems[i] = methodStyle.Render(method)
		}
	}
	methodsRow := lipgloss.JoinHorizontal(lipgloss.Left, methodItems...)
	
	methodSection := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render(methodLabel),
		methodsRow,
	)

	// URL
	urlLabel := "[2] URL:"
	if m.activeSection == 1 {
		urlLabel = activeSectionStyle.Render("[2] URL:")
	}
	urlSection := lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render(urlLabel),
		func() string {
			if m.activeSection == 1 && m.inputMode {
				return activeInputStyle.Width(leftWidth - 20).Render(m.urlInput.View())
			}
			return inputStyle.Width(leftWidth - 20).Render(m.urlInput.View())
		}(),
	)

	// Заголовки
	headersLabel := "[3] Headers:"
	if m.activeSection == 2 {
		headersLabel = activeSectionStyle.Render("[3] Headers:")
	}
	headersView := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render(headersLabel),
		func() string {
			headersText := ""
			for _, h := range m.headers {
				headersText += fmt.Sprintf("  %s: %s\n", h.key, h.value)
			}
			if m.activeSection == 2 && m.inputMode {
				return activeInputStyle.Width(leftWidth - 4).Height(len(m.headers) + 1).Render(
					headersText + m.headerInput.View())
			}
			return inputStyle.Width(leftWidth - 4).Height(len(m.headers) + 1).Render(
				headersText + m.headerInput.View())
		}(),
	)

	// Тело запроса
	bodyLabel := "[4] Body:"
	if m.activeSection == 3 {
		bodyLabel = activeSectionStyle.Render("[4] Body:")
	}
	bodyView := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render(bodyLabel),
		func() string {
			if m.activeSection == 3 && m.inputMode {
				return activeInputStyle.Width(leftWidth - 4).Height(12).Render(m.bodyInput.View())
			}
			return inputStyle.Width(leftWidth - 4).Height(12).Render(m.bodyInput.View())
		}(),
	)
	// Параметры
	paramsLabel := "[5] Params:"
	if m.activeSection == 4 {
		paramsLabel = activeSectionStyle.Render("[5] Params:")
	}
	paramsView := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render(paramsLabel),
		func() string {
			paramsText := ""
			for _, p := range m.params {
				paramsText += fmt.Sprintf("  %s: %s\n", p.key, p.value)
			}
			if m.activeSection == 4 {
				return activeInputStyle.Width(leftWidth - 4).Height(len(m.params) + 1).Render(
					paramsText + m.paramInput.View())
			}
			return inputStyle.Width(leftWidth - 4).Height(len(m.params) + 1).Render(
				paramsText + m.paramInput.View())
		}(),
	)

	// Ответ
	responseLabel := "[6] Response:"
	if m.activeSection == 5 {
		responseLabel = activeSectionStyle.Render("[6] Response:")
	}
	statusInfo := lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Status:"),
		lipgloss.NewStyle().Width(20).Render(m.status),
		labelStyle.Render("Time:"),
		lipgloss.NewStyle().Width(15).Render(m.responseTime),
	)

	responseView := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render(responseLabel),
		func() string {
			if m.activeSection == 5 {
				return activeInputStyle.Width(rightColumn.GetWidth()-4).Height(m.height-12).Render(m.responseVP.View())
			}
			return inputStyle.Width(rightColumn.GetWidth()-4).Height(m.height-12).Render(m.responseVP.View())
		}(),
	)

	// Сообщения о состоянии
	var statusMessage string
	if m.loading {
		statusMessage = "⏳ Sending request..."
	} else if m.errorMsg != "" {
		statusMessage = errorStyle.Render("Error: " + m.errorMsg)
	} else if m.status != "" {
		statusMessage = successStyle.Render("✓ Request completed")
	}

	// Собираем левую колонку
	leftContent := lipgloss.JoinVertical(lipgloss.Left,
		sectionStyle.Render(methodSection),
		sectionStyle.Render(urlSection),
		sectionStyle.Render(headersView),
		sectionStyle.Render(bodyView),
		sectionStyle.Render(paramsView),
		statusMessage,
	)

	// Собираем правую колонку
	rightContent := lipgloss.JoinVertical(lipgloss.Left,
		statusInfo,
		responseView,
	)

	// Основной layout
	mainLayout := lipgloss.JoinHorizontal(lipgloss.Top,
		leftColumn.Render(leftContent),
		lipgloss.NewStyle().Width(2).Render(""),
		rightColumn.Render(rightContent),
	)

	return docStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("REST Client TUI"),
			"",
			mainLayout,
			"\nPress 1-6 to switch sections, TAB/SHIFT+TAB to navigate, ENTER to send request, q to quit",
		),
	)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}

