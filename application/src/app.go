package main

import (
	types "common"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type App struct {
	send   chan types.Message
	recv   chan types.Message
	runner IRunner
	client *Client

	exitCtx context.Context
	exitFn  context.CancelFunc

	state struct {
		webview_opened     bool
		connection_started bool
	}
}

func (self *App) destroy() {
	close(self.send)
	self.client.destroy()

	err := self.runner.RestoreEnv()
	log.Println(err)
}

func newApp() (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())
	app := &App{
		send:    make(chan types.Message, 100),
		recv:    make(chan types.Message, 100),
		exitCtx: ctx,
		exitFn:  cancel,
	}
	var err error

	client, err := newClient(app.send)
	if err != nil {
		return nil, err
	}
	app.client = client

	app.runner, err = NewRunner(app.send)
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (self *App) login(user_login *types.TUserLogin) error {
	err := self.client.login(user_login)
	if err != nil {
		errorMessage := types.NewMessage(types.TErr{
			Message: "Failed to log in user: " + err.Error(),
		})
		self.send <- errorMessage
		return err
	}

	return nil
}


func KillProcess(pid uint32) error {
	cmd := exec.Command("taskkill", "/PID", strconv.Itoa(int(pid)), "/F")
	return cmd.Run()
}


func (self *App) maintainConnection() {
	if self.state.connection_started {
		return
	}
	self.state.connection_started = true
	processes, err := self.runner.ListAllProcess()
	if err != nil {
		fmt.Printf("Error listing processes: %v\n", err)
		return
	}

	fmt.Println("Running Processes (Visible on Taskbar):")
	for pid, windowText := range processes {
		fmt.Printf("PID: %d, Window Title: %s\n", pid, windowText)
	}

	appsToKill := []string{"Chrome", "Firefox", "Brave"}

	for pid, cmdline := range processes {
		for _, app := range appsToKill {
			if strings.Contains(cmdline, app) {
				fmt.Printf("Killing process %d (%s)\n", pid, cmdline)
				if err := KillProcess(pid); err != nil {
					fmt.Printf("Error killing process %d: %v\n", pid, err)
				}
			}
		}
	}

	go self.client.maintainConn()
	go self.handleServerMessages()
}


func (self *App) handleServerMessages() {
	for {
		var msg types.Message
		var ok bool
		select {
		case <-self.exitCtx.Done():
			return
		case msg, ok = <-self.client.server.recv:
			if !ok {
				return
			}
		}

		switch msg.Typ {
		default:
			log.Printf("message type '%s' not handled ('%s')\n", msg.Typ.TSName(), msg.Val)
		}
	}
}

func (self *App) startTest(testData types.TGetTest) error {
	questionPaper, err := self.client.getTest(testData)

	if err != nil {
		return err
	}

	log.Println("Question paper: ", questionPaper)

	routeMessage := types.TLoadRoute{
		Route: "/tests/1",
	}
	message := types.NewMessage(routeMessage)

	self.send <- message

	return nil
}

func (self *App) openWv() {
	var url string
	if build_mode == "DEV" {
		url = fmt.Sprintf("http://localhost:%s/", os.Getenv("DEV_PORT"))
	} else {
		url = fmt.Sprintf("http://localhost:%s/", port)
	}
	self.state.webview_opened = true

	go func() {
		uritaOpenWv(url)
		self.exitFn()
	}()
}

// must be called from the main thread :/
func (self *App) wait() {
	<-self.exitCtx.Done()
}

func (self *App) notifyErr(err error) {
	if err != nil {
		self.send <- types.NewMessage(types.TErr{
			Message: fmt.Sprintf("Error: %s", err),
		})
		log.Printf("Error: %s\n", err)
	}
}

func (self *App) prepareEnv() {
	err := self.runner.SetupEnv()
	self.notifyErr(err)
}
