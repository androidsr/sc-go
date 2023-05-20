package wsocket

/*
func Test_NewWebsocket(t *testing.T) {
	configs, _ := syaml.LoadFile[syaml.PaasRoot]("../sc.yaml")
	router := sgin.New(configs.Paas.Gin)
	socket := New(websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}, time.Second*60, 3, func(w http.ResponseWriter, r *http.Request) string {
		return "sirui"
	})
	router.GET("/ws", func(c *gin.Context) {
		err := socket.Register(c.Writer, c.Request)
		if err != nil {
			c.JSON(http.StatusOK, model.NewFail(5000, err.Error()))
		}
	})
	go func() {
		for {
			m := <-socket.Data
			fmt.Println(m.UserId, string(m.Data))
			socket.Write(m.UserId, websocket.TextMessage, []byte("你好，"+string(m.Data)))
		}
	}()
	router.RunServer()
}
*/
