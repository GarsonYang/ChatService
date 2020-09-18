package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-GarsonYang/servers/gateway/indexes"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-GarsonYang/servers/gateway/sessions"
	"golang.org/x/crypto/bcrypt"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-GarsonYang/servers/gateway/models/users"
)

//TODO: define HTTP handler functions as described in the
//assignment description. Remember to use your handler context
//struct as the receiver on these functions so that you have
//access to things like the session store and user store.
func (ctx *HandlerCtx) UsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "Error: request body must be in JSON", http.StatusUnsupportedMediaType)
			return
		}

		nu := &users.NewUser{}

		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(nu); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := nu.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		u, err := nu.ToUser()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		u, err = ctx.UserStore.Insert(u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		addNewUserToTrie(u, ctx.Root)

		state := &SessionState{
			SessionStartTime: time.Now(),
			AuthedUser:       u,
		}
		_, err = sessions.BeginSession(ctx.SigningKey, ctx.SessionStore, state, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		enc := json.NewEncoder(w)
		if err = enc.Encode(u); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		return
	}

	if r.Method == http.MethodGet {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer") {
			http.Error(w, "Not Authorized", http.StatusUnauthorized)
			return
		}

		if len(r.FormValue("q")) == 0 {
			http.Error(w, "Query string cannot be empty", http.StatusBadRequest)
			return
		}

		prefix := r.FormValue("q")
		max := 20
		IDs := ctx.Root.Find(prefix, max)

		users := []users.User{}
		for _, ID := range IDs {
			u, err := ctx.UserStore.GetByID(ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			users = append(users, *u)
		}

		sort.Slice(users, func(i, j int) bool {
			return users[i].UserName < users[j].UserName
		})

		enc := json.NewEncoder(w)
		if err := enc.Encode(users); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		return
	}

	http.Error(w, "Status Method Not Allowed", http.StatusMethodNotAllowed)
}

func (ctx *HandlerCtx) SpecificUserHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer") {
		http.Error(w, "Not Authorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		base := path.Base(r.URL.Path)
		if base != "me" {
			intID, err := strconv.Atoi(base)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			u, err := ctx.UserStore.GetByID((int64)(intID))
			if err != nil {
				http.Error(w, "Not a valid user ID", http.StatusNotFound)
				return
			}

			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			enc := json.NewEncoder(w)
			if err = enc.Encode(u); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			return
		}

		// retrieve the session state
		sessionState := &SessionState{}
		if _, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, sessionState); err != nil {
			http.Error(w, "The operation is forbidden", http.StatusForbidden)
			return
		}

		u, err := ctx.UserStore.GetByID(sessionState.AuthedUser.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(w)
		if err = enc.Encode(u); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		return
	}

	if r.Method == http.MethodPatch {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "Error: request body must be in JSON", http.StatusUnsupportedMediaType)
			return
		}

		// retrieve the session state
		sessionState := &SessionState{}
		if _, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, sessionState); err != nil {
			http.Error(w, "The operation is forbidden", http.StatusForbidden)
			return
		}

		base := path.Base(r.URL.Path)
		if base != "me" {
			intID, err := strconv.Atoi(base)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			u, err := ctx.UserStore.GetByID((int64)(intID))
			if err != nil {
				http.Error(w, "Not a valid user ID", http.StatusNotFound)
				return
			}

			if sessionState.AuthedUser.ID != u.ID {
				http.Error(w, "The operation is forbidden", http.StatusForbidden)
				return
			}
		}

		update := &users.Updates{}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(update); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		u, err := ctx.UserStore.GetByID(sessionState.AuthedUser.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		removeUserFromTrie(u, ctx.Root)

		u, err = ctx.UserStore.Update(u.ID, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		addNewUserToTrie(u, ctx.Root)

		enc := json.NewEncoder(w)
		if err = enc.Encode(u); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")

		return
	}

	http.Error(w, "Status Method Not Allowed", http.StatusMethodNotAllowed)
}

func (ctx *HandlerCtx) SessionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Status Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "Error: request body must be in JSON", http.StatusUnsupportedMediaType)
		return
	}

	cred := &users.Credentials{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(cred); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u, err := ctx.UserStore.GetByEmail(cred.Email)
	if err != nil {
		rand := make([]byte, len(cred.Password))
		bcrypt.CompareHashAndPassword(rand, []byte(cred.Password))
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	err = u.Authenticate(cred.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	state := &SessionState{
		SessionStartTime: time.Now(),
		AuthedUser:       u,
	}
	_, err = sessions.BeginSession(ctx.SigningKey, ctx.SessionStore, state, w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err = enc.Encode(u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (ctx *HandlerCtx) SpecificSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Status Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	lastPathSeg := path.Base(r.URL.Path)
	if lastPathSeg != "mine" {
		http.Error(w, "The operation is forbidden", http.StatusForbidden)
		return
	}

	_, err := sessions.EndSession(r, ctx.SigningKey, ctx.SessionStore)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	io.WriteString(w, "signed out")
}

func addNewUserToTrie(u *users.User, root *indexes.TrieNode) {
	values := []string{}
	values = append(values, strings.ToLower(u.UserName))
	values = append(values, strings.Split(strings.ToLower(u.FirstName), " ")...)
	values = append(values, strings.Split(strings.ToLower(u.LastName), " ")...)
	for _, v := range values {
		if len(v) > 0 {
			root.Add(v, u.ID)
		}
	}
}

func removeUserFromTrie(u *users.User, root *indexes.TrieNode) {
	values := []string{}
	values = append(values, strings.ToLower(u.UserName))
	values = append(values, strings.Split(strings.ToLower(u.FirstName), " ")...)
	values = append(values, strings.Split(strings.ToLower(u.LastName), " ")...)
	for _, v := range values {
		if len(v) > 0 {
			root.Remove(v, u.ID)
		}
	}
}
