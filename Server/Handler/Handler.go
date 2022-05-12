package Handler

import (
	signing "UKIWcoursework/Server/Signing"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type UserDetails struct {
	Username string
}

var errTokenInvalid error = errors.New("the given token is invalid")
var errTokenExpired error = errors.New("the given token is expired")

type Token struct {
	Username   string
	Expiration string
	Signature  string
	Public_key string
}

func ParseToken(cookie *http.Cookie) (*Token, error) {
	if cookie == nil || cookie.Value == "null" {
		return nil, nil
	}

	unescaped_token, err := url.PathUnescape(cookie.Value)
	if err != nil {
		return nil, err
	}

	json_token := []byte(unescaped_token)

	token := new(Token)
	err = json.Unmarshal(json_token, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func GenerateSignatureToken(token *Token) ([]byte, error) {
	payload := map[string]string{
		"username":   token.Username,
		"expiration": token.Expiration,
	}

	json_payload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return json_payload, err
}

func checkToken(cookie *http.Cookie) (user_details *UserDetails, err error) {
	token, err := ParseToken(cookie)
	if err != nil {
		return nil, err
	}

	if token == nil {
		return nil, nil
	}

	expiration, err := strconv.Atoi(token.Expiration)
	if err != nil {
		return nil, err
	}

	if expiration < int(time.Now().Unix()) {
		fmt.Println("Expired")
		return nil, errTokenExpired
	}

	json_payload, err := GenerateSignatureToken(token)
	if err != nil {
		return nil, err
	}

	verified, err := signing.VerifySignature(string(json_payload), token.Signature, token.Public_key)
	if err != nil {
		return nil, err
	}

	if !verified {
		return nil, errTokenInvalid
	}

	user_details = &UserDetails{token.Username}
	return user_details, nil
}

type Handler struct {
	Middleware    func(w http.ResponseWriter, r *http.Request, user_details *UserDetails) ErrorResponse
	Require_login bool
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("auth_token")
	user_details, err := checkToken(cookie)

	if err == errTokenExpired || err == errTokenInvalid {
		cookie := new(http.Cookie)
		cookie.Name = "auth_token"
		cookie.Value = "null"

		http.SetCookie(w, cookie)
		//we have fixed the error
		err = nil
	}

	if user_details == nil && h.Require_login {
		url := "/login?return=" + r.URL.Path
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
