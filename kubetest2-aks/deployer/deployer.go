/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package deployer implements the kubetest2 kind deployer
package deployer

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucasjones/reggen"
	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"sigs.k8s.io/kubetest2/pkg/exec"
	"sigs.k8s.io/kubetest2/pkg/types"
)

var GitTag string
var randomPostFix, _ = RandString(6)
var AksResourceGroup = "aks-rg-" + randomPostFix
var AksClusterName = "aks-cluster-" + randomPostFix

var home, _ = os.UserHomeDir()

// Name is the name of the deployer
const Name = "aks"

// New implements deployer.New for kind
func New(opts types.Options) (types.Deployer, *pflag.FlagSet) {
	// create a deployer object and set fields that are not flag controlled
	d := &deployer{
		commonOptions: opts,
		logsDir:       filepath.Join(opts.RunDir(), "cluster-logs"),
	}
	// register flags and return
	return d, bindFlags(d)
}

// assert that New implements types.NewDeployer
var _ types.NewDeployer = New

type deployer struct {
	// generic parts
	commonOptions types.Options

	logsDir     string
	kubectlPath string

	OverwriteLogsDir bool `desc:"If set, will overwrite an existing logs directory if one is encountered during dumping of logs. Useful when runnning tests locally."`

	RepoRoot string `desc:"The path to the root of the local kubernetes/cloud-provider-gcp repo. Necessary to call certain scripts. Defaults to the current directory. If operating in legacy mode, this should be set to the local kubernetes/kubernetes repo."`

	KubeconfigPath string `flag:"kubeconfig" desc:"Absolute path to existing kubeconfig for cluster"`
}

func (d *deployer) Up() error {
	klog.V(1).Info("AKS deployer starting Up()")

	env := d.buildEnv()

	defer func() {
		if err := d.DumpClusterLogs(); err != nil {
			klog.Warningf("Dumping cluster logs at the end of Up() failed: %s", err)
		}
	}()

	script := filepath.Join(home, "kubetest2", "kubetest2-aks", "scripts", "kube-up.sh")
	klog.V(2).Infof("About to run script at: %s", script)

	cmd := exec.Command(script)
	cmd.SetEnv(env...)
	exec.InheritOutput(cmd)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error encountered during %s: %s", script, err)
	}

	if isUp, err := d.IsUp(); err != nil {
		klog.Warningf("failed to check if cluster is up: %s", err)
	} else if isUp {
		klog.V(1).Infof("cluster reported as up")
	} else {
		klog.Errorf("cluster reported as down")
	}

	return nil
}

func (d *deployer) Down() error {
	klog.V(1).Info("AKS deployer starting Down()")

	env := d.buildEnv()

	script := filepath.Join(home, "kubetest2", "kubetest2-aks", "scripts", "kube-down.sh")
	klog.V(2).Infof("About to run script at: %s", script)

	cmd := exec.Command(script)
	cmd.SetEnv(env...)
	exec.InheritOutput(cmd)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error encountered during %s: %s", script, err)
	}

	return nil
}

func (d *deployer) IsUp() (up bool, err error) {
	klog.V(1).Info("AKS deployer starting IsUp()")

	env := d.buildEnv()

	if err := d.init(); err != nil {
		return false, fmt.Errorf("isUp failed to init: %s", err)
	}

	// naive assumption: nodes reported = cluster up
	// similar to other deployers' implementations
	args := []string{
		//d.kubectlPath,
		"/usr/local/bin/kubectl",
		"get",
		"nodes",
		"-o=name",
		"--kubeconfig=" + home + "/.kube/" + AksClusterName + ".yaml",
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(env...)
	cmd.SetStderr(os.Stderr)
	lines, err := exec.OutputLines(cmd)
	if err != nil {
		return false, fmt.Errorf("is up failed to get nodes: %s", err)
	}

	return len(lines) > 0, nil
}

//func (d *deployer) DumpClusterLogs() error {
//	return nil
//}

func (d *deployer) Build() error {
	// TODO: build should probably still exist with common options
	return nil
}

func (d *deployer) Kubeconfig() (string, error) {
	// noop deployer is specifically used with an existing cluster and KUBECONFIG
	if d.KubeconfigPath != "" {
		return d.KubeconfigPath, nil
	}
	if kconfig, ok := os.LookupEnv("KUBECONFIG"); ok {
		return kconfig, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".kube", "config"), nil
}

func (d *deployer) Version() string {
	return GitTag
}

// helper used to create & bind a flagset to the deployer
func bindFlags(d *deployer) *pflag.FlagSet {
	flags, err := gpflag.Parse(d)
	if err != nil {
		klog.Fatalf("unable to generate flags from deployer")
		return nil
	}

	klog.InitFlags(nil)
	flags.AddGoFlagSet(flag.CommandLine)

	return flags
}

// assert that deployer implements types.DeployerWithKubeconfig
var _ types.DeployerWithKubeconfig = &deployer{}

// RandString generates n number of random char string
func RandString(n int) (string, error) {
	randomStr, err := reggen.Generate(fmt.Sprintf("[a-z]{%d}", n), 2)

	if err != nil {
		return "", fmt.Errorf("failed to generate a random string, error: %v", err)
	}
	return randomStr, nil
}
