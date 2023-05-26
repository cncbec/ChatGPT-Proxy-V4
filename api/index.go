package api

import (
    "fmt"
    "io"
    "net/http"

    "github.com/bogdanfinn/tls_client"
)

func ForwardRequest(w http.ResponseWriter, r *http.Request) {
    // 建立安全链接的 transport 对象。
    transport, err := tls_client.NewTransport(tls_client.TLSConfig{})
    if err != nil {
        fmt.Printf("Failed to create HTTPS transport due to %v\n", err)
        return
    }

    // 使用 tls_client 提供的 NewCookieJar 函数创建 cookie jar。
    cookieJar, err := tls_client.NewCookieJar()
    if err != nil {
        fmt.Printf("Failed to create cookie jar due to %v\n", err)
        return
    }
    client := &http.Client{
        Transport: transport,
        Jar:       cookieJar,
    }
	
	if r.URL.RawQuery != "" {
		url = "https://chat.openai.com/backend-api" + r.URL.Path + "?" + r.URL.RawQuery
	} else {
		url = "https://chat.openai.com/backend-api" + r.URL.Path
	}

    // 建立请求对象并设置 headers，包括 Origin、Connection、User-Agent 和 Authorization 等。
    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        fmt.Printf("Failed to create the request object due to %v\n", err)
        return
    }
	req.Header.Set("Host", "chat.openai.com")
	req.Header.Set("Origin", "https://chat.openai.com/chat")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Keep-Alive", "timeout=360")
	req.Header.Set("sec-ch-ua", "\"Chromium\";v=\"112\", \"Brave\";v=\"112\", \"Not:A-Brand\";v=\"99\"")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "\"Linux\"")
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sec-gpc", "1")
    req.Header.Set("Connection", "keep-alive")
    req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")

    // 这里假设您已经将源站点的 Authorization 放入自定义头中（Authorization），然后将该头信息转发给目标站点。
    req.Header.Set("Authorization", r.Header.Get("Authorization"))

    // 使用刚才创建的 HTTP client 执行请求。
    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("Failed to send the request to the target site due to %v\n", err)
        return
    }

    // 传递响应 headers 和 body 至调用方，并使用 HTTP 状态码处理响应结果。
    for key, value := range resp.Header {
        for _, hValue := range value {
            w.Header().Add(key, hValue)
        }
    }
    defer resp.Body.Close()
    w.WriteHeader(resp.StatusCode)
    buf := make([]byte, 1024)
    // 将响应体写入流式响应中
    for {
        n, err := resp.Body.Read(buf)
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
