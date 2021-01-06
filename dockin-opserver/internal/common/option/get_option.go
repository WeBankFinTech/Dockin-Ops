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

package option

import (
	"fmt"
	"net/url"
	"os"

	"github.com/webankfintech/dockin-opserver/internal/common/option/utils"
	"github.com/webankfintech/dockin-opserver/internal/common/printer"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
)

type PrintFlags struct {
	JSONYamlPrintFlags *genericclioptions.JSONYamlPrintFlags
	NamePrintFlags     *genericclioptions.NamePrintFlags
	TemplateFlags      *genericclioptions.KubeTemplatePrintFlags

	NoHeaders    *bool
	OutputFormat *string
}

func NewGetPrintFlags() *PrintFlags {
	outputFormat := ""
	noHeaders := false

	return &PrintFlags{
		OutputFormat: &outputFormat,
		NoHeaders:    &noHeaders,

		JSONYamlPrintFlags: genericclioptions.NewJSONYamlPrintFlags(),
		NamePrintFlags:     genericclioptions.NewNamePrintFlags(""),
		TemplateFlags:      genericclioptions.NewKubeTemplatePrintFlags(),
	}
}

type GetOption struct {
	PrintFlags             *PrintFlags
	Raw                    string
	Watch                  bool
	WatchOnly              bool
	ChunkSize              int64
	IsHumanReadablePrinter bool

	OutputWatchEvents bool

	LabelSelector     string
	FieldSelector     string
	AllNamespaces     bool
	Namespace         string
	Name              string
	Type              string
	ExplicitNamespace bool

	NoHeaders      bool
	Sort           bool
	IgnoreNotFound bool
	Export         bool

	//ToPrinter      func(*meta.RESTMapping, *bool, bool, bool) (printers.ResourcePrinterFunc, error)
}

func (option *GetOption) Complete(flags *genericclioptions.ConfigFlags, cmd *cobra.Command, args []string) error {
	if len(option.Raw) > 0 {
		if len(args) > 0 {
			return fmt.Errorf("arguments may not be passed when --raw is specified")
		}
		return nil
	}

	option.Namespace = *flags.Namespace
	if option.AllNamespaces {
		option.ExplicitNamespace = false
	}
	al := len(args)
	if al == 1 {
		option.Type = args[0]
	} else if al == 2 {
		option.Type = args[0]
		option.Name = args[1]
	}

	option.NoHeaders = utils.GetFlagBool(cmd, "no-headers")

	//outputOption := cmd.Flags().Lookup("output").Value.String()
	//if outputOption == "custom-columns" {
	//	option.ServerPrint = false
	//}
	templateArg := ""
	if option.PrintFlags.TemplateFlags != nil && option.PrintFlags.TemplateFlags.TemplateArgument != nil {
		templateArg = *option.PrintFlags.TemplateFlags.TemplateArgument
	}
	if (len(*option.PrintFlags.OutputFormat) == 0 && len(templateArg) == 0) || *option.PrintFlags.OutputFormat == "wide" {
		option.IsHumanReadablePrinter = true
	}

	return nil
}

func (option *GetOption) Validate(cmd *cobra.Command) error {
	if len(option.Raw) > 0 {
		if option.Watch || option.WatchOnly || len(option.LabelSelector) > 0 || option.Export {
			return fmt.Errorf("--raw may not be specified with other flags that filter the server request or alter the output")
		}
		if len(utils.GetFlagString(cmd, "output")) > 0 {
			return utils.UsageErrorf(cmd, "--raw and --output are mutually exclusive")
		}
		if _, err := url.ParseRequestURI(option.Raw); err != nil {
			return utils.UsageErrorf(cmd, "--raw must be a valid URL path: %v", err)
		}
	}

	return nil
}

func (option *GetOption) getUrl() string {
	return fmt.Sprintf("http://%s?command=%s", utils.CommonUrl, "get")
}

func (option *GetOption) Run() error {
	var (
		print printers.ResourcePrinter
		err   error
	)
	body, err := jsoniter.MarshalToString(option)
	if err != nil {
		fmt.Printf("json encode failed %v", err)
	}
	resp, err := utils.HttpPost(option.getUrl(), body)
	if err != nil {
		return err
	}
	if option.IsHumanReadablePrinter {
		printer.PrettyPrint(resp)
		return nil
	}
	if *option.PrintFlags.OutputFormat == "yaml" {
		print, err = option.PrintFlags.JSONYamlPrintFlags.ToPrinter("yaml")
		if err != nil {
			fmt.Printf("to yaml print failed %v\n", err)
		}
	} else if *option.PrintFlags.OutputFormat == "json" {
		print, err = option.PrintFlags.JSONYamlPrintFlags.ToPrinter("json")
		if err != nil {
			fmt.Printf("to json print failed %v\n", err)
		}
	}
	if err != nil {
		printer.PrettyPrint(resp)
		return nil
	}

	print.PrintObj(nil, os.Stdout)
	return nil
}

func (o *GetOption) RequestsHeader() (key, value string) {
	//if o.PrintWithOpenAPICols {
	//	return
	//}
	if !o.IsHumanReadablePrinter {
		return
	}

	group := metav1beta1.GroupName
	version := metav1beta1.SchemeGroupVersion.Version

	value = fmt.Sprintf("application/json;as=Table;v=%s;g=%s, application/json", version, group)
	key = "Accept"
	return

	//if o.Sort {
	//	req.Param("includeObject", "Object")
	//}
}

func (f *PrintFlags) AddFlags(cmd *cobra.Command) {
	f.JSONYamlPrintFlags.AddFlags(cmd)
	f.NamePrintFlags.AddFlags(cmd)
	f.TemplateFlags.AddFlags(cmd)
	if f.OutputFormat != nil {
		cmd.Flags().StringVarP(f.OutputFormat, "output", "o", *f.OutputFormat, "Output format. One of: json|yaml|wide|name|custom-columns=...|custom-columns-file=...|go-template=...|go-template-file=...|jsonpath=...|jsonpath-file=... See custom columns [http://kubernetes.io/docs/user-guide/kubectl-overview/#custom-columns], golang template [http://golang.org/pkg/text/template/#pkg-overview] and jsonpath template [http://kubernetes.io/docs/user-guide/jsonpath].")
	}
	if f.NoHeaders != nil {
		cmd.Flags().BoolVar(f.NoHeaders, "no-headers", *f.NoHeaders, "When using the default or custom-column output format, don't print headers (default print headers).")
	}
}
