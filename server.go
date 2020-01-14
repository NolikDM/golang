package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
)

func get_session_key() string {
	b := ""
	for i := 0; i < 10; i++ {
		b += strconv.Itoa(rand.Intn(9) + 1)
	}
	return b
}

func get_hash_str() string {
	li := ""
	for i := 0; i < 5; i++ {
		li += strconv.Itoa(rand.Intn(10))
	}
	return li
}

type Session_protector struct {
	//struct used to protect web services from unauthorized access
	__hash string
}

func (self Session_protector) __calc_hash(session_key string, val int) string {
	//calculate hash
	switch val {
	case 1:
		result := ""
		ret := ""
		for idx := 0; idx < 5; idx++ {
			result += string(session_key[idx]) /*ЗДЕСЬ ОШИБКА*/
		}
		i, _ := strconv.Atoi(result)
		result = "00" + strconv.Itoa(i%97)
		for idx := len(result) - 2; idx < len(result); idx++ {
			ret += string(result[idx])
		}
		return ret
	case 2:
		result := ""
		for idx := 0; idx < len(session_key); idx++ {
			result += string(session_key[len(session_key)-idx-1])
		}
		return result
	case 3:
		return session_key[len(session_key)-5:] + session_key[0:5]
	case 4:
		num := 0
		for i := 1; i < 9; i++ {
			num += int(session_key[i]) + 41
		}
		return string(num)
	case 5:
		ch := ""
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
	default:
		result, _ := strconv.Atoi(session_key)
		return strconv.Itoa(result + val)
	}
}

func (self Session_protector) next_session_key(session_key string) string {
	result := 0
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
			// will listen for message to process ending in newline (\n)
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

				// output message received
				fmt.Println("Message from client ( id = ", id, " ) Received: ", string(text), "key: ", key1)
				// sample process for string received
				newmessage := strings.ToUpper(text)
				key2 = server_protector.next_session_key(strings.Replace(key1, "\n", "", -1))
				fmt.Print("New key: ", key2, "\n")
				// send new string back to client
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
	n := flag.Int("n", 100, "a number of simultaneous connections")
	flag.Parse()
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
}
