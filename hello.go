package main

//All Go Imports That Are Used
import (
    "fmt"
    "net/http"
    "io/ioutil"
    "html/template"
    "os"
    "bytes"
    "bufio"
    "strings"
    "strconv"
    "path/filepath"
)

//Struct for a User that stores their Username and Password
type User struct
{
    Username string
    Password string
}

//Struct that defines all the stuff a page can have
type Page struct 
{
    Title string
    Body  []byte
    HtmlBody template.HTML
    HtmlForm template.HTML
    FriendList template.HTML
    FriendForm template.HTML
}

//Used to simplify the templating
var templates = template.Must(template.ParseFiles("edit.html", "view.html", "main.html", "board.html"))

//VARIABLE TO KEEP TRACK OF CURRENT USER
var loggedInUser = new(User) 

//GLOBAL NUMBER OF POSTS
var numOfPosts = 0; 

//Saves a file 
func (p *Page) save() error {
    filename := p.Title + ".txt"
    return ioutil.WriteFile(filename, p.Body, 0600)
}

//Saves a User
func saveUser(u *User, w http.ResponseWriter, r *http.Request) {

    filename := "Users/"+u.Username + "/" + u.Username + ".txt"
    formFileName := "Users/"+u.Username + "/"+u.Username+"Form.txt"
    friendFileName := "Users/"+u.Username + "/"+u.Username+"Friend.txt"
    os.MkdirAll("./Users/"+u.Username, 0777)
    os.MkdirAll("./Users/"+u.Username+"/"+"Posts", 0777)
    os.Create("./Users/"+u.Username+"/"+"friendList.txt")
    f, err := os.OpenFile("Users/allUsers.txt", os.O_APPEND|os.O_WRONLY, 0600)
    if err != nil {
        panic(err)
    }
    defer f.Close()
    if _, err = f.WriteString(u.Username+" "+u.Password+"\n"); err != nil {
        panic(err)
    }

    body, err2 := ioutil.ReadFile(filename)
    if err2 != nil {
        nameAndPass := []byte(u.Username + " " + u.Password)
        ioutil.WriteFile(filename, nameAndPass, 0600)
        userForm := []byte("<form action = /userPost/"+u.Username + " method = POST style = position:absolute;left:20px;top:30%><div><textarea name = body3 rows = 10 cols = 20></textarea></div><div><input type = submit value = Post&emsp;On&emsp;"+u.Username+"s&emsp;Page></div></form>")
        ioutil.WriteFile(formFileName, userForm, 0600)
        friendForm := []byte("<form action = /addFriend/"+u.Username+" method = POST><div><input type = submit value = Add&emsp;To&emsp;Friends></div></form>")
        ioutil.WriteFile(friendFileName, friendForm, 0600)
        http.Redirect(w, r, "/view/Users/"+u.Username, http.StatusFound)
    } else {
        fmt.Println(body)
        http.Redirect(w, r, "/", http.StatusFound) 
    }
    
} 

//Loads a regular page
func loadPage(title string) (*Page, error) {
    filename := title + ".txt"
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body, HtmlBody: template.HTML(body)}, nil
}

//Loads a User Page
func loadUserPage(title string, uname string) (*Page, error) {
    filename := title + "/" + uname + ".txt"
    filename2 := title + "/" + uname + "Form.txt"
    filename3 := title + "/" + "friendList.txt"
    filename4 := title + "/" + uname +"Friend.txt"
    var buffer bytes.Buffer
    files, _ := ioutil.ReadDir(title+"/Posts/")
    for _, f := range files {     
            name := f.Name()
            if name != ".DS_Store" {
                files2, _ := ioutil.ReadDir(title+"/Posts/"+name)
                for _, f2 := range files2 {
                    if f2.Name() != ".DS_Store" {
                        name2 := strings.Split(f2.Name(),".")
                        p , err  := loadPage(title+"/Posts/"+name+"/"+name2[0])
                        if err != nil {
                            return nil, nil
                        }
                        buffer.WriteString(string(p.Body))
                    }
                }
                
            }
            
    }    
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    body2, err2 := ioutil.ReadFile(filename2)
    if err2 != nil {
        return nil, err2
    }
    body3, err3 := ioutil.ReadFile(filename3)
    if err3 != nil {
        fmt.Printf("FAIL")
        return nil, err3
    }
    body4, err4 := ioutil.ReadFile(filename4)
    if err4 != nil {
        fmt.Printf("FAIL")
        return nil, err4
    }
    return &Page{Title: title, Body: body, HtmlBody: template.HTML(buffer.String()), HtmlForm: template.HTML(body2), FriendList: template.HTML(body3), FriendForm: template.HTML(body4)}, nil
}

//Function Used to Render HTML templates
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

//Main Page handler
func handler(w http.ResponseWriter, r *http.Request) {
    if loggedInUser.Username != "" {
        http.Redirect(w, r, "/view/Users/"+loggedInUser.Username, http.StatusFound)
    }
    renderTemplate(w, "main", nil)

}

//Viewing A User Page
func viewHandler(w http.ResponseWriter, r *http.Request) {
    title := filepath.Base(r.URL.Path[:])
    title2 := r.URL.Path[len("/view/"):]
    p , err  := loadUserPage(title2, title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    renderTemplate(w, "view", p)
}

//Used to Edit a Page
func editHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/edit/"):]
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}

//Used to Save a File
func saveHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/save/"):]
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body), HtmlBody: template.HTML(body)}
    p.save()
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
//Handles Posts on a Users Page
func userPostHandler(w http.ResponseWriter, r *http.Request) {
    uName := filepath.Base(r.URL.Path[:])
    var countOfPosts = 0
    files, _ := ioutil.ReadDir("./Users/"+uName+"/Posts/")
    for _, f := range files {  
        fmt.Println(f.Name())   
        countOfPosts++
    }
    os.MkdirAll("./Users/"+uName+"/Posts/"+uName+"-"+strconv.Itoa(countOfPosts), 0777)
    body3 := r.FormValue("body3")
    f2, err2 := os.Create("./Users/"+uName+"/Posts/"+uName+"-"+strconv.Itoa(countOfPosts)+"/a.txt")
    if err2 != nil {
        fmt.Println("File Creation failed")
    }
    postText := []byte("<div style = color:#FFFF00;font-size:18px;>\n<a style = text-decoration:none;color:#FFFF00; href = /view/Users/"+loggedInUser.Username+">"+loggedInUser.Username + ":</a><br><span style = color:#00FF00;>"+ body3 + "</span><br>\n</div>\n<div style = background-color:black;color:00ff00>")
    ioutil.WriteFile("./Users/"+uName+"/Posts/"+uName+"-"+strconv.Itoa(countOfPosts)+"/a.txt", postText, 0600)
    f3, err3 := os.Create("./Users/"+uName+"/Posts/"+uName+"-"+strconv.Itoa(countOfPosts)+"/b.txt")
    if err3 != nil {
        fmt.Println("File Creation failed")
    }
    f4, err4 := os.Create("./Users/"+uName+"/Posts/"+uName+"-"+strconv.Itoa(countOfPosts)+"/c.txt")
    if err4 != nil {
        fmt.Println("File Creation failed")
    }
    formText := []byte("<br><form action=/userCommentPost/"+uName+"-"+strconv.Itoa(countOfPosts)+" method=POST><input type=text name=comment"+strconv.Itoa(countOfPosts)+" value=><input type=submit value=Comment></form></div><br>")
    ioutil.WriteFile("./Users/"+uName+"/Posts/"+uName+"-"+strconv.Itoa(countOfPosts)+"/c.txt", formText, 0600)
    http.Redirect(w, r, "/view/Users/"+uName, http.StatusFound)
    fmt.Println(f2.Name() + " " + f3.Name() + " " + f4.Name())
}
//handles Comments on a Users Posts
func userCommentHandler(w http.ResponseWriter, r *http.Request) {
    postNumber := r.URL.Path[len("/userCommentPost/"):]
    stuff := strings.Split(postNumber, "-")
    comment := r.FormValue("comment"+stuff[1])
    f, err := os.OpenFile("Users/"+stuff[0]+"/Posts/"+stuff[0]+"-"+stuff[1]+"/b.txt", os.O_APPEND|os.O_WRONLY, 0600)
    if err != nil {
        panic(err)
    }
    defer f.Close()
    if _, err = f.WriteString("\n&emsp;<a style = text-decoration:none;color:#FFFF00; href = /view/Users/"+loggedInUser.Username+">"+loggedInUser.Username+":</a> <br>&emsp;&emsp;<span style = color:#00FF00;>"+comment+"</span><br>"); err != nil {
        panic(err)
    }
    http.Redirect(w, r, "/view/Users/"+stuff[0], http.StatusFound)
}
//Handles Global DiscussionBoard Posts
func globalPostHandler(w http.ResponseWriter, r *http.Request) {
    numOfPosts++
    os.MkdirAll("./Posts/"+strconv.Itoa(numOfPosts), 0777)
    body2 := r.FormValue("body2")
    f2, err2 := os.Create("./Posts/"+strconv.Itoa(numOfPosts)+"/a.txt")
    if err2 != nil {
        fmt.Println("File Creation failed")
    }
    postText := []byte("<div style = color:#FFFF00;font-size:18px;>\n<a style = text-decoration:none;color:#FFFF00; href = /view/Users/"+loggedInUser.Username+">"+ loggedInUser.Username+ ":</a><br><span style = color:#00FF00;>" + body2 + "</span><br>\n</div>\n<div style = background-color:black;color:00ff00>")
    ioutil.WriteFile("./Posts/"+strconv.Itoa(numOfPosts)+"/a.txt", postText, 0600)
    f3, err3 := os.Create("./Posts/"+strconv.Itoa(numOfPosts)+"/b.txt")
    if err3 != nil {
        fmt.Println("File Creation failed")
    }
    f4, err4 := os.Create("./Posts/"+strconv.Itoa(numOfPosts)+"/c.txt")
    if err4 != nil {
        fmt.Println("File Creation failed")
    }
    formText := []byte("<br><form action=/commentPost/"+strconv.Itoa(numOfPosts)+" method=POST><input type=text name=comment"+strconv.Itoa(numOfPosts)+" value=><input type=submit value=Comment></form></div><br>")
    ioutil.WriteFile("./Posts/"+strconv.Itoa(numOfPosts)+"/c.txt", formText, 0600)
    http.Redirect(w, r, "/board/", http.StatusFound)
    fmt.Println(f2.Name() + " " + f3.Name() + " " + f4.Name())
}
//Handles Global Post Comments
func commentPostHandler(w http.ResponseWriter, r *http.Request) {
    postNumber := r.URL.Path[len("/commentPost/"):]
    comment := r.FormValue("comment"+postNumber)
    f, err := os.OpenFile("Posts/"+postNumber+"/b.txt", os.O_APPEND|os.O_WRONLY, 0600)
    if err != nil {
        panic(err)
    }
    defer f.Close()
    if _, err = f.WriteString("\n&emsp;<a style = text-decoration:none;color:#FFFF00; href = /view/Users/"+loggedInUser.Username+">"+loggedInUser.Username+":</a> <br>&emsp;&emsp;<span style = color:#00FF00;>"+comment+"</span><br>"); err != nil {
        panic(err)
    }
    http.Redirect(w, r, "/board/", http.StatusFound)
}
//Used to handle User Registration
func registerHandler(w http.ResponseWriter, r *http.Request) {
    var first = new(User)
    first.Username = r.FormValue("username")
    first.Password = r.FormValue("password")
    if first.Username == "" || first.Password == ""{
        http.Redirect(w, r, "/", http.StatusFound)
        return
    }
    body, err2 := ioutil.ReadFile("Users/allUsers.txt")
    if err2 != nil {
        fmt.Printf("FAIL")
        http.Redirect(w, r, "/", http.StatusFound)
    }
    s := string(body[:])
        scanner := bufio.NewScanner(strings.NewReader(s)) 
        scanner.Split(bufio.ScanWords)
       for scanner.Scan() {
         name := scanner.Text()
         if name == first.Username {
            http.Redirect(w, r, "/", http.StatusFound)
            fmt.Println("USERNAME MATCHES REGISTRATION FAILED")
            break
         }
         scanner.Scan()
       }    
    fmt.Println("SUCCESSFUL REGISTRATION")
    loggedInUser.Username = first.Username
    loggedInUser.Password = first.Password
    saveUser(first, w, r)
}
//Used to Handle User Login 
func loginHandler(w http.ResponseWriter, r *http.Request) {
    username := r.FormValue("username")
    password := r.FormValue("password")

    filename := "Users/" + username +"/"+ username + ".txt"
    body, err2 := ioutil.ReadFile(filename)
    if err2 != nil {
        fmt.Printf("FAIL\n")
        http.Redirect(w, r, "/", http.StatusFound)
    } else {
        p , err  := loadUserPage("Users/"+username, username)
        if err != nil {
            http.Redirect(w, r, "/", http.StatusFound)
            return
        }
        s := string(body[:])
        scanner := bufio.NewScanner(strings.NewReader(s)) 
        scanner.Scan()
        nameAndPassWord := scanner.Text()
        if (username + " " + password) == nameAndPassWord {
            fmt.Printf("Successful Login\n")
            loggedInUser.Username = username
            loggedInUser.Password = password
            renderTemplate(w, "view", p)
        } else {
            fmt.Printf("FAILED LOGIN\n")
            http.Redirect(w, r, "/", http.StatusFound)
        }
        
    }
}
//Handles the display of the Global DiscussionBoard
func boardHandler(w http.ResponseWriter, r *http.Request) {
    var buffer bytes.Buffer
    numOfPosts = 0
    files, _ := ioutil.ReadDir("./Posts/")
    for _, f := range files {     
            name := f.Name()
            if name != ".DS_Store" {
                numOfPosts++
                files2, _ := ioutil.ReadDir("./Posts/"+name)
                for _, f2 := range files2 {
                    if f2.Name() != ".DS_Store" {
                        name2 := strings.Split(f2.Name(),".")
                        p , err  := loadPage("Posts/"+name+"/"+name2[0])
                        if err != nil {
                            return
                        }
                        buffer.WriteString(string(p.Body))
                    }
                }
                
            }
            
    }
    BoardPage := &Page{Title: "Discussion Board", Body: []byte(buffer.String()), HtmlBody: template.HTML(buffer.String())}
    renderTemplate(w, "board", BoardPage)
}
//Handles Logout
func logOutHandler(w http.ResponseWriter, r *http.Request) {
    loggedInUser.Username = ""
    loggedInUser.Password = ""
    http.Redirect(w, r, "/", http.StatusFound)
}

//Handles adding a friend to a User
func addFriendHandler(w http.ResponseWriter, r *http.Request) {
    userToAdd := filepath.Base(r.URL.Path[:])
    whereToGo := "/view/Users/"+loggedInUser.Username
    if userToAdd == loggedInUser.Username {
        fmt.Println("Trying to Add Self")
        http.Redirect(w, r, whereToGo, http.StatusFound)
        return
    }
    friendFile := "Users/"+loggedInUser.Username+"/friendList.txt"
    body, err2 := ioutil.ReadFile(friendFile)
    if err2 != nil {
        fmt.Println("Failed File Retrieval")
    }
    s := string(body[:])
    scanner := bufio.NewScanner(strings.NewReader(s)) 
    scanner.Split(bufio.ScanWords)
    for scanner.Scan() {
        
        name := scanner.Text()
        if userToAdd == name {
            fmt.Println("Already Friends")
            http.Redirect(w, r, whereToGo, http.StatusFound)
            return
        }
    }
    fmt.Println("ADDING")
    f, err := os.OpenFile(friendFile, os.O_APPEND|os.O_WRONLY, 0600)
    if err != nil {
        panic(err)
    }
    defer f.Close()
    if _, err = f.WriteString("<a style = text-decoration:none;color:#FFFF00; href = /view/Users/"+userToAdd+"> "+userToAdd+" </a><br>\n"); err != nil {
        panic(err)
    }
    http.Redirect(w, r, whereToGo, http.StatusFound)


}

func main() {
    
    http.HandleFunc("/addFriend/", addFriendHandler)
	http.HandleFunc("/view/", viewHandler)
    http.HandleFunc("/logOut/", logOutHandler)
    http.HandleFunc("/commentPost/", commentPostHandler)
    http.HandleFunc("/globalPost/", globalPostHandler)
    http.HandleFunc("/userPost/", userPostHandler)
    http.HandleFunc("/userCommentPost/", userCommentHandler)
    http.HandleFunc("/board/", boardHandler)
    http.HandleFunc("/login/", loginHandler)
    http.HandleFunc("/register/", registerHandler)
    http.HandleFunc("/edit/", editHandler)
    http.HandleFunc("/save/", saveHandler)
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)

}