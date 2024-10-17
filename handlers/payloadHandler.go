package handlers

import (
	"bling_limit/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type ViolationReport struct {
    Codigo     string
    Deposito   string
    Tentativas int64
}

type ViolationNotification struct {
    Codigo     string `json:"codigo"`
    Deposito   string `json:"deposito"`
    Tentativas int64  `json:"tentativas"`
    Message    string `json:"message"`
}

// Variáveis globais para configurações
var (
    MaxAttempts       int
    ExpireTimeSeconds int
    BlockTimeMinutes  int
)

func init() {
    var err error

    // Ler MAX_ATTEMPTS
    MaxAttempts, err = strconv.Atoi(os.Getenv("MAX_ATTEMPTS"))
    if err != nil || MaxAttempts <= 0 {
        MaxAttempts = 3 // Valor padrão
    }

    // Ler EXPIRE_TIME_SECONDS
    ExpireTimeSeconds, err = strconv.Atoi(os.Getenv("EXPIRE_TIME_SECONDS"))
    if err != nil || ExpireTimeSeconds <= 0 {
        ExpireTimeSeconds = 5 // Valor padrão
    }

    // Ler BLOCK_TIME_MINUTES
    BlockTimeMinutes, err = strconv.Atoi(os.Getenv("BLOCK_TIME_MINUTES"))
    if err != nil || BlockTimeMinutes <= 0 {
        BlockTimeMinutes = 1 // Valor padrão
    }

    log.Printf("Configurações: MaxAttempts=%d, ExpireTimeSeconds=%d, BlockTimeMinutes=%d",
        MaxAttempts, ExpireTimeSeconds, BlockTimeMinutes)
}

func sendViolationNotification(violation *ViolationReport) {
    message := fmt.Sprintf("Violação: Código %s excedeu o limite com %d tentativas. (Depósito ID: %s)", violation.Codigo, violation.Tentativas, violation.Deposito)

    callbackURL := os.Getenv("CALLBACK_ENDPOINT")
    if callbackURL == "" {
        log.Println("CALLBACK_ENDPOINT não está definido.")
        return
    }

    notification := ViolationNotification{
        Codigo:     violation.Codigo,
        Deposito:   violation.Deposito,
        Tentativas: violation.Tentativas,
        Message:    message,
    }
    payloadBytes, err := json.Marshal(notification)
    if err != nil {
        log.Printf("Erro ao serializar o payload: %v\n", err)
        return
    }
    body := bytes.NewReader(payloadBytes)

    req, err := http.NewRequest("POST", callbackURL, body)
    if err != nil {
        log.Printf("Erro ao criar a requisição: %v\n", err)
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Erro ao enviar a notificação: %v\n", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        log.Printf("A notificação não foi bem-sucedida: Código de status %d\n", resp.StatusCode)
    }
}

var violations = make(map[string]*ViolationReport)

func PayloadHandler(w http.ResponseWriter, r *http.Request) {
    var payloads []map[string]interface{}

    decoder := json.NewDecoder(r.Body)
    decoder.UseNumber() // Preserva números grandes
    if err := decoder.Decode(&payloads); err != nil {
        log.Printf("Erro ao decodificar o payload: %v\n", err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }
    log.Printf("Recebeu %d payloads\n", len(payloads))
    client := utils.NewRedisClient()

    var responsePayloads []map[string]interface{}

    for _, payload := range payloads {
        // Clonar o payload para modificar sem alterar o original
        responsePayload := make(map[string]interface{})
        for k, v := range payload {
            responsePayload[k] = v
        }

        body, ok := responsePayload["body"].(map[string]interface{})
        if !ok {
            log.Printf("Payload inválido (body não é um map): %v\n", payload)
            continue
        }

        dataValue, ok := body["data"]
        if !ok {
            log.Printf("Payload inválido (data não existe): %v\n", payload)
            continue
        }

        var dataMap map[string]interface{}

        // Tratando o caso em que 'data' é uma string contendo um JSON serializado
        dataStr, ok := dataValue.(string)
        if ok {
            dataDecoder := json.NewDecoder(strings.NewReader(dataStr))
            dataDecoder.UseNumber() // Preserva números grandes
            if err := dataDecoder.Decode(&dataMap); err != nil {
                log.Printf("Erro ao decodificar data string: %v\n", err)
                continue
            }
        } else {
            log.Printf("Payload inválido (data não é uma string): %v\n", payload)
            continue
        }

        codigo, deposito := utils.ExtractCodigoAndDeposito(dataMap)
        if codigo == "" {
            log.Printf("Não foi possível extrair o código do payload: %v\n", payload)
            continue
        }

        blockedKey := "blocked:" + codigo
        blocked, _ := client.Exists(utils.Ctx, blockedKey).Result()
        if blocked > 0 {
            continue
        }

        key := "codigo:" + codigo
        tentativas, _ := client.Incr(utils.Ctx, key).Result()

        if tentativas == 1 {
            client.Expire(utils.Ctx, key, time.Duration(ExpireTimeSeconds)*time.Second)
        }

        log.Printf("Quantidade de tentativas para o código %s: %v\n", codigo, tentativas)

        if tentativas > int64(MaxAttempts) {
            client.Set(utils.Ctx, blockedKey, "1", time.Duration(BlockTimeMinutes)*time.Minute)

            if _, exists := violations[codigo]; !exists {
                violations[codigo] = &ViolationReport{Codigo: codigo, Deposito: deposito, Tentativas: tentativas}
            } else {
                violations[codigo].Tentativas = tentativas
            }

            sendViolationNotification(violations[codigo])

            continue
        }

        // Ajustar os campos dentro de 'deposito' conforme solicitado
        adjustDepositoFields(dataMap)

        // Serializar dataMap para string
        dataBytes, err := json.Marshal(dataMap)
        if err != nil {
            log.Printf("Erro ao serializar dataMap: %v\n", err)
            continue
        }
        adjustedDataString := string(dataBytes)

        // Atualizar o 'data' dentro de 'body' com o JSON ajustado
        body["data"] = adjustedDataString

        // Adicionar o payload ajustado à resposta
        responsePayloads = append(responsePayloads, responsePayload)
    }

    // Enviar a resposta como um array de payloads ajustados
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(responsePayloads); err != nil {
        log.Printf("Erro ao enviar a resposta: %v\n", err)
    }

    // Registrar violações
    for _, report := range violations {
        log.Printf("Violação: Código %s excedeu o limite com %d tentativas. (Depósito ID: %s)\n", report.Codigo, report.Tentativas, report.Deposito)
    }
    violations = make(map[string]*ViolationReport)
}

// Função para ajustar os campos dentro de 'deposito'
func adjustDepositoFields(dataMap map[string]interface{}) {
    retorno, ok := dataMap["retorno"].(map[string]interface{})
    if !ok {
        return
    }

    estoques, ok := retorno["estoques"].([]interface{})
    if !ok {
        return
    }

    for _, estoqueItem := range estoques {
        estoqueMap, ok := estoqueItem.(map[string]interface{})
        if !ok {
            continue
        }

        estoque, ok := estoqueMap["estoque"].(map[string]interface{})
        if !ok {
            continue
        }

        depositos, ok := estoque["depositos"].([]interface{})
        if !ok {
            continue
        }

        for _, depositoItem := range depositos {
            depositoMap, ok := depositoItem.(map[string]interface{})
            if !ok {
                continue
            }

            deposito, ok := depositoMap["deposito"].(map[string]interface{})
            if !ok {
                continue
            }

            // Ajustar 'id' para string
            idValue, ok := deposito["id"]
            if ok {
                switch id := idValue.(type) {
                case float64:
                    deposito["id"] = fmt.Sprintf("%.0f", id)
                case json.Number:
                    deposito["id"] = id.String()
                case string:
                    // id já é string
                default:
                    deposito["id"] = fmt.Sprintf("%v", idValue)
                }
            }

            // Converter 'saldo' (string) para 'saldoVirtual' (int) e remover 'saldo'
            saldoValue, ok := deposito["saldo"]
            if ok {
                saldoStr, ok := saldoValue.(string)
                if ok {
                    saldoInt, err := strconv.Atoi(saldoStr)
                    if err == nil {
                        deposito["saldoVirtual"] = saldoInt
                    } else {
                        // Se não conseguir converter, definir como 0
                        deposito["saldoVirtual"] = 0
                    }
                } else {
                    // Se 'saldo' não é string, tentar converter para número
                    switch saldo := saldoValue.(type) {
                    case float64:
                        deposito["saldoVirtual"] = int(saldo)
                    case json.Number:
                        saldoInt, err := saldo.Int64()
                        if err == nil {
                            deposito["saldoVirtual"] = int(saldoInt)
                        } else {
                            deposito["saldoVirtual"] = 0
                        }
                    default:
                        deposito["saldoVirtual"] = 0
                    }
                }
                // Remover o campo 'saldo'
                delete(deposito, "saldo")
            }

            // Se 'saldoVirtual' ainda for string, convertê-lo para int
            saldoVirtualValue, ok := deposito["saldoVirtual"]
            if ok {
                switch saldoVirtual := saldoVirtualValue.(type) {
                case string:
                    saldoVirtualInt, err := strconv.Atoi(saldoVirtual)
                    if err == nil {
                        deposito["saldoVirtual"] = saldoVirtualInt
                    } else {
                        deposito["saldoVirtual"] = 0
                    }
                case json.Number:
                    saldoVirtualInt, err := saldoVirtual.Int64()
                    if err == nil {
                        deposito["saldoVirtual"] = int(saldoVirtualInt)
                    } else {
                        deposito["saldoVirtual"] = 0
                    }
                case float64:
                    deposito["saldoVirtual"] = int(saldoVirtual)
                }
            }
        }
    }
}
