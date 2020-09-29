/*
Author: Faris Alotaibi
Description: CRUD website for favorite albums. 
	-View all albums in a list at the album page. 
	-Add a new album (name, artist/bandname, rating, addtl comments)
	-Delete an existing album 
	-Update an existing albums Rating/Comments 
*/

package main 

import ("fmt"
		"strings"
		"flag"
		"github.com/kataras/iris/v12"
		)

//I used Artist as a general term to descibe whole bands/solo musicians
type Album struct{
	Name string
	Artist string 
	Rating string 
	Comments string
}

var albumMap = map[string]*Album{}

//Display the home page 
func displayLanding(ctx iris.Context){
    ctx.HTML("<h1>Hello </h1>");
    ctx.HTML("<h2>Welcome to my website!</h2>")
    ctx.HTML(`<a href="/albums"> My favorite albums </a>`)
}

//Page to display all favorite albums
func displayAlbums(ctx iris.Context){
	ctx.HTML(`<h1>My fave albums (ranked in no particular order)</h1>`)
	ctx.HTML(`<ul>%s</ul>`, listAlbumsHTML(albumMap))
	ctx.HTML(`<a href="/albums/create"> Add new album</a><br>
			  <a href="/albums/update"> Update album info</a><br>
			  <a href="/albums/delete"> Remove an album</a><br>
			  <a href="/"> Return</a>`)

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
	if albumMap[albName] == nil{
		newAlbum := &Album{}
		newAlbum.Name = albName
		newAlbum.Artist = ctx.FormValue("artName")
		newAlbum.Rating = ctx.FormValue("Rating")
		newAlbum.Comments = ctx.FormValue("albComms")
		albumMap[newAlbum.Name] = newAlbum
		ctx.HTML(`Added "%s" to the list of favorite albums<br>`, albName)
	}else{ //If the album already exists in the list 
		ctx.HTML(`"%s" is already in the list of favorite albums.<br>
				  Try updating or deleting it instead.<br>`, albName)
	}
	ctx.HTML(`<a href="/albums">Return to albums screen</a>`)
}

//Delete albums and return a delete receipt 
func processDelete(ctx iris.Context){
	albumsToDelete := ctx.FormValues()
	deletedAlbstr:=""
	for _,val := range albumsToDelete["albums"]{
		deletedAlbstr =  deletedAlbstr +`"` + val + `"` + ","
		delete(albumMap, val)
	}

	deletedAlbstr = deletedAlbstr[:len(deletedAlbstr)-1]
	ctx.HTML(`Deleted %s from the album list <br>
		     <a href="/albums">Return to albums screen</a>`, deletedAlbstr)
}

//Display albums that you can update 
func displayUpdate(ctx iris.Context){
	ctx.HTML(`<h1>Update Album Rating/Comments</h1>`)
	if len(albumMap) == 0{
		ctx.HTML(`Nothing to update.<br>
				  <a href="/albums">Return to albums screen</a>`)
	}else{
		htmlstr:=""
		for _, album := range albumMap{
			htmlstr = htmlstr + fmt.Sprintf(`<a href="/albums/update/%[2]s">Update "%[1]s" Rating/Comments</a><br>`, 
				album.Name, strings.ReplaceAll(album.Name," ", "_"))
		/*To get a working link I had to replace all spaces in the album name with underscores
		It gets replaced back to its original form in displayAlbumUpdate*/
		}
		ctx.HTML(htmlstr + `<br>`)
		ctx.HTML(`<a href="/albums">Return to albums screen</a>`)
	}
}

//Similar form to the create, but Album name + Artist fields are marked readonly
func displayAlbumUpdate(ctx iris.Context){
	aName:=ctx.Params().Get("albumname")
	aName = strings.ReplaceAll(aName, "_", " ")
	album:= albumMap[aName]
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
			  </form><br>`, aName, album.Artist, album.Comments)
	ctx.HTML(`<a href="/albums/update">Return to update selection screen</a><br>
		      <a href="/albums">Return to albums screen</a>`)
	
}
//This processes the rating/comment of an album
func processUpdate(ctx iris.Context){
	albname := ctx.FormValue("albName")
	album := albumMap[albname]
	album.Rating = ctx.FormValue("Rating")
	album.Comments = ctx.FormValue("albComms")
	ctx.HTML(`<a href="/albums">Return to albums screen</a>`)
}

//This displays potential albums to delete
func displayDelete(ctx iris.Context){
	ctx.HTML(`<h1>Delete an album (are you sure?)</h1>`)
	if len(albumMap) == 0{ //If there are none, display nothing
		ctx.HTML(`Nothing to delete.<br>
			      <a href="/albums">Return to albums screen</a> `)
	}else{	
		htmlstr:=`<form action="/delAlb" method="post">`
		for _, val:= range albumMap{
			htmlstr = htmlstr + fmt.Sprintf(`<input type="checkbox" id="%[1]s" name="albums" value="%[1]s">
			                             <label for="%[1]s">%[1]s</label><br>`, val.Name)
		}
		htmlstr = htmlstr + `<input type="submit" value="Delete Selections"></form><br>`
		ctx.HTML(htmlstr)
		ctx.HTML(`<a href="/albums">Return to albums screen</a>`)
	}
}

//Takes a single album and formats it to a list item 
func albumToHTMLString (album *Album) string{
	return fmt.Sprintf(`<li> "%s" by %s<br> Rating:%s<br> Comments:"%s"</li>`, album.Name, album.Artist, album.Rating, album.Comments)
}

//Uses the above to generate list items for all albums
func listAlbumsHTML(albumMap map[string]*Album) string{
	retstr:=""
	for _, album:= range albumMap{
		retstr = retstr + albumToHTMLString(album)
	}
	return retstr
}


func main(){
	//Debated on using string vs int vars earlier, realized it ultimately wouldnt matter
	var defaultport int
	flag.IntVar(&defaultport, "listen", 8080, "used to create the server")
	flag.Parse()

	var usedport string
	usedport = fmt.Sprintf(":%d", defaultport)
	
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
    app.Run(iris.Addr(usedport))
}

