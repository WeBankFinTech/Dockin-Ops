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

package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/webankfintech/dockin-opserver/internal/common"

	"github.com/stretchr/testify/assert"
	"github.com/wenzhenxi/gorsa"
)

func TestManager_Initialize(t *testing.T) {
	m := NewManager(nil)
	m.Initialize()
	assert.Equal(t, 2, len(m.ProxyIpRuleClusterMap), "has tctp and cnc")
}

func TestManager_GetProxyClient(t *testing.T) {
	m := NewManager(nil)
	m.Initialize()
	assert.Equal(t, 2, len(m.ProxyIpRuleClusterMap), "has tctp and cnc")
	tctp, err := m.GetProxyClient("127.0.0.1", "default", "tctp")
	assert.NoError(t, err)
	assert.NotNil(t, tctp)

	_, err = m.GetProxyClient("tdtp", "default", "tctp")
	assert.NoError(t, err)
}

type Foo struct {
	Input string
}

type Boo struct {
	Output string
}

func (f *Foo) ToBoo() *Boo {
	fmt.Println("haha")
	return &Boo{Output: f.Input}
}

func Test_Foo(t *testing.T) {
	foo := &Foo{Input: "lp"}
	rkind := reflect.TypeOf(foo.ToBoo).Kind()
	assert.Equal(t, rkind, reflect.Func)

	assert.Equal(t, 1, reflect.TypeOf(foo.ToBoo).NumOut())
	funVal := reflect.ValueOf(foo.ToBoo)
	rVal := funVal.Call([]reflect.Value{})
	rBoo := rVal[0].Interface().(*Boo)
	t.Log(rBoo.Output)
}

func TestInitialize(t *testing.T) {
	m := NewManager(nil)
	m.Initialize()
	for k, v := range m.ProxyIpRuleClusterMap.Items() {
		proxylist := v.([]*ProxyClient)
		for _, vv := range proxylist {
			t.Logf("[ProxyIpRuleClusterMap]: k=%s, namespace=%s, rule=%s, clusterId=%s", k, vv.K8sConfig.Contexts[0].Context.Namespace, vv.K8sConfig.Dockin.Rule, vv.K8sConfig.Dockin.ClusterID)
		}
	}

	for k, v := range m.ProxyClusterNSMap.Items() {
		vv := v.(*ProxyClient)
		t.Logf("[ProxyClusterNSMap]: k=%s, namespace=%s, rule=%s, clusterId=%s", k, vv.K8sConfig.Contexts[0].Context.Namespace, vv.K8sConfig.Dockin.Rule, vv.K8sConfig.Dockin.ClusterID)
	}

	for k, v := range m.ProxyClusterUATMap.Items() {
		vv := v.(*ProxyClient)
		t.Logf("[ProxyClusterUATMap]: k=%s, namespace=%s, rule=%s, clusterId=%s", k, vv.K8sConfig.Contexts[0].Context.Namespace, vv.K8sConfig.Dockin.Rule, vv.K8sConfig.Dockin.ClusterID)
	}
}

func Test_rs_engypt_file(t *testing.T) {
	pubKey := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0m/grbsooecW+tb/NU3U
jsFnzU1fvxy5YdZPJ9gLI8pD8UyejYT147r2ySnZOjS3uVGwPT3UT/yRiAMSn2eR
7g+lc2dLH32AB2SdVd6mTf8R18w5wUTRE7ewu9e9kAsB6phTxJH1OZgyFcLu7+4V
gcA2lfkSOuodMc/TFNW3CGWq0Yd49q9tHFDZ5XxoazL2OUqL7JivAcGz+9WI1DdV
nNJShpR91HL8XzRHVZN8LWEnN0r+nXx0JVeVRQvuWZ3ow2m+1De6onQHoOIcE0IJ
oLuvPl50hpKvN8xD5SliwvGlFSfFVtf3QFihTpqUvLQosccGpYWcMYbKCNRr5dsu
5wIDAQAB
-----END PUBLIC KEY-----`
	privKey := `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDSb+Ctuyih5xb6
1v81TdSOwWfNTV+/HLlh1k8n2AsjykPxTJ6NhPXjuvbJKdk6NLe5UbA9PdRP/JGI
AxKfZ5HuD6VzZ0sffYAHZJ1V3qZN/xHXzDnBRNETt7C7172QCwHqmFPEkfU5mDIV
wu7v7hWBwDaV+RI66h0xz9MU1bcIZarRh3j2r20cUNnlfGhrMvY5SovsmK8BwbP7
1YjUN1Wc0lKGlH3UcvxfNEdVk3wtYSc3Sv6dfHQlV5VFC+5ZnejDab7UN7qidAeg
4hwTQgmgu68+XnSGkq83zEPlKWLC8aUVJ8VW1/dAWKFOmpS8tCixxwalhZwxhsoI
1Gvl2y7nAgMBAAECggEASnI76RpKMKTBU3JWDPSA2xP+9fmGguTVjJA1pqHepwW6
bZYujWBZYPxWrCn66IWX7Z7Bm5jREI8IqTZ1EyGf1bmBTcdgIz7R2Uu2AZfn+7Xe
CRr936rJ0JDunDWhoWDTh7vl/qeoOnzmUx6ISydOQn3OkdXwphkGxQWB5mAJBZXV
sQK6+0BQVAK4vphPHzvV1pghurPWITSYwSJssGBu10AuVAKw1n8aJCG26FQWcx/K
xj+1/DLFJd1l7of8I2BdnkSdfJJA3HLSaQdwj8bGgYAfn42+vjzGjRdZEzsKRdTJ
bsshTSnza11lKFHKDf4yFsIrumwWzcqEH3sHXm+YsQKBgQDxx4DjokTWCRgY73tr
oh3slgJcZXEMVptieItFFlQ7Spa2owcvtYvyH2d2RrGnAaaOOALJIPMjh1AglqQd
g7sGIVeJawA7gVasP1qfO69oSExuxJ4x4ySgiTfXthdI2Z3cLVbR7zDgiwH66Shh
FvtlnDf2ybOal6yZHF43qkuFzwKBgQDe0HM1/E0MIoC0I1y8Q2bYH1iRXp1il/OJ
1tFnrGKCevcUB9h3l0o7Heo4VHdtUnAqzSXSnuw1/SszI5SEEfVKrq2Wy8nIQuT9
tJuqTE0Tck8nCgzGhQx2eYPNmQOIIVzkJlxcfBlXSY1d0qX4xpypA/SjteIitIDV
C9+b6IojaQKBgASdfU1bFJtNUyNuttloH9AbUPI4kX7dzFuF14q7EWKMWvIjjIiR
m5lEljIAyXVZp7dBRHRYZ6u+8n2cwoc5s4E7c7NQ0pFQN7pT/0PY3NFNx/+5SxfC
sTlLRUCd3jXqyYOhbe3V9gXjQWdrufSYfrYC1GKmmQITcRz/GKFRY92rAoGBAKAO
kZSIRzieWGIOvQEoUeqSqebTVq+KhBHSVN7qgGFGv9KNyDwwW8yXsrcARkIr5BN7
Bt6D9x7ZXH0B5B/zXodlb6FRhwPqueBeKyxsXznG9YEPwRmiXc+Fft7kOhtCDB6A
R/zP0MxZM8ngFgXddpAbHVO0xlsz2xAv1VOD+idxAoGAITodcQFB2iBm9+aYZV6v
ukYnywn/m1BH+Vc0rh0riHRRnaVfbBEG7Ho855znlIcUYVtSTlgYeYhtStAue5MY
+sjjERcTQm8NamJCjQ6QKSvOYeBSi+cqvKziUh+rFZ5GaPobL0wwKa2JayYM77R5
Qvj7Mog0l6Orx4k1HzADK2k=
-----END PRIVATE KEY-----
`
	t.Run("d, e", func(t *testing.T) {
		input := "ppppp"
		out, err := gorsa.PublicEncrypt(input, pubKey)
		assert.NoError(t, err)
		t.Log(out)
		out2, err := gorsa.PriKeyDecrypt(out, privKey)
		t.Log(string(out2))
	})

	t.Run("rsa a file", func(t *testing.T) {
		fileName := filepath.Join(common.GetConfPath(), "cluster", "kube-config-tctp.yaml")
		buf, err := ioutil.ReadFile(fileName)
		assert.NoError(t, err)
		out, err := gorsa.PublicEncrypt(string(buf), pubKey)
		assert.NoError(t, err)
		t.Log(out)
		out2, err := gorsa.PriKeyDecrypt(out, privKey)
		t.Log(string(out2))
	})

	t.Run("decrypt given", func(t *testing.T) {
		input := "rf3383DmCFlcPR9dd/GkYIKDmjl0C0UsTyamjwVSLeEA70v/u2fw6DnUJSI6N68NPXo4v18YzE/ujGwtg+lP9pspm7gLrFU2phGRE57eWiTjOACGAHoZ0nSUIijAhW+cCgD2qrAwWcSVRCMF2jrPojnz+33MMOVcztjzXHcZmCFZqNuK0wCCXIEIp+C5tcAzmrYFZPWzixVJcilzCQ0l9SZ5zW+ry01me/LMoo2e4Qcw8SSE1HV0KdpDhSxK4TnDH6p6hvZlA9xr2Dg383uzOGcTK6XWNCdagMd6hVAhJz/70kkANSs1+SB2GZp6xdP0Ip8p3EWeWSRbSOvez8vvQm+1wctAGH2UtCcp8HqKyTUeSsnhzgwGl40nq3eK+LUS6KBpEhXYi79Ckl7ZK40H+2XJdUctKXlN8UB77I0mChfTH0N83op8hkS/t5Ss0feMeTSzLDRGh1ajRpscELnpH2xuPvnmuxfbl0z14iHTDvyI7H9ArsX/S+LV4MXhKIVj/HQTVcU4clgQp4AtaQF1kKQAAcSmOgePIAP7hC+BMWRlh9iG3Su3D2utTnnwUxBiJk9ns9pduvUBzfQUWYrYDR+n/z0yD48nIRSYMyCy96rrP240WLm2EPZu70sfTYjAiai+Xl/1oWotW/IVza1S2gP9ywGv5Ly3DCWXpo+Nvxc="
		out2, err := gorsa.PriKeyDecrypt(input, privKey)
		assert.NoError(t, err)
		t.Log(string(out2))
	})

	t.Run("walk conf file and create rsa file", func(t *testing.T) {
		dir := filepath.Join(common.GetConfPath(), "env")
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(info.Name(), "yaml") {
				return nil
			}

			buf, err := ioutil.ReadFile(path)
			assert.NoError(t, err)
			out, err := gorsa.PublicEncrypt(string(buf), pubKey)
			assert.NoError(t, err)
			newFileName := path + ".perm"
			err = ioutil.WriteFile(newFileName, []byte(out), os.ModePerm)
			assert.NoError(t, err)
			return nil
		})
	})

	t.Run("engypt given yaml", func(t *testing.T) {
		path := ""
		buf, err := ioutil.ReadFile(path)
		assert.NoError(t, err)
		out, err := gorsa.PublicEncrypt(string(buf), pubKey)
		assert.NoError(t, err)
		newFileName := path + ".perm"
		err = ioutil.WriteFile(newFileName, []byte(out), os.ModePerm)
		assert.NoError(t, err)
	})
}
