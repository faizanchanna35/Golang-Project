package test_cases

import (
	"errors"
	"mymodule/database"


	"golang.org/x/crypto/bcrypt"
)

type login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type register struct {

	Name    string `json:"name"`    
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login() (bool, error) {
	db := database.Connect()

	defer db.Close()
	var userId string
	var pass string
	var role_id int

	login := login{
		Email:    "channa321@gmail.com",
		Password: "faizan",
	}

   err := db.QueryRow(`SELECT id, password , role_id FROM users WHERE email = $1`, login.Email).Scan(&userId, &pass, &role_id)
			if err != nil {
				return false, errors.New("invalid username or password")
			}

			error_password := bcrypt.CompareHashAndPassword([]byte(pass), []byte(login.Password))
			if error_password != nil {
				return false, errors.New("invalid username or password")
			}

return true, nil

}

func Register() (bool, error){

   db := database.Connect()
   defer db.Close()

   register := register{
      Name:     "channa",
		Email:    "channafaizan@gmail.com",
		Password: "faizan",
	}

   bytes, err := bcrypt.GenerateFromPassword([]byte(register.Password), 14)
   if err != nil {
      panic(err)
   }

   _, err = db.Exec(`INSERT INTO users (name, email, password) VALUES ($1,$2,$3)`, register.Name, register.Email, bytes)
   if err != nil {
      return false, errors.New("Not insert")

   }
   return true, nil

}
