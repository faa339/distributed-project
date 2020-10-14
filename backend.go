package main; 

import ("fmt"
		"flag"
		"net"
		"os"
		"bufio"
		"strings"
		)

type Album struct{
	Name string
	Artist string 
	Rating string 
	Comments string
}

var albumMap = map[string]*Album{}

func albumToHTMLString (album *Album) string{
	return fmt.Sprintf(`<li> "%s" by %s<br> Rating:%s<br> Comments:"%s"</li>`, album.Name, album.Artist, album.Rating, album.Comments)
}

func G_ALL() string{
	retstr:=""
	for _, album:=range albumMap{
		retstr = retstr + albumToHTMLString(album)		
	}
	return retstr
}

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

func G_ALB(name string) string{
	album := albumMap[name]
	return fmt.Sprintf("%s,%s,%s,%s", album.Name, album.Artist, album.Rating, album.Comments)
}

func CREAT(vals string) string{
	albarr := strings.Split(vals,",")
	if albumMap[albarr[0]] != nil{
		return "-\n"
	}else{
		albumMap[albarr[0]] = &Album{albarr[0], albarr[1], albarr[2], albarr[3]}
		return "\n"
	}
}

func UPDAT(vals string) string{
	updtarr := strings.Split(vals,",")
	album := albumMap[updtarr[0]]
	album.Rating = updtarr[1]
	album.Comments = updtarr[2]
	return "\n"
}

func DELET(vals string) string{
	deleteAlbs := strings.Split(vals,",")
	for _, album:= range deleteAlbs{
		delete(albumMap, album)
	}
	return "\n"
}

func main(){
	//Debated on using string vs int vars earlier, realized it ultimately wouldnt matter
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

	for{
		conn, err := ln.Accept()
		if err!=nil{
			fmt.Fprint(os.Stderr, "Failed to accept")
			os.Exit(1)
		}
		fmt.Println("Conn accepted")
		scanner:=bufio.NewScanner(conn)
		scanner.Scan()
		req := scanner.Text()

		switch req[0:5]{
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
