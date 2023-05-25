package api

import (
	"io"
	"log"
	"os"

	"net/http"
	tls_client "github.com/bogdanfinn/tls-client"


)

var (
	jar     = tls_client.NewCookieJar()
	options = []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(360),
		tls_client.WithClientProfile(tls_client.Chrome_112),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar), // create cookieJar instance and pass it as argument
	}
	client, _  = tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	user_agent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36"
	http_proxy = os.Getenv("http_proxy")
)

func Proxy(w http.ResponseWriter, r *http.Request) {
	client := http.DefaultClient

	jar, _ := r.Cookie("_puid")
	jar.Value = "" // 清空 cookie 的值
	r.AddCookie(jar)

	var url string
	if r.URL.RawQuery != "" {
		url = "https://chat.openai.com/backend-api" + r.URL.Path + "?" + r.URL.RawQuery
	} else {
		url = "https://chat.openai.com/backend-api" + r.URL.Path
	}

	reqURL, _ := url.Parse(url)
	jar.SetCookies(reqURL, []*http.Cookie{})

	req, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Host = "chat.openai.com"
	req.Header.Set("User-Agent", http_proxy)
	req.Header.Set("Host", "chat.openai.com")
	req.Header.Set("Origin", "https://chat.openai.com/chat")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Keep-Alive", "timeout=360")
	req.Header.Set("Authorization", r.Request.Header.Get("Authorization"))
	req.Header.Set("sec-ch-ua", "\"Chromium\";v=\"112\", \"Brave\";v=\"112\", \"Not:A-Brand\";v=\"99\"")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "\"Linux\"")
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sec-gpc", "1")
	req.Header.Set("user-agent", user_agent)

	// 将环境变量 "PUID" 的值作为 Cookie 添加到请求中
	if os.Getenv("PUID") != "" {
		req.AddCookie(&http.Cookie{Name: "_puid", Value: os.Getenv("PUID")})
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 设置响应头部
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// 将响应内容拷贝到客户端的响应中
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response body to client: %v", err)
	}
}



