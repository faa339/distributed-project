/*
Author: Faris Alotaibi
Description: frontend of CRUD website for favorite albums. 
	-View all albums in a list at the album page. 
	-Add a new album (name, artist/bandname, rating, addtl comments)
	-Delete an existing album 
	-Update an existing albums Rating/Comments
pass in --listen [PORTNUM] --backend [HOSTNAME:PORTNUM] 
*/

package main 

import ("fmt"
		"flag"
		"github.com/kataras/iris/v12"
		"net"
		"bufio"
		"strings"
		)

//Kept as a global for easy access to backend
var backendport string

/*
  This function handles getting results from the backend
  There are different kinds of requests for different things 
  done on the frontend
		G_ALL:
             get all info about all albums in the database
		CREAT:
             add album to the database (expect '-' if already in)
        G_NAM:
             get all album names for update/delete display
        G_ALB:
        	 get info about a single album in the database
        UPDAT:
             update album with new rating & comments
        DELET:
             delete album(s) from database 
  Required info per call is passed in the values parameter
*/
func sendReq(operation string, values string) string{
	results := "_" //Default return if connection errors arise
	conn, err:= net.Dial("tcp", backendport)
	if err!=nil{
		fmt.Println("Error connecting to backend")
		return results
	}
	defer conn.Close()
	pscanner := bufio.NewScanner(conn)
	//Backend reads with .Scan too -- need to make sure theres no bad 
	//inputs with \n in them. Sorry if your band  is called "\n and the magic spaces!"
	values = strings.ReplaceAll(values, "\n", " ") 
	passin:= operation + " " + values + "\n" 
	fmt.Fprintf(conn,passin)
	if pscanner.Scan(){
		results=pscanner.Text()
	}
	return results
}


//Display the home page 
func displayLanding(ctx iris.Context){
    ctx.HTML("<h1>Hello </h1>");
    ctx.HTML("<h2>Welcome to my website!</h2>")
    ctx.HTML(`<a href="/albums"> My favorite albums </a>`)
}

//Format album data into multiple list items to display info 
//Albums sent to this are sent as one big string where individual 
//albums are identified by brackets. They are split using a custom 
//split function, but this causes , to be identified as an item too
//so it is subsequently ignored 
func albumsToHTMLLi (albums string) string{
	htmlstr:=""
	albumsList := strings.FieldsFunc(albums, Split)
	for _, albumstring:=range albumsList{
		if albumstring == ","{continue} 
		album := strings.Split(albumstring, ",")
		htmlstr = htmlstr + fmt.Sprintf(`<li> "%s" by %s<br> Rating:%s<br> Comments:"%s"</li>`, 
                               album[0], album[1], album[2], album[3])
	}
	return htmlstr
}

//Custom split function that'll create new tokens on the presence of either bracket
func Split(r rune) bool{
	return r=='[' || r==']' 
}

//Page to display all favorite albums
func displayAlbums(ctx iris.Context){
	ctx.HTML(`<h1>My fave albums (ranked in no particular order)</h1>`)
	result := sendReq("G_ALL", "")
	if result != ""{
		result = result[1:len(result)-1]
		ctx.HTML(`<ul>%s</ul>`, albumsToHTMLLi(result))
		ctx.HTML(`<a href="/albums/create"> Add new album</a><br>
				  <a href="/albums/update"> Update album info</a><br>
				  <a href="/albums/delete"> Remove an album</a><br>
				  <a href="/"> Return</a>`)
	}else{
		ctx.HTML(`Nothing to display.<br>`)
		ctx.HTML(`<a href="/albums/create"> Add new album</a><br>
				  <a href="/albums/update"> Update album info</a><br>
				  <a href="/albums/delete"> Remove an album</a><br>
				  <a href="/"> Return</a>`)
	}

}


//Page to create albums using html form + post request
func displayCreate(ctx iris.Context){
	ctx.HTML(`<h1>Add a new Album</h1>
			  <form action="/addAlb" method="post">
			  	<label for="AlbumName">Album Name:</label>
			  	<input type="text" id="albName" name="albName"><br>
			  	<label for="Artistname">Band/Musician Name:</label>
			  	<input type="text" id="artName" name="artName"><br>
			  	<label for="Rate">Rating:</label>
			  	<input type="radio" id="r1" name="Rating" value="1"><label for="r1">1</label><br>
			  	<input type="radio" id="r2" name="Rating" value="2"><label for="r2">2</label><br>
			  	<input type="radio" id="r3" name="Rating" value="3"><label for="r3">3</label><br>
			  	<input type="radio" id="r4" name="Rating" value="4"><label for="r4">4</label><br>
			  	<input type="radio" id="r5" name="Rating" value="5"><label for="r5">5</label><br>
			  	<label for="Comments">Comments:</label>
			  	<input type="text" id="albComms" name="albComms"><br>
			  	<input type="submit" value="Submit">
			  </form><br>
			  <a href="/albums">Return to albums screen</a>`)
}

//Create a new album and add it to the list 
func processCreate(ctx iris.Context){
	albName:= ctx.FormValue("albName")
	artName:= ctx.FormValue("artName")
	rating:= ctx.FormValue("Rating")
	comms:= ctx.FormValue("albComms")
	album:= fmt.Sprintf(`{%s,%s,%s,%s}`, albName, artName, rating, comms)
	res := sendReq("CREAT", album)
	if res != "-"{
		ctx.HTML(`Added "%s" to the list of favorite albums<br>`, albName)
	}else{
		ctx.HTML(`"%s" is already in the list of favorite albums.<br>
				  Try updating or deleting it instead.<br>`, albName)
	}
	ctx.HTML(`<a href="/albums">Return to albums screen</a>`)
}

//Delete albums and return a delete receipt 
func processDelete(ctx iris.Context){
	albumsToDelete := ctx.FormValues()["albums"]
	if len(albumsToDelete) == 0{
		ctx.HTML(`Nothing selected for deletion.<br> 
				  <a href="/albums/delete">Return to delete selection screen</a><br>
				  <a href="/albums">Return to albums screen</a><br>`)
	}else{
		albumsString:= strings.Join(albumsToDelete, ",")
		sentVal := "{" + albumsString + "}"
		result := sendReq("DELET", sentVal)
		if result == "" {
			ctx.HTML(`Deleted "%s" from the album list <br>
			     <a href="/albums">Return to albums screen</a>`, albumsString)
		}
	}
}

//Display albums that you can update 
func displayUpdate(ctx iris.Context){
	ctx.HTML(`<h1>Update Album Rating/Comments</h1>`)
	albumnames:= sendReq("G_NAM", "")
	if albumnames == "-"{
		ctx.HTML(`Nothing to update.<br>
				  <a href="/albums">Return to albums screen</a>`)
	}else{
		albumnames = albumnames[:len(albumnames)-1]
		names := strings.Split(albumnames, ",")
		htmlstr:=""
		for _,albname := range names{
			htmlstr = htmlstr + fmt.Sprintf(`<a href="/albums/update/%[2]s">Update "%[1]s" Rating/Comments</a><br>`, 
				                             albname, strings.ReplaceAll(albname," ", "_"))
		}
		ctx.HTML(htmlstr + `<br>`)
		ctx.HTML(`<a href="/albums">Return to albums screen</a>`)
	}
}

//Similar form to the create, but Album name + Artist fields are marked readonly
func displayAlbumUpdate(ctx iris.Context){
	aName:=ctx.Params().Get("albumname")
	aName = strings.ReplaceAll(aName, "_", " ")
	sentName := "{" + aName + "}"
	result := sendReq("G_ALB", sentName)
	albvals := strings.Split(result, ",")
	ctx.HTML(`<h2>Updating rating/comments for %s</h2>`, aName)
	ctx.HTML(`<form action="/updateconf" method="post">
			  	<label for="AlbumName">Album Name:</label>
			  	<input type="text" id="albName" name="albName" value="%s" readonly><br>
			  	<label for="Artistname">Band/Musician Name:</label>
			  	<input type="text" id="artName" name="artName" value="%s" readonly><br>
			  	<label for="Rate">Rating:</label>
			  	<input type="radio" id="r1" name="Rating" value="1"><label for="r1">1</label><br>
			  	<input type="radio" id="r2" name="Rating" value="2"><label for="r2">2</label><br>
			  	<input type="radio" id="r3" name="Rating" value="3"><label for="r3">3</label><br>
			  	<input type="radio" id="r4" name="Rating" value="4"><label for="r4">4</label><br>
			  	<input type="radio" id="r5" name="Rating" value="5"><label for="r5">5</label><br>
			  	<label for="Comments">Comments:</label>
			  	<input type="text" id="albComms" name="albComms" value="%s"><br>
			  	<input type="submit" value="Submit">
			  </form><br>`, aName, albvals[1], albvals[3])
	ctx.HTML(`<a href="/albums/update">Return to update selection screen</a><br>
		      <a href="/albums">Return to albums screen</a>`)
}

//This processes the rating/comment of an album
func processUpdate(ctx iris.Context){
	albName := ctx.FormValue("albName")
	rating := ctx.FormValue("Rating")
	comments := ctx.FormValue("albComms")
	values := fmt.Sprintf("{%s,%s,%s}", albName, rating, comments)
	fmt.Println(values)
	res := sendReq("UPDAT", values)
	if res == "_"{
		fmt.Println("Error in retrieving values")
	}else{
		ctx.HTML(`Info for "%s" updated. <br>
			     <a href="/albums">Return to albums screen</a>`, albName)
	}
}

//This displays potential albums to delete
func displayDelete(ctx iris.Context){
	ctx.HTML(`<h1>Delete an album (are you sure?)</h1>`)
	albumnames := sendReq("G_NAM", "")
	if albumnames == "-"{
		ctx.HTML(`Nothing to delete.<br>
			      <a href="/albums">Return to albums screen</a> `)		
	}else{
		albumnames = albumnames[:len(albumnames)-1]
		names := strings.Split(albumnames, ",")
		htmlstr:=`<form action="/delAlb" method="post">`
		for _, albname:= range names{
			htmlstr = htmlstr + fmt.Sprintf(`<input type="checkbox" id="%[1]s" name="albums" value="%[1]s">
			                             <label for="%[1]s">%[1]s</label><br>`, albname)
		}
		htmlstr = htmlstr + `<input type="submit" value="Delete Selections"></form><br>`
		ctx.HTML(htmlstr)
		ctx.HTML(`<a href="/albums">Return to albums screen</a>`)
	}
}

func main(){
	var frontendport string
	flag.StringVar(&frontendport, "listen", "8080", "accept connections here. provide a port number [port]")
	flag.StringVar(&backendport, "backend", ":8090", "conect to backend. provide a host name and port [hostname:port] or just a port [:port]")
	flag.Parse()

	usedFEnd := fmt.Sprintf(":%s", frontendport)
	if len(backendport) == 5{
		backendport = fmt.Sprintf("localhost%s", backendport)
	}

	app := iris.Default()
    app.Get("/", displayLanding)
    app.Get("/albums", displayAlbums)
    app.Get("/albums/update", displayUpdate)
    app.Get("/albums/update/{albumname}", displayAlbumUpdate)
    app.Get("/albums/create", displayCreate)
    app.Get("/albums/delete", displayDelete)
    app.Post("/addAlb", processCreate)
    app.Post("/delAlb", processDelete)
    app.Post("/updateconf", processUpdate)
    app.Run(iris.Addr(usedFEnd))
    
}

