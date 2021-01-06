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

package ctrl

var indexTemplate = `
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>dockin-OPSERVER控制管理台</title>
</head>
<body>

<h1 label="标题居中" style="font-size: 32px; font-weight: bold; padding: 0px 4px 0px 0px; text-align: left; margin: 0px 0px 20px;" dir="rtl">
    DOCKIN-OPSERVER控制管理台<br/>
</h1>
<hr width="60%" align="left"/>


<script language="javascript">
	
	function addRawCmd(){
		var cmdName = document.getElementsByName("cmdRawNameAdd")[0].value
		if (cmdName == "") {
			alert("输入的命令为空")
			return false
		}
		if (!confirm("确认新增交互式命令:" + cmdName)) {
			return false
		}
		var httpClient = new XMLHttpRequest();
		httpClient.open("post", "/v1/dockin/opserver/ctrl/addRawCmd?cmdName=" + cmdName);
		httpClient.send();
		httpClient.onreadystatechange = function() {
			if (this.readyState == 4) {
				if (this.status != 200) {
					alert("http code = " + this.status);
					return false;
				} else {
					alert(this.responseText);
					location.reload();
				}
			}
		}
	}

	function deleteRawCmd(){
		var rawCheckBox = document.getElementsByName("cmdRawNameDelete")
		var cmdList = new Array()
		for (let i = 0; i < rawCheckBox.length; i++) {
			if (rawCheckBox[i].checked) {
				cmdList.push(rawCheckBox[i].value)
			}
		}
		if (cmdList == 0) {
			alert("没有勾选要被删除的命令")
			return false;
		}
		if (!confirm("确认删除交互式命令:" + cmdList.join(","))) {
			return false
		}
		var httpClient = new XMLHttpRequest();
		httpClient.open("post", "/v1/dockin/opserver/ctrl/deleteRawCmd?cmdName=" + cmdList.join(","));
		httpClient.send();
		httpClient.onreadystatechange = function() {
			if (this.readyState == 4) {
				if (this.status != 200) {
					alert("http code = " + this.status);
					return false;
				} else {
					alert(this.responseText);
					location.reload();
				}
			}
		}
	}
	function addCmd(){ 
		var cmdName = document.getElementsByName("cmdNameAdd")[0].value
		if (cmdName == "") {
			alert("输入的命令为空")
			return false
		}
		if (!confirm("确认增加普通命令:" + cmdName)) {
			return false
		}
		var httpClient = new XMLHttpRequest();
		httpClient.open("post", "/v1/dockin/opserver/ctrl/addCmd?cmdName=" + cmdName);
		httpClient.send();
		httpClient.onreadystatechange = function() {
			if (this.readyState == 4) {
				if (this.status != 200) {
					alert("http code = " + this.status);
					return false;
				} else {
					alert(this.responseText);
					location.reload();
				}
			}
		}
	}
	function deleteCmd(){
		var comCheckBox = document.getElementsByName("cmdNameDelete")
		var cmdList = new Array()
		for (let i = 0; i < comCheckBox.length; i++) {
			if (comCheckBox[i].checked) {
				cmdList.push(comCheckBox[i].value)
			}
		}
		if (cmdList == 0) {
			alert("没有勾选要被删除的命令")
			return false;
		}
		if (!confirm("确认删除普通命令:" + cmdList.join(","))) {
			return false
		}
		var httpClient = new XMLHttpRequest();
		httpClient.open("post", "/v1/dockin/opserver/ctrl/deleteCmd?cmdName=" + cmdList.join(","));
		httpClient.send();
		httpClient.onreadystatechange = function() {
			if (this.readyState == 4) {
				if (this.status != 200) {
					alert("http code = " + this.status);
					return false;
				} else {
					alert(this.responseText);
					location.reload();
				}
			}
		}
	}
	function deleteIP(btn) {
		var rule = btn.getAttribute("rule")
		var deleteIPListCheckBox = document.getElementsByName("deleteIPList_" + rule)
		var ipList = new Array()
		for (let i = 0; i < deleteIPListCheckBox.length; i++) {
			if (deleteIPListCheckBox[i].checked) {
				ipList.push(deleteIPListCheckBox[i].value)
			}
		}
		if (ipList == 0) {
			alert("没有勾选要被删除的白名单")
			return false;
		}
		if (!confirm("确认删除白名单:" + ipList.join(","))) {
			return false
		}
		var httpClient = new XMLHttpRequest();
		httpClient.open("post", "/v1/dockin/opserver/ctrl/deleteIP?rule=" + rule + "&ip=" + ipList.join(","));
		httpClient.send();
		httpClient.onreadystatechange = function() {
			if (this.readyState == 4) {
				if (this.status != 200) {
					alert("http code = " + this.status);
					return false;
				} else {
					alert(this.responseText);
					location.reload();
				}
			}
		}
	}
	function addWhitelist() {
		var rule = document.getElementsByName("addWLRuleText")[0].value
		var whiteIp = document.getElementsByName("addWLIPText")[0].value
		if (whiteIp == "") {
			alert("输入的ip为空")
			return false
		}
		if (rule == "") {
			alert("输入的rule为空")
			return false
		}
		if (!confirm("确认新增白名单:rule=" + rule + "ip=" + whiteIp)) {
			return false
		}
		var httpClient = new XMLHttpRequest();
		httpClient.open("post", "/v1/dockin/opserver/ctrl/addWhitelist?ip=" + whiteIp + "&rule=" + rule);
		httpClient.send();
		httpClient.onreadystatechange = function() {
			if (this.readyState == 4) {
				if (this.status != 200) {
					alert("http code = " + this.status);
					return false;
				} else {
					alert(this.responseText);
					location.reload();
				}
			}
		}
	}
	function deleteAccount(btn) {
		var userName = btn.getAttribute("userName")
		var httpClient = new XMLHttpRequest();
		if (!confirm("确认删除账号:" + userName)) {
			return false
		}
		httpClient.open("post", "/v1/dockin/opserver/ctrl/deleteAccount?userName=" + userName);
		httpClient.send();
		httpClient.onreadystatechange = function() {
			if (this.readyState == 4) {
				if (this.status != 200) {
					alert("http code = " + this.status);
					return false;
				} else {
					alert(this.responseText);
					location.reload();
				}
			}
		}
	}
	function addAccount() {
		var userName = document.getElementsByName("userNameAdd")[0].value
		if (userName=="") {
			alert("用户名为空")
			return false
		}
		if (!confirm("确认新增账号:用户名=" + userName)) {
			return false
		}
		var httpClient = new XMLHttpRequest();
		httpClient.open("post", "/v1/dockin/opserver/ctrl/addAccount?userName=" + userName);
		httpClient.send();
		httpClient.onreadystatechange = function() {
			if (this.readyState == 4) {
				if (this.status != 200) {
					alert("http code = " + this.status);
					return false;
				} else {
					alert(this.responseText);
					location.reload();
				}
			}
		}
	}
	function updateVersion(){
		var version = document.getElementsByName("versionUpdate")[0].value
		if (version=="") {
			alert("版本为空")
			return false
		}
		if (!confirm("确认新版本:=" + version)) {
			return false
		}
		var httpClient = new XMLHttpRequest();
		httpClient.open("post", "/v1/dockin/opserver/ctrl/updateVersion?version=" + version);
		httpClient.send();
		httpClient.onreadystatechange = function() {
			if (this.readyState == 4) {
				if (this.status != 200) {
					alert("http code = " + this.status);
					return false;
				} else {
					alert(this.responseText);
					location.reload();
				}
			}
		}
	}
function deleteAccount(btn) {
		var userName = btn.getAttribute("userName")
		var httpClient = new XMLHttpRequest();
		if (!confirm("确认删除账号:" + userName)) {
			return false
		}
		httpClient.open("post", "/v1/dockin/opserver/ctrl/deleteAccount?userName=" + userName);
		httpClient.send();
		httpClient.onreadystatechange = function() {
			if (this.readyState == 4) {
				if (this.status != 200) {
					alert("http code = " + this.status);
					return false;
				} else {
					alert(this.responseText);
					location.reload();
				}
			}
		}
	}
</script>

<h2>
    一、远程执行命令白名单
</h2>
<h3>
    1.&nbsp; 交互式命令<br/>
</h3>

{{range .RawCommand}}
<input type="checkbox" name="cmdRawNameDelete" value="{{.Cmd}}">{{.Cmd}}</input>
{{end}}
</br>
<button type="button" onclick="deleteRawCmd();">删除勾选</button>
</br></br>
<input type="text" name="cmdRawNameAdd" placeholder="输入命令，逗号分隔"/>
<input type="submit" value="新增" onclick="addRawCmd();"/>
</br>
<hr width="60%" align="left"/>
<h3>
    2.&nbsp; 普通命令<br/>
</h3>

{{range .Command}}
<input type="checkbox" name="cmdNameDelete" value="{{.Cmd}}">{{.Cmd}}</input>
{{end}}
</br>
<button type="button" onclick="deleteCmd();">删除勾选</button>
</br></br>
<input type="text" name="cmdNameAdd" placeholder="输入命令，逗号分隔"/>
<input type="submit" value="新增" onclick="addCmd();"/>

<hr width="60%" align="left"/>
<h2>
    二、白名单管理
</h2>
<table border="2px" cellspacing="0px" style="border-collapse:collapse">
    <tbody border="1px" cellspacing="0px" style="border-collapse:collapse">
        <tr class="firstRow">
            <td width="60" valign="top">
                Rule
            </td>
            <td width="500" valign="top">
                ip列表
            </td>
            <td width="200" valign="top">
                操作<br/>
            </td>
        </tr>

		{{range .WhiteList}}
		<tr>
			<td width="60" valign="middle">
				<label>{{.Rule}}</label>
		</td>
			{{$ruleTmp := .Rule}}
			<td width="500" valign="middle">
				{{range .IPList}}
					<input type="checkbox" name="deleteIPList_{{$ruleTmp}}" value="{{.Addr}}">{{.Addr}}</input>&nbsp;&nbsp;
				{{end}}
			</td>
			<td width="200" valign="middle">
				<button type="button" name="deleteWLBtn" rule="{{$ruleTmp}}" onclick="deleteIP(this);">删除勾选</button>
			</td>
		</tr>
		{{end}}
		<tr class>
            <td width="60" valign="top">
                <input type="text" name="addWLRuleText" placeholder="输入rule..."/>
            </td>
            <td width="500" valign="top">
                <input type="text" name="addWLIPText" placeholder="输入IP..."/>
            </td>
            <td width="200" valign="top">
                <button type="button" onclick="addWhitelist();">新增IP白名单</button>
            </td>
        </tr>
    </tbody>
</table>
<hr width="60%" align="left"/>
<h2>
    三、账号管理
</h2>
<table border="2px" cellspacing="0px" style="border-collapse:collapse">
    <tbody border="1px" cellspacing="0px" style="border-collapse:collapse">
        <tr class="firstRow">
            <td width="100" valign="top">
                用户名
            </td>
			<td width="100" valign="top">
                操作
            </td>
        </tr>

		{{range .Account}}
		<tr>
			<td width="100" valign="middle">
				<label>{{.UserName}}</label>
			</td>
			<td width="100" valign="middle">
				<button type="button" name="deleteAccountBtn" userName="{{.UserName}}" onclick="deleteAccount(this);">删除勾选</button>
			</td>
		</tr>
		{{end}}
		<tr>
			<td width="100" valign="middle">
				<input type="text" name="userNameAdd" placeholder="输入用户名"/>
			</td>
			<td width="100" valign="middle">
				<button type="button" name="addAccountBtn" onclick="addAccount();">新增账户</button>
			</td>
		</tr>
    </tbody>
</table>
<h2>
    四、版本管理
</h2>
<table border="2px" cellspacing="0px" style="border-collapse:collapse">
    <tbody border="1px" cellspacing="0px" style="border-collapse:collapse">
        <tr class="firstRow">
            <td width="100" valign="top">
                当前版本
            </td>
			<td width="100" valign="top">
                {{.Version}}
            </td>
        </tr>
		
		<tr>
			<td width="100" valign="middle">
				<input type="text" name="versionUpdate" placeholder="输入最新版本"/>
			</td>
			<td width="100" valign="middle">
				<button type="button" name="updateVersion" onclick="updateVersion();">更新版本</button>
			</td>
		</tr>
    </tbody>
</table>
<br/>
<br/>
<br/>
<br/>
</body>
</html>
`
