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
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/liggitt/tabwriter"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/apimachinery/pkg/watch"
)

var _ ResourcePrinter = &HumanReadablePrinter{}

var (
	statusHandlerEntry = &handlerEntry{
		columnDefinitions: statusColumnDefinitions,
		printFunc:         reflect.ValueOf(printStatus),
	}

	statusColumnDefinitions = []metav1beta1.TableColumnDefinition{
		{Name: "Status", Type: "string"},
		{Name: "Reason", Type: "string"},
		{Name: "Message", Type: "string"},
	}

	defaultHandlerEntry = &handlerEntry{
		columnDefinitions: objectMetaColumnDefinitions,
		printFunc:         reflect.ValueOf(printObjectMeta),
	}

	objectMetaColumnDefinitions = []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Age", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
	}

	withEventTypePrefixColumns = []string{"EVENT"}
	withNamespacePrefixColumns = []string{"NAMESPACE"}
)

func NewTablePrinter(options PrintOptions) *HumanReadablePrinter {
	printer := &HumanReadablePrinter{
		handlerMap: make(map[reflect.Type]*handlerEntry),
		options:    options,
	}
	return printer
}

func printHeader(columnNames []string, w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t")); err != nil {
		return err
	}
	return nil
}

func (h *HumanReadablePrinter) PrintObj(obj runtime.Object, output io.Writer) error {
	w, found := output.(*tabwriter.Writer)
	if !found {
		w = GetNewTabWriter(output)
		output = w
		defer w.Flush()
	}

	var eventType string
	if event, isEvent := obj.(*metav1.WatchEvent); isEvent {
		eventType = event.Type
		obj = event.Object.Object
	}

	if table, ok := obj.(*metav1beta1.Table); ok {
		localOptions := h.options
		if h.printedHeaders && (len(table.ColumnDefinitions) == 0 || reflect.DeepEqual(table.ColumnDefinitions, h.lastColumns)) {
			localOptions.NoHeaders = true
		}

		if len(table.ColumnDefinitions) == 0 {
			table.ColumnDefinitions = h.lastColumns
		} else if !reflect.DeepEqual(table.ColumnDefinitions, h.lastColumns) {
			h.lastColumns = table.ColumnDefinitions
			h.printedHeaders = false
		}

		if len(table.Rows) > 0 {
			h.printedHeaders = true
		}

		if err := decorateTable(table, localOptions); err != nil {
			return err
		}
		if len(eventType) > 0 {
			if err := addColumns(beginning, table,
				[]metav1beta1.TableColumnDefinition{{Name: "Event", Type: "string"}},
				[]cellValueFunc{func(metav1beta1.TableRow) (interface{}, error) { return formatEventType(eventType), nil }},
			); err != nil {
				return err
			}
		}
		return printTable(table, output, localOptions)
	}

	t := reflect.TypeOf(obj)
	if handler := h.handlerMap[t]; handler != nil {
		includeHeaders := h.lastType != t && !h.options.NoHeaders

		if h.lastType != nil && h.lastType != t && !h.options.NoHeaders {
			fmt.Fprintln(output)
		}

		if err := printRowsForHandlerEntry(output, handler, eventType, obj, h.options, includeHeaders); err != nil {
			return err
		}
		h.lastType = t
		return nil
	}

	var handler *handlerEntry
	if _, isStatus := obj.(*metav1.Status); isStatus {
		handler = statusHandlerEntry
	} else {
		handler = defaultHandlerEntry
	}

	includeHeaders := h.lastType != handler && !h.options.NoHeaders

	if h.lastType != nil && h.lastType != handler && !h.options.NoHeaders {
		fmt.Fprintln(output)
	}

	if err := printRowsForHandlerEntry(output, handler, eventType, obj, h.options, includeHeaders); err != nil {
		return err
	}
	h.lastType = handler

	return nil
}

func printTable(table *metav1beta1.Table, output io.Writer, options PrintOptions) error {
	if !options.NoHeaders {
		if len(table.Rows) == 0 {
			return nil
		}

		first := true
		for _, column := range table.ColumnDefinitions {
			if !options.Wide && column.Priority != 0 {
				continue
			}
			if first {
				first = false
			} else {
				fmt.Fprint(output, "\t")
			}
			fmt.Fprint(output, strings.ToUpper(column.Name))
		}
		fmt.Fprintln(output)
	}
	for _, row := range table.Rows {
		first := true
		for i, cell := range row.Cells {
			if i >= len(table.ColumnDefinitions) {
				break
			}
			column := table.ColumnDefinitions[i]
			if !options.Wide && column.Priority != 0 {
				continue
			}
			if first {
				first = false
			} else {
				fmt.Fprint(output, "\t")
			}
			if cell != nil {
				fmt.Fprint(output, cell)
			}
		}
		fmt.Fprintln(output)
	}
	return nil
}

type cellValueFunc func(metav1beta1.TableRow) (interface{}, error)

type columnAddPosition int

const (
	beginning columnAddPosition = 1
	end       columnAddPosition = 2
)

func addColumns(pos columnAddPosition, table *metav1beta1.Table, columns []metav1beta1.TableColumnDefinition, valueFuncs []cellValueFunc) error {
	if len(columns) != len(valueFuncs) {
		return fmt.Errorf("cannot prepend columns, unmatched value functions")
	}
	if len(columns) == 0 {
		return nil
	}

	newRows := make([][]interface{}, len(table.Rows))
	for i := range table.Rows {
		newCells := make([]interface{}, 0, len(columns)+len(table.Rows[i].Cells))

		if pos == end {
			newCells = append(newCells, table.Rows[i].Cells...)
			for len(newCells) < len(table.ColumnDefinitions) {
				newCells = append(newCells, nil)
			}
		}

		for _, f := range valueFuncs {
			newCell, err := f(table.Rows[i])
			if err != nil {
				return err
			}
			newCells = append(newCells, newCell)
		}

		if pos == beginning {
			newCells = append(newCells, table.Rows[i].Cells...)
		}

		newRows[i] = newCells
	}

	newColumns := make([]metav1beta1.TableColumnDefinition, 0, len(columns)+len(table.ColumnDefinitions))
	switch pos {
	case beginning:
		newColumns = append(newColumns, columns...)
		newColumns = append(newColumns, table.ColumnDefinitions...)
	case end:
		newColumns = append(newColumns, table.ColumnDefinitions...)
		newColumns = append(newColumns, columns...)
	default:
		return fmt.Errorf("invalid column add position: %v", pos)
	}
	table.ColumnDefinitions = newColumns
	for i := range table.Rows {
		table.Rows[i].Cells = newRows[i]
	}

	return nil
}

func decorateTable(table *metav1beta1.Table, options PrintOptions) error {
	width := len(table.ColumnDefinitions) + len(options.ColumnLabels)
	if options.WithNamespace {
		width++
	}
	if options.ShowLabels {
		width++
	}

	columns := table.ColumnDefinitions

	nameColumn := -1
	if options.WithKind && !options.Kind.Empty() {
		for i := range columns {
			if columns[i].Format == "name" && columns[i].Type == "string" {
				nameColumn = i
				break
			}
		}
	}

	if width != len(table.ColumnDefinitions) {
		columns = make([]metav1beta1.TableColumnDefinition, 0, width)
		if options.WithNamespace {
			columns = append(columns, metav1beta1.TableColumnDefinition{
				Name: "Namespace",
				Type: "string",
			})
		}
		columns = append(columns, table.ColumnDefinitions...)
		for _, label := range formatLabelHeaders(options.ColumnLabels) {
			columns = append(columns, metav1beta1.TableColumnDefinition{
				Name: label,
				Type: "string",
			})
		}
		if options.ShowLabels {
			columns = append(columns, metav1beta1.TableColumnDefinition{
				Name: "Labels",
				Type: "string",
			})
		}
	}

	rows := table.Rows

	includeLabels := len(options.ColumnLabels) > 0 || options.ShowLabels
	if includeLabels || options.WithNamespace || nameColumn != -1 {
		for i := range rows {
			row := rows[i]

			if nameColumn != -1 {
				row.Cells[nameColumn] = fmt.Sprintf("%s/%s", strings.ToLower(options.Kind.String()), row.Cells[nameColumn])
			}

			var m metav1.Object
			if obj := row.Object.Object; obj != nil {
				if acc, err := meta.Accessor(obj); err == nil {
					m = acc
				}
			}
			if m == nil {
				if options.WithNamespace {
					r := make([]interface{}, 1, width)
					row.Cells = append(r, row.Cells...)
				}
				for j := 0; j < width-len(row.Cells); j++ {
					row.Cells = append(row.Cells, nil)
				}
				rows[i] = row
				continue
			}

			if options.WithNamespace {
				r := make([]interface{}, 1, width)
				r[0] = m.GetNamespace()
				row.Cells = append(r, row.Cells...)
			}
			if includeLabels {
				row.Cells = appendLabelCells(row.Cells, m.GetLabels(), options)
			}
			rows[i] = row
		}
	}

	table.ColumnDefinitions = columns
	table.Rows = rows
	return nil
}

func printRowsForHandlerEntry(output io.Writer, handler *handlerEntry, eventType string, obj runtime.Object, options PrintOptions, includeHeaders bool) error {
	var results []reflect.Value

	args := []reflect.Value{reflect.ValueOf(obj), reflect.ValueOf(options)}
	results = handler.printFunc.Call(args)
	if !results[1].IsNil() {
		return results[1].Interface().(error)
	}

	if includeHeaders {
		var headers []string
		for _, column := range handler.columnDefinitions {
			if column.Priority != 0 && !options.Wide {
				continue
			}
			headers = append(headers, strings.ToUpper(column.Name))
		}
		headers = append(headers, formatLabelHeaders(options.ColumnLabels)...)
		headers = append(headers, formatShowLabelsHeader(options.ShowLabels)...)
		if options.WithNamespace {
			headers = append(withNamespacePrefixColumns, headers...)
		}
		if len(eventType) > 0 {
			headers = append(withEventTypePrefixColumns, headers...)
		}
		printHeader(headers, output)
	}

	if results[1].IsNil() {
		rows := results[0].Interface().([]metav1beta1.TableRow)
		printRows(output, eventType, rows, options)
		return nil
	}
	return results[1].Interface().(error)
}

var formattedEventType = map[string]string{
	string(watch.Added):    "ADDED   ",
	string(watch.Modified): "MODIFIED",
	string(watch.Deleted):  "DELETED ",
	string(watch.Error):    "ERROR   ",
}

func formatEventType(eventType string) string {
	if formatted, ok := formattedEventType[eventType]; ok {
		return formatted
	}
	return string(eventType)
}

func printRows(output io.Writer, eventType string, rows []metav1beta1.TableRow, options PrintOptions) {
	for _, row := range rows {
		if len(eventType) > 0 {
			fmt.Fprint(output, formatEventType(eventType))
			fmt.Fprint(output, "\t")
		}
		if options.WithNamespace {
			if obj := row.Object.Object; obj != nil {
				if m, err := meta.Accessor(obj); err == nil {
					fmt.Fprint(output, m.GetNamespace())
				}
			}
			fmt.Fprint(output, "\t")
		}

		for i, cell := range row.Cells {
			if i != 0 {
				fmt.Fprint(output, "\t")
			} else {
				if options.WithKind && !options.Kind.Empty() {
					fmt.Fprintf(output, "%s/%s", strings.ToLower(options.Kind.String()), cell)
					continue
				}
			}
			fmt.Fprint(output, cell)
		}

		hasLabels := len(options.ColumnLabels) > 0
		if obj := row.Object.Object; obj != nil && (hasLabels || options.ShowLabels) {
			if m, err := meta.Accessor(obj); err == nil {
				for _, value := range labelValues(m.GetLabels(), options) {
					output.Write([]byte("\t"))
					output.Write([]byte(value))
				}
			}
		}

		output.Write([]byte("\n"))
	}
}

func formatLabelHeaders(columnLabels []string) []string {
	formHead := make([]string, len(columnLabels))
	for i, l := range columnLabels {
		p := strings.Split(l, "/")
		formHead[i] = strings.ToUpper((p[len(p)-1]))
	}
	return formHead
}

func formatShowLabelsHeader(showLabels bool) []string {
	if showLabels {
		return []string{"LABELS"}
	}
	return nil
}

func labelValues(itemLabels map[string]string, opts PrintOptions) []string {
	var values []string
	for _, key := range opts.ColumnLabels {
		values = append(values, itemLabels[key])
	}
	if opts.ShowLabels {
		values = append(values, labels.FormatLabels(itemLabels))
	}
	return values
}

func appendLabelCells(values []interface{}, itemLabels map[string]string, opts PrintOptions) []interface{} {
	for _, key := range opts.ColumnLabels {
		values = append(values, itemLabels[key])
	}
	if opts.ShowLabels {
		values = append(values, labels.FormatLabels(itemLabels))
	}
	return values
}

func printStatus(obj runtime.Object, options PrintOptions) ([]metav1beta1.TableRow, error) {
	status, ok := obj.(*metav1.Status)
	if !ok {
		return nil, fmt.Errorf("expected *v1.Status, got %T", obj)
	}
	return []metav1beta1.TableRow{{
		Object: runtime.RawExtension{Object: obj},
		Cells:  []interface{}{status.Status, status.Reason, status.Message},
	}}, nil
}

func printObjectMeta(obj runtime.Object, options PrintOptions) ([]metav1beta1.TableRow, error) {
	if meta.IsListType(obj) {
		rows := make([]metav1beta1.TableRow, 0, 16)
		err := meta.EachListItem(obj, func(obj runtime.Object) error {
			nestedRows, err := printObjectMeta(obj, options)
			if err != nil {
				return err
			}
			rows = append(rows, nestedRows...)
			return nil
		})
		if err != nil {
			return nil, err
		}
		return rows, nil
	}

	rows := make([]metav1beta1.TableRow, 0, 1)
	m, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: obj},
	}
	row.Cells = append(row.Cells, m.GetName(), translateTimestampSince(m.GetCreationTimestamp()))
	rows = append(rows, row)
	return rows, nil
}

func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}
