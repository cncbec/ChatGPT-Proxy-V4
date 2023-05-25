package api

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/tls-client"
)

var (
	UserAgent  = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36"
	HTTPProxy  = os.Getenv("http_proxy")
	CookieJar  = tls_client.NewCookieJar()
	ClientOpts = []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(360),
		tls_client.WithClientProfile(tls_client.Chrome_112),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(CookieJar),
	}
)



func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Remove _cfuvid cookie from session
	CookieJar.SetCookies(r.URL, []*fhttp.Cookie{})

	var url string
	var err error
	var requestMethod string

	if r.URL.RawQuery != "" {
		url = "https://chat.openai.com/backend-api" + r.URL.Path + "?" + r.URL.RawQuery
	} else {
		url = "https://chat.openai.com/backend-api" + r.URL.Path
	}
	requestMethod = r.Method

	request, err := fhttp.NewRequest(requestMethod, url, r.Body)
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
	request.Header.Set("User-Agent", UserAgent)
	if os.Getenv("PUID") != "" {
		request.AddCookie(&fhttp.Cookie{Name: "_puid", Value: os.Getenv("PUID")})
	}

	client, err := fhttp.NewHttpClient(fhttp.NewNoopLogger(), ClientOpts...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := client.Do(request)
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
			w.(http.Flusher).Flush()
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

func main() {
	if HTTPProxy != "" {
		tls_client.DefaultTransport.SetProxy(HTTPProxy)
		log.Println("Proxy set:", HTTPProxy)
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", &ProxyHandler{})
	mux.HandleFunc("/ping", pingHandler)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
