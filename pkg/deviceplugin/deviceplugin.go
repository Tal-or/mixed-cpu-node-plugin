/*
 * Copyright 2023 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package deviceplugin

import (
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"

	"github.com/containerd/nri/pkg/api"
	"github.com/containers/podman/v4/pkg/env"
	"github.com/golang/glog"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
)

const (
	SharedCPUResourceNamespace = "openshift.io"
	SharedCPUResourceName      = "sharedcpu"
	SharedCPUDeviceName        = SharedCPUResourceNamespace + "/" + SharedCPUResourceName
	EnvVarName                 = "OPENSHIFT_SHARED_CPUS"
)

type SharedCpu struct {
	cpus cpuset.CPUSet
}

func (mc *SharedCpu) GetResourceNamespace() string {
	return SharedCPUResourceNamespace
}

func (mc *SharedCpu) Discover(pnl chan dpm.PluginNameList) {
	pnl <- []string{SharedCPUResourceName}
}

func (mc *SharedCpu) NewPlugin(s string) dpm.PluginInterface {
	return pluginImp{
		sharedCpus: &mc.cpus,
		update:     make(chan message),
	}
}

func New(cpus string) (*dpm.Manager, error) {
	sharedCpus, err := cpuset.Parse(cpus)
	if err != nil {
		return nil, err
	}
	mc := &SharedCpu{cpus: sharedCpus}
	return dpm.NewManager(mc), nil
}

// Requested checks whether a given container is requesting the device
func Requested(ctr *api.Container) bool {
	if ctr.Env == nil {
		return false
	}

	envs, err := env.ParseSlice(ctr.Env)
	if err != nil {
		glog.Errorf("failed to parse environment variables for container: %q; err: %v", ctr.Name, err)
		return false
	}

	for k, v := range envs {
		if k == EnvVarName {
			glog.V(4).Infof("shared CPUs ids: %q allocated for container: %q", v, ctr.Name)
			return true
		}
	}
	return false
}
