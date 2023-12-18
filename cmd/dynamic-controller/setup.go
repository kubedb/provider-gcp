// SPDX-FileCopyrightText: 2023 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package dynamic_controller

import (
	"fmt"
	"github.com/crossplane/upjet/pkg/pipeline"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/crossplane/upjet/pkg/config"
	"github.com/crossplane/upjet/pkg/pipeline/templates"
	"github.com/muvaf/typewriter/pkg/wrapper"
	"github.com/pkg/errors"
)

// NewProviderGenerator returns a new ProviderGenerator.
func NewProviderGenerator(rootDir, modulePath string) *ProviderGenerator {
	return &ProviderGenerator{
		ProviderPath:       filepath.Join(rootDir, "cmd", "provider"),
		LocalDirectoryPath: filepath.Join(rootDir, "internal", "controller"),
		LicenseHeaderPath:  filepath.Join(rootDir, "hack", "boilerplate.go.txt"),
		ModulePath:         modulePath,
	}
}

// ProviderGenerator generates controller setup file.
type ProviderGenerator struct {
	ProviderPath       string
	LocalDirectoryPath string
	LicenseHeaderPath  string
	ModulePath         string
}

// Generate writes the setup file and the corresponding provider main file
// using the given list of version packages.
func (sg *ProviderGenerator) Generate(versionPkgMap map[string][]string, mainTemplate string, shortName string, rootGroup string, absRootDir string) error {
	var t *template.Template
	if len(mainTemplate) != 0 {
		tmpl, err := template.New("main").Parse(mainTemplate)
		if err != nil {
			return errors.Wrap(err, "failed to parse the provider main program template")
		}
		t = tmpl
	}
	if t == nil {
		return errors.Wrap(sg.generate(rootGroup, versionPkgMap[config.PackageNameMonolith], shortName, absRootDir), "failed to generate the controller setup file")
	}
	for g, versionPkgList := range versionPkgMap {

		if err := sg.generate(rootGroup, versionPkgList, shortName, absRootDir); err != nil {
			return errors.Wrapf(err, "failed to generate the controller setup file for group: %s", g)
		}
		if err := generateProviderMain(sg.ProviderPath, g, t); err != nil {
			return errors.Wrapf(err, "failed to write main program for group: %s", g)
		}
	}
	return nil
}

func generateProviderMain(providerPath, group string, t *template.Template) error {
	f := filepath.Join(providerPath, group)
	if err := os.MkdirAll(f, 0750); err != nil {
		return errors.Wrapf(err, "failed to mkdir provider main program path: %s", f)
	}
	m, err := os.OpenFile(filepath.Join(filepath.Clean(f), "zz_main.go"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to open provider main program file")
	}
	defer func() {
		if err := m.Close(); err != nil {
			log.Fatalf("Failed to close the templated main %q: %s", f, err.Error())
		}
	}()
	if err := t.Execute(m, map[string]any{
		"Group": group,
	}); err != nil {
		return errors.Wrap(err, "failed to execute provider main program template")
	}
	return nil
}

func (sg *ProviderGenerator) generate(rootGroup string, versionPkgList []string, shortName string, absRootDir string) error {
	setupFile := wrapper.NewFile(filepath.Join(sg.ModulePath, "apis"), "apis", templates.SetupTemplate,
		wrapper.WithGenStatement(pipeline.GenStatement),
		wrapper.WithHeaderPath(sg.LicenseHeaderPath),
	)
	sort.Strings(versionPkgList)
	aliases := make([]string, len(versionPkgList))
	var importData string
	var kindMapData string
	importData += "package controller\n\nimport (\n"
	importData += "\"context\"\n\t\"github.com/crossplane/upjet/pkg/controller\"\n\tapiextensions \"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1\"\n\t\"k8s.io/apimachinery/pkg/runtime/schema\"\n\tctrl \"sigs.k8s.io/controller-runtime\"\n\t\"sigs.k8s.io/controller-runtime/pkg/client\"\n\t\"sigs.k8s.io/controller-runtime/pkg/log\"\n\t\"sync\"\n"
	kindMapData += "\n\nvar (\n"
	kindMapData += "setupFns = map[schema.GroupKind]func(ctrl.Manager, controller.Options) error{\n"
	for i, pkgPath := range versionPkgList {
		aliases[i] = setupFile.Imports.UsePackage(pkgPath)
		rmDot := strings.Split(aliases[i], ".")
		words := strings.Split(pkgPath, "/")
		siz := len(words)
		group := words[siz-2]
		kind := words[siz-1]
		importData += rmDot[0] + " " + "\"" + pkgPath + "\"\n"
		if group == shortName {
			kindMapData += "schema.GroupKind{\"" + rootGroup + "\", "
		} else if group == "controller" {
			kindMapData += "schema.GroupKind{\"" + kind + "." + rootGroup + "\", "
		} else {
			kindMapData += "schema.GroupKind{\"" + group + "." + rootGroup + "\", "
		}
		kindMapData += "\"" + groupKind[kind] + "\"}: " + aliases[i] + "Setup,\n"
	}
	importData += ")\n\n"
	kindMapData += "}\n)\n\n"
	if err := generateControllerFile(importData, kindMapData, absRootDir); err != nil {
		panic(errors.Wrap(err, "cannot create controller"))
	}
	return nil
}

func generateControllerFile(importData string, kindMapData string, absRootDir string) error {
	importData += kindMapData

	filePath := absRootDir + "/internal/controller/zz_dynamic_crd_controller.go" // Replace with the path to your file

	// Attempt to remove the file
	os.Remove(filePath)

	filePath = absRootDir + "/cmd/generator/crd_controller.go.txt" // Replace with the path to your file
	// Read the entire file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	// Convert byte slice to string for printing
	fileContent := string(content)
	importData += fileContent

	filePath = absRootDir + "/internal/controller/zz_dynamic_crd_controller.go"

	// Open or create the file for writing
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Writing data into the file
	_, err = fmt.Fprintf(file, "%s\n", importData)
	if err != nil {
		return err
	}
	return nil
}

/*
var (
	setupFns = map[schema.GroupKind]func(ctrl.Manager, controller.Options) error{
		schema.GroupKind{"azure.kubedb.com", "ResourceGroup"}:  resourcegroup.Setup,
		schema.GroupKind{"azure.kubedb.com", "ProviderConfig"}: providerregistration.Setup,
	}
)
*/
