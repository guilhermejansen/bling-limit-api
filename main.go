package main

import (
	"bling_limit/handlers"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
    logFileName := "server.log"

    logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatalf("Erro ao abrir/criar o arquivo de log: %v", err)
    }
    defer logFile.Close()

    // Configura o logger para escrever tanto no console quanto no arquivo
    multiWriter := io.MultiWriter(os.Stdout, logFile)

    log.SetFlags(0)
    log.SetPrefix("")
    log.SetOutput(newCustomLogger(multiWriter))

    http.HandleFunc("/api/filter", handlers.PayloadHandler)

    fmt.Println("Servidor iniciado na porta 5050")

    if err := http.ListenAndServe(":5050", nil); err != nil {
        log.Fatalf("Erro ao iniciar o servidor: %v", err)
    }
}

type customLogger struct {
    writer io.Writer
}

func newCustomLogger(w io.Writer) *customLogger {
    return &customLogger{writer: w}
}

func (cl *customLogger) Write(bytes []byte) (int, error) {
    now := time.Now()
    timestamp := now.Format("02/01/2006 15:04:05")
    message := fmt.Sprintf("%s %s", timestamp, string(bytes))
    return cl.writer.Write([]byte(message))
}
