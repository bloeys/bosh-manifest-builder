package main

import (
	"bufio"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

type JobSpecProperty struct {
	Default string `yaml:"default,omitempty"`
	Desc    string `yaml:"Description"`
}

type JobSpec struct {
	Name       string
	Properties map[string]*JobSpecProperty
}

type Job struct {
	Name    string
	Release string
	Spec    *JobSpec
}

type InstanceGroup struct {
	Name      string
	Azs       []string
	Instances int
	VmType    string `yaml:"vm_type"`
	Stemcell  string
	Networks  []string
	Jobs      []Job
}

type Deployment struct {
	Name   string
	Groups []*InstanceGroup `yaml:"instance_groups"`
}

type Property struct {
	Name      string
	Children  []*Property
	Contained []string
}

func main() {
	ig := &InstanceGroup{Name: "MyName", Azs: []string{"z1"}, Instances: 1, VmType: "default", Stemcell: "default", Networks: []string{"{name: default}"}, Jobs: []Job{}}
	dep := Deployment{Name: "MyDeployment", Groups: []*InstanceGroup{ig}}

	path := "ceph-objectstorage-broker-boshrelease/jobs"
	jobsDir, _ := ioutil.ReadDir(path)
	for i := 0; i < len(jobsDir); i++ {
		specFile, _ := ioutil.ReadFile(path + "/" + jobsDir[i].Name() + "/spec")
		js := JobSpec{}
		yaml.Unmarshal(specFile, &js)

		ig.Jobs = append(ig.Jobs, Job{Name: js.Name, Release: js.Name, Spec: &js})
	}
	MakeManifest(&dep)
}

func MakeManifest(dep *Deployment) {
	f, _ := os.Create("manifest.yml")
	defer f.Close()
	writer := bufio.NewWriter(f)

	writer.WriteString("---\nname: " + dep.Name + "\n")
	writer.WriteString("\ninstance_groups:\n")
	for i := 0; i < len(dep.Groups); i++ {
		//Print group information
		g := dep.Groups[i]
		writer.WriteString("- name: " + g.Name + "\n")
		writer.WriteString("  azs: " + fmt.Sprint(g.Azs) + "\n")
		writer.WriteString("  instances: " + fmt.Sprint(g.Instances) + "\n")
		writer.WriteString("  vm_type: " + g.VmType + "\n")
		writer.WriteString("  stemcell: " + g.Stemcell + "\n")
		writer.WriteString("  networks: " + fmt.Sprint(g.Networks) + "\n")
		writer.WriteString("  jobs:\n")

		for j := 0; j < len(g.Jobs); j++ {
			jobString := GetJob(g.Jobs[j])
			writer.WriteString(jobString)
		}
	}
	writer.Flush()
}

func GetJob(j Job) string {

	props := map[string]*Property{}
	s := "  - name: " + j.Name + "\n    release: " + j.Release + "\n    properties:\n"
	for k, v := range j.Spec.Properties {
		split := strings.Split(strings.Replace(k, ":", "", 1), ".")
		props[split[0]] = CreatePropertyTree(props[split[0]], split, v.Default)
	}

	indentBase := 4
	for _, v := range props {
		for i := 0; i < len(v.Children); i++ {
			indent := indentBase

			p := v.Children[i]
			indent = indent + 2
			s = s + strings.Repeat(" ", indent) + p.Name + ":\n"

			indent = indent + 2
			for i := 0; i < len(p.Contained); i++ {
				s = s + strings.Repeat(" ", indent) + p.Contained[i] + "\n"
			}
			indent = indent - 2
		}
	}

	return s
}

func CreatePropertyTree(prop *Property, names []string, def string) *Property {
	if prop == nil {
		prop = &Property{Name: names[0], Contained: []string{}, Children: []*Property{}}
	}

	curr := prop
	for i := 1; i < len(names)-1; i++ {
		if !HasProperty(curr.Children, names[i]) {
			curr.Children = append(curr.Children, &Property{Name: names[i], Contained: []string{}, Children: nil})
		}
		curr = curr.Children[len(curr.Children)-1]
	}

	if def == "" {
		curr.Contained = append(curr.Contained, names[len(names)-1]+": \""+def+"\"")
	} else {
		curr.Contained = append(curr.Contained, names[len(names)-1]+": "+def)
	}

	return prop
}

func HasProperty(list []*Property, s string) bool {
	for i := 0; i < len(list); i++ {
		if s == list[i].Name {
			return true
		}
	}
	return false
}
