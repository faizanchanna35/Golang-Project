package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	_ "image"
	"io"
	"io/ioutil"
	"log"
	"mymodule/database"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var store = sessions.NewCookieStore([]byte("super-secret"))

func StaticHandler(tpl Template) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "/register" {
			tpl.Execute(w, nil)
			return
		} else if r.URL.Path == "/dashboard" {

			session, _ := store.Get(r, "session")
			id, ok := session.Values["userId"]
			fmt.Println("ok: ", ok)
			if !ok {
				http.Redirect(w, r, "/signin", http.StatusFound) // http.StatusFound is 302
				return
			}
			fmt.Println(id)

			db := database.Connect()

			var User struct {
				Name    string
				Email   string
				Role_id int
			}
			var name string
			var email string
			var role_id int

			err := db.QueryRow("SELECT name, email, role_id FROM users WHERE id = $1", id).Scan(&name, &email, &role_id)
			if err != nil {
				fmt.Println("error")
			}
			User.Name = name
			User.Email = email
			User.Role_id = role_id
			tpl.Execute(w, User)
			return

			// type Product struct {
			// 	ID       int
			// 	Name     string
			// 	Email    string
			// 	Password string
			// }

			// db := database.Connect()

			// rows, err := db.Query("SELECT * FROM users")
			// if err != nil {
			// 	fmt.Println("error")
			// }
			// defer rows.Close()
			// var products []Product
			// for rows.Next() {
			// 	var p Product
			// 	err = rows.Scan(&p.ID, &p.Name, &p.Email, &p.Password)
			// 	if err != nil {
			// 		panic(err)
			// 	}
			// 	products = append(products, p)
			// }
			// // fmt.Println(products)

		} else if r.URL.Path == "/register_process" {

			// log.Println(name, email, password)
			// var UserRegister struct{
			// 	Name string
			// 	email string
			// 	Password string
			// }

			db := database.Connect()
			defer db.Close()

			r.ParseForm()
			name := r.FormValue("name")
			email := r.FormValue("email")
			password := r.FormValue("password")
			bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
			if err != nil {
				panic(err)
			}

			_, err = db.Exec(`INSERT INTO users (name, email, password) VALUES ($1,$2,$3)`, name, email, bytes)
			if err != nil {
				log.Println("eorrsd", err)
			}
			http.Redirect(w, r, "/signin", http.StatusFound)

		} else if r.URL.Path == "/signin" {
			tpl.Execute(w, nil)
			return
		} else if r.URL.Path == "/signin_process" {

			r.ParseForm()
			email := r.FormValue("email")
			password := r.FormValue("password")

			db := database.Connect()

			defer db.Close()
			var userId string
			var pass string
			var role_id int

			err := db.QueryRow(`SELECT id, password , role_id FROM users WHERE email = $1`, email).Scan(&userId, &pass, &role_id)
			if err != nil {
				http.Redirect(w, r, "/signin", http.StatusFound)
				return
			}

			error_password := bcrypt.CompareHashAndPassword([]byte(pass), []byte(password))
			if error_password != nil {
				http.Redirect(w, r, "/signin", http.StatusFound)
				return
			} else {

				session, _ := store.Get(r, "session")
				session.Values["userId"] = userId
				session.Save(r, w)

				http.Redirect(w, r, "/dashboard", http.StatusFound)
				return

			}

		} else if r.URL.Path == "/All_user" {
			db := database.Connect()
			session, _ := store.Get(r, "session")
			id, ok := session.Values["userId"]
			fmt.Println("ok: ", ok)
			if !ok {
				http.Redirect(w, r, "/signin", http.StatusFound) // http.StatusFound is 302
				return
			}
			fmt.Println(id)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
            defer cancel()
			var role_id int
			 err := db.QueryRowContext(ctx,"SELECT role_id FROM users WHERE id=$1", id).Scan(&role_id)
			if err != nil {
				panic(err)
			}
        
		
		  	type Product struct {
				SessionId interface{}
				Role_id   int
				ID        int
				Name      string
				Email     string
				Password  string
				Login_role_id int
			}

			

			rows, err := db.Query("SELECT * FROM users")
			if err != nil {
				fmt.Println("error")
			}
			defer rows.Close()
			var products []Product

			for rows.Next() {
				var p Product
				err = rows.Scan(&p.ID, &p.Name, &p.Email, &p.Password, &p.Role_id)
				if err != nil {
					panic(err)
				}
				p.SessionId = id
				p.Login_role_id = role_id

				products = append(products, p)
			}

			// fmt.Println(products)
			tpl.Execute(w, products)
			return

		} else if r.URL.Path == "/edit" {
			r.ParseForm()

			id := r.FormValue("id")

			db := database.Connect()

			defer db.Close()

			var data struct {
				Role_id  int
				ID       int
				Name     string
				Email    string
				Password string
			}
			var role_id int
			var index int
			var name string
			var email string
			var password string

			err := db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&index, &name, &email, &password, &role_id)

			if err != nil {
				log.Println("error")
				http.Redirect(w, r, "/signin", http.StatusFound)

				return
			}
			data.ID = index
			data.Email = email
			data.Name = name
			data.Password = password

			tpl.Execute(w, data)
			return
		} else if r.URL.Path == "/edit_process" {

			r.ParseForm()
			id := r.FormValue("id")
			name := r.FormValue("name")
			email := r.FormValue("email")
			password := r.FormValue("password")

			bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)

			if err != nil {
				panic(err)
			}

			db := database.Connect()

			defer db.Close()
			value, err := db.Exec(`UPDATE users SET name = $1, email = $2, password = $3 WHERE id = $4`, name, email, bytes, id)
			if err != nil {
				panic(err)
			}

			if value != nil {
				http.Redirect(w, r, "/All_user", http.StatusFound)
				return
			}

		} else if r.URL.Path == "/delete" {
			r.ParseForm()

			id := r.FormValue("id")

			db := database.Connect()

			defer db.Close()
			value, err := db.Exec(`DELETE FROM users WHERE id=$1;`, id)
			if err != nil {
				panic(err)
			}
			if value != nil {
				http.Redirect(w, r, "/All_user", http.StatusFound)
				return
			}

		} else if r.URL.Path == "/logout" {
			fmt.Println("logouting .........! ")
			session, _ := store.Get(r, "session")
			delete(session.Values, "userId")
			session.Save(r, w)
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		} else if r.URL.Path == "/manage_role" {
			session, _ := store.Get(r, "session")
			id, ok := session.Values["userId"]
			fmt.Println("ok: ", ok)
			if !ok {
				http.Redirect(w, r, "/signin", http.StatusFound) // http.StatusFound is 302
				return
			}
			fmt.Println(id)
			type Product struct {
				SessionId interface{}
				ID        int
				Name      string
				Email     string
				Password  string
				Role_id   int
			}

			db := database.Connect()

			rows, err := db.Query("SELECT * FROM users")
			if err != nil {
				fmt.Println("error")
			}
			defer rows.Close()
			var products []Product

			for rows.Next() {
				var p Product
				err = rows.Scan(&p.ID, &p.Name, &p.Email, &p.Password, &p.Role_id)
				if err != nil {
					panic(err)
				}
                p.SessionId = id
				products = append(products, p)     
			}
			tpl.Execute(w, products)
			return

		} else if r.URL.Path == "/edit_role_process" {
			fmt.Println("edit_role_process")
			r.ParseForm()
			user_id := r.FormValue("user_id")
			id := r.FormValue("edit-" + user_id)
			db := database.Connect()
			value, err := db.Exec(`UPDATE users SET role_id = $1 WHERE id = $2`, id, user_id)
			if err != nil {
				panic(err)
			}

			if value != nil {
				http.Redirect(w, r, "/manage_role", http.StatusFound) // http.StatusFound is 302
				return
			}
		} else if r.URL.Path == "/All_posts" {
            db := database.Connect()
			session, _ := store.Get(r, "session")
			id, ok := session.Values["userId"]
			fmt.Println("ok: ", ok)
			if !ok {
				http.Redirect(w, r, "/signin", http.StatusFound) // http.StatusFound is 302
				return
			}
			fmt.Println(id)
			
			var role_id int
			 err := db.QueryRow("SELECT role_id FROM users WHERE id=$1", id).Scan(&role_id)
			if err != nil {
				panic(err)
			}
			
        
			type Product struct {
				SessionId   interface{}
				ID          int
				Title       string
				Description string

				User_id  int
				Username string
				Image    string
				Login_role_id int
			}

			

			rows, err := db.Query("SELECT * FROM posts")
			if err != nil {
				fmt.Println("error")
			}
			defer rows.Close()
			var products []Product

			for rows.Next() {
				var p Product
				err = rows.Scan(&p.ID, &p.Title, &p.Description, &p.User_id, &p.Image)
				if err != nil {
					panic(err)
				}

				// rows, err := db.Query("SELECT name FROM users WHERE id = $1", p.User_id)
				//  if err != nil {
				// 	fmt.Println("error")
				// }
				// rows.Scan(&p.Username)
                p.Login_role_id = role_id

				products = append(products, p)
			}

			// fmt.Println(products)
			tpl.Execute(w, products)
			return

		} else if r.URL.Path == "/"{

			type Product struct {
				SessionId   interface{}
				ID          int
				Title       string
				Description string

				User_id  int
				Username string
				Image    string
			}

			db := database.Connect()

			rows, err := db.Query("SELECT * FROM posts")
			if err != nil {
				fmt.Println("error")
			}
			defer rows.Close()
			var products []Product

			for rows.Next() {
				var p Product
				err = rows.Scan(&p.ID, &p.Title, &p.Description, &p.User_id, &p.Image)
				if err != nil {
					panic(err)
				}

				// rows, err := db.Query("SELECT name FROM users WHERE id = $1", p.User_id)
				//  if err != nil {
				// 	fmt.Println("error")
				// }
				// rows.Scan(&p.Username)

				products = append(products, p)
			}

			// fmt.Println(products)
			tpl.Execute(w, products)
			return 
 
		} else if r.URL.Path == "/insert_post_process" {

			session, _ := store.Get(r, "session")
			id, ok := session.Values["userId"]

			fmt.Println("ok: ", ok)
			if !ok {
				http.Redirect(w, r, "/signin", http.StatusFound) // http.StatusFound is 302
				return
			}
			fmt.Println(id)
			result, err := strconv.Atoi(id.(string))
			if err != nil {
				log.Println("eorrsd", err)
			}

			file, header, err := r.FormFile("image")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer file.Close()

			// get the folder path where you want to store the uploaded image
			uploadDir := "./static/uploads/"

			// create the folder if it doesn't exist
			if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
				if err = os.MkdirAll(uploadDir, 0755); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			// create a new file in the upload folder with the same name as the uploaded file
			dst, err := os.Create(filepath.Join(uploadDir, header.Filename))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			// copy the uploaded file to the new file
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			log.Println("done")
			// file, header, err := r.FormFile("image")
			// if err != nil {
			// 	http.Error(w, err.Error(), http.StatusBadRequest)
			// 	return
			// }
			// defer file.Close()

			// write the file to disk
			// f, err := os.Create("./uploads" + header.Filename)
			// if err != nil {
			// 	http.Error(w, err.Error(), http.StatusInternalServerError)
			// 	return
			// }
			// defer f.Close()

			// if _, err := io.Copy(f, file); err != nil {
			// 	http.Error(w, err.Error(), http.StatusInternalServerError)
			// 	return
			// }

			// add filename to this user's cookie

			// add filename to this user's cookie

			// file, header, err := r.FormFile("image")
			// if err != nil {
			// 	http.Error(w, err.Error(), http.StatusBadRequest)
			// 	return
			// }
			// defer file.Close()

			// out, err := os.Create("./uploads" + header.Filename)
			// if err != nil {
			// 	http.Error(w, err.Error(), http.StatusInternalServerError)
			// 	return
			// }
			// defer out.Close()
			db := database.Connect()
			defer db.Close()

			r.ParseForm()

			title := r.FormValue("title")
			description := r.FormValue("description")

			_, err = db.Exec(`INSERT INTO posts (title, description, user_id, image) VALUES ($1,$2,$3,$4)`, title, description, result, header.Filename)
			if err != nil {
				log.Println("eorrsd", err)
			}
			http.Redirect(w, r, "/All_posts", http.StatusFound)

		} else if r.URL.Path == "/edit_post" {
			r.ParseForm()

			id := r.FormValue("id")

			db := database.Connect()

			defer db.Close()

			var data struct {
				User_id     int
				ID          int
				Title       string
				Description string
				Image       string
			}
			var user_id      int
			var index        int
			var title        string
			var description  string
			var image        string

			err := db.QueryRow("SELECT * FROM posts WHERE id = $1", id).Scan(&index, &title, &description, &user_id, &image)

			if err != nil {
				log.Println("error")
				http.Redirect(w, r, "/signin", http.StatusFound)

				return
			}
			data.ID = index
			data.Title = title
			data.Description = description
			data.Image = image
 
			tpl.Execute(w, data)
			return
		} else if r.URL.Path == "/edit_post_process" {
			db := database.Connect()
			r.ParseForm()
			
			file, header, err := r.FormFile("image")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer file.Close()

			// get the folder path where you want to store the uploaded image
			uploadDir := "./static/uploads/"

			// create the folder if it doesn't exist
			if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
				if err = os.MkdirAll(uploadDir, 0755); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			// create a new file in the upload folder with the same name as the uploaded file
			dst, err := os.Create(filepath.Join(uploadDir, header.Filename))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			// copy the uploaded file to the new file
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

	
			id := r.FormValue("id")
			
			title := r.FormValue("title")
			description := r.FormValue("description")

			defer db.Close()
			value, err := db.Exec(`UPDATE posts SET title = $1, description = $2, image = $3 WHERE id = $4`, title, description, header.Filename, id)

			if err != nil {
				panic(err)
			}

			if value != nil {
				http.Redirect(w, r, "/All_posts", http.StatusFound)
				return
			}  

		} else if r.URL.Path == "/delete_posts" {
			r.ParseForm()

			id := r.FormValue("id")

			db := database.Connect()

			defer db.Close()
			value, err := db.Exec(`DELETE FROM posts WHERE id=$1;`, id)
			if err != nil {
				panic(err)
			}
			if value != nil {
				http.Redirect(w, r, "/All_posts", http.StatusFound)
				return
			}

		} else if r.URL.Path == "/contact_process" {

			db := database.Connect()
			defer db.Close()

			r.ParseForm()
			name := r.FormValue("name")
			email := r.FormValue("email")
			description := r.FormValue("description")

			_, err := db.Exec(`INSERT INTO contact (name, email, description) VALUES ($1,$2,$3)`, name, email, description)

			if err != nil {
				log.Println("eorrsd", err)
			}
			username := "mfchanna6@gmail.com"
			password := "ztgjvdsyqkislqiz"
			auth := smtp.PlainAuth("", username, password, "smtp.gmail.com")

			// Set up the email message.
			to := []string{email}
			msg := []byte("To: " + email + "\r\n" +
				"Subject: " + name + "\r\n" +
				"\r\n" +
				description + " .\r\n")

			// Send the email message.
			err = smtp.SendMail("smtp.gmail.com:587", auth, username, to, msg)
			if err != nil {
				panic(err)
			}

			http.Redirect(w, r, "/contact", http.StatusFound)

		} else if r.URL.Path == "/weather" {

			type weatherData struct {
				Name string `json:"name"`
				Main struct {
					Kelvin float64 `json:"temp"`
				} `json:"main"`
			}
			r.ParseForm()

			weather2 := r.FormValue("weather")
			var city string
			if weather2 == "" {
				city = "jamshoro"
			} else {
				city = weather2
			}

			apiKey := "b56c890cb4dcaf50820f3b461515751f"
			url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s", city, apiKey)

			resp, err := http.Get(url)
			if err != nil {
				fmt.Println("Error getting weather data: ", err)
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response body: ", err)
				return
			}

			var weather weatherData
			err = json.Unmarshal(body, &weather)
			if err != nil {
				fmt.Println("Error unmarshaling JSON response: ", err)
				return
			}

			tmpl, err := template.New("weather").Parse("The weather in {{.Name}} is {{.Temp}} degrees Kelvin.")
			if err != nil {
				fmt.Println("Error parsing template: ", err)
				return
			}

			err = tmpl.Execute(os.Stdout, struct {
				Name string
				Temp float64
			}{
				Name: weather.Name,
				Temp: weather.Main.Kelvin,
			})
			if err != nil {
				fmt.Println("Error executing template: ", err)
				return
			}

			tpl.Execute(w, weather)
			return

		}
		tpl.Execute(w, nil)

	}

}

// ztgjvdsyqkislqiz
