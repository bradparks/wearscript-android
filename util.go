package main
import (
	"github.com/gorilla/sessions"
	"code.google.com/p/goauth2/oauth"
	"encoding/json"
	"fmt"
	"os"
	"io"
	"net/http"
)

// Cookie store used to store the user's ID in the current session.
var store = sessions.NewCookieStore([]byte(secret))

// OAuth2.0 configuration variables.
func config(host string) *oauth.Config {
	r := &oauth.Config{
		ClientId:       clientId,
		ClientSecret:   clientSecret,
		Scope:          scopes,
		AuthURL:        "https://accounts.google.com/o/oauth2/auth",
		TokenURL:       "https://accounts.google.com/o/oauth2/token",
		AccessType:     "offline",
		ApprovalPrompt: "force",
	}
	if len(host) > 0 {
		r.RedirectURL = fullUrl + "/oauth2callback"
	}
	return r
}

func storeUserID(w http.ResponseWriter, r *http.Request, userId string) error {
	session, err := store.Get(r, sessionName)
	if err != nil {
		fmt.Println("Couldn't get sessionName")
		return err
	}
	fmt.Println("Saves session")
	session.Values["userId"] = userId
	return session.Save(r, w)
}

// userID retrieves the current user's ID from the session's cookies.
func userID(r *http.Request) (string, error) {
	session, err := store.Get(r, sessionName)
	if err != nil {
		return "", err
	}
	userId := session.Values["userId"]
	if userId != nil {
		fmt.Println("Got session: " + userId.(string))
		return userId.(string), nil
	}
	return "", nil
}

func storeCredential(userId string, token *oauth.Token, userInfo string) error {
	// Store the tokens in the datastore.
	err := setUserAttribute(userId, "user_info", userInfo)
	if err != nil {
		return err
	}
	val, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return setUserAttribute(userId, "oauth_token", string(val))
}

func authTransport(userId string) *oauth.Transport {
	flags, err := getUserFlags(userId, "flags")
	if err != nil {
		return nil
	}
	if !hasFlag(flags, "user") {
		return nil
	}
	val, err := getUserAttribute(userId, "oauth_token")
	if err != nil {
		return nil
	}
	tok := new(oauth.Token)
	err = json.Unmarshal([]byte(val), tok)
	if err != nil {
		return nil
	}
	return &oauth.Transport{
		Config:    config(""),
		Token:     tok,
	}
}

func deleteCredential(userId string) error {
	return deleteUserAttribute(userId, "oauth_token")
}

func WriteFile(filename string, data string) {
	fo, err := os.Create(filename)
	if err != nil { fmt.Println("Couldn't create file") }
	defer func() {
		if err := fo.Close(); err != nil {
			fmt.Println("Couldn't close file")
		}
	}()
	if _, err := fo.Write([]byte(data)); err != nil {
		fmt.Println("Couldn't write file")
	}
}

func ReadFile(filename string) (string, error) {
	fi, err := os.Open(filename)
	if err != nil {
		fmt.Println("Couldn't open file: " + filename)
		return "", err
	}
	defer func() {
		if err := fi.Close(); err != nil {
			fmt.Println("Couldn't close file:" + filename)
		}
	}()
	buf := make([]byte, 1024)
	data := ""
	for {
		// read a chunk
		n, err := fi.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("Couldn't read file: " + filename)
			return "", err
		}
		if n == 0 { break }
		data = data + string(buf[:n])
	}
	return data, nil
}
