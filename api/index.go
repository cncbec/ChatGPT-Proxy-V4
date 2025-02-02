package api

import (
	"os"
    "fmt"
    "io"
    yhttp "net/http"
	http "github.com/bogdanfinn/fhttp"
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

func ForwardRequest(w yhttp.ResponseWriter, r *yhttp.Request) {
	jar.SetCookies(r.URL, []*http.Cookie{})
	
	var url string
	var err error
	var request_method string
	var request *http.Request
	var response *http.Response
	
	
	if r.URL.RawQuery != "" {
		url = "https://chat.openai.com/backend-api" + r.URL.Path + "?" + r.URL.RawQuery
	} else {
		url = "https://chat.openai.com/backend-api" + r.URL.Path
	}
	
	
	request_method = r.Method

	request, err = http.NewRequest(request_method, url, r.Body)
	if err != nil {
		fmt.Printf("Failed to create HTTPS transport due to %v\n", err)
		return
    }
	request.Header.Set("Host", "chat.openai.com")
	request.Header.Set("Origin", "https://chat.openai.com/chat")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Keep-Alive", "timeout=360")
	request.Header.Set("Authorization", r.Header.Get("Authorization"))
	request.Header.Set("sec-ch-ua", "\"Chromium\";v=\"112\", \"Brave\";v=\"112\", \"Not:A-Brand\";v=\"99\"")
	request.Header.Set("sec-ch-ua-mobile", "?0")
	request.Header.Set("sec-ch-ua-platform", "\"Linux\"")
	request.Header.Set("sec-fetch-dest", "empty")
	request.Header.Set("sec-fetch-mode", "cors")
	request.Header.Set("sec-fetch-site", "same-origin")
	request.Header.Set("sec-gpc", "1")
	request.Header.Set("user-agent", user_agent)
	if os.Getenv("PUID") != "" {
		request.AddCookie(&http.Cookie{Name: "_puid", Value: os.Getenv("PUID")})
	}

	response, err = client.Do(request)
	if err != nil {
		fmt.Printf("Failed to create cookie jar due to %v\n", err)
		return
    }
	


    defer response.Body.Close()
	w.Header().Add("Content-Type", response.Header.Get("Content-Type"))
    w.WriteHeader(response.StatusCode)
    buf := make([]byte, 1024)
    // 将响应体写入流式响应中
    for {
        n, err := response.Body.Read(buf)
        if err != nil && err != io.EOF {
            fmt.Printf("Failed to read response Body due to %v\n", err)
            break
        }
        if n == 0 {
            break
        }
        _, err = w.Write(buf[:n])
        if err != nil {
            fmt.Printf("Failed to write response Body due to %v\n", err)
            break
        }
    }
}
