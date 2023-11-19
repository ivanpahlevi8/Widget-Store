package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// create some object that hold web socket connetciont
type WebSocketConnection struct {
	*websocket.Conn
}

// create some object to hold user request
type WebSocketRequest struct {
	Action      string `json:"action"`
	Message     string `json:"message"`
	MessageType string `json:"message_type"`
	Username    string `json:"user_name"`
	UserID      int    `json:"user_id"`
	Conn        *WebSocketConnection
}

// cretae object as response
type WebSocketResponse struct {
	Action  string `json:"action"`
	Message string `json:"message"`
	UserID  int    `json:"user_id"`
}

/**
websocket response akan digunakan untuk memberikan response ke masing - masing user yang sedang login
di web yang berbeda. Jika salah satu user yang masih login, kemudian userID dari user tersebut sama dengan userID
yang terdapat pada WebSocketResponse, maka user tersebut akan dilakukan proses logout dari web.
UserID dari WebSocketResponse dapat berisi id dari user yang telah dihapus karena response tersebut akan dibuat
saat admin menghapus user lainnya, sehingga user yang terhapus dapat diambil idnya dan dilakuakn pembuatan response.
Hal ini sesuia dengan tujuan penggunaan websocket yang digunakan untuk meloggout user terhapus yang masih login
*/

// make updgraded connection
var upgradedConn = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

/**
upgradeConn merupakan variable yang digunakan untuk memulai koneksi pada websocket
dinamakan upgradeConn karena saat m,emulia koneksi, koneksi akan diupgrade menjadi koneksi full duplex atau dua arah
*/

// create clinet data to hold client connection
var clients = make(map[WebSocketConnection]string)

// create channel variable to hold payload from user request
var channelRequest = make(chan WebSocketRequest)

// create function to start connection
func (app *Application) WebsocketEndPoint(w http.ResponseWriter, r *http.Request) {
	// create connection with upgrade connection
	wsConn, err := upgradedConn.Upgrade(w, r, nil)

	// check for an error
	if err != nil {
		log.Println("error when creating web socket connection")
		app.errorLog.Printf("Error when creating connection to websocket : %s\n", err)
		return
	}

	// create WebSocketConnection object
	webSocketConn := WebSocketConnection{
		Conn: wsConn,
	}

	// add cline who connect to webscoket
	clients[webSocketConn] = ""

	// create response back
	responseWs := WebSocketResponse{
		Message: "Success connected with Web Socket End Point",
	}

	// send response
	err = wsConn.WriteJSON(responseWs)

	// check for an error
	if err != nil {
		log.Println("error when send json response using Web Socket")
		app.errorLog.Printf("Error when send json response using Web Socket : %s\n", err)
		return
	}

	// listenting to socket always in background
	go app.ListenForWS(&webSocketConn)
}

// create function to be running on background
func (app *Application) ListenForWS(conn *WebSocketConnection) {
	// creqate function to be called when programs accidently stops
	defer func() {
		r := recover()
		if r != nil {
			// there is an error
			app.errorLog.Println("Error happen makes application accidently stops : ", fmt.Sprintf("%v", r))
		}
	}()

	// create request payload
	var requestPayloadUser WebSocketRequest

	// do inifnte loop to read data always
	for {
		// read data from websocket
		err := conn.ReadJSON(&requestPayloadUser)

		// check for an error
		if err != nil {
			log.Println("errro when reading web socket payload : ", err)
			app.errorLog.Println(err)
			return
		}

		// if success, sending data to chan
		channelRequest <- requestPayloadUser
	}

}

// create function to be called always to process data in channel
func (app *Application) ListenToWSChannel() {
	// create response object
	var responsePayload WebSocketResponse

	// loop forever in background
	for {
		// get request data from channel
		reqChann := <-channelRequest

		// check action from request
		switch reqChann.Action {
		case "deleteUser":
			fmt.Println("Delete user : ", reqChann)
			// deleting user action
			// action to take by user is logout
			responsePayload.Action = "logout"
			responsePayload.Message = "Your account has been deleted"

			// send response to all user
			app.BroadCastingToAllUser(responsePayload)
		default:
			log.Println("Unidetified action from request...")
		}
	}
}

// create function to broadcasting response to all user
func (app *Application) BroadCastingToAllUser(payloadResponse WebSocketResponse) {
	// loop through all user
	for client := range clients {
		// send payload response to cliend
		err := client.WriteJSON(payloadResponse)

		// check for an error
		if err != nil {
			log.Println("error when broadcasting response to all user : ", err)
			// close clien connection
			_ = client.Close()
			// delete error client from all client
			delete(clients, client)
		}
	}
}
