package genx

import (
	"fmt"
	"reflect"
)

// SerializeYAMLToJSON gera uma função Go que converte dados YAML para JSON.
// A função gerada tem o nome ConvertYAMLToJSON seguido pelo nome do tipo de dados.
// Parâmetros:
// - v: Interface que representa o tipo de dados a ser convertido.
// Retorna:
// - Uma string contendo a definição da função Go gerada.
func SerializeYAMLToJSON(v interface{}) string {
	typ := reflect.TypeOf(v).Elem()
	funcName := fmt.Sprintf("ConvertYAMLToJSON%s", typ.Name())
	return fmt.Sprintf(`func %s(yamlData []byte) (string, error) {
		var data %s
		if err := yaml.Unmarshal(yamlData, &data); err != nil {
			return "", err
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		return string(jsonData), nil
	}`, funcName, typ.Name())
}

// SerializeJSONToYAML gera uma função Go que converte dados JSON para YAML.
// A função gerada tem o nome ConvertJSONToYAML seguido pelo nome do tipo de dados.
// Parâmetros:
// - v: Interface que representa o tipo de dados a ser convertido.
// Retorna:
// - Uma string contendo a definição da função Go gerada.
func SerializeJSONToYAML(v interface{}) string {
	typ := reflect.TypeOf(v).Elem()
	funcName := fmt.Sprintf("ConvertJSONToYAML%s", typ.Name())
	return fmt.Sprintf(`func %s(jsonData []byte) (string, error) {
		var data %s
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return "", err
		}
		yamlData, err := yaml.Marshal(data)
		if err != nil {
			return "", err
		}
		return string(yamlData), nil
	}`, funcName, typ.Name())
}

// SerializeSQLInsert gera uma função Go que cria uma consulta SQL INSERT para um tipo de dados específico.
// A função gerada tem o nome GenerateSQLInsert seguido pelo nome do tipo de dados.
// Parâmetros:
// - v: Interface que representa o tipo de dados a ser inserido.
// Retorna:
// - Uma string contendo a definição da função Go gerada.
func SerializeSQLInsert(v interface{}) string {
	typ := reflect.TypeOf(v).Elem()
	funcName := fmt.Sprintf("GenerateSQLInsert%s", typ.Name())
	result := fmt.Sprintf(`func %s(data %s) string {
		query := "INSERT INTO %s (`, funcName, typ.Name(), typ.Name())

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if i > 0 {
			result += ", "
		}
		result += field.Name
	}

	result += ") VALUES ("

	for i := 0; i < typ.NumField(); i++ {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("'%v'", reflect.ValueOf(v).Field(i).Interface())
	}

	result += `)"
	return query
}`
	return result
}

// SerializationDemo demonstra o uso das funções de serialização geradas.
// Cria um exemplo de tipo de dados e imprime as funções geradas para conversão e inserção.
func SerializationDemo() {
	type Example struct {
		Name string `yaml:"name" json:"name"`
		Age  int    `yaml:"age" json:"age"`
	}

	example := Example{}
	fmt.Println(SerializeYAMLToJSON(&example))
	fmt.Println(SerializeJSONToYAML(&example))
	fmt.Println(SerializeSQLInsert(&example))
}
