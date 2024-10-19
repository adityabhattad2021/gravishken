package main

import (
	"bytes"
	types "common"
	TEST "common/models/test"
	user "common/models/user"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

var server_url string

type Client struct {
	client http.Client
	jwt    string
	currentUser *user.User
	availableTests []TEST.Test

	server struct {
		conn         *websocket.Conn
		send         chan types.Message
		recv         chan types.Message
		conn_started bool
	}

	exit struct {
		ctx     context.Context
		destroy context.CancelFunc
	}

	frontend struct {
		send chan<- types.Message
	}
}

func newClient(send chan<- types.Message) (*Client, error) {
	self := &Client{}
	self.frontend.send = send

	self.client = http.Client{}

	self.server.send = make(chan types.Message, 100)
	self.server.recv = make(chan types.Message, 100)

	ctx, destroy := context.WithCancel(context.Background())
	self.exit.ctx = ctx
	self.exit.destroy = destroy

	return self, nil
}

func (self *Client) destroy() {
	self.closeServerConn()
	close(self.server.send)
	self.exit.destroy()
}

func (self *Client) notifyErr(err error) {
	if err != nil {
		self.frontend.send <- types.NewMessage(types.TErr{
			Message: fmt.Sprintf("Error: %s", err),
		})
		log.Printf("Error: %s\n", err)
	}
}

func (self *Client) closeServerConn() {
	if self.server.conn == nil {
		return
	}
	self.server.conn.Close()
}

func (self *Client) login(user_login *types.TUserLogin) error {
	login_req := user.UserLoginRequest{
		Username:     user_login.Username,
		Password:     user_login.Password,
		TestPassword: user_login.TestCode,
	}

	json_data, err := json.Marshal(login_req)
	if err != nil {
		return err
	}

	url := server_url + "/user/login"
	// Log the url
	log.Println(url, string(json_data))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json_data))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json")

	resp, err := self.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return fmt.Errorf("%s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result struct {
        Message string      `json:"message"`
        User    user.User   `json:"user"`
        Tests   []TEST.Test `json:"tests"`
    }
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	self.jwt = "TODO"
    self.currentUser = &result.User
    self.availableTests = result.Tests

	return nil
}

func (self *Client) maintainConn() {
	
	for {
		ctx, close := context.WithCancel(context.Background())

		err := self.connect(ctx, close)

		if err != nil {
			log.Println(err)
			close()
		}

		// block till connection breaks
		<-ctx.Done()

		msg := "server disconnected. trying reconnection in 5 seconds..."
		self.notifyErr(fmt.Errorf(msg))
		log.Println(msg)
		select {
		case <-self.exit.ctx.Done():
			log.Println("terminating connection with server")
			return
		case <-time.After(time.Second * 5):
			continue
		}
	}
}

func (self *Client) connect(exit context.Context, cancel context.CancelFunc) error {
	url, err := url.Parse(server_url)
	if err != nil {
		return err
	}
	url.Scheme = "ws"
	url.Path = "/ws"

	header := http.Header{}
	header.Add("Authorization", "Bearer "+self.jwt)

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), header)
	if err != nil {
		return err
	}
	self.server.conn = conn

	go func() {
		defer cancel()
		for {
			select {
			case <-exit.Done():
				return
			case msg, ok := <-self.server.send:
				if !ok {
					return
				}
				log.Println(msg)
				self.server.conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
				err := self.server.conn.WriteJSON(msg)
				if err != nil {
					log.Println(err)
					return
				}

			}
		}
	}()
	go func() {
		defer cancel()
		for {
			var msg types.Message
			err := self.server.conn.ReadJSON(&msg)
			if err != nil {
				log.Println(err)
				return
			}
			self.server.recv <- msg
		}
	}()

	return nil
}

func (self *Client) getTest(testData types.TGetTest) (TEST.Test, error) {
	test_code := testData.TestPassword
	url := server_url + "/test/get_question_paper/" + test_code
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return TEST.Test{}, err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("test_code", test_code)
	req.Header.Set("Authorization", "Bearer "+self.jwt)

	resp, err := self.client.Do(req)
	if err != nil {
		return TEST.Test{}, err
	}
	defer resp.Body.Close()

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return TEST.Test{}, fmt.Errorf("%s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TEST.Test{}, err
	}

	var result struct {
		Message  string    `json:"message"`
		Response TEST.Test `json:"response"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return TEST.Test{}, err
	}

	return result.Response, nil
}
