package controller

import (
	"PaintingExchange/internal/model"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
	"log"
	"net/http"
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// online 当前在线的websocket连接
var online = sync.Map{}

// wlock 连接写互斥锁
var wlock = sync.Map{}

// HandleWebsocket websocket服务端
// @Summary websocket服务端(无法在swagger中测试)
// @Description 通过此端点建立 WebSocket 连接。连接后，进行实时聊天交流。
// @Tags chat
// @Accept json
// @Produce json
// @Param Sec-WebSocket-Protocol header string true "子协议填写JWT(不需要Bearer),用于身份验证"
// @Success 101 {string} string "WebSocket 连接建立成功"
// @Success 200 {object} model.Message
// @Failure 401 {object} string "未授权，JWT无效或已过期"
// @Router /chat [get]
func HandleWebsocket(ctx iris.Context) {
	w := ctx.ResponseWriter()
	r := ctx.Request()
	db := ctx.Value("db").(*gorm.DB)

	//判断请求是否为websocket升级请求
	if websocket.IsWebSocketUpgrade(r) {
		conn, err := upgrader.Upgrade(w, r, w.Header())
		if err != nil {
			log.Println("websocket客户端连接失败", err)
		}
		user, _ := ctx.User().GetRaw()
		username := user.(iris.SimpleUser).Username
		if prev, ok := online.Load(username); ok {
			log.Println("用户", username, "重复登录")
			prev.(*websocket.Conn).Close()
			log.Println("已关闭用户", username, "此前的连接")
		}
		log.Println("用户", username, "连接ws聊天室成功")
		online.Store(username, conn)
		wlock.Store(username, &sync.Mutex{})

		// 历史聊天记录
		//db.Raw("SELECT * FROM (SELECT *, ROW_NUMBER() OVER (PARTITION BY `from` ORDER BY time DESC) AS rn FROM messages WHERE `to` = ?) AS subq WHERE rn <= 10", username).Scan(&initMess)
		var historyFrom []model.Message
		db.Where("`from`=?", username).Find(&historyFrom)
		for _, mess := range historyFrom {
			go func(mess model.Message) {
				fmt.Println("发送用户初始聊天记录", mess)
				if err := sendHistoryMessage(mess, username); err != nil {
					log.Println("初始聊天记录发送失败", err)
				}
			}(mess)
		}
		var historyTo []model.Message
		db.Where("`to`=?", username).Find(&historyTo)
		for _, mess := range historyTo {
			go func(mess model.Message) {
				fmt.Println("发送用户初始聊天记录", mess)
				if err := sendHistoryMessage(mess, username); err != nil {
					log.Println("初始聊天记录发送失败", err)
				}
			}(mess)
		}

		// websocket 连接断开
		conn.SetCloseHandler(func(code int, text string) error {
			online.Delete(username)
			wlock.Delete(username)
			message := websocket.FormatCloseMessage(code, "")
			if err := conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second)); err != nil {
				return err
			}
			log.Println(username, "websocket连接断开")
			return nil
		})

		go func() {
			for {
				t, message, err := conn.ReadMessage()

				if t == -1 {
					conn.Close()
					break
				}

				if err != nil {
					log.Println("websocket消息接收失败", err)
					continue
				}

				var messobj model.Message
				if err = json.Unmarshal(message, &messobj); err != nil {
					log.Println("json消息解析失败", err)
					continue
				}
				messobj.From = username
				messobj.Time = time.Now()

				fmt.Println(username, "用户", messobj.From, "发送消息至", messobj.To, ",内容:", messobj.Content)
				if err := sendMessage(messobj); err != nil {
					log.Println("消息发送失败", err)
				} else {
					db.Create(&messobj)
				}
			}
		}()
	} else {
		log.Println("websocket客户端连接失败")
	}
}

// sendMessage 发送聊天记录
func sendMessage(message model.Message) error {
	return sendHistoryMessage(message, message.To)
}

// sendHistoryMessage 发送历史聊天记录
func sendHistoryMessage(message model.Message, username string) error {
	targetA, ok := online.Load(username)
	if !ok {
		return nil
	}
	wlockA, ok := wlock.Load(username)
	if !ok {
		return errors.New("用户" + username + "连接异常")
	}
	target := targetA.(*websocket.Conn)
	wlock := wlockA.(*sync.Mutex)
	wlock.Lock()
	defer wlock.Unlock()
	jsons, _ := json.Marshal(message)
	err := target.WriteMessage(1, jsons)
	return err
}
