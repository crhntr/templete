package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"mime"
	"path/filepath"
	"reflect"
	"strings"
	"text/template/parse"

	"github.com/spf13/pflag"
)

func main() {
	var fileExtensions []string
	pflag.StringArrayVarP(&fileExtensions, "template-file-extension", "e", []string{".template"}, "file extensions to parse parse as templates")
	pflag.Parse()

	ts := template.New("")

	filePaths := pflag.Args()
	if len(filePaths) == 0 {
		filePaths = []string{"."}
	}
	for _, filePath := range filePaths {
		root, err := filepath.Abs(filePath)
		if err != nil {
			log.Fatal(err)
		}
		walkErr := filepath.Walk(root, func(filePath string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			templateExtension, matches := templateFileExtension(fileExtensions, filePath)
			if !matches {
				return nil
			}

			filePathWithoutTemplateExtension := strings.TrimSuffix(filePath, templateExtension)

			mediaType := mime.TypeByExtension(filepath.Ext(filePathWithoutTemplateExtension))

			log.Println(filePath, filePathWithoutTemplateExtension, mediaType)

			ts, err = ts.ParseFiles(filePath)
			if err != nil {
				return err
			}

			return nil
		})
		if walkErr != nil {
			log.Fatal(walkErr)
		}
	}

	iterateTemplate(ts)
}

func templateFileExtension(fileExtensions []string, filePath string) (string, bool) {
	for _, fe := range fileExtensions {
		if fe == filepath.Ext(filePath) {
			return fe, true
		}
	}
	return "", false
}

func iterateTemplate(templates *template.Template) {
	for _, ts := range templates.Templates() {

		fmt.Println(templateString(ts))
		// fmt.Println(ts.Name())

		// findTemplateNodes(ts)
	}
}

func findTemplateNodes(ts *template.Template) {
	if ts.Tree == nil || ts.Tree.Root == nil {
		return
	}
	if ts.Tree.Root.NodeType != parse.NodeList {
		panic("unexpected node type")
	}
	for _, child := range ts.Tree.Root.Nodes {
		templateNode, isTemplate := child.(*parse.TemplateNode)
		if !isTemplate {
			continue
		}
		tmpl := ts.Lookup(templateNode.Name)
		if tmpl == nil {
			panic("template not found")
		}

		x := fmt.Sprintf("%s", tmpl.Tree)
		if strings.Contains(x, "/* ") {
			fmt.Print("\n=====================\n\n\n")
			fmt.Println(tmpl)
			fmt.Print("\n\n\n=====================\n\n")
		}
	}
}

func templateString(ts *template.Template) string {
	if ts == nil || ts.Tree == nil {
		return ""
	}
	return reflect.ValueOf(ts.Tree).Elem().FieldByName("text").String()
}
