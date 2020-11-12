/*
Author: Faris Alotaibi
Description: Backend of CRUD website for favorite albums. 
	-Handles all data transactions from the frontend. 

pass in --listen [PORTNUM]

Note: I added \n to the end of many returns due to how 
Scan() works. It hangs up if it doesn't detect an ending character
(or an empty line) so appending "\n" to returns sent back through
Fprint kept things moving.

*/

package main; 

import ("fmt"
		"flag"
		"net"
		"os"
		"bufio"
		"strings"
		"sync"
		)

//Custom struct for albums -- I use artist to denote 
//individual singers or full bands
type Album struct{
	Name string
	Artist string 
	Rating string 
	Comments string
}

var albumMap = map[string]*Album{}
var albumMutex sync.RWMutex


// This function handles sending all album info 
// to a frontend client.
func G_ALL(c net.Conn) {
	retstr:="{"
	albumMutex.RLock()
	for _, album:=range albumMap{
		retstr = retstr + fmt.Sprintf("[%s,%s,%s,%s],",
			     album.Name, album.Artist, album.Rating, album.Comments)		
	}
	albumMutex.RUnlock()
	retstr = retstr[:len(retstr)-1]
	retstr = retstr + "}\n"
	fmt.Fprint(c, retstr)
	c.Close()
	fmt.Println("G_ALL sent")
}

// This function handles sending all album 
// names to a client
func G_NAM(c net.Conn) {
	retstr:=""
	albumMutex.RLock()
	if len(albumMap) == 0{
		retstr = "-\n"
		fmt.Fprintf(c, retstr)
	}else{
		for _, album:= range albumMap{
			retstr = retstr + album.Name + ","
		}
		fmt.Fprint(c, retstr)
	}
	albumMutex.RUnlock()
	c.Close()
	fmt.Println("G_NAM sent")
}

// This function handles sending a specific
// album's information to a client
// album name -> send album data  
func G_ALB(name string, c net.Conn) {
	albumMutex.RLock()
	album := albumMap[name]
	albumMutex.RUnlock()
	retstr:=fmt.Sprintf("%s,%s,%s,%s", album.Name, album.Artist, album.Rating, album.Comments)
	fmt.Fprint(c, retstr)
	c.Close()
	fmt.Println("G_ALB sent")
}

// This function handles creating a new 
// album in the map with arguments 
// album information -> add to map 
func CREAT(vals string, c net.Conn) {
	albarr := strings.Split(vals,",")
	retstr:=""
	albumMutex.Lock()
	if albumMap[albarr[0]] != nil{
		retstr="-\n"
		fmt.Fprint(c, retstr)
	}else{
		albumMap[albarr[0]] = &Album{albarr[0], albarr[1], albarr[2], albarr[3]}
		retstr="\n"
		fmt.Fprint(c, retstr)
	}
	albumMutex.Unlock()
	c.Close()
	fmt.Println("CREAT sent")
}

// This function handles updating an albums 
// comments & rating 
// new info -> update info 
func UPDAT(vals string, c net.Conn) {
	updtarr := strings.Split(vals,",")
	albumMutex.Lock()
	album := albumMap[updtarr[0]]
	album.Rating = updtarr[1]
	album.Comments = updtarr[2]
	albumMutex.Unlock()
	fmt.Fprint(c, "\n")
	c.Close()
	fmt.Println("UPDAT sent")
}

// This function removes an album from the map 
// album name -> remove from map 
func DELET(vals string, c net.Conn) {
	deleteAlbs := strings.Split(vals,",")
	albumMutex.Lock()
	for _, album:= range deleteAlbs{
		delete(albumMap, album)
	}
	albumMutex.Unlock()
	fmt.Fprint(c,"\n")
	c.Close()
	fmt.Println("DELET sent")
}

// This function handles requests for 
// individual connections to speed 
// things up
func ConnHandle(c net.Conn){
	scanner:=bufio.NewScanner(c)
	scanner.Scan()
	req := scanner.Text()
	if req == ""{ //Ping from frontend, ignore it. 
		c.Close()
		return
	}
	switch req[0:5]{ //Figure out our operation
	case "G_ALL":
		go G_ALL(c)
	case "CREAT":
		go CREAT(req[7:len(req)-1], c)
	case "G_NAM":
		go G_NAM(c)
	case "G_ALB":
		go G_ALB(req[7:len(req)-1], c)
	case "UPDAT":
		go UPDAT(req[7:len(req)-1], c)
	case "DELET":
		go DELET(req[7:len(req)-1], c)
	}
}

func main(){
	var defaultport string
	flag.StringVar(&defaultport, "listen", "8090", "used to create the server")
	flag.Parse()

	//determine if default or not
	if len(defaultport) == 4{
		 defaultport = fmt.Sprintf(":%s", defaultport)
	}
	
	albumMap["The Beginning Of Times"] = &Album{"The Beginning Of Times",
												 "Amorphis",
												 "5",
												 "Very good"}

	albumMap["The Boy With The Arab Strap"] = &Album{"The Boy With The Arab Strap",
												 	 "Belle & Sebastian",
												 	 "5",
												 	 "Awesome"}


	albumMap["South Somewhere Else"] = &Album{"South Somewhere Else",
											  "Nana Grizol",
											   "5",
											  "Very mellow"}

	ln, err := net.Listen("tcp", defaultport)
	if err != nil{
		fmt.Println("Couldnt bind socket ", err)
		os.Exit(1)
	}

	for{ //Continously accept connections, then send it off to do a request
		conn, err := ln.Accept()
		if err!=nil{
			fmt.Fprint(os.Stderr, "Failed to accept")
			os.Exit(1)
		}
		go ConnHandle(conn)
	}
}
