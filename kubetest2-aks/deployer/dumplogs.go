/*
Copyright 2020 The Kubernetes Authors.

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

package deployer

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"

	"sigs.k8s.io/kubetest2/pkg/exec"
)

func (d *deployer) DumpClusterLogs() error {
	klog.V(1).Info("AKS deployer starting DumpClusterLogs()")


	klog.V(2).Info("making logs directory")
	if err := d.makeLogsDir(); err != nil {
		return fmt.Errorf("couldn't make logs dir: %s", err)
	}

	if err := d.kubectlDump(); err != nil {
		return fmt.Errorf("failed to dump cluster info with kubectl: %s", err)
	}

	return nil
}

func (d *deployer) makeLogsDir() error {
	_, err := os.Stat(d.logsDir)

	if os.IsNotExist(err) {
		err := os.Mkdir(d.logsDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create %s: %s", d.logsDir, err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("unexpected exception when making cluster logs directory: %s", err)
	}

	// file definitely exists, overwrite if requested

	if d.OverwriteLogsDir {
		klog.V(2).Infof("logs directory %s already exists, removing and recreating", d.logsDir)

		if err := os.RemoveAll(d.logsDir); err != nil {
			return fmt.Errorf("failed to delete existing logs directory: %s", err)
		}

		err := os.Mkdir(d.logsDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create %s: %s", d.logsDir, err)
		}
		return nil
	}

	return fmt.Errorf("cluster logs directory %s already exists, please clean up manually or use the overwrite flag before continuing", d.logsDir)
}


func (d *deployer) kubectlDump() error {
	env := d.buildEnv()
	outfile, err := os.Create(filepath.Join(d.logsDir, "cluster-info.log"))
	if err != nil {
		return fmt.Errorf("failed to create cluster-info log file: %s", err)
	}
	defer outfile.Close()

	args := []string{
		//d.kubectlPath,
		"/usr/local/bin/kubectl",
		"cluster-info",
		"dump",
		"--kubeconfig="+home+"/.kube/"+AksClusterName+".yaml",
	}
	klog.V(2).Infof("About to run: %s", args)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(env...)
	cmd.SetStderr(os.Stderr)
	cmd.SetStdout(outfile)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("couldn't use kubectl to dump cluster info: %s", err)
	}

	return nil
}

