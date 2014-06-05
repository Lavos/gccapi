package main

import (
	"os/exec"
	"log"
	"io"
	"net/http"
	// "bytes"
	"os"
)

type CompressHandler struct {
	JavaPath string
}

func (c *CompressHandler) ServeHTTP (rep http.ResponseWriter, req *http.Request) {
	log.Print("Running command.")
	cmd := exec.Command(c.JavaPath, "-jar", "compiler.jar", "-W", "QUIET")

	in, err := cmd.StdinPipe()
	out, err := cmd.StdoutPipe()

	/* var stderr bytes.Buffer
	cmd.Stderr = &stderr */

	if err != nil {
		log.Print("Could not create pipes.")
		http.Error(rep, "Could not open connection to pipes.", 500)
		return
	}

	defer in.Close()
	defer out.Close()

	err = cmd.Start()

	if err != nil {
		log.Print("Could not start compiler.")
		http.Error(rep, "Could not start compiler.", 500)
		return
	}

	io.Copy(in, req.Body)
	in.Close()

	io.Copy(rep, out)

	err = cmd.Wait()

	if err != nil {
		log.Print("Compiler barfed.")
		// log.Print(stderr.String())
		http.Error(rep, "Compiler exited.", 500)
		return
	}

	log.Print("Successfully responded.")
}

func main () {
	javaPath := os.Getenv("JAVAPATH")

	if javaPath == "" {
		javaPath = "/usr/bin/java"
	}

	port := os.Getenv("GCCAPIPORT")

	if port == "" {
		port = ":7000"
	}

	http.Handle("/compress", &CompressHandler{javaPath})
	log.Printf("Started server on port %s.", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
