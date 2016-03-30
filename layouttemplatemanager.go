package pagoda

import (
	"io"
	"text/template"
)

// LayoutTemplateManager loads and executes templates with a layout page
type LayoutTemplateManager struct {
	*TemplateManager
	layoutTemplate string
	funcs          template.FuncMap
}

func getLayoutTemplateManager(templateManager *TemplateManager, layoutTemplateName string) *LayoutTemplateManager {
	return &LayoutTemplateManager{templateManager, layoutTemplateName, template.FuncMap{}}
}

// GetTemplate gets a template from the templateFolder based on the templateName
func (layoutTemplateManager *LayoutTemplateManager) GetTemplate(templateName string) (tpl *template.Template, err error) {
	templateName = layoutTemplateManager.getTemplateName(templateName)

	// try to get the layout root template from cache, otherwise create a new one
	rootTemplate := layoutTemplateManager.layoutTemplates[templateName]
	if rootTemplate == nil {
		rootTemplate = layoutTemplateManager.createRootTemplate().Funcs(layoutTemplateManager.funcs)
		layoutTemplateManager.layoutTemplates[templateName] = rootTemplate
	}

	// add layout placeholder func
	funcs := template.FuncMap{
		"pagoda_layout_placeholder": func(data interface{}) string {
			return layoutTemplateManager.execSubTemplate(templateName, data)
		},
	}

	layoutTemplateName := layoutTemplateManager.getTemplateName(layoutTemplateManager.layoutTemplate)
	return layoutTemplateManager.getTemplate(layoutTemplateName, rootTemplate, funcs)
}

// Execute a template named templateName
func (layoutTemplateManager *LayoutTemplateManager) Execute(templateName string, writer io.Writer, data interface{}) (err error) {
	tpl, err := layoutTemplateManager.GetTemplate(templateName)

	if err == nil {
		err = tpl.Execute(writer, data)
	}
	return
}

// Funcs adds template functions
func (layoutTemplateManager *LayoutTemplateManager) Funcs(funcs template.FuncMap) {
	for _, rootTemplate := range layoutTemplateManager.layoutTemplates {
		rootTemplate.Funcs(funcs)
	}
	for name, function := range funcs {
		layoutTemplateManager.funcs[name] = function
	}
}
