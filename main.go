package main 

import ("fmt"
		//"net"
		//"net/http"
		//"html/template"
		"strings"
		"flag"
		"github.com/kataras/iris/v12"
		)

type Album struct{
	Name string
	Artist string 
	Rating string 
	Comments string
}

func displayLanding(ctx iris.Context){
    ctx.HTML("<h1>Hello </h1>");
    ctx.HTML("<h2>Welcome to my website!</h2>")
    ctx.HTML(`<a href="/albums"> My favorite albums </a>`)
}

func displayAlbums(ctx iris.Context){
	ctx.HTML(`<h1>My fave albums (ranked in no particular order)</h1>`)
	ctx.HTML(`<ul>%s</ul>`, listAlbumsHTML(albumMap))
	   
	ctx.HTML(`<a href="/albums/create"> Add new album</a><br>
			  <a href="/albums/update"> Update album info</a><br>
			  <a href="/albums/delete"> Remove an album</a><br>
			  <a href="/"> Return</a>`)

}

func displayCreate(ctx iris.Context){
	ctx.HTML(`<h1>Add a new Album</h1>
			  <form action="/addAlb" method="post">
			  	<label for="AlbumName">Album Name:</label>
			  	<input type="text" id="albName" name="albName"><br>
			  	<label for="Artistname">Artist Name:</label>
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

func processCreate(ctx iris.Context){
	newAlbum := &Album{}
	newAlbum.Name = ctx.FormValue("albName")
	newAlbum.Artist = ctx.FormValue("artName")
	newAlbum.Rating = ctx.FormValue("Rating")
	newAlbum.Comments = ctx.FormValue("albComms")
	albumMap[newAlbum.Name] = newAlbum
	ctx.HTML(`<a href="/albums">Return to albums screen</a>`)
}

func processDelete(ctx iris.Context){
	fmt.Println("Entering processDelete")
	albumsToDelete := ctx.FormValues()
	for _,val := range albumsToDelete["albums"]{
		albumMap[val] = nil
	}
	ctx.HTML(`<a href="/albums">Return to albums screen</a>`)
}

func displayUpdate(ctx iris.Context){
	ctx.HTML(`<h1>Update Album Rating/Comments</h1>`)
	htmlstr:=""
	for _, album := range albumMap{
		if album != nil{
			htmlstr = htmlstr + fmt.Sprintf(`<a href="/albums/update/%[2]s">Update %[1]s's Rating/Comments</a><br>`, 
				album.Name, strings.ReplaceAll(album.Name," ", "_"))
		}
	}
	ctx.HTML(htmlstr + `<br>`)
	ctx.HTML(`<a href="/albums">Return to albums screen</a>`)

}

func displayAlbumUpdate(ctx iris.Context){
	aName:=ctx.Params().Get("albumname")
	aName = strings.ReplaceAll(aName, "_", " ")
	album:= albumMap[aName]
	ctx.HTML(`<h2>Updating rating/comments for %s</h2>`, aName)
	ctx.HTML(`<form action="/updateconf" method="post">
			  	<label for="AlbumName">Album Name:</label>
			  	<input type="text" id="albName" name="albName" value="%s" readonly><br>
			  	<label for="Artistname">Artist Name:</label>
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

func processUpdate(ctx iris.Context){
	albname := ctx.FormValue("albName")
	album := albumMap[albname]
	album.Rating = ctx.FormValue("Rating")
	album.Comments = ctx.FormValue("albComms")
	ctx.HTML(`<a href="/albums">Return to albums screen</a>`)
}

func displayDelete(ctx iris.Context){
	ctx.HTML(`<h1>Delete an album (are you sure?)</h1>`)
	htmlstr:=`<form action="/delAlb" method="post">`
	for _, val:= range albumMap{
		if val != nil{
			htmlstr = htmlstr + fmt.Sprintf(`<input type="checkbox" id="%[1]s" name="albums" value="%[1]s">
			                             <label for="%[1]s">%[1]s</label><br>`, val.Name)
		}
	}
	htmlstr = htmlstr + `<input type="submit" value="Delete Selections"></form><br>`
	ctx.HTML(htmlstr)
	ctx.HTML(`<a href="/albums">Return to albums screen</a>`)
}


func albumToHTMLString (album *Album) string{
	return fmt.Sprintf(`<li> "%s" by %s<br> Rating:%s<br> Comments:"%s"</li>`, album.Name, album.Artist, album.Rating, album.Comments)
}

func listAlbumsHTML(albumMap map[string]*Album) string{
	retstr:=""
	for _, album:= range albumMap{
		if album != nil{
			retstr = retstr + albumToHTMLString(album)
		}
	}
	return retstr
}

var albumMap = map[string]*Album{}

func main(){
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

