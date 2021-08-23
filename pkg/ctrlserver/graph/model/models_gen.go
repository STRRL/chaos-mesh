// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"

	v1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type Fd struct {
	Fd     string `json:"fd"`
	Target string `json:"target"`
}

type Namespace struct {
	Ns          string                      `json:"ns"`
	Component   []*v1.Pod                   `json:"component"`
	Pod         *v1.Pod                     `json:"pod"`
	Pods        []*v1.Pod                   `json:"pods"`
	Stress      *v1alpha1.StressChaos       `json:"stress"`
	Stresses    []*v1alpha1.StressChaos     `json:"stresses"`
	Io          *v1alpha1.IOChaos           `json:"io"`
	Ios         []*v1alpha1.IOChaos         `json:"ios"`
	Podio       *v1alpha1.PodIOChaos        `json:"podio"`
	Podios      []*v1alpha1.PodIOChaos      `json:"podios"`
	HTTP        *v1alpha1.HTTPChaos         `json:"http"`
	HTTPS       []*v1alpha1.HTTPChaos       `json:"https"`
	Podhttp     *v1alpha1.PodHttpChaos      `json:"podhttp"`
	Podhttps    []*v1alpha1.PodHttpChaos    `json:"podhttps"`
	Network     *v1alpha1.NetworkChaos      `json:"network"`
	Networks    []*v1alpha1.NetworkChaos    `json:"networks"`
	Podnetwork  *v1alpha1.PodNetworkChaos   `json:"podnetwork"`
	Podnetworks []*v1alpha1.PodNetworkChaos `json:"podnetworks"`
}

type Process struct {
	Pod     *v1.Pod `json:"pod"`
	Pid     string  `json:"pid"`
	Command string  `json:"command"`
	Fds     []*Fd   `json:"fds"`
}

type Component string

const (
	ComponentManager   Component = "MANAGER"
	ComponentDaemon    Component = "DAEMON"
	ComponentDashboard Component = "DASHBOARD"
	ComponentDNSServer Component = "DNSSERVER"
)

var AllComponent = []Component{
	ComponentManager,
	ComponentDaemon,
	ComponentDashboard,
	ComponentDNSServer,
}

func (e Component) IsValid() bool {
	switch e {
	case ComponentManager, ComponentDaemon, ComponentDashboard, ComponentDNSServer:
		return true
	}
	return false
}

func (e Component) String() string {
	return string(e)
}

func (e *Component) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Component(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Component", str)
	}
	return nil
}

func (e Component) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
