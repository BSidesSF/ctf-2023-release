package main

import (
	"bufio"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	html "html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	babel "github.com/jvatic/goja-babel"
	"golang.org/x/crypto/bcrypt"
)

const (
	BASE_NAME            = "base.html"
	DEFAULT_DSN          = "lastpwned:hacktheplanet@tcp(127.0.0.1:3306)/lastpwned"
	DEFAULT_TOKEN_SECRET = "tossed salads and scrambled eggs"
	TOKEN_HEADER         = "X-Auth-Token"
	TOKEN_EXPIRES        = 24 * time.Hour
)

type contextKey string

var (
	useTemplateCache = true
	templateCache    map[string]*html.Template
	templateDir      = "./templates"
	staticDir        = "./static"
	listenAddr       = "0.0.0.0:3000"
	templateFuncMap  map[string]any
	dbHandle         *sql.DB
	dbHandleOnce     sync.Once
	tokenSecret      []byte
	usernameKey      = contextKey("username")
	compileJSX       = true
	jsxCache         map[string]string
	babelOptions     = map[string]interface{}{
		"plugins": []string{
			"syntax-jsx",
			"transform-react-jsx",
			"transform-react-display-name",
			"transform-react-jsx-self",
			"transform-react-jsx-source",
		},
	}
)

func main() {
	// static server
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	http.Handle("/jsx/", http.StripPrefix("/jsx/", http.HandlerFunc(JSXServer)))

	// handlers
	http.HandleFunc("/api/login", BGContextAdapter(JSONMethodAdapter(LoginHandler)))
	http.HandleFunc("/api/register", BGContextAdapter(JSONMethodAdapter(RegisterHandler)))
	http.HandleFunc("/api/keybag", AuthRequiredAdapter(KeybagHandler))
	http.HandleFunc("/api/keybag/history", AuthRequiredAdapter(KeybagHistoryListHandler))
	historyPrefix := "/api/keybag/history/"
	http.Handle(historyPrefix, http.StripPrefix(
		historyPrefix, http.HandlerFunc(AuthRequiredAdapter(KeybagHistoryGetHandler))))
	http.HandleFunc("/api/", http.NotFound)
	http.HandleFunc("/favicon.ico", http.NotFound)
	http.HandleFunc("/", indexHandler)

	// listening
	log.Printf("Starting listening on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func init() {
	useTemplateCache = os.Getenv("RELOAD_TEMPLATES") == ""
	compileJSX = os.Getenv("DEV_JSX") == ""

	templateDir = Getenv("TEMPLATE_DIR", templateDir)
	staticDir = Getenv("STATIC_DIR", staticDir)
	listenAddr = Getenv("LISTEN_ADDR", listenAddr)
	tokenSecret = []byte(Getenv("TOKEN_SECRET", DEFAULT_TOKEN_SECRET))

	templateCache = make(map[string]*html.Template)
	templateFuncMap = make(map[string]any)
	templateFuncMap["jsx"] = JSXSource
	jsxCache = make(map[string]string)
	if err := babel.Init(4); err != nil {
		panic(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", nil)
}

func renderTemplate(w http.ResponseWriter, name string, data any) {
	tmpl, err := GetTemplate(name)
	if err != nil {
		log.Printf("Error getting template %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err = tmpl.ExecuteTemplate(w, BASE_NAME, data); err != nil {
		log.Printf("Error rendering template %s: %v", name, err)
	}
}

func GetTemplate(name string) (*html.Template, error) {
	if useTemplateCache {
		if tmpl, ok := templateCache[name]; ok {
			return tmpl, nil
		}
	}
	tmpl, err := loadTemplate(name)
	if err != nil {
		return nil, err
	}
	log.Printf("Loaded %s %s", name, tmpl.DefinedTemplates())
	templateCache[name] = tmpl
	return tmpl, nil
}

func loadTemplate(name string) (*html.Template, error) {
	log.Printf("Loading template %s", name)
	tmpls := []string{
		path.Join(templateDir, BASE_NAME),
		path.Join(templateDir, name),
	}
	return html.New(BASE_NAME).Funcs(templateFuncMap).ParseFiles(tmpls...)
}

func Getenv(name, def string) string {
	if val, ok := os.LookupEnv(name); ok {
		return val
	}
	return def
}

func JSXServer(w http.ResponseWriter, r *http.Request) {
	name := path.Base(r.URL.Path)
	if path.Ext(name) != ".jsx" {
		log.Printf("jsx handler called for non-jsx: %s", name)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if val, ok := jsxCache[name]; ok {
		w.Header().Add("Content-type", "text/javascript")
		io.WriteString(w, val)
		return
	}
	fp, err := os.Open(path.Join(staticDir, name))
	if err != nil {
		log.Printf("error opening jsx source: %v", err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	defer fp.Close()
	rdr := bufio.NewReader(fp)
	jsRdr, err := babel.Transform(rdr, babelOptions)
	if err != nil {
		log.Printf("error performing babel transform: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	var b strings.Builder
	if _, err := io.Copy(&b, jsRdr); err != nil {
		log.Printf("error getting js source: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	js := b.String()
	jsxCache[name] = js
	w.Header().Add("Content-type", "text/javascript")
	io.WriteString(w, js)
}

func JSXSource(name string) html.HTML {
	if compileJSX {
		return html.HTML(fmt.Sprintf(`<script src="/jsx/%s" type="text/javascript"></script>`, name))
	}
	return html.HTML(fmt.Sprintf(`<script src="/static/%s" type="text/jsx"></script>`, name))
}

func GetDB() *sql.DB {
	dbHandleOnce.Do(func() {
		dsn := Getenv("MYSQL_DSN", DEFAULT_DSN)
		log.Printf("connecting to %s", dsn)
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}
		dbHandle = db
	})
	return dbHandle
}

func BGContextAdapter(f func(context.Context, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f(context.Background(), w, r)
	}
}

func AuthContextAdapter(f func(context.Context, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		token := r.Header.Get(TOKEN_HEADER)
		if token != "" {
			username, err := VerifyAuthToken(token)
			if err != nil {
				log.Printf("error verifying token: %v", err)
				http.Error(w, "Unauthorized", http.StatusForbidden)
				return
			}
			ctx = context.WithValue(ctx, usernameKey, username)
		}
		f(ctx, w, r)
	}
}

func AuthRequiredAdapter(f func(context.Context, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return AuthContextAdapter(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if v := ctx.Value(usernameKey); v == nil {
			log.Printf("auth required but not found")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		f(ctx, w, r)
	})
}

func JSONMethodAdapter[T any, V any](handler func(context.Context, http.ResponseWriter, *http.Request, *T) V) func(context.Context, http.ResponseWriter, *http.Request) {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Print("Non-post method for JSON")
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()
		if r.Header.Get("Content-type") != "application/json" {
			log.Print("Not getting JSON")
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		var data T
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&data); err != nil {
			log.Printf("Error decoding JSON body: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		rv := handler(ctx, w, r, &data)
		SendJSON(w, rv)
	}
}

type StatusCodeProvider interface {
	StatusCode() int
}

func SendJSON(w http.ResponseWriter, data any) {
	if p, ok := data.(StatusCodeProvider); ok {
		c := p.StatusCode()
		if c == 0 {
			c = 100
		}
		w.WriteHeader(c)
	}
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error encoding json: %v", err)
	}
}

// Auth tokens

func MakeAuthToken(username string) string {
	username = NormalizeUsername(username)
	exp := time.Now().Add(TOKEN_EXPIRES).Unix()
	toSign := fmt.Sprintf("%s:%d", username, exp)
	h := hmac.New(sha256.New, tokenSecret)
	h.Write([]byte(toSign))
	mac := h.Sum(nil)
	encodedMAC := base64.RawStdEncoding.EncodeToString(mac)
	return base64.RawStdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", toSign, encodedMAC)))
}

func VerifyAuthToken(token string) (string, error) {
	rawToken, err := base64.RawStdEncoding.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("error decoding auth token: %w", err)
	}
	pieces := strings.Split(string(rawToken), ":")
	if len(pieces) != 3 {
		return "", fmt.Errorf("token expected 3 pieces, got %d", len(pieces))
	}
	tstamp, err := strconv.ParseInt(pieces[1], 10, 64)
	if err != nil {
		return "", fmt.Errorf("error parsing timestamp: %w", err)
	}
	ts := time.Unix(tstamp, 0)
	if time.Now().After(ts) {
		return "", fmt.Errorf("auth token is expired")
	}
	toSign := fmt.Sprintf("%s:%s", pieces[0], pieces[1])
	h := hmac.New(sha256.New, tokenSecret)
	h.Write([]byte(toSign))
	mac := h.Sum(nil)
	gotMac, err := base64.RawStdEncoding.DecodeString(pieces[2])
	if err != nil {
		return "", fmt.Errorf("error decoding auth sig: %w", err)
	}
	if hmac.Equal(mac, gotMac) {
		return pieces[0], nil
	}
	return "", fmt.Errorf("bad hmac signature")
}

func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func GetUsernameOrUnauthorized(ctx context.Context, w http.ResponseWriter) string {
	if v := ctx.Value(usernameKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	log.Printf("unable to get username from context")
	http.Error(w, "Forbidden", http.StatusForbidden)
	return ""
}

// API Handlers

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Username string `json:"username,omitempty"`
	Success  bool   `json:"success"`
	Message  string `json:"message,omitempty"`
	Token    string `json:"token,omitempty"`
}

func LoginHandler(_ context.Context, w http.ResponseWriter, r *http.Request, data *LoginRequest) LoginResponse {
	db := GetDB()
	failedStr := "Invalid Username/Password."

	query := `SELECT username, passhash FROM users WHERE username = ?`
	row := db.QueryRow(query, NormalizeUsername(data.Username))
	if err := row.Err(); err != nil {
		log.Printf("Error getting username/passhash: %v", err)
		return LoginResponse{
			Success: false,
			Message: failedStr,
		}
	}
	var username string
	var passhash string
	if err := row.Scan(&username, &passhash); err != nil {
		log.Printf("Error getting username/passhash: %v", err)
		return LoginResponse{
			Success: false,
			Message: failedStr,
		}
	}

	// Verify the hash
	if err := bcrypt.CompareHashAndPassword([]byte(passhash), []byte(data.Password)); err != nil {
		log.Printf("Error comparing password hashes: %v", err)
		return LoginResponse{
			Success: false,
			Message: failedStr,
		}
	}

	log.Printf("successful login for %s", username)
	token := MakeAuthToken(username)
	return LoginResponse{
		Username: username,
		Success:  true,
		Message:  "Logged in.",
		Token:    token,
	}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Confirm  string `json:"confirm"`
}

type RegisterResponse struct {
	Username string `json:"username,omitempty"`
	Success  bool   `json:"success"`
	Message  string `json:"message,omitempty"`
	Token    string `json:"token,omitempty"`
}

func RegisterHandler(_ context.Context, w http.ResponseWriter, r *http.Request, data *RegisterRequest) RegisterResponse {
	if data.Password != data.Confirm {
		return RegisterResponse{
			Success: false,
			Message: "Passwords don't match",
		}
	}

	if len(data.Username) < 5 {
		return RegisterResponse{
			Success: false,
			Message: "Username too short",
		}
	}
	if len(data.Password) < 5 {
		return RegisterResponse{
			Success: false,
			Message: "Password too short",
		}
	}

	passhash, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error hashing password: %v", err)
		return RegisterResponse{
			Success: false,
			Message: "Password hashing error",
		}
	}

	username := NormalizeUsername(data.Username)

	db := GetDB()
	tx, err := db.Begin()
	if err != nil {
		log.Printf("error starting transaction: %v", err)
		return RegisterResponse{
			Success: false,
			Message: "internal error",
		}
	}
	defer tx.Rollback()
	query := `INSERT INTO users (username, passhash) VALUES (?, ?)`
	if _, err := tx.Exec(query, username, string(passhash)); err != nil {
		log.Printf("error inserting new record: %v", err)
		return RegisterResponse{
			Success: false,
			Message: "could not register",
		}
	}

	query = `INSERT INTO keybags (username, generation) VALUES (?, 0)`
	if _, err := tx.Exec(query, username); err != nil {
		log.Printf("error inserting empty keybag: %v", err)
		return RegisterResponse{
			Success: false,
			Message: "could not register",
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("error committing transaction: %v", err)
		return RegisterResponse{
			Success: false,
			Message: "could not register",
		}
	}

	return RegisterResponse{
		Username: username,
		Success:  true,
		Token:    MakeAuthToken(username),
		Message:  "Success",
	}
}

type KeybagUpdateRequest struct {
	Generation int    `json:"generation"`
	KeyHash    string `json:"keyhash"`
	Iterations int    `json:"iterations"`
	KeyBag     []byte `json:"keybag"`
}

type KeybagUpdateResponse struct {
	Success   bool               `json:"success"`
	Message   string             `json:"message"`
	NewKeybag *KeybagGetResponse `json:"updated,omitempty"`
}

type KeybagGetResponse struct {
	Generation int    `json:"generation"`
	KeyHash    string `json:"keyhash"`
	Iterations int    `json:"iterations"`
	CTime      string `json:"created"`
	KeyBag     []byte `json:"keybag"`
}

func KeybagHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		JSONMethodAdapter(KeybagUpdateHandler)(ctx, w, r)
		return
	case http.MethodGet:
		// intentional fall out of switch
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// handle GET
	username := GetUsernameOrUnauthorized(ctx, w)
	if username == "" {
		return
	}
	resp, err := getKeybagForUser(ctx, username)
	if err != nil {
		log.Printf("error getting keybag: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	SendJSON(w, resp)
}

func KeybagUpdateHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, data *KeybagUpdateRequest) KeybagUpdateResponse {
	username := GetUsernameOrUnauthorized(ctx, w)
	if username == "" {
		return KeybagUpdateResponse{}
	}
	db := GetDB()
	query := `INSERT INTO keybags (username, generation, keyhash, iter, keybag) VALUES(?, ?, ?, ?, ?)`
	if _, err := db.Exec(query, username, data.Generation+1, data.KeyHash, data.Iterations, data.KeyBag); err != nil {
		log.Printf("Error updating keybag: %v", err)
		return KeybagUpdateResponse{
			Success: false,
			Message: "Failed Updating Keybag",
		}
	}
	kb, err := getKeybagForUser(ctx, username)
	if err != nil {
		return KeybagUpdateResponse{
			Success: false,
			Message: "Failed Getting Updated Keybag",
		}
	}
	return KeybagUpdateResponse{
		Success:   true,
		Message:   "Updated",
		NewKeybag: &kb,
	}
}

func getKeybagForUser(ctx context.Context, username string) (KeybagGetResponse, error) {
	db := GetDB()
	query := `SELECT generation, keyhash, iter, ctime, keybag FROM keybags WHERE username = ? ORDER BY generation DESC LIMIT 1`
	row := db.QueryRow(query, username)
	return extractKeybag(ctx, row)
}

func extractKeybag(ctx context.Context, row *sql.Row) (KeybagGetResponse, error) {
	resp := KeybagGetResponse{}
	var keyhash sql.NullString
	if err := row.Scan(&resp.Generation, &keyhash, &resp.Iterations, &resp.CTime, &resp.KeyBag); err != nil {
		return resp, err
	}
	if keyhash.Valid {
		resp.KeyHash = keyhash.String
	}
	return resp, nil
}

type KeybagHistoryEntry struct {
	Generation int    `json:"generation"`
	CTime      string `json:"created"`
}

type KeybagHistoryResponse struct {
	Success bool                 `json:"success"`
	Message string               `json:"message,omitempty"`
	Entries []KeybagHistoryEntry `json:"entries,omitempty"`
}

func KeybagHistoryListHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	username := GetUsernameOrUnauthorized(ctx, w)
	if username == "" {
		return
	}
	db := GetDB()
	query := `SELECT generation, ctime FROM keybags WHERE username = ? ORDER BY generation DESC LIMIT 20`
	rows, err := db.Query(query, username)
	if err != nil {
		log.Printf("error querying for history: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var entries []KeybagHistoryEntry
	for rows.Next() {
		var e KeybagHistoryEntry
		if err := rows.Scan(&e.Generation, &e.CTime); err != nil {
			log.Printf("error scanning history: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		entries = append(entries, e)
	}
	rv := KeybagHistoryResponse{
		Entries: entries,
		Success: true,
	}
	SendJSON(w, rv)
}

func KeybagHistoryGetHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	username, genStr, found := strings.Cut(path, "/")
	if !found {
		log.Printf("no / in history request")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	gen64, err := strconv.ParseInt(genStr, 10, 32)
	if err != nil {
		log.Printf("error parsing int from %s: %v", genStr, err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	gen := int(gen64)

	db := GetDB()
	query := `SELECT generation, keyhash, iter, ctime, keybag FROM keybags WHERE username = ? AND generation = ? LIMIT 1`
	row := db.QueryRow(query, username, gen)
	resp, err := extractKeybag(ctx, row)
	if err == sql.ErrNoRows {
		log.Printf("no keybag for %s/%d", username, gen)
		w.WriteHeader(http.StatusNotFound)
		resp := struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}{
			Success: false,
			Message: "keybag not found",
		}
		SendJSON(w, resp)
		return
	}
	if err != nil {
		log.Printf("error loading keybag %s/%d: %v", username, gen, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	SendJSON(w, resp)
}
