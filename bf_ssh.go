/* Simple SSH Bruteforcing Tool*/

/*
TO-DO
- User wordlist
- Passwords wordlist
- Probar con canales
- Fichero de salida .txt o .csv
- Comprobacion de los multiples errores
- Para el listado de user y pass, si no encuentra nada que no pinte todos los intentos"
- Comprobar que el puerto es un entero entre 1 y 65535
*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"os"
	"time"
)

/* Flags Globales*/
var (
	host      = flag.String("l", "", "Single host in format IP:Port")
	list_host = flag.String("L", "", "List of host in format IP:Port")
	user      = flag.String("u", "", "SSH User")
	pwd       = flag.String("p", "", "SSH Password")
	// SSH Dial Timeout
	tmout = flag.Duration("t", 300*time.Millisecond, "Timeout")
	out   = flag.String("o", "", "Output file")
)

func main() {
	fmt.Printf("##########################################\n")
	fmt.Printf("######### GO SSH BRUTE -- v0.0.1 #########\n")
	fmt.Printf("##########################################\n")

	flag.Parse()
	if *user == "" && *pwd == "" && (*host == "" || *list_host == "") {
		fmt.Printf("\nERROR - Must complete input params\n")
		flag.PrintDefaults()
		fmt.Printf("Example: %s -H host-list.txt -u root -p T3mp0ra1 -t 500ms > output.txt\n\n", os.Args[0])
		os.Exit(1)
	}

    /* Timestamp */
    fmt.Printf("Date: %s", time.Now().Format("02.01.2006 15:04:05\n"))

	/* Si se pasa como parametro un listado de hosts */
	if *list_host != "" {
		/* Leemos la lista */
		hosts, err := readList(*list_host)
		if err != nil {
			log.Fatal("Can't read hosts file")
		}
		/* Recorremos el listado de hosts */
		timeS := *tmout
		for _, host := range hosts {
			/* Incrementamos el tiempo de espera por cada hilo */
			timeS += *tmout
			/* Direccion IP y puerto del host */
			ip, port, _ := net.SplitHostPort(host)
			/* Comprobamos los parametros */
			if net.ParseIP(ip) == nil || port == "" {
				fmt.Printf("Bad IP:Port Format -- %s:%s\n", ip, port)
			} else {
				/* Si todo es correcto lanzamos la conexion */
				go sshConn(ip, port)
			}
		}
		/* Dormimos el programa principal hasta que
		   acabe el proceso */
		time.Sleep(timeS)

	} else if *host != "" {
		/* Si se pasa como parametro un host */
		ip, port, _ := net.SplitHostPort(*host)
		if net.ParseIP(ip) == nil || port == "" {
			fmt.Printf("Bad IP:Port Format -- %s:%s %s\n", ip, port)
		} else {
			/* Llamamos al escaner */
			sshConn(ip, port)
		}
	}
}

func readList(list string) (lst []string, err error) {
	/* Abrimos la lista indicada */
	file, err := os.Open(list)
	if err != nil {
		return
	}
	defer file.Close()
	/* Leemos cada entrada de la lista */
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lst = append(lst, scanner.Text())
	}
	return
}

func sshIsUp(ip, port string) (isUp bool, addr string) {
	isUp = false
	/* Comprobamos si el servicio esta escuchando */
	addr = net.JoinHostPort(ip, port)
	_, err := net.DialTimeout("tcp", addr, *tmout)
	/* Si el puerto está cerrado */
	if err != nil {
		fmt.Printf("\n\033[1;91mFAILED --> Port %s/tcp is closed on %s\033[0m\n", port, ip)
	} else {
		/* Si el servicio en el puerto indicado esta a la escucha */
		isUp = true
	}
	return
}

func sshConn(ip, port string) {
	/* Comprobamos que el puerto esta a la escucha */
	isUp, addr := sshIsUp(ip, port)
	if isUp == true {
		/* SSH Client Config */
		config := &ssh.ClientConfig{
			User:            *user,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Auth:            []ssh.AuthMethod{ssh.Password(*pwd)},
			Timeout:         *tmout,
		}
		/* Configuramos los valores que no se hayan cumplimentado */
		config.SetDefaults()
		/* SSH Connection */
		_, err := ssh.Dial("tcp", addr, config)
		if err != nil {
			fmt.Printf("\n\033[1;91mFAILED --> host: %s   login: %s   password: %s\033[0m\n", addr, *user, *pwd)
		} else {
			fmt.Printf("\n\033[1;92mSUCCESS --> host: %s   login: %s   password: %s\033[0m\n", addr, *user, *pwd)
		}
	}
}
