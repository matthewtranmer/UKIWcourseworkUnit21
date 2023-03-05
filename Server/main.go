package main

import (
	handler "UKIWcoursework/Server/Handler"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type Pages struct {
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
		fmt.Println(time.Now().Local().String() + " " + r.URL.Path + " Page Not Found")
		return handler.HTTPerror{Code: 404, Err: nil}
	}

	err := p.executeTemplates(w, "home.html", DefaultTemplateData{user_details})
	if err != nil {
		return handler.HTTPerror{Code: 500, Err: err}
	}

	return nil
}

func loginUser(w http.ResponseWriter, username string) {
	cookie := new(http.Cookie)
	cookie.Name = "auth_token"
	cookie.Value = "logged_in"
	cookie.Path = "/"

	http.SetCookie(w, cookie)
}

type LoginTemplateData struct {
	User_details  *handler.UserDetails
	Error         bool
	Error_message string
}

func (p *Pages) login(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	fmt.Println("Called login")

	if r.Method == "POST" {
		r.ParseForm()

		username := r.PostForm["username"][0]

		if username != "matthew" {
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

		raw_password := r.PostForm["password"][0]
		//matthew - testing only
		password_hash := "$2a$12$8l0Y3aEgv0Qyq4M87BBjaO7XxMN6lem6TKMphA8Tod5TfD.TFu.Ou"

		err := bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(raw_password))
		if err == nil {
			fmt.Println("Authenticated")

			loginUser(w, username)

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

func (p *Pages) myaccount(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	err := p.executeTemplates(w, "myaccount.html", DefaultTemplateData{user_details})
	if err != nil {
		return handler.HTTPerror{Code: 500, Err: err}
	}

	return nil
}

func (p *Pages) logout(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	cookie := new(http.Cookie)
	cookie.Name = "auth_token"
	cookie.Value = "null"
	cookie.Path = "/"

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

func (p *Pages) shops(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	err := p.executeTemplates(w, "shops.html", DefaultTemplateData{user_details})
	if err != nil {
		return handler.HTTPerror{Code: 500, Err: err}
	}

	return nil
}

type BoatTemplateData struct {
	User_details *handler.UserDetails
	Boats        []Boat
}

type Boat struct {
	Name  string
	Image string
	Price string
}

func (p *Pages) boats(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	boat_storage := []Boat{{"Boat 0", "0", "£600"}, {"Boat 1", "1", "£1900"}, {"Boat 2", "2", "£1400"}, {"Boat 3", "3", "£6000"}, {"Boat 4", "4", "£1000"}, {"Boat 5", "5", "£49,999"}, {"Boat 6", "6", "£4,000"}}
	err := p.executeTemplates(w, "boats.html", BoatTemplateData{user_details, boat_storage})
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

func (p *Pages) signup(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	err := p.executeTemplates(w, "signup.html", DefaultTemplateData{user_details})
	if err != nil {
		return handler.HTTPerror{Code: 500, Err: err}
	}

	return nil
}

func (p *Pages) claydonmarina(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	err := p.executeTemplates(w, "claydonmarina.html", DefaultTemplateData{user_details})
	if err != nil {
		return handler.HTTPerror{Code: 500, Err: err}
	}

	return nil
}

func (p *Pages) bills(w http.ResponseWriter, r *http.Request, user_details *handler.UserDetails) handler.ErrorResponse {
	err := p.executeTemplates(w, "bills.html", DefaultTemplateData{user_details})
	if err != nil {
		return handler.HTTPerror{Code: 500, Err: err}
	}

	return nil
}

func main() {
	log.SetOutput(os.Stdout)

	pages := new(Pages)
	pages.template_path = "templates/"

	//testng only
	//fs := http.FileServer(http.Dir("/home/matthew/Websites/UKIWcoursework/static"))
	fs := http.FileServer(http.Dir("C:/Users/Matthew/Desktop/github.com/matthewtranmer/UKIWcourseworkUnit21/static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))

	//General Services
	http.Handle("/", handler.Handler{Middleware: pages.home, Require_login: false})
	http.Handle("/about", handler.Handler{Middleware: pages.about, Require_login: false})
	http.Handle("/claydonmarina", handler.Handler{Middleware: pages.claydonmarina, Require_login: false})
	http.Handle("/sales/shops", handler.Handler{Middleware: pages.shops, Require_login: false})
	http.Handle("/sales/boats", handler.Handler{Middleware: pages.boats, Require_login: false})
	http.Handle("/search", handler.Handler{Middleware: pages.search, Require_login: false})

	//Acount Services
	http.Handle("/accounts/signup", handler.Handler{Middleware: pages.signup, Require_login: false})
	http.Handle("/accounts/login", handler.Handler{Middleware: pages.login, Require_login: false})
	http.Handle("/accounts/myaccount", handler.Handler{Middleware: pages.myaccount, Require_login: true})
	http.Handle("/accounts/logout", handler.Handler{Middleware: pages.logout, Require_login: true})
	http.Handle("/accounts/bills", handler.Handler{Middleware: pages.bills, Require_login: true})

	fmt.Println("Server Started!")
	http.ListenAndServe("127.0.0.1:8000", nil)
}
