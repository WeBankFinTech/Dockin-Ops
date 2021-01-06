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

package printer

import (
	"io"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ResourcePrinter interface {
	PrintObj(runtime.Object, io.Writer) error
}

type ResourcePrinterFunc func(runtime.Object, io.Writer) error

func (fn ResourcePrinterFunc) PrintObj(obj runtime.Object, w io.Writer) error {
	return fn(obj, w)
}

type PrintOptions struct {
	NoHeaders          bool
	WithNamespace      bool
	WithKind           bool
	Wide               bool
	ShowLabels         bool
	AbsoluteTimestamps bool
	Kind               schema.GroupKind
	ColumnLabels       []string

	SortBy string

	AllowMissingKeys bool
}
