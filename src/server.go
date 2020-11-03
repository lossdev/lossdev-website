package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

// type conversion needed for context.WithValue()
type key int

var (
	httpLogger *log.Logger
	// port for ListenAndServeTLS to bind - 443 requires elevated privilege
	// don't set these variables directly, the production switch will toggle
	// between 8000 and 443 depending on false / true respectively
	serverPort = "8000"
	requestID  = 8000
	// log paths
	httpLogPath   = "http.log"
	accessLogPath = "access.log"
	keyPath       = "../lossdev.key"
	pemPath       = "../lossdev.pem"
)

func main() {
	// set production mode if -p is provided
	var production bool
	flag.BoolVar(&production, "p", false, "Set Production Mode")
	flag.Parse()

	if production {
		uid, err := checkUID()
		if err != nil {
			log.Println(err)
			return
		} else if uid != 0 {
			log.Println("Fatal: Must be run as root in production mode")
			return
		}
		serverPort = "443"
		requestID = 443
		// create /var/log/lossdev if it doesn't exist
		attemptCreateLogDir()
		httpLogPath = "/var/log/lossdev/http.log"
		accessLogPath = "/var/log/lossdev/access.log"
		keyPath = "/var/www/html/lossdev.key"
		pemPath = "/var/www/html/lossdev.pem"
	}

	// TLS config options - only use tls >= TLS1.2 and set suite ciphers
	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// initialize log files
	httpLogFile, err := os.OpenFile(httpLogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Error opening http log")
		return
	}
	accessLogFile, err := os.OpenFile(accessLogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Error opening access log")
	}
	defer httpLogFile.Close()
	defer accessLogFile.Close()

	// initialize loggers
	httpLogger = log.New(httpLogFile, "", log.LstdFlags)
	accessLogger := log.New(accessLogFile, "", log.LstdFlags)

	httpLogger.Println("[STARTUP]")

	// server options
	srv := &http.Server{
		Addr:         ":" + serverPort,
		Handler:      tracingHandler(nextRequestID)(loggingHandler(accessLogger)(routeHandler())),
		ErrorLog:     httpLogger,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	// signal catchers
	doneChannel := make(chan bool)
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, os.Interrupt, syscall.SIGTERM)

	// graceful shutdown goroutine
	go func() {
		<-quitChannel

		httpLogger.Println("[SHUTDOWN]")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		srv.SetKeepAlivesEnabled(false)

		if err := srv.Shutdown(ctx); err != nil {
			httpLogger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}

		close(doneChannel)
	}()

	httpLogger.Println("[READY] Serving at https://localhost:" + serverPort)
	go http.ListenAndServe(":80", http.HandlerFunc(redirect))
	log.Fatal(srv.ListenAndServeTLS(pemPath, keyPath))
	<-doneChannel
}

// checkUID evaluates if the UID of the cantina-server process is root.
// If it isn't, then the server won't run
func checkUID() (int, error) {
	u, err := user.Current()
	if err != nil {
		return -1, err
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return -1, err
	}
	return uid, nil
}

// attemptCreateLogDir will run every startup. It simply attempts to create
// the log directory so it is present for later operations
func attemptCreateLogDir() {
	_ = os.Mkdir("/var/log/lossdev", 0644)
}

// add logging events to accessLog.log
func loggingHandler(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rID, ok := r.Context().Value(requestID).(string)
				if !ok {
					rID = "UNKNOWN"
				}
				logger.Println(rID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())

			}()
			next.ServeHTTP(w, r)
		})
	}
}

// aids in populating accessLog
func tracingHandler(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rID := r.Header.Get("X-Request-Id")
			if rID == "" {
				rID = nextRequestID()
			}
			k := key(requestID)
			ctx := context.WithValue(r.Context(), k, rID)
			w.Header().Set("X-Request-Id", rID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// main routing function. Add rules below
func routeHandler() *http.ServeMux {
	router := http.NewServeMux()
	router.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("../css"))))
	router.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("../assets"))))
	router.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("../js"))))
	http.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir("../html"))))
	router.HandleFunc("/", indexHandler)
	// robots.txt
	router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./robots.txt")
	})
	return router
}

// redirect http to https
func redirect(w http.ResponseWriter, req *http.Request) {
	target := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}

// index (/) handler
func indexHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

	t := template.Must(template.ParseFiles("../html/index.html"))
	err := t.Execute(w, nil)
	if err != nil {
		httpLogger.Println(err)
	}
}
