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

//This function returns all info about every album in the map
func G_ALL() string{
	retstr:="{"
	for _, album:=range albumMap{
		retstr = retstr + fmt.Sprintf("[%s,%s,%s,%s],",
			     album.Name, album.Artist, album.Rating, album.Comments)		
	}
	retstr = retstr[:len(retstr)-1]
	retstr = retstr + "}"
	return retstr
}

//This function returns all album names in the map
func G_NAM() string{
	retstr:=""
	if len(albumMap) == 0{
		return "-"
	}
	for _, album:= range albumMap{
		retstr = retstr + album.Name + ","
	}
	return retstr
}

//This function returns all info about a specific album 
func G_ALB(name string) string{
	album := albumMap[name]
	return fmt.Sprintf("%s,%s,%s,%s", album.Name, album.Artist, album.Rating, album.Comments)
}

//This function adds an album to the map
func CREAT(vals string) string{
	albarr := strings.Split(vals,",")
	if albumMap[albarr[0]] != nil{
		return "-\n"
	}else{
		albumMap[albarr[0]] = &Album{albarr[0], albarr[1], albarr[2], albarr[3]}
		return "\n"
	}
}

//This function updates an albums ratings & comments in the map 
func UPDAT(vals string) string{
	updtarr := strings.Split(vals,",")
	album := albumMap[updtarr[0]]
	album.Rating = updtarr[1]
	album.Comments = updtarr[2]
	return "\n"
}

//This function removes an album from the map
func DELET(vals string) string{
	deleteAlbs := strings.Split(vals,",")
	for _, album:= range deleteAlbs{
		delete(albumMap, album)
	}
	return "\n"
}

func main(){
	var defaultport string
	flag.StringVar(&defaultport, "listen", "8090", "used to create the server")
	flag.Parse()

	var usedport string
	usedport = fmt.Sprintf(":%s", defaultport)
	
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

	ln, err := net.Listen("tcp", usedport)
	if err != nil{
		fmt.Println("Couldnt bind socket ", err)
		os.Exit(1)
	}

	for{ //Continously accept connections until close
		conn, err := ln.Accept()
		if err!=nil{
			fmt.Fprint(os.Stderr, "Failed to accept")
			os.Exit(1)
		}
		fmt.Println("Conn accepted")
		scanner:=bufio.NewScanner(conn)
		scanner.Scan()
		req := scanner.Text()

		switch req[0:5]{ //Figure out our operation
		case "G_ALL":
			result:= G_ALL() + "\n"
			fmt.Fprint(conn, result)
			fmt.Println("G_ALL sent")
			conn.Close()
		case "CREAT":
			vals := req[7:len(req)-1]
			result := CREAT(vals)
			fmt.Fprint(conn, result)
			conn.Close()
		case "G_NAM":
			result:= G_NAM() + "\n"
			fmt.Fprint(conn, result)
			fmt.Println("G_NAM sent")
			conn.Close()
		case "G_ALB":
			name:= req[7:len(req)-1]
			result := G_ALB(name) + "\n"
			fmt.Fprint(conn, result)
			fmt.Println("G_ALB sent")
			conn.Close()
		case "UPDAT":
			vals:= req[7:len(req)-1]
			result := UPDAT(vals)
			fmt.Fprint(conn, result)
			fmt.Println("UPDAT sent")
			conn.Close()
		case "DELET":
			vals:= req[7:len(req)-1]
			result := DELET(vals)
			fmt.Fprint(conn, result)
			fmt.Println("DELET sent")
			conn.Close()
		}
	}
}
