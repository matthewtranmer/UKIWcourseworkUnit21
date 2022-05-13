package Handler

import (
	"log"
	"net/http"
)

type UserDetails struct {
	Username string
}

func checkToken(cookie *http.Cookie) (user_details *UserDetails, err error) {
	if cookie == nil {
		return nil, nil
	}

	if cookie.Value == "logged_in" {
		user_details = &UserDetails{"matthew"}
		return user_details, nil
	}

	return nil, nil
}

type Handler struct {
	Middleware    func(w http.ResponseWriter, r *http.Request, user_details *UserDetails) ErrorResponse
	Require_login bool
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("auth_token")
	user_details, err := checkToken(cookie)

	if user_details == nil && h.Require_login {
		url := "/accounts/login?return=" + r.URL.Path
		http.Redirect(w, r, url, http.StatusSeeOther)
		return
	}

	var http_error ErrorResponse
	if err == nil {
		http_error = h.Middleware(w, r, user_details)
		if http_error == nil {
			return
		}
	}

	if err != nil {
		http_error = HTTPerror{500, err}
	}

	if http_error.GetLogError() != nil {
		log.Println(http_error.GetLogError())
	}

	//handle error
	w.Header().Add("content-type", "text/html")
	w.WriteHeader(http_error.GetCode())

	message := "<h1>" + http_error.GetError() + "</h1>"
	w.Write([]byte(message))
}

type ErrorResponse interface {
	GetCode() int
	GetError() string
	GetLogError() error
}

type HTTPerror struct {
	Code int
	Err  error
}

func (e HTTPerror) GetLogError() error {
	return e.Err
}

func (e HTTPerror) GetCode() int {
	return e.Code
}

func (e HTTPerror) GetError() string {
	switch e.Code {
	case 404:
		return "404 - Page Not Found"
	case 500:
		return "500 - Internal Server Error"
	}

	return ""
}
