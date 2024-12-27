package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"xixunyunsign/utils"
)

var (
	account   string
	password  string
	school_id string
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录到系统",
	Run: func(cmd *cobra.Command, args []string) {
		login()
	},
}

func init() {
	LoginCmd.Flags().StringVarP(&account, "account", "a", "", "账号")
	LoginCmd.Flags().StringVarP(&password, "password", "p", "", "密码")
	LoginCmd.Flags().StringVarP(&school_id, "school_id", "i", "7", "学校id")
	LoginCmd.MarkFlagRequired("account")
	LoginCmd.MarkFlagRequired("password")
}

func login() {
	apiURL := "https://api.xixunyun.com/login/api"

	data := url.Values{}
	data.Set("app_version", "5.1.3")
	data.Set("registration_id", "")
	data.Set("uuid", "fd9dc13a49cc850c")
	data.Set("request_source", "3")
	data.Set("platform", "2")
	data.Set("mac", "7C:F3:1B:BB:F1:C4")
	data.Set("password", password)
	data.Set("system", "10")
	data.Set("school_id", school_id)
	data.Set("model", "LM-G820")
	data.Set("app_id", "cn.vanber.xixunyun.saas")
	data.Set("account", account)
	data.Set("key", "")

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Println("创建请求失败:", err)
		return
	}

	query := req.URL.Query()
	//query.Add("token", "2e5569c4598355e9e3cabcd5806d8754")
	query.Add("from", "app")
	query.Add("version", "5.1.3")
	query.Add("platform", "android")
	query.Add("entrance_year", "0")
	query.Add("graduate_year", "0")
	query.Add("school_id", school_id)
	req.URL.RawQuery = query.Encode()

	req.Header.Set("User-Agent", "okhttp/3.8.0")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Set("Cookie", "PHPSESSID=sjgggpe71m53qv1o9dor0uurg4")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["code"].(float64) != 20000 {
		fmt.Println("登录失败:", result["message"])
		return
	}

	dataMap := result["data"].(map[string]interface{})
	token := dataMap["token"].(string)

	// 保存到数据库
	err = utils.SaveUser(
		account,
		password,
		token,
		"", "", // 经纬度信息留空
		getStringFromResult(dataMap, "bind_phone"),
		getStringFromResult(dataMap, "user_number"),
		getStringFromResult(dataMap, "user_name"),
		dataMap["school_id"].(float64),
		getStringFromResult(dataMap, "sex"),
		getStringFromResult(dataMap, "class_name"),
		getStringFromResult(dataMap, "entrance_year"),
		getStringFromResult(dataMap, "graduation_year"),
	)
	if err != nil {
		fmt.Println("保存用户信息失败:", err)
		return
	}
	fmt.Println("登录成功！")
}

func getStringFromResult(dataMap map[string]interface{}, key string) string {
	if value, ok := dataMap[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return "" // 如果字段不存在或类型不匹配，返回空字符串
}
