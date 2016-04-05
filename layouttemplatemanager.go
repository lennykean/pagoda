package pagoda

import (
	"io"
	"text/template"
)

// LayoutTemplateManager loads and executes templates with a layout page
type LayoutTemplateManager struct {
	*TemplateManager
	layoutTemplate string
}

func getLayoutTemplateManager(templateManager *TemplateManager, layoutTemplateName string) *LayoutTemplateManager {
	return &LayoutTemplateManager{templateManager, layoutTemplateName}
}

// GetTemplate gets a template from the templateFolder based on the templateName
func (layoutTemplateManager *LayoutTemplateManager) GetTemplate(templateName string) (tpl *template.Template, err error) {
	templateName = layoutTemplateManager.getTemplateName(templateName)

	// add layout placeholder func
	funcs := template.FuncMap{
		"pagoda_layout_placeholder": func(data interface{}) string {
			return layoutTemplateManager.execSubTemplate(templateName, data)
		},
	}
	layoutTemplateManager.Funcs(funcs)

	layoutTemplateName := layoutTemplateManager.getTemplateName(layoutTemplateManager.layoutTemplate)
	return layoutTemplateManager.getTemplate(layoutTemplateName, templateName)
}

// Execute a template named templateName
func (layoutTemplateManager *LayoutTemplateManager) Execute(templateName string, writer io.Writer, data interface{}) (err error) {
	tpl, err := layoutTemplateManager.GetTemplate(templateName)

	if err == nil {
		err = tpl.Execute(writer, data)
	}
	return
}
