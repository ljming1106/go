package main

import fmt

func echo(c net.Conn, shout string, delay time.Duration){
    fmt.Fprintln(c,"\t",strings.ToUpper(shout))
    time.Sleep(delay)
    fmt.Fprintln(c,"\t",shout)
    time.Sleep(delay)
    fmt.Fprintln(c,"\t",strings.ToLower(shout))
}

func handleConn(c net.Conn){
	input := butio.NewScanner(c)
	for input.Scan(){
		echo(c,input.Text(),1*time.Second)
	}
	// 注意：忽略input.Err()中可能的错误
	c.Close()
}
