package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-GarsonYang/servers/gateway/indexes"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-GarsonYang/servers/gateway/models/users"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-GarsonYang/servers/gateway/sessions"

	"github.com/go-redis/redis"

	_ "github.com/go-sql-driver/mysql"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-GarsonYang/servers/gateway/handlers"
)

const sessionDuration = (time.Duration)(10 * 6 * 10000000000)

type Director func(r *http.Request)

func CustomDirector(targets []string, ctx *handlers.HandlerCtx) Director {
	var counter int32
	counter = 0
	return func(r *http.Request) {
		sessionState := &handlers.SessionState{}
		sid, _ := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, sessionState)
		if sid != sessions.InvalidSessionID {
			user := sessionState.AuthedUser
			u, _ := json.Marshal(user)
			r.Header.Set("x-user", (string)(u))
		} else {
			r.Header.Del("x-user")
		}

		t := targets[counter%(int32)(len(targets))]
		targ, err := url.Parse(t)
		if err != nil {
			panic(err)
		}
		atomic.AddInt32(&counter, 1)

		r.Host = targ.Host
		r.URL.Host = targ.Host
		r.URL.Scheme = targ.Scheme
	}
}

//main is the main entry point for the server
func main() {
	addr := os.Getenv("ADDR")
	if len(addr) == 0 {
		addr = ":443"
	}

	tlsCertPath := os.Getenv("TLSCERT")
	if len(tlsCertPath) == 0 {
		log.Fatal("error accessing TLS public certificate")
	}

	tlsKeyPath := os.Getenv("TLSKEY")
	if len(tlsKeyPath) == 0 {
		log.Fatal("error accessing TLS private key")
	}

	sessionKey := os.Getenv("SESSIONKEY")

	redisAddr := os.Getenv("REDISADDR")
	if len(redisAddr) == 0 {
		redisAddr = "redistest:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
	redisStore := sessions.NewRedisStore(rdb, sessionDuration)

	dsn := os.Getenv("DSN")
	if len(dsn) == 0 {
		dsn = "root:password@tcp(mysqltest:3306)/demo"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("error opening database: ", err)
	}
	defer db.Close()
	sqlStore := users.NewSqlStore(db)

	root := indexes.NewTrieNode()
	users.LoadToTrie(db, root)

	ctx := &handlers.HandlerCtx{
		SigningKey:   sessionKey,
		SessionStore: redisStore,
		UserStore:    sqlStore,
		Root:         root,
		Notifier: &handlers.Notifier{
			ConnectionMap: make(map[int64][]*websocket.Conn),
		},
	}

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel")
	}
	defer ch.Close()

	ctx.ConnectToRabbitAndListen(ch)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/users", ctx.UsersHandler)
	mux.HandleFunc("/v1/users/", ctx.SpecificUserHandler)
	mux.HandleFunc("/v1/sessions", ctx.SessionsHandler)
	mux.HandleFunc("/v1/sessions/", ctx.SpecificSessionHandler)
	mux.HandleFunc("/ws", ctx.WebSocketConnectionHandler)

	SUMMARYADDR := os.Getenv("SUMMARYADDR")
	if len(SUMMARYADDR) == 0 {
		SUMMARYADDR = "http://summaryservice:5001"
	}

	summaryAddresses := strings.Split(SUMMARYADDR, ",")
	summaryProxy := &httputil.ReverseProxy{Director: CustomDirector(summaryAddresses, ctx)}
	mux.Handle("/v1/summary/", summaryProxy)

	MESSAGINGADDR := os.Getenv("MESSAGESADDR")
	if len(MESSAGINGADDR) == 0 {
		MESSAGINGADDR = "http://messagingservice:4001"
	}
	messagingAddresses := strings.Split(MESSAGINGADDR, ",")
	messagingProxy := &httputil.ReverseProxy{Director: CustomDirector(messagingAddresses, ctx)}
	mux.Handle("/v1/channels", messagingProxy)
	mux.Handle("/v1/channels/", messagingProxy)
	mux.Handle("/v1/messages/", messagingProxy)

	wrappedMux := &handlers.CORS{
		Handler: mux,
	}

	log.Printf("server is listening to port %s", addr)
	log.Fatal(http.ListenAndServeTLS(addr, tlsCertPath, tlsKeyPath, wrappedMux))
}
