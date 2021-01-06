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
	"io/ioutil"

	"github.com/webankfintech/dockin-opsctl/internal/utils/aes"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/webankfintech/dockin-opsctl/internal/common"
	"github.com/webankfintech/dockin-opsctl/internal/common/protocol"
	"github.com/webankfintech/dockin-opsctl/internal/log"
	"github.com/webankfintech/dockin-opsctl/internal/utils"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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
	Command string
	Type    string
	Name    string

	Rule          string
	Namespace     string
	AllNamespaces bool
	OutputFormat  string
}

func (option *GetOption) Complete(flags *genericclioptions.ConfigFlags, cmd *cobra.Command, args []string) error {
	l := len(args)
	if l < 1 {
		return errors.Errorf("%s\n%s", NoTypeOrNameErr, GetCommandSuggest)
	} else if l == 1 {
		option.Type = args[0]
	} else {
		option.Type = args[0]
		option.Name = args[1]
	}

	log.Debugf("cmdLine params:%s", args)
	option.Namespace, _ = cmd.Flags().GetString("namespace")
	option.Rule, _ = cmd.Flags().GetString("rule")

	return nil
}

func (option *GetOption) Run() error {
	proto := protocol.NewProto()
	proto.Command = option.Command
	proto.Resource = option.Type
	proto.Name = option.Name
	if option.OutputFormat != "" {
		proto.PrintType = option.OutputFormat
	}

	if option.Namespace != "" {
		proto.Params["namespace"] = option.Namespace
	}

	if option.Rule != "" {
		proto.Params["rule"] = option.Rule
	}

	if option.AllNamespaces == true {
		proto.Params["all-namespaces"] = true
	}

	data, err := jsoniter.MarshalToString(proto)
	if err != nil {
		log.Debugf("json marshal %s err:%s", proto.String(), err.Error())
		return err
	}

	aes, err := aes.NewAes(common.ResKey)
	if err != nil {
		nerr := errors.Errorf("encrypt failed err=%s", err.Error())
		log.Debugf(nerr.Error())
		return nerr
	}

	encode, err := aes.AesEncrypt(data)
	if err != nil {
		nerr := errors.Errorf("decrypt err=%s", err.Error())
		log.Debugf(nerr.Error())
		return nerr
	}

	log.Debugf(data)
	params := make(map[string]string)
	params["params"] = encode

	resp, err := utils.HttpGet(common.GetCommonUrlByCmd("echo"), params, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)

	fmt.Println(string(b))
	return nil
}

func (o *GetOption) RequestsHeader() (key, value string) {

	group := metav1beta1.GroupName
	version := metav1beta1.SchemeGroupVersion.Version

	value = fmt.Sprintf("application/json;as=Table;v=%s;g=%s, application/json", version, group)
	key = "Accept"
	return

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
