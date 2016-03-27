package pagoda

import (
    "bytes"
    "io"
    "io/ioutil"
    "github.com/fsnotify/fsnotify"
    "strings"
    "text/template"
)

// TemplateManager autmoatically loads, retrieves and executes templates
type TemplateManager struct {
    templateFolder string
    templates map[string]*template.Template
    funcs template.FuncMap
    watcher *fsnotify.Watcher
} 

// NewTemplateManager creates a new TemplateManager based on templateFolder
func NewTemplateManager(templateFolder string) (templateManager *TemplateManager, err error) {
    templateManager = &TemplateManager{
        templateFolder: templateFolder,
        templates: make(map[string]*template.Template),        
    }    
    templateManager.funcs = template.FuncMap{
        "pagoda_template": templateManager.execSubTemplate,
    }
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        templateManager = nil
    } else {          
        templateManager.watcher = watcher
        go templateManager.watchTemplates()
    }    
    return 
}

func (templateManager *TemplateManager) watchTemplates() {
    for {
        select {
            case event := <-templateManager.watcher.Events:
                if event.Op & fsnotify.Write == fsnotify.Write {
                    // invalidate cache if the file changes
                    templateID := templateManager.getTemplateIDFromTemplatePath(event.Name)
                    delete(templateManager.templates, templateID)
                }
        }
    }
}

func (templateManager *TemplateManager) getTemplateIDFromTemplatePath(templatePath string) string {
    templateID := templatePath[len(templateManager.templateFolder):] 
                    
    if strings.HasPrefix(templateID, "/") {
        templateID = templateID[1:]
    }                                        
    if strings.HasSuffix(templateID, ".html") {
        templateID = templateID[:len(templateID) - 5]
    }
    return templateID
}

func (templateManager *TemplateManager) getTemplateIDFromTemplateName(templateName string) string {
    templateID := templateName
    if (strings.HasSuffix(templateID, ".html")) {
        templateID = templateID[:len(templateID) - 5]
    }  
    return templateID
}

func (templateManager *TemplateManager) execSubTemplate(templateName string, args ...interface{}) string {  
    tpl, err := templateManager.GetTemplate(templateName)
    if err != nil {
        return ""
    }
    
    var data interface{};
    if len(args) > 0 {
        data = args[0]
    }
    
    buffer := bytes.Buffer{}
    tpl.Execute(&buffer, data)
    
    return buffer.String()
}

// GetTemplate gets a template from the templateFolder based on the templateName
func (templateManager *TemplateManager) GetTemplate(templateName string) (tpl *template.Template, err error) {
    templateID := templateManager.getTemplateIDFromTemplateName(templateName)
    
    // try to get template from cache
    cachedTpl := templateManager.templates[templateID];
    if cachedTpl != nil {
        tpl = cachedTpl
        return 
    }   
            
    // get template path    
    templatePath := templateManager.templateFolder
    if (!strings.HasSuffix(templateManager.templateFolder, "/")) {
        templatePath += "/"
    } 
    templatePath += templateID + ".html"
    
    // find/parse template file       
    file, err := ioutil.ReadFile(templatePath)
    if err == nil {                
        tpl, err = template.New(templateName).Funcs(templateManager.funcs).Parse(string(file))
    }
    if (err == nil) {
        templateManager.watcher.Add(templatePath)
        templateManager.templates[templateID] = tpl
    }   
    return
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
func (templateManager *TemplateManager) Close() {
    templateManager.watcher.Close()
}