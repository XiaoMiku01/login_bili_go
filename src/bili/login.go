package bili

import (
	"fmt"
	"io"
	"net/http"
	"os"
	re "regexp"
	"strings"
	"time"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	gjson "github.com/tidwall/gjson"
)

const user_agent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36 Edg/97.0.1072.69`

var CK string

func get_login_key_and_login_url() (login_key string, login_url string) {
	url := "https://passport.bilibili.com/x/passport-login/web/qrcode/generate"
	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", user_agent)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	data := gjson.ParseBytes(body)
	login_key = data.Get("data.qrcode_key").String()
	login_url = data.Get("data.url").String()
	return
}

func verify_login(login_key string) {
	for {
		url := "https://passport.bilibili.com/x/passport-login/web/qrcode/poll"
		client := http.Client{}
		url += "?" + "qrcode_key=" + login_key
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", user_agent)
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		data := gjson.ParseBytes(body)
		if data.Get("data.url").String() != "" {
			var cookie_content = []byte{}
			cookie := make(map[string]string)
			for _, v := range resp.Header["Set-Cookie"] {
				kv := strings.Split(v, ";")[0]
				kv_arr := strings.Split(kv, "=")
				cookie[kv_arr[0]] = kv_arr[1]
			}
			filename := cookie["DedeUserID"] + "_cookie.txt"
			cookie_content = []byte(`DedeUserID=` + cookie["DedeUserID"] + `;DedeUserID__ckMd5=` + cookie["DedeUserID__ckMd5"] + `;Expires=` + cookie["Expires"] + `;SESSDATA=` + cookie["SESSDATA"] + `;bili_jct=` + cookie["bili_jct"] + `;`)

			err := os.WriteFile(filename, cookie_content, 0644)
			if err != nil {
				panic(err)
			}
			s := fmt.Sprintf("扫码成功, cookie如下,已自动保存在当前目录下 %v 文件:", filename)
			fmt.Println(s)
			fmt.Println(string(cookie_content))
			CK = string(cookie_content)
			break
		}
		time.Sleep(time.Second * 3)
	}
}

func is_login() (bool, gjson.Result, string, string) {
	url := "https://api.bilibili.com/x/web-interface/nav"
	cookie_str := string(CK)
	reg := re.MustCompile(`bili_jct=([0-9a-zA-Z]+);`)
	csrf := reg.FindStringSubmatch(cookie_str)[1]
	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("Cookie", cookie_str)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	data := gjson.ParseBytes(body)
	return data.Get("code").Int() == 0, data, cookie_str, csrf
}

func Login() (string, string) {
	for {
		// 清空 输出
		fmt.Print("\033[H\033[2J")
		login_key, login_url := get_login_key_and_login_url()
		qrcode := qrcodeTerminal.New()
		qrcode.Get([]byte(login_url)).Print()
		fmt.Println("若依然无法扫描，请将以下链接复制到B站打开并确认(任意私信一个人,最好是B站官号，发送链接即可打开)")
		fmt.Println(login_url)
		verify_login(login_key)
		is_login, data, cookie_str, csrf := is_login()
		if is_login {
			uname := data.Get("data.uname").String()
			fmt.Println(uname + "已登录")
			return cookie_str, csrf
		}
	}
}
