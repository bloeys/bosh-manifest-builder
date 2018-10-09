package main

import (
	// "bufio"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	// "os"
)

type Job struct {
	Name       string
	Release    string
	Properties map[string]string
}

type JobSpecProperty struct {
	Default string `yaml:"default,omitempty"`
	Desc    string `yaml:"Description"`
}

type JobSpec struct {
	Name       string
	Packages   []string
	Properties map[string]JobSpecProperty
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
	Groups []InstanceGroup `yaml:"instance_groups"`
}

func main() {
	// ig := InstanceGroup{Name: "MyName", Azs: []string{"z1"}, Instances: 1, VmType: "default", Stemcell: "default", Networks: []string{"default"}, Jobs: []Job{}}
	// dep := Deployment{Name: "MyDeployment", Groups: []InstanceGroup{ig}}

	path := "ceph-objectstorage-broker-boshrelease/jobs"
	jobsDir, _ := ioutil.ReadDir(path)
	for i := 0; i < len(jobsDir); i++ {
		f, _ := ioutil.ReadFile(path + "/" + jobsDir[i].Name() + "/spec")
		js := JobSpec{}
		yaml.Unmarshal(f, &js)

		y, _ := yaml.Marshal(&js)
		fmt.Println(string(y))
	}
	// y, _ := yaml.Marshal(&dep)
	// fmt.Println("---\n" + string(y))
}
