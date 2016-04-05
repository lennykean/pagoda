package pagoda

import (
	"bytes"
	"github.com/fsnotify/fsnotify"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"
)

type watcher interface {
	Add(string) error
	Close() error
}

const defaultRootTemplateName = ":DEFAULT_ROOT_TEMPLATE:"

// TemplateManager automatically loads, retrieves and executes templates
type TemplateManager struct {
	templateFolder string
	rootTemplates  map[string]*template.Template
	watcher        watcher
	watchEvents    chan fsnotify.Event
	readFile       func(string) ([]byte, error)
	funcs          template.FuncMap
}

// NewTemplateManager creates a new TemplateManager based on templateFolder
func NewTemplateManager(templateFolder string) (templateManager *TemplateManager, err error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		templateManager = nil
	} else {
		templateManager = newTemplateManager(templateFolder, watcher, watcher.Events)
	}
	return
}

func newTemplateManager(templateFolder string, watcher watcher, watchEvents chan fsnotify.Event) *TemplateManager {
	templateManager := &TemplateManager{
		templateFolder: templateFolder,
		rootTemplates:  make(map[string]*template.Template),
		watcher:        watcher,
		watchEvents:    watchEvents,
		readFile:       ioutil.ReadFile,
	}
	templateManager.funcs = template.FuncMap{
		"pagoda_template": templateManager.execSubTemplate,
	}
	go templateManager.watchTemplates()

	return templateManager
}

func (templateManager *TemplateManager) createRootTemplate(name string) {
	templateManager.rootTemplates[name] = template.New(name).Funcs(templateManager.funcs)
}

func (templateManager *TemplateManager) watchTemplates() {
	for {
		select {
		case event := <-templateManager.watchEvents:
			if event.Op&fsnotify.Write == fsnotify.Write {
				// invalidate cache if the file changes
				templateManager.rootTemplates = make(map[string]*template.Template)
			}
		}
	}
}

func (templateManager *TemplateManager) getTemplateName(templateName string) string {
	templateID := templateName
	if strings.HasSuffix(templateID, ".html") {
		templateID = templateID[:len(templateID)-5]
	}
	return templateID
}

func (templateManager *TemplateManager) execSubTemplate(templateName string, args ...interface{}) string {
	tpl, err := templateManager.GetTemplate(templateName)
	if err != nil {
		return ""
	}

	var data interface{}
	if len(args) > 0 {
		data = args[0]
	}

	buffer := bytes.Buffer{}
	tpl.Execute(&buffer, data)

	return buffer.String()
}

func (templateManager *TemplateManager) getTemplate(templateName string, rootTemplateName string) (tpl *template.Template, err error) {
	// try to find the root template in cache, create a new one if not found
	if templateManager.rootTemplates[rootTemplateName] == nil {
		templateManager.createRootTemplate(rootTemplateName)
	}

	// try to get template from cache
	cachedTpl := templateManager.rootTemplates[rootTemplateName].Lookup(templateName)
	if cachedTpl != nil {
		tpl = cachedTpl
		return
	}

	// get template path
	templatePath := filepath.Join(templateManager.templateFolder, templateName+".html")

	// find/parse template file
	file, err := templateManager.readFile(templatePath)
	if err == nil {
		tpl, err = templateManager.rootTemplates[rootTemplateName].New(templateName).Parse(string(file))
	}
	if err == nil {
		templateManager.watcher.Add(templatePath)
	}
	return
}

// Funcs adds template functions
func (templateManager *TemplateManager) Funcs(funcs template.FuncMap) {
	for _, rootTemplate := range templateManager.rootTemplates {
		rootTemplate.Funcs(funcs)
	}
	for name, function := range funcs {
		templateManager.funcs[name] = function
	}
}

// GetTemplate gets a template from the templateFolder based on the templateName
func (templateManager *TemplateManager) GetTemplate(templateName string) (tpl *template.Template, err error) {
	templateName = templateManager.getTemplateName(templateName)
	return templateManager.getTemplate(templateName, defaultRootTemplateName)
}

// UseLayoutTemplate allows templates to be wrapped with a layout template
func (templateManager *TemplateManager) UseLayoutTemplate(layoutTemplateName string) *LayoutTemplateManager {
	return getLayoutTemplateManager(templateManager, layoutTemplateName)
}

// Execute a template named templateName
func (templateManager *TemplateManager) Execute(templateName string, writer io.Writer, data interface{}) (err error) {
	tpl, err := templateManager.GetTemplate(templateName)

	if err == nil {
		err = tpl.Execute(writer, data)
	}
	return
}

// Close cleans up resources
func (templateManager *TemplateManager) Close() error {
	return templateManager.watcher.Close()
}
