// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package timechaos

import (
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/action"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/timechaos/timeskew"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/timechaos/timestop"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
)

type Impl struct {
	fx.In

	TimeSkew *timeskew.Impl `action:"time-skew"`
	TimeStop *timestop.Impl `action:"time-stop"`
}

func NewImpl(impl Impl) *common.ChaosImplPair {
	delegate := action.New(&impl)
	return &common.ChaosImplPair{
		Name:   "timechaos",
		Object: &v1alpha1.TimeChaos{},
		Impl:   &delegate,
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
	timeskew.NewImpl,
	timestop.NewImpl,
)
