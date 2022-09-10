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
)


func (d *deployer) init() error {
	return nil
}


func (d *deployer) buildEnv() []string {
	// The base env currently does not inherit the current os env (except for PATH)
	// because (for now) it doesn't have to. In future, this may have to change when
	// support is added for k/k's kube-up.sh and kube-down.sh which support a wide
	// variety of environment variables. Before doing so, it is worth investigating
	// inheriting the os env vs. adding flags to this deployer on a case-by-case
	// basis to support individual environment configurations.
	var env []string

	// path is necessary for scripts to find az, etc
	// can be removed if env is inherited from the os
	env = append(env, fmt.Sprintf("PATH=%s", os.Getenv("PATH")))

	// kube-up.sh, kube-down.sh scipts uses
	// the following environment variables as a parameter
	// for az commands
	env = append(env, fmt.Sprintf("AZ_LOCATION=%s", "westus2"))
	env = append(env, fmt.Sprintf("HOME=%s", home))
        env = append(env, fmt.Sprintf("AZ_RESOURCE_GROUP=%s", AksResourceGroup))
        env = append(env, fmt.Sprintf("AZ_CLUSTER_NAME=%s", AksClusterName))
        env = append(env, fmt.Sprintf("KUBECONFIG=%s", home + "/.kube/" + AksClusterName +     ".yaml"))

	return env
}

