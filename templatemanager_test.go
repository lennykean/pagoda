package pagoda

import (
	"bytes"
	"github.com/fsnotify/fsnotify"
	"testing"
)

type mockWatcher struct {
	watchedList []string
	isClosed    bool
}

func (mockWatcher *mockWatcher) Add(name string) error {
	mockWatcher.watchedList = append(mockWatcher.watchedList, name)
	return nil
}
func (mockWatcher *mockWatcher) Close() error {
	mockWatcher.isClosed = true
	return nil
}

func mockReadFile(string) ([]byte, error) {
	return []byte("<div>test</div>"), nil
}

func Test_NewTemplateManager(t *testing.T) {
	// Arrange
	events := make(chan fsnotify.Event)
	mockWatcher := mockWatcher{}

	// Act
	templateManager := newTemplateManager("testFolder", &mockWatcher, events)

	// Assert
	if templateManager == nil || len(templateManager.funcs) != 1 || templateManager.watcher != &mockWatcher {
		t.Error("failed to create template manager")
	}
}

func Test_GetTemplate(t *testing.T) {
	// Arrange
	events := make(chan fsnotify.Event)
	mockWatcher := mockWatcher{}
	templateManager := newTemplateManager("testFolder", &mockWatcher, events)
	templateManager.readFile = mockReadFile

	// Act
	templateManager.GetTemplate("testTemplate")

	// Assert
	if mockWatcher.watchedList[0] != "testFolder/testTemplate.html" {
		t.Error("failed to watch template file for changes")
	}
	if templateManager.rootTemplate.Lookup("testTemplate") == nil {
		t.Error("failed to add template to cache")
	}
}

func Test_GetTemplate_Caches(t *testing.T) {
	// Arrange
	events := make(chan fsnotify.Event)
	mockWatcher := mockWatcher{}
	templateManager := newTemplateManager("testFolder", &mockWatcher, events)
	templateManager.readFile = mockReadFile

	// Act
	t1, _ := templateManager.GetTemplate("testTemplate")
	t2, _ := templateManager.GetTemplate("testTemplate")

	// Assert
	if t1 != t2 {
		t.Fail()
	}
}

func Test_Invalidate_Cache_On_Change(t *testing.T) {
	// Arrange
	events := make(chan fsnotify.Event)
	mockWatcher := mockWatcher{}
	templateManager := newTemplateManager("testFolder", &mockWatcher, events)
	templateManager.readFile = mockReadFile

	templateManager.GetTemplate("testTemplate")

	// Act
	events <- fsnotify.Event{
		Op:   fsnotify.Write,
		Name: "testFolder/testTemplate.html",
	}

	// Assert
	if templateManager.rootTemplate.Lookup("testTemplate") != nil {
		t.Fail()
	}
}

func Test_Execute(t *testing.T) {
	// Arrange
	events := make(chan fsnotify.Event)
	mockWatcher := mockWatcher{}
	templateManager := newTemplateManager("testFolder", &mockWatcher, events)
	templateManager.readFile = func(name string) ([]byte, error) {
		return []byte("<div>{{.}}</div>"), nil
	}
	buffer := bytes.Buffer{}

	// Act
	templateManager.Execute("testTemplate", &buffer, "test")
	result := buffer.String()

	// Assert
	expected := "<div>test</div>"
	if result != expected {
		t.Errorf("template execution failed, expected: %v, got %v", expected, result)
	}
}

func Test_Execute_SubTemplate(t *testing.T) {
	// Arrange
	events := make(chan fsnotify.Event)
	mockWatcher := mockWatcher{}
	templateManager := newTemplateManager("testFolder", &mockWatcher, events)
	templateManager.readFile = func(name string) ([]byte, error) {
		if name == "testFolder/testTemplate.html" {
			return []byte("<div>{{pagoda_template \"subTemplate\" .}}</div>"), nil
		}
		if name == "testFolder/subTemplate.html" {
			return []byte("<h1>{{.}}</h1>"), nil
		}
		return []byte{}, nil
	}

	// Act
	buffer := bytes.Buffer{}
	templateManager.Execute("testTemplate", &buffer, "test")
	result := buffer.String()

	// Assert
	expected := "<div><h1>test</h1></div>"
	if result != expected {
		t.Errorf("template execution failed, expected: %v, got %v", expected, result)
	}
}

func Test_Execute_LayoutTemplate(t *testing.T) {
	// Arrange
	events := make(chan fsnotify.Event)
	mockWatcher := mockWatcher{}
	templateManager := newTemplateManager("testFolder", &mockWatcher, events)
	templateManager.readFile = func(name string) ([]byte, error) {
		if name == "testFolder/layout.html" {
			return []byte("<html><body>{{pagoda_layout_placeholder .}}</body></html>"), nil
		}
		if name == "testFolder/testTemplate.html" {
			return []byte("<h1>{{.}}</h1>"), nil
		}
		return []byte{}, nil
	}
	layoutTemplateManger := templateManager.UseLayoutTemplate("layout")

	// Act
	buffer := bytes.Buffer{}
	layoutTemplateManger.Execute("testTemplate", &buffer, "test")
	result := buffer.String()

	// Assert
	expected := "<html><body><h1>test</h1></body></html>"
	if result != expected {
		t.Errorf("template execution failed, expected: %v, got %v", expected, result)
	}
}

func Test_Define_Across_Templates(t *testing.T) {
	// Arrange
	events := make(chan fsnotify.Event)
	mockWatcher := mockWatcher{}
	templateManager := newTemplateManager("testFolder", &mockWatcher, events)
	templateManager.readFile = func(name string) ([]byte, error) {
		if name == "testFolder/testTemplate.html" {
			return []byte("<div><div>{{pagoda_template \"subTemplate\"}}</div><div>{{pagoda_template \"definedTemplate\"}}</div></div>"), nil
		}
		if name == "testFolder/subTemplate.html" {
			return []byte("{{define \"definedTemplate\"}}defined-template-render{{end}}sub-template-render"), nil
		}
		return []byte{}, nil
	}

	// Act
	buffer := bytes.Buffer{}
	templateManager.Execute("testTemplate", &buffer, nil)
	result := buffer.String()

	// Assert
	expected := "<div><div>sub-template-render</div><div>defined-template-render</div></div>"
	if result != expected {
		t.Errorf("template execution failed, expected: %v, got %v", expected, result)
	}
}
