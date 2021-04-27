package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"hdc-task-manager/common"
	"io/ioutil"
	"net/http"
	"strconv"
)

//HTTPPost post request
func HTTPPost(url string, requestBody string) (map[string]interface{}, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(requestBody)))
	defer common.Catchs()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("Post request failed, err: ", err, "body: ", requestBody, "url:", url)
		return nil, err
	}
	defer resp.Body.Close()
	logs.Info("response Status:", resp.Status, "url: ", url)
	logs.Info("response Headers:", resp.Header, "url: ", url)
	status, _ := strconv.Atoi(resp.Status)
	if status > 300 {
		logs.Error("Post request failed, err: ", err, "body: ", requestBody, "url:", url)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	if err != nil || body == nil {
		logs.Error("post failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	logs.Info("post successed!, body: ", string(body))
	var iss map[string]interface{}
	err = json.Unmarshal(body, &iss)
	if err != nil {
		logs.Error(err, string(body))
		return nil, err
	}
	logs.Info(iss)
	return iss, nil
}

//HTTPPatch patch request
func HTTPPatch(url string, requestBody string) (map[string]interface{}, error) {
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte(requestBody)))
	defer common.Catchs()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("PATCH request failed, err: ", err, "body: ", requestBody, "url:", url)
		return nil, err
	}
	defer resp.Body.Close()
	logs.Info("response Status:", resp.Status, "url: ", url)
	logs.Info("response Headers:", resp.Header, "url: ", url)
	status, _ := strconv.Atoi(resp.Status)
	if status > 300 {
		logs.Error("Patch request failed, err: ", err, "body: ", requestBody, "url:", url)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	if err != nil || body == nil {
		logs.Error("PATCH failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	logs.Info("PATCH successed!, body: ", string(body))
	var iss map[string]interface{}
	err = json.Unmarshal(body, &iss)
	if err != nil {
		logs.Error(err, string(body))
		return nil, err
	}
	logs.Info(iss)
	return iss, nil
}

//HTTPPut patch request
func HTTPPut(url string, requestBody string) ([]map[string]interface{}, error) {
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(requestBody)))
	defer common.Catchs()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("PATCH request failed, err: ", err, "body: ", requestBody, "url:", url)
		return nil, err
	}
	defer resp.Body.Close()
	logs.Info("response Status:", resp.Status, "url: ", url)
	logs.Info("response Headers:", resp.Header, "url: ", url)
	status, _ := strconv.Atoi(resp.Status)
	if status > 300 {
		logs.Error("Patch request failed, err: ", err, "body: ", requestBody, "url:", url)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	if err != nil || body == nil{
		logs.Error("PUT failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	logs.Info("PUT successed!, body: ", string(body))
	var iss []map[string]interface{}
	err = json.Unmarshal(body, &iss)
	if err != nil {
		logs.Error(err, string(body))
		return nil, err
	}
	return iss, nil
}