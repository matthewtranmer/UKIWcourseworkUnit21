package main

import (
	handler "UKIWcoursework/Server/Handler"
	signing "UKIWcoursework/Server/Signing"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type Pages struct {
	db            *sql.DB
	template_path string
}

func (p *Pages) executeTemplates(w http.ResponseWriter, template_name string, data interface{}) error {
	document, err := template.ParseFiles(p.template_path+"base.html", p.template_path+template_name)
	if err != nil {
		return err
	}

	err = document.Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}

type DefaultTemplateData struct {
	User_details *handler.UserDetails
}

func (p *Pages) home(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	if r.URL.Path != "/" {
		fmt.Println(time.Now().Local().String() + " Page Not Found")
		return handler.HTTPerror{Code: 404, Err: nil}
	}

	fmt.Println("Called Home")

	err := p.executeTemplates(w, "home.html", DefaultTemplateData{user_details})
	if err != nil {
		return handler.HTTPerror{Code: 500, Err: err}
	}

	return nil
}

func loginUser(w http.ResponseWriter, username string) error {
	//TESTING EXPIRATION TIME
	expiration := time.Now().Unix() + 555555

	payload := map[string]string{
		"username":   username,
		"expiration": strconv.Itoa(int(expiration)),
	}

	json_payload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	signature, public_key, err := signing.GenerateSignature(string(json_payload))
	if err != nil {
		return err
	}

	token := handler.Token{
		Username:   username,
		Expiration: payload["expiration"],
		Signature:  signature,
		Public_key: public_key,
	}

	json_token, err := json.Marshal(token)
	if err != nil {
		return err
	}

	cookie := new(http.Cookie)
	cookie.Name = "auth_token"
	cookie.Value = url.PathEscape(string(json_token))

	http.SetCookie(w, cookie)
	return nil
}

type LoginTemplateData struct {
	User_details  *handler.UserDetails
	Error         bool
	Error_message string
}

func (p *Pages) login(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	fmt.Println("Called login")

	if r.Method == "POST" {
		stmt, err := p.db.Prepare("SELECT Password FROM UserData WHERE Username = ?")
		if err != nil {
			return handler.HTTPerror{Code: 500, Err: err}
		}

		err = r.ParseForm()
		if err != nil {
			return handler.HTTPerror{Code: 500, Err: err}
		}

		username := r.PostForm["username"][0]
		raw_password := r.PostForm["password"][0]
		database_hash := new(string)

		err = stmt.QueryRow(username).Scan(database_hash)
		if err != nil {
			data := LoginTemplateData{
				user_details,
				true,
				"The username you entered does not exist!",
			}

			err := p.executeTemplates(w, "login.html", data)
			if err != nil {
				return handler.HTTPerror{Code: 500, Err: err}
			}
			return nil
		}

		err = bcrypt.CompareHashAndPassword([]byte(*database_hash), []byte(raw_password))
		if err == nil {
			fmt.Println("Authenticated")

			err = loginUser(w, username)
			if err != nil {
				return handler.HTTPerror{Code: 500, Err: err}
			}

			r.ParseForm()
			redirect_url := r.Form.Get("return")

			if redirect_url == "" {
				redirect_url = "/"
			}

			http.Redirect(w, r, redirect_url, http.StatusSeeOther)
			return nil
		}

		data := LoginTemplateData{
			user_details,
			true,
			"The password you entered was invalid!",
		}
		err = p.executeTemplates(w, "login.html", data)
		if err != nil {
			return handler.HTTPerror{Code: 500, Err: err}
		}
		return nil
	}

	data := LoginTemplateData{
		user_details,
		false,
		"",
	}

	err := p.executeTemplates(w, "login.html", data)
	if err != nil {
		return handler.HTTPerror{Code: 500, Err: err}
	}
	return nil
}

func saveFormImage(r *http.Request, key string, path string, username string) error {
	form_file, headers, err := r.FormFile(key)
	if err.Error() == "http: no such file" {
		default_img, err := os.ReadFile(path + "/default/default.jpeg")
		if err != nil {
			return err
		}

		err = os.WriteFile(path+username+".jpeg", default_img, 0644)
		if err != nil {
			return err
		}

		return nil
	}

	if err != nil {
		return err
	}

	buffer := make([]byte, headers.Size)
	_, err = form_file.Read(buffer)
	if err != nil {
		return err
	}

	content_type := http.DetectContentType(buffer)

	if content_type == "image/png" {
		img, _ := png.Decode(bytes.NewReader(buffer))

		jpg_buf := new(bytes.Buffer)
		jpeg.Encode(jpg_buf, img, nil)

		buffer = jpg_buf.Bytes()
	} else if content_type != "image/jpeg" {
		return errors.New("File type unsupported")
	}

	file, err := os.OpenFile(path+username+".jpeg", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	_, err = file.Write(buffer)
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

func (p *Pages) signup(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	fmt.Println("Called signup")

	if r.Method == "POST" {
		stmt, err := p.db.Prepare("INSERT INTO UserData (Username, Password, Email, DOB, FirstName, LastName) VALUES (?, ?, ?, ?, ?, ?)")
		if err != nil {
			return handler.HTTPerror{Code: 500, Err: err}
		}

		err = r.ParseMultipartForm(1048576)
		if err != nil {
			return handler.HTTPerror{Code: 500, Err: err}
		}

		//will cause error if not sent
		DOB := r.PostForm["dob-year"][0] + "-" + r.PostForm["dob-month"][0] + "-" + r.PostForm["dob-day"][0]
		password_hash, err := bcrypt.GenerateFromPassword([]byte(r.PostForm["password"][0]), 12)
		if err != nil {
			return handler.HTTPerror{Code: 500, Err: err}
		}

		_, err = stmt.Exec(
			r.PostForm["username"][0],
			string(password_hash),
			r.PostForm["email"][0],
			DOB,
			r.PostForm["firstname"][0],
			r.PostForm["lastname"][0],
		)
		if err != nil {
			return handler.HTTPerror{Code: 500, Err: err}
		}

		defer stmt.Close()

		path := "/home/matthew/Websites/UKIWcoursework/static/profilepictures/"
		err = saveFormImage(r, "pfp", path, r.PostForm["username"][0])
		if err != nil {
			return handler.HTTPerror{Code: 500, Err: err}
		}

		loginUser(w, r.PostForm["username"][0])
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}

	err := p.executeTemplates(w, "signup.html", DefaultTemplateData{user_details})
	if err != nil {
		return handler.HTTPerror{Code: 500, Err: err}
	}

	return nil
}

func (p *Pages) myaccount(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	err := p.executeTemplates(w, "myaccount.html", DefaultTemplateData{user_details})
	if err != nil {
		return handler.HTTPerror{Code: 500, Err: err}
	}

	return nil
}

func (p *Pages) logout(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	cookie, _ := r.Cookie("auth_token")
	token, _ := handler.ParseToken(cookie)
	payload, _ := handler.GenerateSignatureToken(token)

	signing.BlacklistSignature(string(payload), token.Signature, token.Public_key)

	cookie = new(http.Cookie)
	cookie.Name = "auth_token"
	cookie.Value = "null"

	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func (p *Pages) about(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	err := p.executeTemplates(w, "about.html", DefaultTemplateData{user_details})
	if err != nil {
		return handler.HTTPerror{Code: 500, Err: err}
	}

	return nil
}


func (p *Pages) ourmarinas(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	err := p.executeTemplates(w, "ourmarinas.html", DefaultTemplateData{user_details})
	if err != nil {
		return handler.HTTPerror{Code: 500, Err: err}
	}

	return nil
}

func (p *Pages) shops(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	err := p.executeTemplates(w, "shops.html", DefaultTemplateData{user_details})
	if err != nil {
		return handler.HTTPerror{Code: 500, Err: err}
	}

	return nil
}

func (p *Pages) boats(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	err := p.executeTemplates(w, "boats.html", DefaultTemplateData{user_details})
	if err != nil {
		return handler.HTTPerror{Code: 500, Err: err}
	}

	return nil
}

func (p *Pages) search(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	err := p.executeTemplates(w, "search.html", DefaultTemplateData{user_details})
	if err != nil {
		return handler.HTTPerror{Code: 500, Err: err}
	}

	return nil
}

func main() {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}

	writer := io.MultiWriter(file, os.Stdout)
	log.SetOutput(writer)

	pages := new(Pages)
	pages.db, err = sql.Open("mysql", "matthew:MysqlPassword111@tcp(127.0.0.1:3306)/UKIW")
	if err != nil {
		panic(err)
	}

	pages.template_path = "templates/"

	//testng only
	//fs := http.FileServer(http.Dir("/home/matthew/Websites/UKIWcoursework/static"))
	//http.Handle("/static/", http.StripPrefix("/static", fs))

	//General Services
	http.Handle("/", handler.Handler{Middleware: pages.home, Require_login: false})
	http.Handle("/about", handler.Handler{Middleware: pages.about, Require_login: false})
	http.Handle("/ourmarinas", handler.Handler{Middleware: pages.ourmarinas, Require_login: false})
	http.Handle("/sales/shops", handler.Handler{Middleware: pages.shops, Require_login: false})
	http.Handle("/sales/boats", handler.Handler{Middleware: pages.boats, Require_login: false})
	http.Handle("/search", handler.Handler{Middleware: pages.search, Require_login: false})
	
	//Acount Services
	http.Handle("/accounts/signup", handler.Handler{Middleware: pages.signup, Require_login: false})
	http.Handle("/accounts/login", handler.Handler{Middleware: pages.login, Require_login: false})
	http.Handle("/accounts/myaccount", handler.Handler{Middleware: pages.myaccount, Require_login: true})
	http.Handle("/accounts/logout", handler.Handler{Middleware: pages.logout, Require_login: true})

	fmt.Println("Server Started!")
	http.ListenAndServe("127.0.0.1:8000", nil)

}
