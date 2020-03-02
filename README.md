# bf_ssh
Simple SSH brute-forcer in Go❤️

1. Get it!
```
go get https://github.com/jfagoagas/bf_ssh
```

2. Use it!
```
bf_ssh --help
```
```
Usage:
  -L string
        List of host in format IP:Port
  -P string
        List of passwords, one per line
  -U string
        List of users, one per line
  -l string
        Host in format IP:Port
  -o string
        Output file
  -p string
        Password
  -t duration
        SSH Dial Timeout (default 300ms)
  -u string
        Username

Modes:
Single Mode --> bf_ssh -L  <host-list> -U <user-list> -P <pass-list>
Multi Mode  --> bf_ssh -l <host> -u <user> -p <pass>
Note: options -t <500ms> and -o <out-file> are optional
```
## TO-DO
- [ ] Implement restore file
- [ ] Sorted output file (only succesful ones)
- [ ] Login and password combo file (login:passwd)
- [ ] Test ssh --> host does not support password auth
- [ ] Check split & joins ip:port
- [ ] Test host is up --> once per login or only one per host
