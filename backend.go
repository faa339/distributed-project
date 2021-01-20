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
		"math/rand"
		"time"
		"strconv"
		)

//Custom struct for albums -- I use artist to denote 
//individual singers or full bands
type Album struct{
	Name string
	Artist string 
	Rating string 
	Comments string
}

//Struct to be used for raft
type RaftNode struct{
	Port string
	ElectionTerm int 
	State int 
}

type Log struct{
	Command string 
	Committed bool
}
//some constants for node state
var LEADER int = 0
var CANDIDATE int = 1
var FOLLOWER int = 2

//Data store variables
var albumMap = map[string]*Album{}
var albumMutex sync.RWMutex

//Raft variables -- note that we store the leaders
//port for recognition later 
var myState RaftNode
var nodeList = []string{}
var opLog = []Log{}
var LeaderPort string

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
	// fmt.Println("G_ALL sent")
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
	// fmt.Println("G_NAM sent")
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
	// fmt.Println("G_ALB sent")
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
	// fmt.Println("CREAT sent")
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
	// fmt.Println("UPDAT sent")
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
	// fmt.Println("DELET sent")
}


// This function gets consensus for ops that 
// require it (UPDAT, CREAT, DELET) 
func getConsensus(c net.Conn, op string) bool{
	latestLog:= Log{op, false}
	opLog = append(opLog, latestLog)
	formatOp := strings.ReplaceAll(op, " ", "_")
	sendingStr := fmt.Sprintf("PREPP %s {%s_%t}\n", myState.Port, formatOp, false)
	commitChan := make(chan string, len(nodeList))
	//Get consensus from each node
	// fmt.Println("Getting consensus")
	for _, node := range nodeList{
		go func(){
			conn, err:=net.Dial("tcp", node)
			if err!=nil{
				commitChan<-"N"
			}else{
				fmt.Fprint(conn, sendingStr)
				pscanner := bufio.NewScanner(conn)
				if pscanner.Scan(){
					res:=pscanner.Text()
					commitChan<-res
				conn.Close()
				}
			}
		}()
	}
	//Check to see how many can commit 
	commitCount:=0
	for i:=0; i<len(nodeList);i++{
		res:= <-commitChan 
		if res == "C"{
			commitCount++
		}
	}
    fmt.Println(commitCount)
	//If we can, do it and tell the nodes to do it
	if commitCount > (len(nodeList) / 2){
		// fmt.Println("Got consensus")
		switch op[0:5]{
		case "CREAT":
			go CREAT(op[7:len(op)-1], c)
		case "UPDAT":
			go UPDAT(op[7:len(op)-1], c)
		case "DELET":
			go DELET(op[7:len(op)-1], c)
		}
		opLog[len(opLog)-1].Committed = true
		for _, node:= range nodeList{
			go func(){
				conn, err:=net.Dial("tcp", node)
				if err!=nil{
					//it's down--when it gets back up it'll get updated
					return
				}else{
					fmt.Fprint(conn, "COMMT\n")
					conn.Close()
				}
			}()
		}
		return true
	}else{//If we can't commit
		return false
	}
}


//This will send the entire clients log 
//Should only be called by the leader
func sendLog() string{
	logStr := "["
	for i:=0; i<len(opLog); i++{
		entry := fmt.Sprintf("{%s,%t} ", opLog[i].Command, opLog[i].Committed)
		logStr = logStr + entry
	}
	logStr = logStr + "]\n"
	return logStr
}

//This will get a log from the leader 
//Should be called by followers that are behind
func getLog(){
	leaderC, err := net.Dial("tcp", LeaderPort)
	if err!=nil{
		//Figure this out 
	}
	pscanner := bufio.NewScanner(leaderC)
	fmt.Fprint(leaderC,"LOGRQ\n")
	logStr:=""
	if pscanner.Scan(){
		logStr = pscanner.Text()
	}
	logStr = logStr[1:len(logStr)-1]
	logEntries := strings.Split(logStr," ")
	logEntries = logEntries[:len(logEntries)-1]
	//Remake this -- we want to replace all of our log with this new one 
	//for safety (only append committed entries) 
	opLog = []Log{}
	for i:=0; i<len(logEntries); i++{
		entryPrep := logEntries[i][1:len(logEntries[i])-1]
		entry:= strings.Split(entryPrep, ",")
		commitState, _:= strconv.ParseBool(entry[1])
		if commitState{
			opLog = append(opLog, Log{entry[0], commitState})
		}
	}
}

//Deals with commiting log changes
func commtHandle(c net.Conn){
	operation:= opLog[len(opLog)-1].Command
	switch operation[0:5]{
	case "CREAT":
		go CREAT(operation[7:len(operation)-1], c)
	case "UPDAT":
		go UPDAT(operation[7:len(operation)-1], c)
	case "DELET":
		go DELET(operation[7:len(operation)-1], c)
	}
	opLog[len(opLog)-1].Committed = true
}

//Deal with incoming messages (for any end)
//Specifially handles raft related stuff like update logs
//and voting in elections
func msgProcess(c net.Conn){
	// TODO: Move this out of here -- we don't always need to enter this
	comm := ""
	pscanner := bufio.NewScanner(c)
	if pscanner.Scan(){
		comm = pscanner.Text()
		
	}
	if comm==""{
		c.Close()
		return
	}
	// fmt.Println("Received ", comm)
	switch comm[0:5]{
	case "VOTME": //Someone wants to become leader
		val, _ := strconv.Atoi(string(comm[6]))
		if (val > myState.ElectionTerm){
			fmt.Fprint(c, "VOTED\n")
		}else{
			fmt.Fprint(c, "\n")
		}
		c.Close()
	case "LEADE": //Someone declaring themselves leader
		val, _ := strconv.Atoi(string(comm[6]))
		if  val < myState.ElectionTerm{
			ret := fmt.Sprintf("FIX__ %s %d\n", LeaderPort, myState.ElectionTerm)
			fmt.Fprint(c, ret)
		}else{
			myState.State = FOLLOWER
			myState.ElectionTerm,_ = strconv.Atoi(string(comm[6]))
			LeaderPort = comm[8:]
			// fmt.Fprint(c, "\n")
		}
		c.Close()
	case "FIX__": //Someones telling us our log is outdated 
		vals := strings.Split(comm, " ")
		LeaderPort = vals[1]
		myState.ElectionTerm, _ = strconv.Atoi(vals[2])
		getLog()
		c.Close()
	case "LOGRQ": //Someones trying to get their log updated 
		logTxt := sendLog()
		fmt.Fprint(c, logTxt)
		c.Close()
	case "PREPP": //Leader said lets try to commit 
		vals:= strings.Split(comm[6:len(comm)], " ")
		// fmt.Println(vals[0], LeaderPort)
		if  vals[0]!= LeaderPort{
			fmt.Fprint(c, "N\n")
			// fmt.Println("Msg not from leader; no consensus")
		}else{
			// fmt.Println("prepared to commit")
			rawOp:= vals[1]
			thisOp:= strings.ReplaceAll(rawOp, "_", " ")
			opLog = append(opLog, Log{thisOp, false})
			fmt.Fprint(c, "C\n")
		}
		c.Close()
	case "COMMT":
		commtHandle(c)
	case "L____":
		retstr:= fmt.Sprintf("%d %s", myState.ElectionTerm, LeaderPort)
		fmt.Fprint(c, retstr)
		c.Close()
	case "HBCHK":
		myLead := comm[8:len(comm)]
		// fmt.Println(myLead)
		if LeaderPort == ""{
			LeaderPort = myLead
		}
		c.Close()
		// term, _:= strconv.Atoi(string(comm[6]))
		// port := comm[8:len(comm)]
		// if (term>myState.ElectionTerm){
		// 	myState.ElectionTerm = term
		// 	LeaderPort = port
		// 	c.Close()
		// }else{
		// 	c.Close()
		// 	c, err:= net.Dial("tcp", port)
		// 	if err==nil{
		// 		sentStr:= fmt.Sprintf("LEADE %d %s\n", myState.ElectionTerm, myState.Port)
		// 		fmt.Fprint(c, sentStr)
		// 		c.Close()
		// 	}
		// }
	case "G_ALL":
		go G_ALL(c)
	case "G_NAM":
		go G_NAM(c)
	case "G_ALB":
		go G_ALB(comm[7:len(comm)-1], c)
	default:
		if myState.State != LEADER{
			retstr:= fmt.Sprintf("%d %s", myState.ElectionTerm, LeaderPort)
			fmt.Fprint(c, retstr)
			c.Close()
		}
		res := getConsensus(c, comm)
		if res==false{
			fmt.Fprint(c, "/")
			c.Close()
		}
	}
}

//Function to handle an election
func ThisElection(){
	myState.State = CANDIDATE
	myState.ElectionTerm++
	// fmt.Println("starting election ", myState.ElectionTerm)
	voteCount := 1
	for _, node := range nodeList{
		// fmt.Println("connecting to ", node)
		conn, err:= net.Dial("tcp", node)
		if err != nil{
			continue
		}else{
			votestr := fmt.Sprintf("VOTME %d\n", myState.ElectionTerm)
			fmt.Fprint(conn, votestr)
			pscanner:=bufio.NewScanner(conn)
			if pscanner.Scan(){
				res := pscanner.Text()
				if res=="VOTED"{
					voteCount=voteCount + 1 
				}
			}
			conn.Close()
		}
	}
	//if we become the leader 
	if voteCount > (len(nodeList) / 2){
		myState.State = LEADER
		for _, node := range nodeList{
			if node==myState.Port{continue}
			conn, err:= net.Dial("tcp", node)
			if err!=nil{
				continue
			}else{
				res := fmt.Sprintf("LEADE %d %s\n", myState.ElectionTerm, myState.Port)
				fmt.Fprint(conn,res)
				conn.Close()
			}
		}
		LeaderPort = myState.Port
	}else{//If we don't
		myState.State=FOLLOWER
		// fmt.Println("not leader")
	}
}

//Returns some generated time 
//I'm not a fan of one-liners but this is much better than the alt
func timeGen() time.Duration{
	return time.Duration(rand.Intn(151) + 150) * time.Millisecond
}

//Let the follower's know the leader is ok!
func HB(){
	sentmsg:= fmt.Sprintf("HBCHK %d %s\n", myState.ElectionTerm, myState.Port)
	for _, node:= range nodeList{
		if myState.State == LEADER{
			conn, err := net.Dial("tcp", node)
			if err == nil{
				fmt.Fprint(conn, sentmsg)	
				conn.Close()
			}
		}
	}
}

//This function handles Raft initialization and 
//further program execution
func DoRaft(){
	rand.Seed(time.Now().UnixNano()) //Need to make sure we get unique rand#s
	//Setup a listener on our port
	ln, err := net.Listen("tcp", myState.Port)
	if err != nil{
		fmt.Fprintf(os.Stderr, "Couldnt bind socket: %s", err)
		os.Exit(1)
	}
	//We need a channel at least as big as the number
	//of other backends but lets make it as big as poss
	connChan := make(chan net.Conn, len(nodeList)* 10)
	go func(){//Function to accept conns and send them back
		for{
			conn, err := ln.Accept()
			// fmt.Println("Accepted new conn ", conn.LocalAddr().String())
			if err!=nil{
				fmt.Fprint(os.Stderr, "Failed to accept")
				os.Exit(1)
			}
			connChan <- conn
		}
	}()

	var timeoutTimer *time.Timer
	currTimeout := timeGen()
	var HBTimer *time.Timer 
	HBTimeout:= time.Duration(50) * time.Millisecond
	for{
		if myState.State == LEADER{
			timeoutTimer.Stop()
			go func(){<-timeoutTimer.C}() //Drain the channel just in case
			HBTimer = time.NewTimer(HBTimeout)
			// fmt.Println("Leader loop beginning")
			// fmt.Println("I am leader:", myState.Port)
			for myState.State==LEADER{
				select{
				case <- HBTimer.C:
					HB()
					HBTimer.Reset(HBTimeout)
				case potentialFE := <-connChan:
					//Still need to send heartbeats -- do something
					HBTimer.Stop()
					msgProcess(potentialFE)
					if myState.State == FOLLOWER{
						break
					}
					HBTimer.Reset(HBTimeout)
				}
			}
		}else{
			currTimeout = timeGen()
			timeoutTimer = time.NewTimer(currTimeout)
			for myState.State==FOLLOWER{
				// fmt.Println("Leader is ", LeaderPort)
				select{	
				case <-timeoutTimer.C:
					LeaderPort=""
					ThisElection()
					if myState.State == LEADER{
						break
					}
					currTimeout=timeGen()
					timeoutTimer.Reset(currTimeout)
				case conn:= <- connChan:
					timeoutTimer.Stop()
					msgProcess(conn)
					timeoutTimer.Reset(currTimeout)
				}
			}
		}
	}
}

func main(){
	var defaultport string
	var backends string
	flag.StringVar(&defaultport, "listen", "", "used to create the server")
	flag.StringVar(&backends, "backend", "", "used to connect to other backends")
	flag.Parse()

	//determine if default or not
	if defaultport == ""{
		fmt.Fprintf(os.Stderr, "You must provide a port to listen on.\n")
		os.Exit(1)
	}
	if len(defaultport) == 4{
		 defaultport = fmt.Sprintf("127.0.0.1:%s", defaultport)
	}
	if backends==""{
		fmt.Fprint(os.Stderr, "You must provide a comma separated list of other backendsn\n")
		os.Exit(1)
	}else{
		nodeList = strings.Split(backends, ",")
		for i:=0; i<len(nodeList); i++{
			if len(nodeList[i]) <= 5{
				nodeList[i] = fmt.Sprintf("127.0.0.1%s", nodeList[i])
			}
		}
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

	//Initialize Node state. Setting up random timeout 
	myState = RaftNode{defaultport, 0, FOLLOWER}
	//Raft time! 
    DoRaft()
}
