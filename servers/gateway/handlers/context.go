package handlers

import (
	"github.com/UW-Info-441-Winter-Quarter-2020/homework-GarsonYang/servers/gateway/indexes"
	"github.com/UW-Info-441-Winter-Quarter-2020/homework-GarsonYang/servers/gateway/models/users"
	"github.com/UW-Info-441-Winter-Quarter-2020/homework-GarsonYang/servers/gateway/sessions"
)

//TODO: define a handler context struct that
//will be a receiver on any of your HTTP
//handler functions that need access to
//globals, such as the key used for signing
//and verifying SessionIDs, the session store
//and the user store
type HandlerCtx struct {
	SigningKey   string
	SessionStore sessions.Store
	UserStore    users.Store
	Root         *indexes.TrieNode
	Notifier     *Notifier
}
