package main

import (
	"github.com/cloudwego/hertz/pkg/common/json"
	"github.com/gorilla/websocket"
	"io"
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

type CreateHostResult struct {
	Result string `json:"result"`
}

func main() {

	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) != 3 {
		log.Fatal("Missing parameters [base-url] [subdomain] [local-app-port]")
	}

	baseUrl := argsWithoutProg[0]
	subdomain := argsWithoutProg[1]
	localAppPort := argsWithoutProg[2]

	resp, err := http.Get("https://config." + baseUrl + "/create-host/" + subdomain)
	if err != nil {
		log.Fatal("Connect to create host failed:", err)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		log.Fatal(err)
	}

	obj := CreateHostResult{}
	err = json.Unmarshal(body, &obj)

	c, _, err := websocket.DefaultDialer.Dial("ws://"+baseUrl+":"+obj.Result, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	defer c.Close()

	httpUrl := "https://" + subdomain + "." + baseUrl + "/"
	log.Println("Accepting connections on:", httpUrl)

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		obj := RequestInfo{}
		err = json.Unmarshal(message, &obj)

		// TODO: handle failing request (maybe server was not started yet)
		request, err := http.NewRequest(obj.Method, "http://localhost:"+localAppPort+obj.Path, strings.NewReader(obj.Body))

		headers := string(obj.Headers)
		addHeaders(headers, request)

		response, err := http.DefaultClient.Do(request)

		if err != nil {
			log.Println(obj.Method, obj.Path, "-", err)
			return
		}

		body, err := io.ReadAll(response.Body)
		response.Body.Close()

		responseInfo := ResponseInfo{
			StatusCode: response.StatusCode,
			Headers:    response.Header,
			Body:       string(body),
		}

		log.Println(obj.Method, obj.Path, "-", response.StatusCode)

		serialisedResponse, err := json.Marshal(responseInfo)

		c.WriteMessage(websocket.TextMessage, serialisedResponse)
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
