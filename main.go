package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
)

// JSON 데이터를 받을 구조체 정의
type Message struct {
	Command string `json:"command"`
	Path    string `json:"path"`
	Filter  string `json:"filter"`
}

var AIBoothConn net.Conn

func main() {
	//PhotoBoothAppStart()
	fmt.Println("SERVER START")
	// TCP 프로토콜을 사용하여 1998 포트에서 서버를 시작합니다.
	listener, err := net.Listen("tcp", ":3001")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()

	for {
		// 클라이언트의 연결을 대기하고, 연결이 수립되면 처리합니다.
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err.Error())
			return
		}
		// 클라이언트와 연결된 소켓에서 데이터를 읽어옵니다.
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	log.Println("New conneciton : ", conn.LocalAddr().String())
	for {
		// 버퍼 생성
		buf := make([]byte, 1024)

		// 소켓으로부터 데이터를 읽어옵니다.
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed by client:", conn.RemoteAddr())
			} else {
				fmt.Println("Error:", err)
			}
			return
		}

		PacketHandler(buf, n, conn)
	}
}

func PacketHandler(buf []byte, n int, conn net.Conn) {
	log.Println("Packet Receved:", string(buf))

	if string(buf)[:9] == "python_o_" { //python_메시지 패킷이라면
		log.Println("Python success : ", string(buf)[9:])
		AIBoothConn.Write([]byte("ai_converting_finished"))
		return
	} else if string(buf)[:9] == "python_x_" {
		log.Println("Python fail error message : ", string(buf)[9:])
		AIBoothConn.Write([]byte("ai_converting_Fail_" + string(buf)[9:]))
		return
	}

	// PhotoPacket 형식 json Unmarshal 시도
	var message Message
	err := json.Unmarshal(buf[:n], &message)
	if err != nil {
		fmt.Println("Error decoding JSON:", err.Error())
		return
	} else {
		AIBoothConn = conn
		if PhotoPacket(message) {
			fmt.Println("AI Convert Python Finished")
			conn.Write([]byte("ai_converting_finished"))
			AIPythonSuccesss()
		} else {
			fmt.Println("AI Convert Python fail")
			conn.Write([]byte("ai_converting_Fail_AI Convert Python fail"))
		}
	}
}

func PhotoPacket(message Message) bool {
	fmt.Println("PhotoPacket : ", message)

	result := false
	//python workflow_api.py --input-directory E:\outputs --output-directory E:\outputs
	switch message.Filter {
	case "1": //Beauty Filter
		result = PythonCode("beauty_workflow.py")
	case "2": //Figure Filter
		result = PythonCode("figure_workflow.py")
	case "3": //PopAnimation Filter
		result = PythonCode("pop_workflow.py")
	case "4": // Age Filter
		result = PythonCode("workflow_api.py")
	case "5":
		result = PythonCode("workflow_api.py")
	case "6":
		result = PythonCode("workflow_api.py")
	default:
		log.Println("Unknown Filter", message.Filter)
		result = PythonCode("workflow_api.py")
	}
	return result
}

func PythonCode(message string) bool {
	cmd := exec.Command("python", message)

	// 파이썬 스크립트의 표준 출력 파이프 설정
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error creating stdout pipe:", err)
		return false
	}

	// 파이썬 스크립트 실행
	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting command:", err)
		return false
	}

	// 파이썬 스크립트의 출력을 실시간으로 읽어 화면에 표시합니다.
	go func() {
		reader := bufio.NewReader(stdout)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("Error reading command output:", err)
				return
			}
			fmt.Print(line)
		}
	}()

	// 파이썬 스크립트의 실행을 기다립니다.
	if err := cmd.Wait(); err != nil {
		fmt.Println("Error waiting for command:", err)
		return false
	}
	fmt.Println("Command execution completed.")
	return true
}

func PhotoBoothAppStart() {
	go func() {
		exePath := "C:\\PPRK\\pprk.exe"
		cmd := exec.Command(exePath)
		err := cmd.Run()
		if err != nil {
			fmt.Println("PhotoBooth App Err:", err)
		} else {
			fmt.Println("PhotoBooth App Start")
		}
	}()
}

func SendPacketToBooth(pkt string) {
	if AIBoothConn != nil {
		AIBoothConn.Write([]byte(pkt))
	} else {
		log.Println("No Booth Conn")
	}
}

func AIPythonSuccesss() {

}
