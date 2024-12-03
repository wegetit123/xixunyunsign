package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"github.com/spf13/cobra"
	"xixuanyunsign/utils"
)

var (
	address      string
	address_name string
	latitude     string
	longitude    string
	remark       string
	comment      string
)

var SignCmd = &cobra.Command{
	Use:   "sign",
	Short: "执行签到",
	Run: func(cmd *cobra.Command, args []string) {
		signIn()
	},
}

func init() {
	SignCmd.Flags().StringVarP(&account, "account", "a", "", "账号")
	SignCmd.Flags().StringVarP(&address, "address", "", "", "地址(具体名称_小字部分)")
	SignCmd.Flags().StringVarP(&address_name, "address_name", "", "", "地址名称")
	SignCmd.Flags().StringVarP(&latitude, "latitude", "", "", "纬度")
	SignCmd.Flags().StringVarP(&longitude, "longitude", "", "", "经度")
	SignCmd.Flags().StringVarP(&remark, "remark", "", "0", "备注")
	SignCmd.Flags().StringVarP(&comment, "comment", "", "", "评论")
	SignCmd.MarkFlagRequired("account")
	SignCmd.MarkFlagRequired("address")
}

func signIn() {
	token, dbLatitude, dbLongitude, err := utils.GetUser(account)
	if err != nil || token == "" {
		fmt.Println("未找到该账号的 token，请先登录。")
		return
	}

	// 如果命令行未提供 latitude 和 longitude，则使用数据库中的值
	if latitude == "" {
		latitude = dbLatitude
	}
	if longitude == "" {
		longitude = dbLongitude
	}

	if latitude == "" || longitude == "" {
		fmt.Println("未提供经纬度信息，且数据库中不存在，请先查询签到信息或手动提供经纬度。")
		return
	}
	// 使用公钥加密 latitude 和 longitude
	encryptedLatitude, err := rsaEncrypt([]byte(latitude))
	if err != nil {
		fmt.Println("加密纬度失败:", err)
		return
	}

	encryptedLongitude, err := rsaEncrypt([]byte(longitude))
	if err != nil {
		fmt.Println("加密经度失败:", err)
		return
	}

	apiURL := "https://api.xixunyun.com/signin_rsa"

	data := url.Values{}
	data.Set("address", address)
	data.Set("province", "")
	data.Set("city", "")
	data.Set("latitude", encryptedLatitude)
	data.Set("longitude", encryptedLongitude)
	data.Set("remark", remark)
	data.Set("comment", comment)
	data.Set("address_name", address_name)
	data.Set("change_sign_resource", "0")

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Println("创建请求失败:", err)
		return
	}

	query := req.URL.Query()
	query.Add("token", token)
	query.Add("from", "app")
	query.Add("version", "5.1.3")
	query.Add("platform", "android")
	query.Add("entrance_year", "0")
	query.Add("graduate_year", "0")
	query.Add("school_id", "7")
	req.URL.RawQuery = query.Encode()

	req.Header.Set("User-Agent", "okhttp/3.8.0")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", "PHPSESSID=qkc555lu6050h43e204crialf0")

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
		fmt.Println("签到失败:", result["message"])
		return
	}

	fmt.Println("签到成功！")
}

// rsaEncrypt 使用提供的公钥对数据进行加密
func rsaEncrypt(origData []byte) (string, error) {
	publicKey := `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDlYsiV3DsG+t8OFMLyhdmG2P2J
4GJwmwb1rKKcDZmTxEphPiYTeFIg4IFEiqDCATAPHs8UHypphZTK6LlzANyTzl9L
jQS6BYVQk81LhQ29dxyrXgwkRw9RdWaMPtcXRD4h6ovx6FQjwQlBM5vaHaJOHhEo
rHOSyd/deTvcS+hRSQIDAQAB
-----END PUBLIC KEY-----`

	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return "", errors.New("公钥解码失败")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	pub := pubInterface.(*rsa.PublicKey)

	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
	if err != nil {
		return "", err
	}

	// 使用 base64 编码
	encryptedString := base64.StdEncoding.EncodeToString(encryptedData)
	return encryptedString, nil
}
