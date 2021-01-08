// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package node

import "math/rand"

type NodeNameGenerator interface {
	GenerateNodeName(templateName string) string
}

type basicNodeNameGenerator struct {
}

const pool = "abcdefghijklmnopqrstuvwxyz0123456789"
const suffixLength = 5

func NewBasicNodeNameGenerator() *basicNodeNameGenerator {
	return &basicNodeNameGenerator{}
}

func (it *basicNodeNameGenerator) GenerateNodeName(templateName string) string {
	suffix := ""
	for i := 0; i < suffixLength; i++ {
		suffix += string(pool[rand.Intn(len(pool))])
	}
	return templateName + "-" + suffix
}