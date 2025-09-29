package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
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
	TabSaved
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

var MethodNames = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

// --- Структуры данных ---

type Param struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SavedRequest определяет структуру для сохранения запроса в JSON
type SavedRequest struct {
	Name    string     `json:"name"`
	Method  HTTPMethod `json:"method"`
	URL     string     `json:"url"`
	Body    string     `json:"body"`
	Headers []Header   `json:"headers"`
	Params  []Param    `json:"params"`
}

// Implement list.Item interface for SavedRequest
func (sr SavedRequest) Title() string { return sr.Name }
func (sr SavedRequest) Description() string {
	return fmt.Sprintf("[%s] %s", MethodNames[sr.Method], sr.URL)
}
func (sr SavedRequest) FilterValue() string { return sr.Name }

type ResponseData struct {
	Body       string
	Status     string
	Time       string
	StatusCode int
}

type ErrorData struct {
	Message string
}

// AppModel представляет основное состояние приложения
type AppModel struct {
	// Компоненты
	urlInput      textinput.Model
	bodyInput     textarea.Model
	responseVP    viewport.Model
	paramInput    textinput.Model
	headerInput   textinput.Model
	savedList     list.Model
	saveNameInput textinput.Model

	// Данные
	params        []Param
	headers       []Header
	savedRequests []list.Item // []SavedRequest
	configPath    string

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
	isSaving       bool
	isDeleting     bool

	// Размеры
	width  int
	height int
}

// --- Инициализация ---

func NewAppModel() *AppModel {
	urlInput := textinput.New()
	urlInput.Placeholder = "https://api.example.com/endpoint"
	urlInput.CharLimit = 500

	bodyInput := textarea.New()
	bodyInput.Placeholder = "{\"key\": \"value\"}"
	bodyInput.ShowLineNumbers = false
	bodyInput.FocusedStyle.CursorLine = lipgloss.NewStyle()

	responseVP := viewport.New(10, 10)

	paramInput := textinput.New()
	paramInput.Placeholder = "key=value"
	paramInput.CharLimit = 100

	headerInput := textinput.New()
	headerInput.Placeholder = "Content-Type=application/json"
	headerInput.CharLimit = 100

	saveNameInput := textinput.New()
	saveNameInput.Placeholder = "My Awesome Request"
	saveNameInput.CharLimit = 100

	savedList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	savedList.Title = "Сохраненные запросы"
	savedList.SetShowStatusBar(false)

	configPath, _ := getConfigPath()

	m := &AppModel{
		urlInput:       urlInput,
		bodyInput:      bodyInput,
		responseVP:     responseVP,
		paramInput:     paramInput,
		headerInput:    headerInput,
		savedList:      savedList,
		saveNameInput:  saveNameInput,
		params:         []Param{},
		headers:        []Header{{Key: "Content-Type", Value: "application/json"}},
		configPath:     configPath,
		activeTab:      TabRequest,
		activeSection:  SectionMethod,
		selectedMethod: MethodGET,
	}

	m.loadRequests()
	return m
}

// --- Логика Сохранения/Загрузки ---

func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	postuiDir := filepath.Join(configDir, "postui")
	if err := os.MkdirAll(postuiDir, 0750); err != nil {
		return "", err
	}
	return filepath.Join(postuiDir, "requests.json"), nil
}

func (m *AppModel) loadRequests() error {
	data, err := ioutil.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var savedRequests []SavedRequest
	if err := json.Unmarshal(data, &savedRequests); err != nil {
		return err
	}

	items := make([]list.Item, len(savedRequests))
	for i, sr := range savedRequests {
		items[i] = sr
	}

	m.savedRequests = items
	m.savedList.SetItems(m.savedRequests)
	return nil
}

func (m *AppModel) saveRequests() error {
	savedRequests := make([]SavedRequest, len(m.savedRequests))
	for i, item := range m.savedRequests {
		savedRequests[i] = item.(SavedRequest)
	}

	data, err := json.MarshalIndent(savedRequests, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(m.configPath, data, 0644)
}

func (m *AppModel) AddNewSavedRequest(name string) {
	newReq := SavedRequest{
		Name:    name,
		Method:  m.selectedMethod,
		URL:     m.urlInput.Value(),
		Body:    m.bodyInput.Value(),
		Headers: m.headers,
		Params:  m.params,
	}
	m.savedRequests = append(m.savedRequests, newReq)
	m.savedList.SetItems(m.savedRequests)
	m.saveRequests()
}

func (m *AppModel) LoadRequestFromSaved() {
	if item, ok := m.savedList.SelectedItem().(SavedRequest); ok {
		m.selectedMethod = item.Method
		m.urlInput.SetValue(item.URL)
		m.bodyInput.SetValue(item.Body)
		m.headers = item.Headers
		m.params = item.Params
		m.activeTab = TabRequest
	}
}

func (m *AppModel) DeleteSelectedRequest() {
	if len(m.savedRequests) > 0 {
		idx := m.savedList.Index()
		m.savedRequests = append(m.savedRequests[:idx], m.savedRequests[idx+1:]...)
		m.savedList.SetItems(m.savedRequests)
		m.saveRequests()
	}
}

// --- Обновление состояния ---

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
	m.savedList.SetSize(contentWidth, contentHeight)

	m.urlInput.Width = contentWidth - 14
	m.paramInput.Width = contentWidth - 14
	m.headerInput.Width = contentWidth - 14
	m.bodyInput.SetWidth(contentWidth - 14 - 2)
	m.saveNameInput.Width = contentWidth - 20

	occupiedHeight := len(m.headers) + len(m.params) + 17
	bodyHeight := contentHeight - occupiedHeight
	if bodyHeight < 3 {
		bodyHeight = 3
	}
	m.bodyInput.SetHeight(bodyHeight)
}

func (m *AppModel) SetResponseData(data ResponseData) {
	m.loading = false
	m.response = FormatJSON(data.Body)
	m.status = fmt.Sprintf("%s (%d)", data.Status, data.StatusCode)
	m.responseTime = data.Time
	m.responseVP.SetContent(m.response)
	m.errorMsg = ""
	m.activeTab = TabResponse
}

func (m *AppModel) SetError(err ErrorData) {
	m.loading = false
	m.errorMsg = err.Message
	m.response = ""
	m.status = "Error"
	m.responseTime = ""
	m.responseVP.SetContent(err.Message)
	m.activeTab = TabResponse
}

func (m *AppModel) GetCurrentMethod() string {
	return MethodNames[m.selectedMethod]
}

func (m *AppModel) AddParam(key, value string) {
	m.params = append(m.params, Param{Key: key, Value: value})
}

func (m *AppModel) RemoveLastParam() {
	if len(m.params) > 0 {
		m.params = m.params[:len(m.params)-1]
	}
}

func (m *AppModel) AddHeader(key, value string) {
	m.headers = append(m.headers, Header{Key: key, Value: value})
}

func (m *AppModel) RemoveLastHeader() {
	if len(m.headers) > 0 {
		m.headers = m.headers[:len(m.headers)-1]
	}
}

func (m *AppModel) SetLoading(loading bool) {
	m.loading = loading
}

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

func (m *AppModel) GetSavedList() *list.Model {
	return &m.savedList
}

func (m *AppModel) GetSaveNameInput() *textinput.Model {
	return &m.saveNameInput
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

func (m *AppModel) IsSaving() bool {
	return m.isSaving
}

func (m *AppModel) SetIsSaving(saving bool) {
	m.isSaving = saving
}

func (m *AppModel) IsDeleting() bool {
	return m.isDeleting
}

func (m *AppModel) SetIsDeleting(deleting bool) {
	m.isDeleting = deleting
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
