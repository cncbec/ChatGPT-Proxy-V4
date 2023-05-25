package api

import (
	"io"
	"log"
	"net/http"
	"os"
	tls_client "github.com/bogdanfinn/tls-client"
)

var (
	jar     = tls_client.NewCookieJar()
	options = []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(360),
		tls_client.WithClientProfile(tls_client.Chrome_112),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
	}
	client, _  = tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	user_agent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36"
)

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// Remove _cfuvid cookie from session
	jar.SetCookies(r.URL, []*http.Cookie{})

	var url string
	var err error
	var requestMethod string
	var request *http.Request
	var response *http.Response

	if r.URL.RawQuery != "" {
		url = "https://chat.openai.com/backend-api" + r.URL.Path + "?" + r.URL.RawQuery
	} else {
		url = "https://chat.openai.com/backend-api" + r.URL.Path
	}
	requestMethod = r.Method

	request, err = http.NewRequest(requestMethod, url, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()
	w.Header().Set("Content-Type", response.Header.Get("Content-Type"))
	w.WriteHeader(response.StatusCode)

	buf := make([]byte, 4096)
	for {
		n, err := response.Body.Read(buf)
		if n > 0 {
			_, writeErr := w.Write(buf[:n])
			if writeErr != nil {
				log.Printf("Error writing to client: %v", writeErr)
				break
			}
			w.(http.Flusher).Flush() // flush buffer to make sure the data is sent to client in time.
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading from response body: %v", err)
			break
		}
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "pong"}`))
}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/":
		proxyHandler(w, r)
	case "/ping":
		pingHandler(w, r)
	default:
		http.NotFound(w, r)
	}
}
