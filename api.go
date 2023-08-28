package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v5"
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

	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin))
	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleGetAccountByID), s.store))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleDeleteAccount), s.store)).Methods("DELETE")
	router.HandleFunc("/set-admin/{accountNumber}", withJWTAuth(makeHTTPHandleFunc(s.handleSetAdmin), s.store)).Methods("POST")
	router.HandleFunc("/transfer-history", withJWTAuth(makeHTTPHandleFunc(s.handleTransferHistory), s.store)).Methods("GET")

	router.HandleFunc("/transfer", withJWTAuth(makeHTTPHandleFunc(s.handleTransfer), s.store))

	log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	acc, err := s.store.GetAccountByNumber(int(req.Number))
	if err != nil {
		return err
	}

	if !acc.ValidPassword(req.Password) {
		return fmt.Errorf("not authenticated")
	}

	token, err := createJWT(acc)
	if err != nil {
		return err
	}

	resp := LoginResponse{
		Token:  token,
		Number: acc.Number,
	}

	return WriteJSON(w, http.StatusOK, resp)
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		id, err := getID(r)
		if err != nil {
			return err
		}

		account, err := s.store.GetAccountByID(id)
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

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	tokenString := r.Header.Get("x-jwt-token")
	token, err := validateJWT(tokenString)
	if err != nil {
		return err
	}
	claims := token.Claims.(jwt.MapClaims)
	isAdmin := claims["isAdmin"].(bool)

	req := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	// Check if the required fields are provided
	if req.FirstName == "" || req.LastName == "" {
		return fmt.Errorf("first name and last name are required")
	}

	if req.IsAdmin && !isAdmin {
		return fmt.Errorf("permission denied: only admin can create accounts with isAdmin=true")
	}

	account, err := NewAccount(req.FirstName, req.LastName, req.Password, req.Balance)
	if err != nil {
		return err
	}

	if err := s.store.CreateAccount(account, account.IsAdmin); err != nil {
		return err
	}

	if isAdmin && req.IsAdmin {
		account.IsAdmin = true
		if err := s.store.UpdateAccount(account); err != nil {
			return err
		}
	}

	return WriteJSON(w, http.StatusOK, req)
}

func (s *APIServer) handleSetAdmin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}

	accountNumberStr := mux.Vars(r)["accountNumber"]
	accountNumber, err := strconv.ParseInt(accountNumberStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid account number given: %s", accountNumberStr)
	}

	tokenString := r.Header.Get("x-jwt-token")
	token, err := validateJWT(tokenString)
	if err != nil {
		return err
	}
	claims := token.Claims.(jwt.MapClaims)
	loggedInAccountNumber := int64(claims["accountNumber"].(float64))
	isLoggedInAccountAdmin := claims["isAdmin"].(bool)

	if !isLoggedInAccountAdmin {
		return fmt.Errorf("permission denied: you must be an admin to set admin status")
	}

	loggedInAccount, err := s.store.GetAccountByNumber(int(loggedInAccountNumber))
	if err != nil {
		return err
	}

	if !loggedInAccount.IsAdmin {
		return fmt.Errorf("permission denied: you must be an admin to set admin status")
	}

	account, err := s.store.GetAccountByNumber(int(accountNumber))
	if err != nil {
		return err
	}

	account.IsAdmin = true // Only admins can set other accounts as admin
	if err := s.store.UpdateAccount(account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	if err := s.store.DeleteAccount(int64(id)); err != nil {
		return err
	}
	fmt.Printf("Account with ID %d deleted\n", id)
	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": int(id)})
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}
	defer r.Body.Close()

	// Extract the JWT token from the request headers
	tokenString := r.Header.Get("x-jwt-token")

	// Validate the JWT token
	token, err := validateJWT(tokenString)
	if err != nil {
		return err
	}
	claims := token.Claims.(jwt.MapClaims)
	// Get the account number from the token claims
	accountNumber := int64(claims["accountNumber"].(float64))

	// Compare the account number from the token with the sender's account number
	if accountNumber != transferReq.FromAccount {
		return fmt.Errorf("not authorized: mismatched account number")
	}

	fromAccount, err := s.store.GetAccountByNumber(int(transferReq.FromAccount))
	if err != nil {
		return err
	}

	toAccount, err := s.store.GetAccountByNumber(int(transferReq.ToAccount))
	if err != nil {
		return err
	}

	if fromAccount.Balance < transferReq.Amount {
		return fmt.Errorf("insufficient balance for transfer")
	}

	fromAccount.Balance -= transferReq.Amount
	toAccount.Balance += transferReq.Amount

	// Update the balances in the account table
	if err := s.store.UpdateAccountBalance(int(fromAccount.Number), fromAccount.Balance); err != nil {
		return err
	}

	if err := s.store.UpdateAccountBalance(int(toAccount.Number), toAccount.Balance); err != nil {
		return err
	}

	// Record the transfer in the transfer table
	if err := s.store.CreateTransfer(fromAccount.Number, toAccount.Number, transferReq.Amount); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, transferReq)
}

func (s *APIServer) handleTransferHistory(w http.ResponseWriter, r *http.Request) error {
	// Check if the user is an admin
	tokenString := r.Header.Get("x-jwt-token")
	token, err := validateJWT(tokenString)
	if err != nil {
		return err
	}
	claims := token.Claims.(jwt.MapClaims)
	isAdmin := claims["isAdmin"].(bool)

	if !isAdmin {
		return fmt.Errorf("permission denied: only admin can view transfer history")
	}

	// Fetch transfer history from the transfer table
	transfers, err := s.store.GetTransferHistory()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, transfers)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func createJWT(account *Account) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
		"isAdmin":       account.IsAdmin,
	}

	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))

}

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})
}

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT auth middleware")

		tokenString := r.Header.Get("x-jwt-token")

		token, err := validateJWT(tokenString)
		if err != nil {
			permissionDenied(w)
			return
		}
		claims := token.Claims.(jwt.MapClaims)

		if !token.Valid {
			permissionDenied(w)
			return
		}

		isAdmin := claims["isAdmin"].(bool)

		if isAdmin {
			handlerFunc(w, r)
		} else {
			permissionDenied(w)
		}
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}
	return id, nil
}

// func getNumber(r *http.Request) (int64, error) {
// 	numberStr := mux.Vars(r)["number"]
// 	number, err := strconv.ParseInt(numberStr, 10, 64)
// 	if err != nil {
// 		return 0, fmt.Errorf("invalid number given: %s", numberStr)
// 	}
// 	return number, nil
// }
