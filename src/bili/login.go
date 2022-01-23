package bili

import (
	"fmt"
	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	gjson "github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	re "regexp"
	"strings"
	"time"
)

const user_agent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36 Edg/97.0.1072.69`

func get_login_key_and_login_url() (login_key string, login_url string) {
	url := "https://passport.bilibili.com/qrcode/getLoginUrl"
	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", user_agent)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	data := gjson.ParseBytes(body)
	login_key = data.Get("data.oauthKey").String()
	login_url = data.Get("data.url").String()
	return
}

func get_live_buvid() string {
	url := "https://api.live.bilibili.com/gift/v3/live/gift_config"
	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", user_agent)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	SetCookie := resp.Header.Get("Set-Cookie")
	reg := re.MustCompile(`LIVE_BUVID=(AUTO[0-9]+)`)
	live_buvid := reg.FindStringSubmatch(SetCookie)[1]
	return live_buvid
}

func verify_login(login_key string) {
	for {
		url := "https://passport.bilibili.com/qrcode/getLoginInfo"
		client := http.Client{}
		req, _ := http.NewRequest("POST", url, nil)
		req.Header.Set("User-Agent", user_agent)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Body = ioutil.NopCloser(strings.NewReader("oauthKey=" + login_key))
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		data := gjson.ParseBytes(body)
		if data.Get("status").Bool() {
			url := data.Get("data.url").String()
			reg := re.MustCompile(`DedeUserID=(\d+)&DedeUserID__ckMd5=([0-9a-zA-Z]+)&Expires=(\d+)&SESSDATA=([0-9a-zA-Z%]+)&bili_jct=([0-9a-zA-Z]+)&`)
			cookie := make(map[string]string)
			cookie["DedeUserID"] = reg.FindStringSubmatch(url)[1]
			cookie["DedeUserID__ckMd5"] = reg.FindStringSubmatch(url)[2]
			cookie["Expires"] = reg.FindStringSubmatch(url)[3]
			cookie["SESSDATA"] = reg.FindStringSubmatch(url)[4]
			cookie["bili_jct"] = reg.FindStringSubmatch(url)[5]
			cookie["LIVE_BUVID"] = get_live_buvid()
			cookie_content := []byte(`DedeUserID=` + cookie["DedeUserID"] + `;DedeUserID__ckMd5=` + cookie["DedeUserID__ckMd5"] + `;Expires=` + cookie["Expires"] + `;SESSDATA=` + cookie["SESSDATA"] + `;bili_jct=` + cookie["bili_jct"] + `;LIVE_BUVID=` + cookie["LIVE_BUVID"])
			filename := "cookie.txt"
			err := ioutil.WriteFile(filename, cookie_content, 0644)
			if err != nil {
				panic(err)
			}
			s := fmt.Sprintf("扫码成功, cookie如下,已自动保存在当前目录下 %v 文件:", filename)
			fmt.Println(s)
			fmt.Println(string(cookie_content))
			break
		}
		time.Sleep(time.Second * 3)
	}
}

func is_login() (bool, gjson.Result) {
	url := "https://api.bilibili.com/x/web-interface/nav"
	cookie, err := ioutil.ReadFile("cookie.txt")
	if err != nil {
		return false, gjson.Result{}
	}
	cookie_str := string(cookie)
	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", user_agent)
	req.Header.Set("Cookie", cookie_str)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	data := gjson.ParseBytes(body)
	return data.Get("code").Int() == 0, data
}

func Login() {
	for {
		is_login, data := is_login()
		if is_login {
			uname := data.Get("data.uname").String()
			fmt.Println(uname + "已登录")
			fmt.Scanf("%s", "")
			break
		}
		fmt.Println("未登录,或cookie已过期,请扫码登录")
		fmt.Println("请最大化窗口，以确保二维码完整显示，回车继续")
		fmt.Scanf("%s", "")
		login_key, login_url := get_login_key_and_login_url()
		qrcode := qrcodeTerminal.New()
		qrcode.Get([]byte(login_url)).Print()
		fmt.Println("若依然无法扫描，请将以下链接复制到B站打开并确认(任意私信一个人,最好是B站官号，发送链接即可打开)")
		fmt.Println(login_url)
		verify_login(login_key)
	}
}
