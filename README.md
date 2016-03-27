# Pagoda

[![GoDoc](https://godoc.org/github.com/lennykean/pagoda?status.svg)](https://godoc.org/github.com/lennykean/pagoda) 
[![Go Report Card](https://goreportcard.com/badge/github.com/lennykean/pagoda)](https://goreportcard.com/report/github.com/lennykean/pagoda) 
[![Coverage](http://gocover.io/_badge/github.com/lennykean/pagoda)](http://gocover.io/github.com/lennykean/pagoda) 

A simple template manager for go

## Features
* Automatic template retrieval
* Simplified template execution
* Template caching
* Template change detection

# Examples

##Executing a template named home/index.html from your template directory
``` Go
templateManager, _ := pagoda.NewTemplateManager(myTemplateDirectory)

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    templateManager.Execute("home/index", w, nil)        
})
```
##Executing a template with sub-templates
``` Go
templateManager, _ := pagoda.NewTemplateManager(myTemplateDirectory)

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    templateManager.Execute("index", w, []string{"Pagani Huayra", "Ferrari LaFerrari", "McLaren P1"})
})
```
index.html

**Note: when loading sub-templates, use ```pagoda_template``` instead of ```template```** 
``` html            
<html>
    <body>
        <h1>My Garage</h1>
        <ul>
        {{ range . }}
            {{ pagoda_template "car" . }}
        {{ end }}
        </ul>
    </body>
</html>
```
##Layout pages

Pagoda's layout feature allows you to use a site-wide template to maintain a consistant layout across all pages

``` Go
templateManager, _ := pagoda.NewTemplateManager(myTemplateDirectory)

layoutTemplateManager := templateManager.UseLayoutTemplate("layout")

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    layoutTemplateManager.Execute("home", w, []string{"Pagani Huayra", "Ferrari LaFerrari", "McLaren P1"})
})
```
layout.html

**Use ```{{pagoda_layout_placeholder .}}``` to set the location you want your template to render in the layout template ** 
``` html
<html>
    <head>
        <title></title>
    <head>
    <body>
        {{ pagoda_layout_placeholder . }}
    </body>
</html>
```
home.html 
``` html
<h1>My Garage</h1>
<ul>
{{ range . }}
    <li>{{.}}</li>
{{ end }}
</ul>
```

