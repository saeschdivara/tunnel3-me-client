package main

import (
	"github.com/cloudwego/hertz/pkg/common/json"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type RequestInfo struct {
	Method  string
	Body    string
	Headers []byte
	Path    string
}

type ResponseInfo struct {
	Body       string
	Headers    map[string][]string
	StatusCode int
}

func main() {

	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) != 2 {
		log.Fatal("Missing parameters [full-websocket-url] [local-app-port]")
	}

	localAppPort := argsWithoutProg[1]

	c, _, err := websocket.DefaultDialer.Dial(argsWithoutProg[0], nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		obj := RequestInfo{}
		err = json.Unmarshal(message, &obj)

		request, err := http.NewRequest(obj.Method, "http://localhost:"+localAppPort+obj.Path, strings.NewReader(obj.Body))

		headers := string(obj.Headers)
		addHeaders(headers, request)

		response, err := http.DefaultClient.Do(request)

		if err != nil {
			log.Println("send:", err)
			return
		}

		body, err := ioutil.ReadAll(response.Body)
		response.Body.Close()

		responseInfo := ResponseInfo{
			StatusCode: response.StatusCode,
			Headers:    response.Header,
			Body:       string(body),
		}

		serialisedResponse, err := json.Marshal(responseInfo)

		c.WriteMessage(websocket.TextMessage, serialisedResponse)

		log.Println("response: ", response)
	}
}

func addHeaders(headers string, request *http.Request) {
	for _, headerValue := range strings.Split(headers, "\r\n") {
		splitHeader := strings.Split(headerValue, ": ")

		if len(splitHeader) == 2 {
			request.Header.Add(splitHeader[0], splitHeader[1])
		}
	}
}
