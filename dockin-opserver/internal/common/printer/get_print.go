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
	"bytes"
	"fmt"
	"text/tabwriter"

	utils2 "github.com/webankfintech/dockin-opserver/internal/common/option/utils"

	jsoniter "github.com/json-iterator/go"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
)

func PrintTable(table *metav1beta1.Table) string {
	buf := &bytes.Buffer{}
	tw := tabwriter.NewWriter(buf, 5, 8, 1, ' ', 0)
	printer := NewTablePrinter(PrintOptions{Wide:true})
	err := printer.PrintObj(table, tw)
	utils2.CheckErr(err)
	tw.Flush()
	return buf.String()
}


func PrettyPrint(bytes []byte) {
	table := &metav1beta1.Table{}
	if err := jsoniter.Unmarshal(bytes, table); err != nil {
		fmt.Printf("unmarshal result failed, %v\n", err)
	}
	fmt.Print(PrintTable(table))
}
