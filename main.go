package main

import (
	"bufio"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"go.bug.st/serial"
)

type Message struct {
	msg  string
	user string
}

func (m Message) serialize() string {
	return fmt.Sprintf("@b@%s|%s", m.user, m.msg)
}
func (m Message) deserialize(data string) Message {
	rawdata := strings.TrimSuffix(strings.TrimPrefix(data, "@b@"), "\n")
	splitData := strings.Split(rawdata, "|")
	id := splitData[0]
	msg := splitData[1]
	return Message{msg: msg, user: id}
}

func main() {
	fmt.Println("Hello, World!")
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		com := template.HTML(portFinder())
		tmpl := template.Must(template.ParseFiles("test.html"))

		tmpl.Execute(w, map[string]interface{}{"com": com})
	}
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		port := r.URL.Query().Get("port")
		if port == "" {
			fmt.Println("No port specified")
			return
		}
		ID := r.FormValue("ID")
		MSG := r.FormValue("MSG")
		msg := Message{msg: MSG, user: ID}
		writeToSerial(port, msg)
		tmppl, _ := template.New("h").Parse("<div>Message sent</div>")
		tmppl.Execute(w, nil)

	})
	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		port := r.URL.Query().Get("port")
		if port == "" {
			fmt.Println("No port specified")
			return
		}
		fmt.Print("Port: ")
		fmt.Println(port)
		response := readFromSerial(port)
		htmlstr := response
		tmpl, _ := template.New("h").Parse(htmlstr)
		tmpl.Execute(w, nil)
	})
	http.HandleFunc("/ports", func(w http.ResponseWriter, r *http.Request) {
		response := portFinder()
		htmlstr := response
		tmpl, _ := template.New("h").Parse(htmlstr)
		tmpl.Execute(w, nil)
	})
	http.HandleFunc("/", helloHandler)
	fmt.Println("server running on 'http://localhost:8080'")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func portFinder() string {
	ports, _ := serial.GetPortsList()
	if ports == nil {
		return "No ports found"
	}
	list := "<select class = 'box' id='port-select'multiple>"

	for _, port := range ports {
		list += fmt.Sprintf("<option value='%s'> %s</option>", port, port)
		fmt.Println(port)
	}
	list += "</select>"
	return fmt.Sprint(list)
}

func readFromSerial(com string) string {
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	port, err := serial.Open(com, mode) // Use COM3 port
	if err != nil {
		fmt.Println("Error opening port: ", err)
		return "Error opening port"
	}
	defer port.Close()
	fmt.Println("Connected to ESP32")

	port.SetReadTimeout(5 * time.Second)

	_, err = port.Write([]byte("request\n"))
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(port)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messageChannel := make(chan string)
	errorChannel := make(chan string)

	go func() {
		for {
			n, err := reader.ReadString('\n')
			if err != nil {
				if err.Error() == "timeout" {
					errorChannel <- "timeout"
					return
				}
			}
			if len(n) > 0 {
				fmt.Println(n)
			}
			if strings.HasPrefix(n, "@b@") {
				msg := Message{}.deserialize(n)

				fmt.Println(msg.msg)
				fmt.Println(msg.user)
				messageChannel <- fmt.Sprintf("<div>ID: %s<br/>MSG: %s</div>", msg.user, msg.msg)
				return

			}
		}
	}()

	select {
	case <-ctx.Done():
		fmt.Println("Timeout")
		return "Timeout"
	case msg := <-messageChannel:

		return msg
	case err := <-errorChannel:
		if err == "timeout" {
			fmt.Println("Timeout")
			return "Timeout"
		}
		log.Fatal(err)
		return err
	}
}
func writeToSerial(com string, msg Message) {
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	port, err := serial.Open(com, mode) // Use COM3 port
	if err != nil {
		fmt.Println("Error opening port: ", err)
		return
	}
	defer port.Close()
	fmt.Println("Connected to ESP32")
	_, err = port.Write([]byte("post\n"))
	if err != nil {
		log.Fatal(err)
	}
	ser := msg.serialize()
	fmt.Println(ser)
	_, err = port.Write([]byte(ser))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Message sent")
	return
	/* 	reader := bufio.NewReader(port)
	   	for {
	   		n, err := reader.ReadString('\n')
	   		if err != nil {
	   			log.Fatal(err)
	   		}
	   		if len(n) > 0 {
	   			fmt.Println(n)
	   		}
	   	}
	*/
}
