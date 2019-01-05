package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	yaml "gopkg.in/yaml.v2"
)

const (
	headerTemplateFile  = "ro-header.tpl"
	groupReadmeTemplate = "wg-readme.tpl"
	cmReadmeTemplate    = "cm-readme.tpl"

	groupReadmeFile = "README.md"
	cmReadmeFile    = "README.md"

	cmDir          = "community"
	templateDir    = "tools/generators/community/templates"
	groupsYamlFile = "community/wgs.yaml"
)

// Person -
type Person struct {
	Name   string
	GitHub string
}

// Subproject -
type Subproject struct {
	Name        string
	Description string
	Owners      []string
}

// RoadmapTask -
type RoadmapTask struct {
	Name        string
	Description string
	Issue       string
}

// RoadmapSubproject -
type RoadmapSubproject struct {
	Name        string
	Description string
	Tasks       []RoadmapTask
}

// Roadmap -
type Roadmap struct {
	Name        string
	Description string
	Subprojects []RoadmapSubproject
}

// Group -
type Group struct {
	Name        string
	Dir         string
	Description string
	Label       string
	Chairs      []Person
	Subprojects []Subproject
	Roadmaps    []Roadmap
}

func writeTemplate(templatePath, outputPath string, data interface{}) error {
	t, err := template.
		New(filepath.Base(templatePath)).
		ParseFiles(
			filepath.Join(templateDir, headerTemplateFile),
			filepath.Join(templateDir, templatePath),
		)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Truncate(0)

	err = t.Execute(f, data)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	yamlData, err := ioutil.ReadFile(groupsYamlFile)
	if err != nil {
		log.Fatal(err)
	}

	groups := []Group{}
	err = yaml.Unmarshal(yamlData, &groups)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Generating group READMEs")
	for _, group := range groups {
		fmt.Printf("> %s (%s)\n", group.Name, group.Label)
		fn := filepath.Join(cmDir, group.Label, groupReadmeFile)
		err := writeTemplate(groupReadmeTemplate, fn, group)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Generating community README")
	cmReadmePath := filepath.Join(cmDir, cmReadmeFile)
	err = writeTemplate(cmReadmeTemplate, cmReadmePath, groups)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done")
}
