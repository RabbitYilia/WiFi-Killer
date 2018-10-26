package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/satori/go.uuid"
)

var Result map[string]string
var Process map[string]*exec.Cmd
var StationMap map[string]string
var ClientMap map[string]string
var ModeMap map[string]string

func main() {
	/*Init Service*/
	Process = make(map[string]*exec.Cmd)
	StationMap = make(map[string]string)
	ClientMap = make(map[string]string)
	ModeMap = make(map[string]string)
	Result = make(map[string]string)
	router := httprouter.New()
	router.GET("/", index)
	router.GET("/list", list)
	router.GET("/start", start)
	router.GET("/scan", scan)
	router.POST("/attack", attack)
	router.POST("/View", View)
	http.ListenAndServe(":8088", router)
}

func list(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	HTML := `
<!DOCTYPE html>
<html>
    <head>
        <meta charset="UTF-8">
		<meta http-equiv="refresh" content="5">
        <title>WiFi Killer-Job Add</title>
    </head>
    <body>
	<table>
	<tr>
	<td>TaskID</td><td>StationMAC</td><td>ClientMAC</td><td>State</td><td>View</td><td>STOP</td>
	</tr>
	`
	for TaskID, cmd := range Process {
		if TaskID == "0" {
			continue
		}
		HTML += "<tr>"
		State := "Run"
		_, err := cmd.StdoutPipe()
		if err != nil {
			State = "Stop"
		}
		HTML += "<td>" + TaskID + "</td><td>" + StationMap[TaskID] + "</td><td>" + ClientMap[TaskID] + "</td><td>" + State + "</td><td>" + createFuncButton(TaskID, "View") + "</td><td>" + createFuncButton(TaskID, "Delete") + "</td>"
		HTML += "</tr>"
	}
	HTML += `
	    </table>
		<a href="/">
        <button>Home Page</button>
        </a>
    </body>
</html>`
	fmt.Fprint(w, HTML)
}

func View(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	err := r.ParseForm()
	if err != nil {
		w.Header().Set("Content-Type", "application/text; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error Bad Request")
		return
	}

	fmt.Fprint(w, `
<!DOCTYPE html>
<html>
    <head>
        <meta charset="UTF-8">
        <title>WiFi Killer - Result</title>
		<meta http-equiv="refresh" content="5">
    </head>
    <body>
	    <p>`+Result[r.Form.Get("TaskID")]+`</p>
		<a href="/start">
        <button>Home Page</button>
        </a>
    </body>
</html>`)
}

func UpdateResult(TaskID string, reader *bufio.Reader) {
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		Result[TaskID] += line
	}
}
func attack(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	err := r.ParseForm()
	if err != nil {
		w.Header().Set("Content-Type", "application/text; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error Bad Request")
		return
	}
	TaskID := r.Form.Get("TaskID")
	ModeMap[TaskID] = r.Form.Get("Mode")
	StationMap[TaskID] = r.Form.Get("StationMAC")
	ClientMap[TaskID] = r.Form.Get("ClientMAC")
	cmd := exec.Command("ping")
	Process[TaskID] = cmd
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Start()
	reader := bufio.NewReader(stdout)
	go UpdateResult(TaskID, reader)
	HTML := `
<!DOCTYPE html>
<html>
    <head>
        <meta charset="UTF-8">
		<meta http-equiv="refresh" content="5">
		<meta http-equiv="refresh" content="5;url=/list"> 
        <title>WiFi Killer-Job Add</title>
    </head>
    <body>
	    <p>Job Add Successfully,Auto jump in 5 sec</p>
		<a href="/list">
        <button>Check Job</button>
        </a>
    </body>
</html>`
	fmt.Fprint(w, HTML)
}
func scan(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	HTML := `
<!DOCTYPE html>
<html>
    <head>
        <meta charset="UTF-8">
        <title>WiFi	 Killer - Scan</title>
		<meta http-equiv="refresh" content="5">
    </head>
    <body>`
	_, ok := Process["0"]
	if !ok {
		cmd := exec.Command("ping", "127.0.0.1", "-t")
		Process["0"] = cmd
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			HTML += `
	    <p>Scan Failed,Device is not ready.</p>
		<a href="/">
        <button>Home Page</button>
        </a>
    </body>
</html>`
			fmt.Fprint(w, HTML)
			log.Fatal(err)
		}
		cmd.Start()
		reader := bufio.NewReader(stdout)
		go UpdateResult("0", reader)
		HTML += `<p>Preparing,Please Wait</p>
    </body>
</html>`
		fmt.Fprint(w, HTML)
	} else {
		Result["0"] += ""
		StationMAC := "00:00:00:00:00:00"
		ClientMAC := "FF:FF:FF:FF:FF:FF"
		HTML += "<p>" + StationMAC + "|" + ClientMAC + "</p>\n"
		HTML += "<table>\n<tr>"
		HTML += "<td>" + createScanItemButton(StationMAC, ClientMAC, "0|Deauthentication") + "</td>"
		HTML += "<td>" + createScanItemButton(StationMAC, ClientMAC, "3|ARP replay") + "</td>"
		HTML += "</tr>\n</table>"
		HTML += `
    </body>
</html>`
		fmt.Fprint(w, HTML)
	}
}

func createFuncButton(TaskID string, Func string) string {
	uid, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}
	HTML := "<form id=\"" + uid.String() + "\" method=\"post\" action=\"/" + Func + "\">"
	HTML += "<input type=\"hidden\" name=\"TaskID\" value=\"" + TaskID + "\" />"
	HTML += "<button onclick=\"document.getElementById('" + uid.String() + "').submit();\">" + Func + "</button>"
	HTML += "</form>"
	return HTML
}

func createScanItemButton(StationMAC string, ClientMAC string, Mode string) string {
	uid, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}
	HTML := "<form id=\"" + uid.String() + "\" method=\"post\" action=\"/attack\">"
	HTML += "<input type=\"hidden\" name=\"TaskID\" value=\"" + uid.String() + "\" />"
	HTML += "<input type=\"hidden\" name=\"StationMAC\" value=\"" + StationMAC + "\" />"
	HTML += "<input type=\"hidden\" name=\"ClientMAC\" value=\"" + ClientMAC + "\" />"
	HTML += "<input type=\"hidden\" name=\"Mode\" value=\"" + strings.Split(Mode, "|")[0] + "\" />"
	HTML += "<button onclick=\"document.getElementById('" + uid.String() + "').submit();\">" + Mode + "</button>"
	HTML += "</form>"
	return HTML
}
func index(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	fmt.Fprint(w, `
<!DOCTYPE html>
<html>
    <head>
        <meta charset="UTF-8">
        <title>WiFi Killer</title>
    </head>
    <body>
	    <p>Press Button to Start</p>
		<a href="/start">
        <button>Start</button>
        </a>
    </body>
</html>`)
}

func start(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	HTML := `
<!DOCTYPE html>
<html>
    <head>
        <meta charset="UTF-8">
        <title>WiFi Killer - Start</title>
    </head>
    <body>`
	cmd := exec.Command("ping", "127.0.0.1")
	buf, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
		HTML += `
	    <p>Start Failed,Device is not ready.</p>
		<a href="/">
        <button>Home Page</button>
        </a>
    </body>
</html>`
		fmt.Fprint(w, HTML)
		os.Exit(1)
	}
	result := string(buf)
	result += ""
	if true {
		HTML += `
	    <p>Already Start</p>
		<a href="/scan">
        <button>Scan</button>
        </a>
    </body>
</html>`
	} else {
		cmd = exec.Command("ping", "127.0.0.1")
		buf, err = cmd.Output()
		if err != nil {
			log.Fatal(err)
			HTML += `
	    <p>Start Failed,Device is not ready.</p>
		<a href="/">
        <button>Home Page</button>
        </a>
    </body>
</html>`
			fmt.Fprint(w, HTML)
			os.Exit(1)
		}
		result = string(buf)
		result += ""
		if true {
			HTML += `
	    <p>Start Success</p>
		<a href="/scan">
        <button>Scan</button>
        </a>
    </body>
</html>`
		} else {
			HTML += `
	    <p>Start Failed,Device is not ready.</p>
		<a href="/">
        <button>Home Page</button>
        </a>
    </body>
</html>`
			fmt.Fprint(w, HTML)
			os.Exit(1)
		}
	}
	fmt.Fprint(w, HTML)
}
