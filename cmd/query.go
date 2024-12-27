package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
	"xixunyunsign/utils"
)

func init() {
	QueryCmd.Flags().StringVarP(&account, "account", "a", "", "账号")
	QueryCmd.MarkFlagRequired("account")
}

var QueryCmd = &cobra.Command{
	Use:   "query",
	Short: "查询签到信息",
	Run: func(cmd *cobra.Command, args []string) {
		querySignIn()
	},
}

func querySignIn() {
	token, _, _, err := utils.GetUser(account)
	if err != nil || token == "" {
		fmt.Println("未找到该账号的 token，请先登录。")
		return
	}
	userData, err := utils.GetAdditionalUserData(account)
	if err != nil {
		fmt.Println("获取用户额外信息失败:", err)
		return
	}

	apiURL := "https://api.xixunyun.com/signin40/homepage"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		fmt.Println("创建请求失败:", err)
		return
	}

	query := req.URL.Query()
	query.Add("month_date", "2024-12")
	query.Add("token", token)
	query.Add("from", "app")
	query.Add("version", "5.1.3")
	query.Add("platform", "android")
	query.Add("entrance_year", userData["entrance_year"])
	query.Add("graduate_year", userData["graduation_year"])
	query.Add("school_id", userData["school_id"])
	req.URL.RawQuery = query.Encode()

	req.Header.Set("User-Agent", "okhttp/3.8.0")
	req.Header.Set("Accept-Encoding", "gzip")
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
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("解析响应失败:", err)
		return
	}

	code, ok := result["code"].(float64)
	if !ok || code != 20000 {
		message, _ := result["message"].(string)
		fmt.Printf("查询失败: %s\n", message)
		return
	}

	fmt.Println("查询成功！")

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		fmt.Println("解析数据失败：无效的响应结构")
		return
	}

	signResourcesInfo, ok := data["sign_resources_info"].(map[string]interface{})
	if !ok {
		fmt.Println("解析签到资源信息失败：无效的响应结构")
		return
	}

	midLatitude := fmt.Sprintf("%v", signResourcesInfo["mid_sign_latitude"])
	midLongitude := fmt.Sprintf("%v", signResourcesInfo["mid_sign_longitude"])

	// 更新数据库中的经纬度信息
	err = utils.UpdateCoordinates(account, midLatitude, midLongitude)
	if err != nil {
		fmt.Println("保存经纬度信息失败:", err)
		return
	}

	fmt.Println("应签到位置的经纬度已更新。")
}
