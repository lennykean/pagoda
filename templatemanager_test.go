package pagoda

import (
	"bytes"
	"github.com/fsnotify/fsnotify"
	"os"
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
	if templateManager == nil || templateManager.watcher != &mockWatcher {
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
	if mockWatcher.watchedList[0] != "testFolder"+string(os.PathSeparator)+"testTemplate.html" {
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
		Name: "testFolder" + string(os.PathSeparator) + "testTemplate.html",
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
		if name == "testFolder"+string(os.PathSeparator)+"testTemplate.html" {
			return []byte("<div>{{pagoda_template \"subTemplate\" .}}</div>"), nil
		}
		if name == "testFolder"+string(os.PathSeparator)+"subTemplate.html" {
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
		if name == "testFolder"+string(os.PathSeparator)+"layout.html" {
			return []byte("<html><body>{{pagoda_layout_placeholder .}}</body></html>"), nil
		}
		if name == "testFolder"+string(os.PathSeparator)+"testTemplate.html" {
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

func Test_LayoutTemplate_Does_Not_Cache_Inner_Template(t *testing.T) {
	// Arrange
	events := make(chan fsnotify.Event)
	mockWatcher := mockWatcher{}
	templateManager := newTemplateManager("testFolder", &mockWatcher, events)
	templateManager.readFile = func(name string) ([]byte, error) {
		if name == "testFolder"+string(os.PathSeparator)+"layout.html" {
			return []byte("<html><body>{{pagoda_layout_placeholder .}}</body></html>"), nil
		}
		if name == "testFolder"+string(os.PathSeparator)+"testTemplate1.html" {
			return []byte("<h1>{{.}}</h1>"), nil
		}
		if name == "testFolder"+string(os.PathSeparator)+"testTemplate2.html" {
			return []byte("<h2>{{.}}</h2>"), nil
		}
		return []byte{}, nil
	}
	layoutTemplateManger := templateManager.UseLayoutTemplate("layout")

	// Act
	buffer1 := bytes.Buffer{}
	buffer2 := bytes.Buffer{}
	buffer3 := bytes.Buffer{}
	layoutTemplateManger.Execute("testTemplate1", &buffer1, "test")
	layoutTemplateManger.Execute("testTemplate2", &buffer2, "test")
	layoutTemplateManger.Execute("testTemplate1", &buffer3, "test")
	result1 := buffer1.String()
	result2 := buffer2.String()
	result3 := buffer3.String()

	// Assert
	expectedt1 := "<html><body><h1>test</h1></body></html>"
	expectedt2 := "<html><body><h2>test</h2></body></html>"
	if result1 != expectedt1 {
		t.Errorf("template execution 1 failed, expected: %v, got %v", expectedt1, result1)
	}
	if result2 != expectedt2 {
		t.Errorf("template execution 2 failed, expected: %v, got %v", expectedt2, result2)
	}
	if result3 != expectedt1 {
		t.Errorf("template execution 3 failed, expected: %v, got %v", expectedt1, result3)
	}
}

func Test_Define_Across_Templates(t *testing.T) {
	// Arrange
	events := make(chan fsnotify.Event)
	mockWatcher := mockWatcher{}
	templateManager := newTemplateManager("testFolder", &mockWatcher, events)
	templateManager.readFile = func(name string) ([]byte, error) {
		if name == "testFolder"+string(os.PathSeparator)+"testTemplate.html" {
			return []byte("<div><div>{{pagoda_template \"subTemplate\"}}</div><div>{{pagoda_template \"definedTemplate\"}}</div></div>"), nil
		}
		if name == "testFolder"+string(os.PathSeparator)+"subTemplate.html" {
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
