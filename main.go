package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

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
	id := strings.Split(rawdata, "|")[0]
	msg := strings.Split(rawdata, "|")[1]
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
		ID := r.FormValue("ID")
		MSG := r.FormValue("MSG")
		htmlstr := fmt.Sprintf("<div><h1>ID: %s</h1><h1>MSG: %s</h1></div>", ID, MSG)
		tmpl, _ := template.New("h").Parse(htmlstr)
		tmpl.Execute(w, nil)
	})
	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		response := readFromSerial()
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
	list := "<ul>"

	for _, port := range ports {
		list += fmt.Sprintf("<li>%s</li>", port)
		fmt.Println(port)
	}
	list += "</ul>"
	return fmt.Sprint(list)
}

func readFromSerial() string {
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	port, err := serial.Open("COM3", mode) // Use COM3 port
	if err != nil {
		fmt.Println("Error opening port: ", err)
		return "Error opening port"
	}
	defer port.Close()
	fmt.Println("Connected to ESP32")
	_, err = port.Write([]byte("request\n"))
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(port)
	for {
		n, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if len(n) > 0 {
			fmt.Println(n)
		}
		if strings.HasPrefix(n, "@b@") {
			msg := Message{}.deserialize(n)

			fmt.Println(msg.msg)
			fmt.Println(msg.user)
			return fmt.Sprintf("<div>ID: %s<br/>MSG: %s</div>", msg.user, msg.msg)

		}
	}
}
