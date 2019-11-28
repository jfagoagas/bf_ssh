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
	host      = flag.String("l", "", "Single Host IP:Port")
	list_host = flag.String("L", "", "List of Host IP:Port")
	user      = flag.String("u", "", "SSH User")
	pwd       = flag.String("p", "", "SSH Password")
	// don't set timer too low, you may bypass the right password, for me it works with 150ms, some other systems needs more than 300ms.
	tmout = flag.Duration("t", 300*time.Millisecond, "SSH Timeout Dial Response (ex:300ms), don't set this too low")
	out   = flag.String("o", "", "Output file")
)

/* Canal de comunicaciion */
var channel = make(chan string)

func usage() {
	fmt.Printf("\nUsage: %s [-l HOST:PORT] [-L HOST LIST] [-u USER] [-p PASSWORD] [-t TMOUT] [-o OUTPUT FILE]\n", os.Args[0])
    fmt.Printf("Examples: %s -H host-list.txt -u root -p T3mp0ra1 -t 500ms > output.txt\n\n", os.Args[0])
	os.Exit(1)
}

func main() {
	fmt.Printf("##########################################\n")
	fmt.Printf("######### GO SSH BRUTE -- v0.0.1 #########\n")
	fmt.Printf("##########################################\n")
	fmt.Printf("Date: %s", time.Now().Format("02.01.2006 15:04:05\n"))

	flag.Parse()
	if *user == "" && *pwd == "" && (*host == "" || *list_host == "") {
		//fmt.Printf("\nERROR - Must complete input params\n")
		//flag.PrintDefaults()
        usage()
		os.Exit(1)
	}

	/* Fichero de salida */
    /*
	var outfile *os.File
	if *out == "" {
		outfile = os.Stdout
	} else {
		outfile, err = os.Create(*out)
		if err != nil {
			log.Fatal("Can't create file for writing, exiting.")
		}
		defer outfile.Close()
	}
    */

	/* Si se pasa como parametro un listado de hosts */
	if *list_host != "" {
		readHostList()
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

func readHostList() {
    /* Leemos la lista de host */
	file, err := os.Open(*list_host)
	if err != nil {
		log.Fatal(err)
	}
	/* Leemos cada entrada del fichero */
	scanner := bufio.NewScanner(file)
	timeS := *tmout
	for scanner.Scan() {
		timeS += *tmout
		/* Direccion IP y puerto del host */
		ip, port, _ := net.SplitHostPort(scanner.Text())
		/* Comprobamos los parametros */
		if net.ParseIP(ip) == nil || port == "" {
			fmt.Printf("Bad IP:Port Format -- %s:%s\n", ip, port)
		} else {
			go sshConn(ip, port)
		}
	}
	/* Dormimos el programa principal hasta que
	   acabe el proceso */
	time.Sleep(timeS)
	/* Cerramos el fichero */
	file.Close()
}

func sshIsUp(ip, port string) (bool, string) {
	isUp := false
    /* Comprobamos si el servicio esta escuchando */
	addr := net.JoinHostPort(ip, port)
	_, err := net.DialTimeout("tcp", addr, *tmout)
	/* Si el puerto está cerrado */
	if err != nil {
		fmt.Printf("\n\033[1;91mFAILED --> Port %s/tcp is closed on %s\033[0m\n", port, ip)
	} else {
		/* Si el servicio en el puerto indicado esta a la escucha */
		isUp = true
	}
	return isUp, addr
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
