package dynamic_controller

import (
	"fmt"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/upjet/pkg/config"
	"github.com/crossplane/upjet/pkg/examples"
	"github.com/crossplane/upjet/pkg/pipeline"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var (
	groupKind map[string]string
)

func GenerateController(pc *config.Provider, absRootDir string) {
	groupKind = make(map[string]string)
	if len(os.Args) < 2 || os.Args[1] == "" {
		panic("root directory is required to be given as argument")
	}
	rootDir := os.Args[1]
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		panic(fmt.Sprintf("cannot calculate the absolute path with %s", rootDir))
	}
	resourcesGroups := map[string]map[string]map[string]*config.Resource{}
	for name, resource := range pc.Resources {
		group := pc.RootGroup
		if resource.ShortGroup != "" {
			group = strings.ToLower(resource.ShortGroup) + "." + pc.RootGroup
		}
		if len(resourcesGroups[group]) == 0 {
			resourcesGroups[group] = map[string]map[string]*config.Resource{}
		}
		if len(resourcesGroups[group][resource.Version]) == 0 {
			resourcesGroups[group][resource.Version] = map[string]*config.Resource{}
		}
		resourcesGroups[group][resource.Version][name] = resource
		kind := strings.ToLower(resource.Kind)
		groupKind[kind] = resource.Kind
	}

	exampleGen := examples.NewGenerator(rootDir, pc.ModulePath, pc.ShortName, pc.Resources)
	if err := exampleGen.SetReferenceTypes(pc.Resources); err != nil {
		panic(errors.Wrap(err, "cannot set reference types for resources"))
	}
	// Add ProviderConfig API package to the list of API version packages.
	apiVersionPkgList := make([]string, 0)
	for _, p := range pc.BasePackages.APIVersion {

		apiVersionPkgList = append(apiVersionPkgList, filepath.Join(pc.ModulePath, p))
	}
	// Add ProviderConfig controller package to the list of controller packages.
	controllerPkgMap := make(map[string][]string)
	// new API takes precedence
	for p, g := range pc.BasePackages.ControllerMap {
		path := filepath.Join(pc.ModulePath, p)
		controllerPkgMap[g] = append(controllerPkgMap[g], path)
		controllerPkgMap[config.PackageNameMonolith] = append(controllerPkgMap[config.PackageNameMonolith], path)
	}
	//nolint:staticcheck
	for _, p := range pc.BasePackages.Controller {
		path := filepath.Join(pc.ModulePath, p)
		found := false
		for _, p := range controllerPkgMap[config.PackageNameConfig] {
			if path == p {
				found = true
				break
			}
		}
		if !found {
			controllerPkgMap[config.PackageNameConfig] = append(controllerPkgMap[config.PackageNameConfig], path)
		}
		found = false
		for _, p := range controllerPkgMap[config.PackageNameMonolith] {
			if path == p {
				found = true
				break
			}
		}
		if !found {
			controllerPkgMap[config.PackageNameMonolith] = append(controllerPkgMap[config.PackageNameMonolith], path)
		}
	}
	count := 0
	for group, versions := range resourcesGroups {
		for version, resources := range versions {
			versionGen := pipeline.NewVersionGenerator(rootDir, pc.ModulePath, group, version)
			crdGen := pipeline.NewCRDGenerator(versionGen.Package(), rootDir, pc.ShortName, group, version)
			var _ = pipeline.NewTerraformedGenerator(versionGen.Package(), rootDir, group, version)
			ctrlGen := pipeline.NewControllerGenerator(rootDir, pc.ModulePath, group)

			for _, name := range sortedResources(resources) {
				_, err := crdGen.Generate(resources[name])
				if err != nil {
					panic(errors.Wrapf(err, "cannot generate crd for resource %s", name))
				}
				featuresPkgPath := ""
				if pc.FeaturesPackage != "" {
					featuresPkgPath = filepath.Join(pc.ModulePath, pc.FeaturesPackage)
				}
				ctrlPkgPath, err := ctrlGen.Generate(resources[name], versionGen.Package().Path(), featuresPkgPath)
				if err != nil {
					panic(errors.Wrapf(err, "cannot generate controller for resource %s", name))
				}
				sGroup := strings.Split(group, ".")[0]
				controllerPkgMap[sGroup] = append(controllerPkgMap[sGroup], ctrlPkgPath)
				controllerPkgMap[config.PackageNameMonolith] = append(controllerPkgMap[config.PackageNameMonolith], ctrlPkgPath)

				count++
			}
			if err := versionGen.Generate(); err != nil {
				panic(errors.Wrap(err, "cannot generate version files"))
			}
			apiVersionPkgList = append(apiVersionPkgList, versionGen.Package().Path())
		}
	}
	if err := NewProviderGenerator(rootDir, pc.ModulePath).Generate(controllerPkgMap, pc.MainTemplate, pc.ShortName, pc.RootGroup, absRootDir); err != nil {
		panic(errors.Wrap(err, "cannot generate setup file"))
	}

	internalCmd := exec.Command("bash", "-c", "goimports -w $(find . -iname 'zz_*')")
	internalCmd.Dir = filepath.Clean(filepath.Join(rootDir, "internal"))
	if out, err := internalCmd.CombinedOutput(); err != nil {
		panic(errors.Wrap(err, "cannot run goimports for internal folder: "+string(out)))
	}
}

func sortedResources(m map[string]*config.Resource) []string {
	result := make([]string, len(m))
	i := 0
	for g := range m {
		result[i] = g
		i++
	}
	sort.Strings(result)
	return result
}
