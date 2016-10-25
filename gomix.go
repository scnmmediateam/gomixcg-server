/*
	Enter at own risk! Preferred intoxicant recommended

	Kako bo ime program? Semafor server?
	Premislite, dajte ideje. Semafor mi ni všeč.  /Dejan


*/

package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
	"os"
	"github.com/googollee/go-socket.io"
	"github.com/tarm/serial"
)

var comPort = "4"

//Small
var inputScoreboardS = "3"
var overlayScoreboardS = "2"

//Large
var inputScoreboardL = "2"
var overlayScoreboardL = "1"

var isRunning = true

var httpclient = http.Client{
	Timeout: (time.Duration(5 * time.Millisecond)),
}

//var casparIP = "127.0.0.1"

//Vmix config
var vmixUsed = false
var vmixIP = "127.0.0.1"
var vmixPort = "8088"
var vmixSocket = vmixIP + ":" + vmixPort

//Casper config
var casparUsed = true
var casperEnstablished = false
var casparIP = "88.200.86.101"
var casparPort = "5250"
var casparSocket = casparIP + ":" + casparPort
var casparConnection, _ = net.Dial("tcp", casparSocket)

var webServerPort = ":8081"

var nameHomeShort = ""
var nameGuestShort = ""

func HexParser() {
	c := &serial.Config{Name: comPort, Baud: 9600}
	s, err := serial.OpenPort(c)
	hexString := ""
	hexStringPointer := &hexString
	for 1 == 1 {
		if err == nil {
			buf := make([]byte, 128)
			n, err := s.Read(buf)

			if err == nil {
				hexString += hex.EncodeToString(buf[:n])
				hexStringPointer = &hexString
				commands := SplitIntoCommands(hexStringPointer)

				for i := 0; i < len(commands); i++ {
					command, err := hex.DecodeString(commands[i])
					if err == nil {
						ParseCommand(string(command))
					}
				}
			}
		}
	}
}

func main() {
	log.Println("builded")
	casparConnection, _ = net.Dial("tcp", casparSocket)
	casperEnstablished = ( casparConnection != nil)

	vmixSocket = vmixIP + ":" + vmixPort
	comPort = "COM" + comPort
	

	go HexParser()
	go WebServer()
	CommandLine()
	//fmt.Scanln()
	

	/*/ DEBUG HEX
	hexString := "d3 34 31 54 50 3a 31 30 3a 30 30 2e 20 20 2f 31 43 5f 00 d3 34 31 46 48 53 3a 30 30 42 73 00 d3 34 31 46 47 53 3a 30 30 42 72 00 d3 34 31 53 45 52 3a 48 30 2c 47 30 43 77 00 d3 34 31 41 54 3a 20 20 2f 52 42 48 00"
	hexString = strings.Replace(hexString, " ", "", -1)
	hexStringPointer := &hexString

	commands := SplitIntoCommands(hexStringPointer)

	for i := 0; i < len(commands); i++ {
		log.Println("command: " + commands[i])
		command, err := hex.DecodeString(commands[i])
		if err == nil {
			ParseCommand(string(command))
		}
	}
	//*/
}

func ByteArrayToAsciiString(c []byte) string {
	n := -1
	for i, b := range c {
		if b == 0 {
			break
		}
		n = i
	}
	return string(c[:n+1])
}

func WebServer() {
	socketServer, err := socketio.NewServer(nil)
	http.Handle("/socket.io/", socketServer)
	http.Handle("/", http.FileServer(http.Dir("./client")))


	log.Println("test")
	if err == nil {
		log.Println("test2")



		socketServer.On("connection", func(so socketio.Socket) {
			log.Println("Socket ima povezavo")

			so.On("vgasniGrafike", func(msg string) {
				log.Println("Grafike nerfed!")
				SendCommandCaspar("CG 2-3 STOP 1 \r\n")
			})
			so.On("prizgiGrafike", func(msg string) {
				log.Println("Grafike pls!")
				SendCommandCaspar("CG 2-3 ADD 1 \"uraNogometMalaZgoraj\" 1 \"<templateData>" +
					"<componentData id=\\\"imeKratkoDomaci\\\"><data id=\\\"text\\\" value=\\\"" + nameHomeShort + "\\\"/></componentData>" +
					"<componentData id=\\\"imeKratkoGosti\\\"><data id=\\\"text\\\" value=\\\"" + nameGuestShort + "\\\"/></componentData>" +
					"</templateData>\"\r\n")
			})
			so.On("vnosImeDomaciKratko", func(imeKratkoDomaciws string) {
				nameHomeShort=imeKratkoDomaciws
				log.Println("Vnasam: " + imeKratkoDomaciws)
				SendCommandCaspar("CG 2-3 UPDATE 1 \"<templateData>" +
					"<componentData id=\\\"imeKratkoDomaci\\\"><data id=\\\"text\\\" value=\\\"" + nameHomeShort + "\\\"/></componentData>" +
					"</templateData>\"\r\n")
			})
			so.On("vnosImeGostiKratko", func(imeKratkoGostiws string) {
				nameGuestShort=imeKratkoGostiws
				log.Println("Vnasam: " + nameGuestShort)

				SendCommandCaspar("CG 2-3 UPDATE 1 \"<templateData>" +
					"<componentData id=\\\"imeKratkoGosti\\\"><data id=\\\"text\\\" value=\\\"" + nameGuestShort + "\\\"/></componentData>" +
					"</templateData>\"\r\n")
			})
			so.On("disconnection", func() {
				log.Println("Pepe je izgubil povezavo!")
			})
		})
		socketServer.On("error", func(so socketio.Socket, err error) {
			log.Println("error: ", err)
		})


	log.Fatal(http.ListenAndServe(":8081", nil))
	} else {
		panic(err)
	}
}

func SplitIntoCommands(hexStringPointer *string) []string {
	hexString := *hexStringPointer
	startIndex := 0
	endIndex := 0
	command := ""
	commands := []string{}
	for i := 0; i < len(hexString); i++ {
		if i+5 < len(hexString) && string(hexString[i:i+6]) == "d33431" {
			startIndex = i
		}
		if startIndex != -1 && i < len(hexString)-1 && string(hexString[i:i+2]) == "00" {
			if i < len(hexString)-2 && string(hexString[i:i+3]) == "000" {
				endIndex = i + 1
			} else {
				endIndex = i
				command = string(hexString[startIndex+6 : endIndex])
				commands = append(commands, command)
				*hexStringPointer = string(hexString[0:startIndex]) + string(hexString[endIndex+2:len(hexString)])
				hexString = *hexStringPointer
				i = -1
				startIndex = -1
			}
		}
	}

	if startIndex != -1 {
		*hexStringPointer = string(hexString[startIndex:len(hexString)])
	}
	return commands
}

func ParseCommand(command string) {

	if len(command) > 2 {
		header := strings.Split(command, ":")[0]
		log.Println(header)
		if header == "TP" {
			min := string(command[3:5])
			sec := string(command[6:8])
			msec := string(command[9:11])
			period := string(command[12:13])
			minInt, err := strconv.Atoi(min)
			if err == nil {
				if minInt == 0 {
					if casparUsed {
						SendCommandCaspar("CG 2-3 UPDATE 1 \"<templateData>" +
							"<componentData id=\\\"cas\\\"><data id=\\\"text\\\" value=\\\"" + sec + ":" + msec + "\\\"/></componentData>" +
							"</templateData>\"\r\n")
					}
					if vmixUsed {
						go SendCommandVMIX("http://" + vmixSocket + "/api/?Function=SetText&Input=" + inputScoreboardL + "&SelectedName=cas&Value=" + sec + ":" + msec)
						go SendCommandVMIX("http://" + vmixSocket + "/api/?Function=SetText&Input=" + inputScoreboardS + "&SelectedName=cas&Value=" + sec + ":" + msec)
					}

				} else {
					if casparUsed {
						SendCommandCaspar("CG 2-3 UPDATE 1 \"<templateData>" +
							"<componentData id=\\\"cas\\\"><data id=\\\"text\\\" value=\\\"" + min + ":" + sec + "\\\"/></componentData>" +
							"</templateData>\"\r\n")
					}
					if vmixUsed {
						go SendCommandVMIX("http://" + vmixSocket + "/api/?Function=SetText&Input=" + inputScoreboardL + "&SelectedName=cas&Value=" + min + ":" + sec)
						go SendCommandVMIX("http://" + vmixSocket + "/api/?Function=SetText&Input=" + inputScoreboardS + "&SelectedName=cas&Value=" + min + ":" + sec)
					}
				}

				if period != "s" && period != "S" && period != "o" && period != "O" && period != "P" && period != "p" {
					if casparUsed {
						SendCommandCaspar("CG 2-3 UPDATE 1 \"<templateData>" +
							"<componentData id=\\\"perioda\\\"><data id=\\\"text\\\" value=\\\"" + period + "\\\"/></componentData>" +
							"</templateData>\"\r\n")
					}
					if vmixUsed {
						SendCommandVMIX("http://" + vmixSocket + "/api/?Function=SetText&Input=" + inputScoreboardL + "&SelectedName=cetrtina&Value=" + period)
						SendCommandVMIX("http://" + vmixSocket + "/api/?Function=SetText&Input=" + inputScoreboardS + "&SelectedName=cetrtina&Value=" + period)
					}
				}
			}
		} else if header == "R" {
			home, err := strconv.Atoi(string(command[3:6]))
			if err == nil {
				guests, err := strconv.Atoi(string(command[8:11]))
				if err == nil {
					if casparUsed {
						SendCommandCaspar("CG 2-3 UPDATE 1 \"<templateData>" +
							"<componentData id=\\\"rezultat\\\"><data id=\\\"text\\\" value=\\\"" + strconv.Itoa(home) + "-" + strconv.Itoa(guests) + "\\\"/></componentData>" +
							"</templateData>\"\r\n")
					}
					if vmixUsed {
						SendCommandVMIX("http://" + vmixSocket + "/api/?Function=SetText&Input=" + inputScoreboardL + "&SelectedName=rezultat&Value=" + strconv.Itoa(home) + "-" + strconv.Itoa(guests))
						SendCommandVMIX("http://" + vmixSocket + "/api/?Function=SetText&Input=" + inputScoreboardS + "&SelectedName=rezultat&Value=" + strconv.Itoa(home) + "-" + strconv.Itoa(guests))
					}
				}
			}
		} else if header == "AT" {
			attackTime, err := strconv.Atoi(string(command[3:5]))
			if err == nil && attackTime <= 10 {
				if vmixUsed {
					SendCommandVMIX("http://" + vmixSocket + "/api/?Function=SetText&Input=" + inputScoreboardL + "&SelectedName=napad&Value=" + strconv.Itoa(attackTime))
					SendCommandVMIX("http://" + vmixSocket + "/api/?Function=SetText&Input=" + inputScoreboardS + "&SelectedName=napad&Value=" + strconv.Itoa(attackTime))
				}

			} else if string(command[3:5]) == "  " {
				if vmixUsed {
					SendCommandVMIX("http://" + vmixSocket + "/api/?Function=SetText&Input=" + inputScoreboardL + "&SelectedName=napad&Value=")
					SendCommandVMIX("http://" + vmixSocket + "/api/?Function=SetText&Input=" + inputScoreboardS + "&SelectedName=napad&Value=")
				}
			}
		} else if header == "RT" {
			hr := string(command[3:5])
			min := string(command[6:8])
			sec := string(command[9:11])

			//logCommand := "Global time (hr:min) - " + hr + ":" + min


			if vmixUsed {
				go SendCommandVMIX("http://" + vmixSocket + "/api/?Function=SetText&Input=" + inputScoreboardL + "&SelectedName=cas&Value=" + hr + ":" + min)
				go SendCommandVMIX("http://" + vmixSocket + "/api/?Function=SetText&Input=" + inputScoreboardS + "&SelectedName=cas&Value=" + hr + ":" + min)
			}
			if casparUsed {
				go SendCommandCaspar("CG 2-3 UPDATE 1 \"<templateData>" +
							"<componentData id=\\\"cas\\\"><data id=\\\"text\\\" value=\\\"" + min + ":" + sec + "\\\"/></componentData>" +
							"</templateData>\"\r\n")
			}
		}
	}
}

func CommandLine() {
	for isRunning == true {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">>> ")
		command, _ := reader.ReadString('\r')
		commandSplit := strings.Split( command[0:len(command)-1], " ")
		switch string(commandSplit[0]) {
			case "exit":
				isRunning = false
				break
			case "vmix":
				//Turn vmix on
				if Contains(commandSplit,"-on"){
					casparUsed = true;
				}

				//Turn vmix off
				if Contains(commandSplit,"-off"){
					casparUsed = false;
				}

				//IP
				if Contains(commandSplit,"-ip") && IndexOf(commandSplit,"-ip")+1 < len(commandSplit) {
					vmixIP = commandSplit[IndexOf(commandSplit,"-ip")+1]
					vmixSocket = vmixIP + ":" + vmixPort
				}
				//port
				if Contains(commandSplit,"-port") && IndexOf(commandSplit,"-port")+1 < len(commandSplit) {
					vmixPort = commandSplit[IndexOf(commandSplit,"-port")+1]
					vmixSocket = vmixIP + ":" + vmixPort
				}

				//print config
				if Contains(commandSplit,"-config") {
					log.Println("--- VMIX CONFIG ---")
					log.Println("Is used: "+strconv.FormatBool(vmixUsed))
					log.Println("IP: "+vmixIP)
					log.Println("Port: "+vmixPort)
					log.Println("Socket: "+vmixSocket)
				}
				break
			case "caspar":
				//Turn caspar on
				if Contains(commandSplit,"-on"){
					casparUsed = true;
				}

				//Turn caspar off
				if Contains(commandSplit,"-off"){
					casparUsed = false;
				}
				//IP
				if Contains(commandSplit,"-ip") && IndexOf(commandSplit,"-ip")+1 < len(commandSplit) {
					casparIP = commandSplit[IndexOf(commandSplit,"-ip")+1]
					casparSocket = casparIP + ":" + casparPort
				}
				//port
				if Contains(commandSplit,"-port") && IndexOf(commandSplit,"-port")+1 < len(commandSplit) {
					casparPort = commandSplit[IndexOf(commandSplit,"-port")+1]
					casparSocket = casparIP + ":" + casparPort
				}

				//Connect to casper
				if Contains(commandSplit,"-connect") {
					casparConnection, _ = net.Dial("tcp", casparSocket)
					casperEnstablished = ( casparConnection != nil)
					if casparConnection != nil {
						log.Println("SUCCESS - Connection to caspar succesfully enstablished!")
					} else {
						log.Println("ERROR - Valid casper connection couldn't be enstablished!")
					}
				}
				//print config
				if Contains(commandSplit,"-config") {
					log.Println("--- CASPAR CONFIG ---")
					log.Println("Is used: "+strconv.FormatBool(casparUsed))
					log.Println("IP: "+casparIP)
					log.Println("Port: "+casparPort)
					log.Println("Socket: "+casparSocket)
				}
				break
		}
	}
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func IndexOf(s []string, e string) int {
	for i, a := range s {
		if a == e {
			return i
		}
	}
	return -1
}

func SendCommandVMIX(command string) {
	httpclient.Get(command)
}

func SendCommandCaspar(command string) {
	if( casperEnstablished ){
		fmt.Fprintf(casparConnection, command)
		status, _ := bufio.NewReader(casparConnection).ReadString('\n')
		log.Println(status) //načeloma vrne 202 OK, napiše pa tudi napako, če ni v redu ukaz, na žalost samo z 400,401,402...
	} else {
			log.Println("ERROR - Valid casper connection isn't enstablished")
	}
}
