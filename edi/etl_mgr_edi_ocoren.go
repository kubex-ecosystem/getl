package edi

import (
	"fmt"
	"strings"
)

// OCOREN representa um registro de Ocorrências nas Entregas.
type OCOREN struct {
	Field1 string
	Field2 string
	Field3 string
	// Adicione mais campos conforme necessário
}

// Parse analisa uma linha de texto e preenche os campos do OCOREN.
// line: a linha de texto a ser analisada.
// Retorna um erro se a linha for muito curta ou se ocorrer algum problema durante a análise.
func (o *OCOREN) Parse(line string) error {
	if len(line) < 30 { // Ajuste conforme o tamanho esperado
		return fmt.Errorf("linha muito curta")
	}
	o.Field1 = strings.TrimSpace(line[0:10])
	o.Field2 = strings.TrimSpace(line[10:20])
	o.Field3 = strings.TrimSpace(line[20:30])
	// Preencha mais campos conforme necessário
	return nil
}
