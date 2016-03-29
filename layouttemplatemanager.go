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

func (layoutTemplateManager *LayoutTemplateManager) getLayoutTemplateID(templateName string) string {        
    layoutTemplateName := layoutTemplateManager.getTemplateName(layoutTemplateManager.layoutTemplate)
    templateName = layoutTemplateManager.getTemplateName(templateName)
    
    return layoutTemplateName + ":" + templateName
}

// GetTemplate gets a template from the templateFolder based on the templateName
func (layoutTemplateManager *LayoutTemplateManager) GetTemplate(templateName string) (tpl *template.Template, err error) {
	funcs := layoutTemplateManager.funcs
	funcs["pagoda_layout_placeholder"] = func(data interface{}) string {
		return layoutTemplateManager.execSubTemplate(templateName, data)
	}
    layoutTemplateName := layoutTemplateManager.getTemplateName(layoutTemplateManager.layoutTemplate)
    layoutTemplateID := layoutTemplateManager.getLayoutTemplateID(templateName)
    
	return layoutTemplateManager.getTemplate(layoutTemplateName, layoutTemplateID, funcs)
}

// Execute a template named templateName
func (layoutTemplateManager *LayoutTemplateManager) Execute(templateName string, writer io.Writer, data interface{}) (err error) {
	tpl, err := layoutTemplateManager.GetTemplate(templateName)

	if err == nil {
		err = tpl.Execute(writer, data)
	}
	return
}
