package models

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

// Section представляет различные секции интерфейса
type Section int

const (
	SectionMethod Section = iota
	SectionURL
	SectionHeaders
	SectionBody
	SectionParams
	SectionResponse
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
	bodyInput.SetWidth(50)
	bodyInput.SetHeight(13)
	bodyInput.ShowLineNumbers = false

	responseVP := viewport.New(50, 15)

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
		activeSection:  SectionMethod,
		selectedMethod: MethodGET,
		inputMode:      false,
	}
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

// SetResponseData обновляет данные ответа и содержимое viewport
func (m *AppModel) SetResponseData(data ResponseData) {
	m.loading = false
	m.response = FormatJSON(data.Body)
	m.status = fmt.Sprintf("%s (%d)", data.Status, data.StatusCode)
	m.responseTime = data.Time
	m.responseVP.SetContent(m.response)
	m.errorMsg = ""
}

// SetError обновляет состояние ошибки
func (m *AppModel) SetError(err ErrorData) {
	m.loading = false
	m.errorMsg = err.Message
	m.response = ""
	m.status = "Error"
	m.responseTime = ""
	m.responseVP.SetContent("")
}

// SetLoading устанавливает состояние загрузки
func (m *AppModel) SetLoading(loading bool) {
	m.loading = loading
}

// UpdateDimensions обновляет размеры окна и настраивает размеры компонентов
func (m *AppModel) UpdateDimensions(width, height int) {
	m.width = width
	m.height = height

	leftWidth := (width - 4) * 2 / 3
	m.responseVP.Width = width - leftWidth - 6

	responseHeight := 15
	if height > 30 {
		responseHeight = height / 3
		if responseHeight > 20 {
			responseHeight = 20
		}
	}
	m.responseVP.Height = responseHeight

	m.urlInput.Width = leftWidth - 4
	m.bodyInput.SetWidth(leftWidth - 4)
	m.paramInput.Width = leftWidth - 4
	m.headerInput.Width = leftWidth - 4
}

// Геттеры для доступа к приватным полям

// URLInputValue возвращает значение URL поля ввода
func (m *AppModel) URLInputValue() string {
	return m.urlInput.Value()
}

// BodyInputValue возвращает значение поля ввода тела запроса
func (m *AppModel) BodyInputValue() string {
	return m.bodyInput.Value()
}

// GetHeaders возвращает список заголовков
func (m *AppModel) GetHeaders() []Header {
	return m.headers
}

// GetParams возвращает список параметров
func (m *AppModel) GetParams() []Param {
	return m.params
}

// GetResponseVP возвращает viewport для ответа
func (m *AppModel) GetResponseVP() *viewport.Model {
	return &m.responseVP
}

// GetURLInput возвращает поле ввода URL
func (m *AppModel) GetURLInput() *textinput.Model {
	return &m.urlInput
}

// GetBodyInput возвращает поле ввода тела запроса
func (m *AppModel) GetBodyInput() *textarea.Model {
	return &m.bodyInput
}

// GetParamInput возвращает поле ввода параметров
func (m *AppModel) GetParamInput() *textinput.Model {
	return &m.paramInput
}

// GetHeaderInput возвращает поле ввода заголовков
func (m *AppModel) GetHeaderInput() *textinput.Model {
	return &m.headerInput
}

// GetActiveSection возвращает активную секцию
func (m *AppModel) GetActiveSection() Section {
	return m.activeSection
}

// SetActiveSection устанавливает активную секцию
func (m *AppModel) SetActiveSection(section Section) {
	m.activeSection = section
}

// GetSelectedMethod возвращает выбранный метод
func (m *AppModel) GetSelectedMethod() HTTPMethod {
	return m.selectedMethod
}

// SetSelectedMethod устанавливает выбранный метод
func (m *AppModel) SetSelectedMethod(method HTTPMethod) {
	m.selectedMethod = method
}

// GetInputMode возвращает режим ввода
func (m *AppModel) GetInputMode() bool {
	return m.inputMode
}

// SetInputMode устанавливает режим ввода
func (m *AppModel) SetInputMode(mode bool) {
	m.inputMode = mode
}

// GetLoading возвращает состояние загрузки
func (m *AppModel) GetLoading() bool {
	return m.loading
}

// GetResponse возвращает ответ
func (m *AppModel) GetResponse() string {
	return m.response
}

// GetStatus возвращает статус
func (m *AppModel) GetStatus() string {
	return m.status
}

// GetResponseTime возвращает время ответа
func (m *AppModel) GetResponseTime() string {
	return m.responseTime
}

// GetErrorMsg возвращает сообщение об ошибке
func (m *AppModel) GetErrorMsg() string {
	return m.errorMsg
}

// GetDimensions возвращает размеры окна
func (m *AppModel) GetDimensions() (int, int) {
	return m.width, m.height
}

// FormatJSON форматирует JSON строку с правильными отступами
func FormatJSON(input string) string {
	if input == "" {
		return ""
	}

	// Пытаемся отформатировать как JSON
	var formatted bytes.Buffer
	err := json.Indent(&formatted, []byte(input), "", "  ")
	if err != nil {
		// Если не JSON, возвращаем как есть
		return input
	}
	return formatted.String()
}
