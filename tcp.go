package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings" // нужно для обработки образца
	"time"
)

func get_session_key() string {
	// генерируем 10 рандомных строк
	result := ""
	for i := 0; i < 10; i++ {
		result += string(strconv.Itoa(int(9*rand.Float64()) + 1)[0])
	}
	return result
}

func get_hash_str() string {
	li := ""
	for i := 0; i < 5; i++ {
		li += strconv.Itoa(int(int((6 * rand.Float64()) + 1)))
	}
	return li
}

type Session_protector struct {
	// Структура, используемая для защиты web-служб от несанкционированного доступа
	__hash string
}

func (self Session_protector) __calc_hash(session_key string, val int) string {
	//calculate hash
	if val == 1 {
		result := ""
		ret := ""
		for idx := 0; idx < 5; idx++ {
			result += string(session_key[idx]) // здесь ошибка
		}
		i, _ := strconv.Atoi(result)
		result = "00" + strconv.Itoa(i%97)
		for idx := len(result) - 2; idx < len(result); idx++ {
			ret += string(result[idx])
		}
		return ret
	}
	if val == 2 {
		result := ""
		for idx := 0; idx < len(session_key); idx++ {
			result += string(session_key[len(session_key)-idx-1])
		}
		return result
	}
	if val == 3 {
		result := ""
		ret := ""
		for idx := 0; idx < 5; idx++ {
			result += string(session_key[idx]) // здесь ошибка
		}
		for idx := 5; idx < len(session_key); idx++ {
			ret += string(session_key[idx])
		}
		return ret + result
	}
	if val == 4 {
		result := 0
		for idx := 1; idx < 8; idx++ {
			num, _ := strconv.Atoi(string(session_key[idx])) // здесь ошибка
			result += num + 41
		}
		return strconv.Itoa(result)
	}
	if val == 5 {
		var ch string
		result := 0
		for idx := 0; idx < len(session_key); idx++ {
			ch = string(int(int(session_key[idx]) ^ 43))
			if _, err := strconv.Atoi(ch); err != nil {
				ch = string(int(ch[0]))
			}
			num, _ := strconv.Atoi(ch)
			result += num
		}
		return strconv.Itoa(result)
	}
	result, _ := strconv.Atoi(session_key)
	return strconv.Itoa(result + val)
}

func (self Session_protector) next_session_key(session_key string) string {
	// генерируем следующий сеансовый ключ
	// проверка хэша
	if self.__hash == "" {
		fmt.Println("hash is empty")
		return get_session_key()
	}
	for idx := 0; idx < len(self.__hash); idx++ {
		i := string(self.__hash[idx])
		if _, err := strconv.Atoi(i); err != nil {
			fmt.Println("Here is letter")
			return get_session_key()
		}
	}
	result := 0
	ret := ""
	for idx := 0; idx < len(self.__hash); idx++ {
		num, _ := strconv.Atoi(string(self.__hash[idx]))
		k, _ := strconv.Atoi(self.__calc_hash(session_key, num))
		result += k
	}
	for idx := 0; idx < 10 && idx < len(strconv.Itoa(result)); idx++ {
		ret += string((strconv.Itoa(result))[idx])
	}
	m := ""
	ret = "0000000000" + ret
	for idx := len(ret) - 10; idx < len(ret); idx++ {
		m += string(ret[idx])
	}
	return m
}

func run_connection(conn *net.Conn, id int, point *int) {
	// run loop forever (or until ctrl-c)
	text, serr := bufio.NewReader(*conn).ReadString('\n')
	if serr == nil {
		serv_hash_string := ""
		key1 := ""
		for i := 0; i < 5; i++ {
			serv_hash_string += string(text[i])
		}
		for i := 5; i < 15; i++ {
			key1 += string(text[i])
		}
		fmt.Println(serv_hash_string, key1)
		server_protector := Session_protector{strings.Replace(serv_hash_string, "\n", "", -1)}
		key2 := server_protector.next_session_key(key1)
		(*conn).Write([]byte(key2 + "\n"))
		for {
			// будем прослушивать все сообщения разделенные \n
			message, err := bufio.NewReader(*conn).ReadString('\n')
			if err == nil {
				key1 = ""
				text = ""
				for i := len(message) - 11; i < len(message)-1; i++ {
					key1 += string(message[i])
				}
				for i := 0; i < len(message)-11; i++ {
					text += string(message[i])
				}
				// печатаем полученное сообщение
				fmt.Println("Message from client ( id = ", id, " ) Received: ", string(text), "key: ", key1)
				// процесс выборки для полученной строки
				newmessage := strings.ToUpper(text)
				key2 = server_protector.next_session_key(strings.Replace(key1, "\n", "", -1))
				fmt.Print("New key: ", key2, "\n")
				// отправить новую строку обратно клиенту
				(*conn).Write([]byte(newmessage + key2 + "\n"))
			} else {
				(*conn).Close()
				*point -= 1
				fmt.Println("Client ( id =", id, ") Disconnected!")
				break
			}
		}
	} else {
		(*conn).Close()
		*point -= 1
		fmt.Println("Client ( id =", id, ") Disconnected!")
	}
}

func main() {
	port := flag.String("port", ":8081", "a server listening port")
	IP := flag.String("ip:port", "", "a client connection port")
	n := flag.Int("n", 100, "a number of simultaneous connections")
	flag.Parse()
	if *IP == "" {
		fmt.Println("Launching server...")
		var id = 1
		ln, _ := net.Listen("tcp", *port)
		point := 1
		for {
			conn, _ := ln.Accept()
			if point <= *n {
				point += 1
				fmt.Println("New client ( id =", id, ") Connected!")
				go run_connection(&conn, id, &point)
				id += 1
			} else {
				conn.Close()
			}
		}
	} else {
		rand.Seed(time.Now().UnixNano())
		conn, err := net.Dial("tcp", *IP)
		if err != nil {
			fmt.Println("Server not found. Try later again.")
			for {
			}
		} else {
			cl_hash_string := get_hash_str()
			key1 := get_session_key()
			fmt.Print(cl_hash_string + "\n")
			fmt.Fprintf(conn, cl_hash_string+key1+"\n")
			client_protector := Session_protector{cl_hash_string}
			key2, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Println("Server closed connection. Try later again.")
				// for {
				// }
			}
			key1 = client_protector.next_session_key(key1)
			key1 = client_protector.next_session_key(key1)
			for {
				fmt.Print("Text to send: ")
				text := ""
				// отправляем в сокет
				fmt.Fprintf(conn, strings.Replace(text, "\n", "", -1)+key1+"\n")
				// ждем ответа
				fmt.Println("Waiting for answer...")
				message, err := bufio.NewReader(conn).ReadString('\n')
				if err != nil {
					fmt.Println("Server closed connection. Try later again.")
					for {
					}
				}
				key2 = ""
				text = ""
				for i := len(message) - 11; i < len(message)-1; i++ {
					key2 += string(message[i])
				}
				for i := 0; i < len(message)-11; i++ {
					text += string(message[i])
				}
				key1 = client_protector.next_session_key(key1)
				fmt.Println("Message from server: "+text, "key: ", key1, " ", key2)
				key1 = client_protector.next_session_key(key1)
			}
		}
	}
}
