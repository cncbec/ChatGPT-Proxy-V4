package api



func Handler(w http.ResponseWriter, r *http.Request) {

	if http_proxy != "" {
		client.SetProxy(http_proxy)
		println("Proxy set:" + http_proxy)
	}

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "9090"
	}
	handlerss := gin.Default()
	handlerss.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	handlerss.Any("/api/*path", proxy)

	gin.SetMode(gin.ReleaseMode)
	endless.ListenAndServe(os.Getenv("HOST")+":"+PORT, handlerss)
}

