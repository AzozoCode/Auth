package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.Use(loggingMiddleware)

	router.HandleFunc("/account", makeHttpHandleFunc(s.handleAccount))

	router.HandleFunc("/accounts", withJWTAuth(makeHttpHandleFunc(s.handleGetAccounts)))

	router.HandleFunc("/account/{id}", makeHttpHandleFunc(s.handleGetAccountByID))

	router.HandleFunc("/transfer", makeHttpHandleFunc(s.handleTransfer)).Methods("POST")

	log.Println("JSON API server running on port:", s.listenAddr)
	log.Fatal(http.ListenAndServe(s.listenAddr, router))
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {

	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	if r.Method == "PUT" {
		return s.handleTransfer(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		id, err := getID(r)

		if err != nil {
			return err
		}

		//database call here
		account, err := s.store.GetAccountById(id)

		if err != nil {
			return err
		}

		return WriteJSON(w, http.StatusOK, account)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)

}

func (s *APIServer) handleGetAccounts(w http.ResponseWriter, r *http.Request) error {

	accounts, err := s.store.GetAccounts()

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, accounts)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {

	createAccReq := new(CreateAccountRequest)

	if err := json.NewDecoder(r.Body).Decode(createAccReq); err != nil {
		return err
	}

	defer r.Body.Close()

	account := NewAccount(createAccReq.FirstName, createAccReq.LastName)

	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	tokenString, err := creatJWTAuth(account)

	if err != nil {
		return err
	}

	fmt.Println("JWT token: ", tokenString)

	return WriteJSON(w, http.StatusCreated, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {

	id, err := getID(r)

	if err != nil {
		return err
	}

	if err = s.store.DeleteAccount(id); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})

}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {

	var account = new(Account)
	err := json.NewDecoder(r.Body).Decode(&account)

	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}

	r.Body.Close()

	if s.store.UpdateAccount(account) != nil {

		return WriteJSON(w, http.StatusInternalServerError, ApiError{Error: `an error occurred transfering amount`})
	}

	return WriteJSON(w, http.StatusOK, map[string]string{"message": "success"})

}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type ApiError struct {
	Error string `json:"error"`
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received request: ", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func getID(r *http.Request) (int, error) {

	idStr := mux.Vars(r)["id"]

	intId, err := strconv.Atoi(idStr)

	if err != nil {
		return 0, fmt.Errorf("invalid id given %s", idStr)
	}

	return intId, nil
}

func withJWTAuth(handleFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT auth middleware")

		tokenString := r.Header.Get("Authorization")

		fmt.Println("token: ", tokenString)
		_, err := validateJWT(tokenString)

		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "invalid token"})
			return
		}
		handleFunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {

	secret := os.Getenv("JWT_SECRET")
	return jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %s", t.Header["alg"])
		}

		return []byte(secret), nil
	})
}

func creatJWTAuth(account *Account) (string, error) {

	subByte, err := json.Marshal(account)

	if err != nil {
		return "", fmt.Errorf("error serializing")
	}

	claims := &jwt.MapClaims{
		"sub":           string(subByte),
		"exp":           time.Now().Add(time.Hour * 24 * 30).Unix(),
		"accountNumber": account.Number,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))

}
