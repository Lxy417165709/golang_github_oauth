package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

type Client struct {
	ClientId     string
	ClientSecret string
	RedirectUrl  string
}

var client = Client{
	ClientId:     "7e5fe351bc9b131c6f2a",
	ClientSecret: "9fd22c13ae790685c59e3fb4a9b444b75b506a5b",
	RedirectUrl:  "http://localhost:9090/oauth/redirect",
}

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"` // 这个字段下面没用到
	Scope       string `json:"scope"`      // 这个字段下面也没用到
}

// 网站的主页，里面包含一个Github第三方登录的链接
// 第一步: 用户点击该链接后，用户会跳转到Github登录页面									(用户进行操作)
// 第二步: 用户登录成功后，将会跳转到 client.RedirectUrl，并在该Url后面拼接一个code		(用户浏览器重定向)
func Hello(w http.ResponseWriter, r *http.Request) {
	// 解析指定文件生成模板对象
	var temp *template.Template
	var err error
	if temp, err = template.ParseFiles("views/hello.html"); err != nil {
		fmt.Println("读取文件失败，错误信息为:", err)
		return
	}

	// 利用给定数据渲染模板(html页面)，并将结果写入w，返回给前端
	if err = temp.Execute(w, client); err != nil {
		fmt.Println("读取渲染html页面失败，错误信息为:", err)
		return
	}
}

// 第三步: 当用户访问 client.RedirectUrl 时，服务器会根据这个url，获得code
// 第四步: 服务器通过 code，向Github请求获取用户的token
// 第五步: 获取到token后，服务器再通过 token 中的 AccessToken 向Github 请求用户的个人信息
// 第六步: 用户信息返回后，服务器将用户信息返回前端，展示给用户看
func Oauth(w http.ResponseWriter, r *http.Request) {

	// 第三、四步
	var token *Token
	var err error
	if token, err = GetToken(GetUrl(r.URL.Query().Get("code"))); err != nil {
		fmt.Println(err)
		return
	}

	// 第五步
	var userInfo map[string]interface{}
	if userInfo, err = GetUserInfo(token); err != nil {
		fmt.Println("获取用户信息失败，错误信息为:", err)
		return
	}

	//  第六步
	var userInfoBytes []byte
	if userInfoBytes, err = json.Marshal(userInfo); err != nil {
		fmt.Println("在将用户信息(map)转为用户信息([]byte)时发生错误，错误信息为:", err)
		return
	}

	if _, err = w.Write(userInfoBytes); err != nil {
		fmt.Println("在将用户信息([]byte)返回前端时发生错误，错误信息为:", err)
		return
	}

}
func GetUrl(code string) string {
	return fmt.Sprintf(
		"https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		client.ClientId, client.ClientSecret, code,
	)
}
func GetToken(url string) (*Token, error) {


	var req *http.Request
	var err error
	if req, err = http.NewRequest(http.MethodGet, url, nil);err!=nil{
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	
	var httpClient = http.Client{}
	var res *http.Response
	if res, err = httpClient.Do(req);err!=nil{
		return nil, err
	}

	// 将相应解析为token
	var token Token
	if err = json.NewDecoder(res.Body).Decode(&token); err != nil {
		return nil, err
	}
	return &token, nil
}
func GetUserInfo(token *Token) (map[string]interface{}, error) {

	var userInfoUrl = "https://api.github.com/user"
	var req *http.Request
	var err error
	if req, err = http.NewRequest(http.MethodGet, userInfoUrl, nil); err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token.AccessToken))

	var client = http.Client{}
	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return nil, err
	}

	// 将响应数据写入userInfo中
	var userInfo = make(map[string]interface{})
	if err = json.NewDecoder(res.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	return userInfo, nil
}

func main() {
	http.HandleFunc("/", Hello)
	http.HandleFunc("/oauth/redirect", Oauth)

	if err := http.ListenAndServe(":9090", nil); err != nil {
		fmt.Println("监听失败，错误信息为:", err)
		return
	}
}
