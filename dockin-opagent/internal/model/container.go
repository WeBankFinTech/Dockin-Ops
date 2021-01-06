/*
 * Copyright (C) @2021 Webank Group Holding Limited
 * <p>
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * <p>
 * http://www.apache.org/licenses/LICENSE-2.0
 * <p>
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 */

package model

type Container struct {
	ID         string `json:"Id"`
	Names      []string
	Image      string
	ImageID    string
	Command    string
	Created    int64
	Ports      []Port
	SizeRw     int64 `json:",omitempty"`
	SizeRootFs int64 `json:",omitempty"`
	Labels     map[string]string
	State      string
	Status     string
	HostConfig struct {
		NetworkMode string `json:",omitempty"`
	}
	NetworkSettings *SummaryNetworkSettings
	Mounts          []MountPoint
}

type Port struct {

	// IP
	IP string `json:"IP,omitempty"`

	// Port on the container
	// Required: true
	PrivatePort uint16 `json:"PrivatePort"`

	// Port exposed on the host
	PublicPort uint16 `json:"PublicPort,omitempty"`

	// type
	// Required: true
	Type string `json:"Type"`
}

type MountPoint struct {
	//Type        mount.Type `json:",omitempty"`
	Name        string `json:",omitempty"`
	Source      string
	Destination string
	Driver      string `json:",omitempty"`
	Mode        string
	RW          bool
	//Propagation mount.Propagation
}

type SummaryNetworkSettings struct {
	Networks map[string]*EndpointSettings
}

type EndpointSettings struct {
	// Configurations
	IPAMConfig *EndpointIPAMConfig
	Links      []string
	Aliases    []string
	// Operational data
	NetworkID           string
	EndpointID          string
	Gateway             string
	IPAddress           string
	IPPrefixLen         int
	IPv6Gateway         string
	GlobalIPv6Address   string
	GlobalIPv6PrefixLen int
	MacAddress          string
}

type EndpointIPAMConfig struct {
	IPv4Address  string   `json:",omitempty"`
	IPv6Address  string   `json:",omitempty"`
	LinkLocalIPs []string `json:",omitempty"`
}
