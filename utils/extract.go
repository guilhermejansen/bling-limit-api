package utils

import (
	"bytes"
	"encoding/json"
	"log"
)

type EstoqueData struct {
    Retorno struct {
        Estoques []struct {
            Estoque struct {
                Codigo    string `json:"codigo"`
                Depositos []struct {
                    Deposito struct {
                        ID json.Number `json:"id"`
                    } `json:"deposito"`
                } `json:"depositos"`
            } `json:"estoque"`
        } `json:"estoques"`
    } `json:"retorno"`
}

func ExtractCodigoAndDeposito(data map[string]interface{}) (string, string) {
    dataBytes, err := json.Marshal(data)
    if err != nil {
        log.Printf("Erro ao serializar data: %v", err)
        return "", ""
    }

    var result EstoqueData

    // Usando um decoder com UseNumber() para preservar números grandes
    decoder := json.NewDecoder(bytes.NewReader(dataBytes))
    decoder.UseNumber()

    if err := decoder.Decode(&result); err != nil {
        log.Printf("Estrutura JSON: %v\n", string(dataBytes))
        log.Printf("Erro ao extrair código e depósito: %v", err)
        return "", ""
    }

    if len(result.Retorno.Estoques) > 0 && len(result.Retorno.Estoques[0].Estoque.Depositos) > 0 {
        codigo := result.Retorno.Estoques[0].Estoque.Codigo
        depositoID := result.Retorno.Estoques[0].Estoque.Depositos[0].Deposito.ID.String()
        return codigo, depositoID
    }

    return "", ""
}
