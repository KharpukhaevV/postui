package models

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// Tab представляет активную вкладку в интерфейсе
type Tab int

const (
	TabRequest Tab = iota
	TabResponse
)

// Section представляет различные секции интерфейса
type Section int

const (
	SectionMethod Section = iota
	SectionURL
	SectionHeaders
	SectionBody
	SectionParams
)

// HTTPMethod представляет доступные HTTP методы
type HTTPMethod int

const (
	MethodGET HTTPMethod = iota
	MethodPOST
	MethodPUT
	MethodDELETE
	MethodPATCH
	MethodHEAD
	MethodOPTIONS
)

// MethodNames сопоставляет HTTP методы с их строковыми представлениями
var MethodNames = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

// Param представляет параметр запроса
type Param struct {
	Key   string
	Value string
}

// Header представляет HTTP заголовок
type Header struct {
	Key   string
	Value string
}

// ResponseData содержит информацию об HTTP ответе
type ResponseData struct {
	Body       string
	Status     string
	Time       string
	StatusCode int
}

// ErrorData содержит информацию об ошибке
type ErrorData struct {
	Message string
}

// AppModel представляет основное состояние приложения
type AppModel struct {
	// Компоненты ввода
	urlInput    textinput.Model
	bodyInput   textarea.Model
	responseVP  viewport.Model
	paramInput  textinput.Model
	headerInput textinput.Model

	// Данные
	params  []Param
	headers []Header

	// Состояние
	activeTab      Tab
	activeSection  Section
	selectedMethod HTTPMethod
	loading        bool
	response       string
	status         string
	responseTime   string
	errorMsg       string
	inputMode      bool

	// Размеры
	width  int
	height int
}

// NewAppModel создает новую модель приложения со значениями по умолчанию
func NewAppModel() *AppModel {
	urlInput := textinput.New()
	urlInput.Placeholder = "https://api.example.com/endpoint"
	urlInput.CharLimit = 500

	bodyInput := textarea.New()
	bodyInput.Placeholder = "{\n  \"key\": \"value\"\n}"
	bodyInput.ShowLineNumbers = false
	// Убираем подсветку активной строки в поле Body
	bodyInput.FocusedStyle.CursorLine = lipgloss.NewStyle()

	responseVP := viewport.New(10, 10)

	paramInput := textinput.New()
	paramInput.Placeholder = "key=value"
	paramInput.CharLimit = 100

	headerInput := textinput.New()
	headerInput.Placeholder = "Content-Type=application/json"
	headerInput.CharLimit = 100

	return &AppModel{
		urlInput:       urlInput,
		bodyInput:      bodyInput,
		responseVP:     responseVP,
		paramInput:     paramInput,
		headerInput:    headerInput,
		params:         []Param{},
		headers:        []Header{},
		activeTab:      TabRequest,
		activeSection:  SectionMethod,
		selectedMethod: MethodGET,
		inputMode:      false,
	}
}

// UpdateDimensions обновляет размеры окна для нового вкладочного интерфейса
func (m *AppModel) UpdateDimensions(width, height int) {
	m.width = width
	m.height = height

	contentHeight := height - 5
	if contentHeight < 10 {
		contentHeight = 10
	}

	contentWidth := width - 4

	m.responseVP.Width = contentWidth
	m.responseVP.Height = contentHeight

	m.urlInput.Width = contentWidth - 14
	m.paramInput.Width = contentWidth - 14
	m.headerInput.Width = contentWidth - 14
	m.bodyInput.SetWidth(contentWidth - 14)

	occupiedHeight := len(m.headers) + len(m.params) + 17
	bodyHeight := contentHeight - occupiedHeight
	if bodyHeight < 3 {
		bodyHeight = 3
	}
	m.bodyInput.SetHeight(bodyHeight)
}

// GetCurrentMethod возвращает выбранный HTTP метод в виде строки
func (m *AppModel) GetCurrentMethod() string {
	return MethodNames[m.selectedMethod]
}

// AddParam добавляет новый параметр запроса
func (m *AppModel) AddParam(key, value string) {
	m.params = append(m.params, Param{Key: key, Value: value})
}

// RemoveLastParam удаляет последний добавленный параметр
func (m *AppModel) RemoveLastParam() {
	if len(m.params) > 0 {
		m.params = m.params[:len(m.params)-1]
	}
}

// AddHeader добавляет новый HTTP заголовок
func (m *AppModel) AddHeader(key, value string) {
	m.headers = append(m.headers, Header{Key: key, Value: value})
}

// RemoveLastHeader удаляет последний добавленный заголовок
func (m *AppModel) RemoveLastHeader() {
	if len(m.headers) > 0 {
		m.headers = m.headers[:len(m.headers)-1]
	}
}

// SetResponseData обновляет данные ответа и переключает на вкладку ответа
func (m *AppModel) SetResponseData(data ResponseData) {
	m.loading = false
	m.response = FormatJSON(data.Body)
	m.status = fmt.Sprintf("%s (%d)", data.Status, data.StatusCode)
	m.responseTime = data.Time
	m.responseVP.SetContent(m.response)
	m.errorMsg = ""
	m.activeTab = TabResponse // Автоматически переключаемся на вкладку ответа
}

// SetError обновляет состояние ошибки и переключает на вкладку ответа
func (m *AppModel) SetError(err ErrorData) {
	m.loading = false
	m.errorMsg = err.Message
	m.response = ""
	m.status = "Error"
	m.responseTime = ""
	m.responseVP.SetContent(err.Message)
	m.activeTab = TabResponse // Автоматически переключаемся на вкладку ответа
}

// SetLoading устанавливает состояние загрузки
func (m *AppModel) SetLoading(loading bool) {
	m.loading = loading
}

// Геттеры и Сеттеры

func (m *AppModel) GetActiveTab() Tab {
	return m.activeTab
}

func (m *AppModel) SetActiveTab(tab Tab) {
	m.activeTab = tab
}

func (m *AppModel) URLInputValue() string {
	return m.urlInput.Value()
}

func (m *AppModel) BodyInputValue() string {
	return m.bodyInput.Value()
}

func (m *AppModel) GetHeaders() []Header {
	return m.headers
}

func (m *AppModel) GetParams() []Param {
	return m.params
}

func (m *AppModel) GetResponseVP() *viewport.Model {
	return &m.responseVP
}

func (m *AppModel) GetURLInput() *textinput.Model {
	return &m.urlInput
}

func (m *AppModel) GetBodyInput() *textarea.Model {
	return &m.bodyInput
}

func (m *AppModel) GetParamInput() *textinput.Model {
	return &m.paramInput
}

func (m *AppModel) GetHeaderInput() *textinput.Model {
	return &m.headerInput
}

func (m *AppModel) GetActiveSection() Section {
	return m.activeSection
}

func (m *AppModel) SetActiveSection(section Section) {
	m.activeSection = section
}

func (m *AppModel) GetSelectedMethod() HTTPMethod {
	return m.selectedMethod
}

func (m *AppModel) SetSelectedMethod(method HTTPMethod) {
	m.selectedMethod = method
}

func (m *AppModel) GetInputMode() bool {
	return m.inputMode
}

func (m *AppModel) SetInputMode(mode bool) {
	m.inputMode = mode
}

func (m *AppModel) GetLoading() bool {
	return m.loading
}

func (m *AppModel) GetResponse() string {
	return m.response
}

func (m *AppModel) GetStatus() string {
	return m.status
}

func (m *AppModel) GetResponseTime() string {
	return m.responseTime
}

func (m *AppModel) GetErrorMsg() string {
	return m.errorMsg
}

func (m *AppModel) GetDimensions() (int, int) {
	return m.width, m.height
}

// FormatJSON форматирует JSON строку с правильными отступами
func FormatJSON(input string) string {
	if input == "" {
		return ""
	}

	var formatted bytes.Buffer
	err := json.Indent(&formatted, []byte(input), "", "  ")
	if err != nil {
		return input
	}
	return formatted.String()
}
