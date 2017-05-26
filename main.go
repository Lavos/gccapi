package main

import (
	"os/exec"
	"log"
	"io"
	"net/http"
	"os"
	"os/signal"
	"context"
	"encoding/json"
	"bytes"
	"time"
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Configuration struct {
	JavaPath string `default:"/usr/bin/java"`
	JarPath string `default:"./closure-compiler.jar"`
	Port int64 `default:"9000"`
}

var (
	c Configuration
)

type ErrorResponse struct {
	ErrorMessage string `json:"error_message"`
	CompilerBody string `json:"compiler_body,omitempty"`
}

func ServeHTTP (w http.ResponseWriter, req *http.Request) {
	log.Printf("%#v", req)

	start := time.Now()

	log.Print("Running command.")
	cmd := exec.Command(c.JavaPath, "-jar", c.JarPath, "-W", "QUIET", "--compilation_level", "SIMPLE_OPTIMIZATIONS")

	in, err := cmd.StdinPipe()
	out, err := cmd.StdoutPipe()
	errout, err := cmd.StderrPipe()

	if err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, ErrorResponse{
			ErrorMessage: "Could not create command pipes.",
		})

		return
	}

	defer in.Close()
	defer out.Close()
	defer errout.Close()

	err = cmd.Start()

	if err != nil {
		SendErrorResponse(w, http.StatusInternalServerError, ErrorResponse{
			ErrorMessage: "Could not start command.",
		})

		return
	}

	stdbuf := new(bytes.Buffer)
	errbuf := new(bytes.Buffer)

	io.Copy(in, req.Body)
	in.Close()

	go func(){
		io.Copy(stdbuf, out)
	}()

	go func(){
		io.Copy(errbuf, errout)
	}()

	err = cmd.Wait()

	if err != nil {
		SendErrorResponse(w, http.StatusBadRequest, ErrorResponse{
			ErrorMessage: "Compiler error.",
			CompilerBody: errbuf.String(),
		})

		return
	}

	w.Header().Set("Content-Type", "text/javascript")
	w.Header().Set("Compile-Time", fmt.Sprintf("%s", time.Now().Sub(start)))
	w.WriteHeader(http.StatusOK)
	stdbuf.WriteTo(w)
}

func SendErrorResponse(w http.ResponseWriter, statusCode int, er ErrorResponse) {
	log.Print("%#v", er)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	json.NewEncoder(w).Encode(er)
}

func main () {
	envconfig.MustProcess("GCCAPI", &c)
	log.Printf("%#v", c)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", c.Port),
		Handler:        http.HandlerFunc(ServeHTTP),
	}

	go func(){
		err := s.ListenAndServe()
		log.Fatal(err)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	// Block until a signal is received.
	log.Printf("Got signal %s, shutting down.\n", <-sig)

	s.Shutdown(context.TODO())

	log.Printf("Server shutdown successful.")
}
