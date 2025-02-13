package genx

import (
	"fmt"
	"reflect"
)

// DeserializeJSONTo gera uma função para desserializar dados JSON em uma struct.
// v: a interface contendo os dados JSON.
// Retorna uma string contendo a função gerada.
func DeserializeJSONTo(v interface{}) string {
	typ := reflect.TypeOf(v).Elem()
	funcName := fmt.Sprintf("DeserializeJSONTo%s", typ.Name())
	return fmt.Sprintf(`func %s(jsonData []byte) (%s, error) {
					var data %s
					if err := json.Unmarshal(jsonData, &data); err != nil {
						return data, err
					}
					return data, nil
				}`, funcName, typ.Name(), typ.Name())
}

// DeserializeYAMLTo gera uma função para desserializar dados YAML em uma struct.
// v: a interface contendo os dados YAML.
// Retorna uma string contendo a função gerada.
func DeserializeYAMLTo(v interface{}) string {
	typ := reflect.TypeOf(v).Elem()
	funcName := fmt.Sprintf("DeserializeYAMLTo%s", typ.Name())
	return fmt.Sprintf(`func %s(yamlData []byte) (%s, error) {
					var data %s
					if err := yaml.Unmarshal(yamlData, &data); err != nil {
						return data, err
					}
					return data, nil
				}`, funcName, typ.Name(), typ.Name())
}

// DeserializeSQLTo gera uma função para desserializar dados SQL em uma struct.
// v: a interface contendo os dados SQL.
// Retorna uma string contendo a função gerada.
func DeserializeSQLTo(v interface{}) string {
	typ := reflect.TypeOf(v).Elem()
	funcName := fmt.Sprintf("DeserializeSQLTo%s", typ.Name())
	result := fmt.Sprintf(`func %s(row map[string]interface{}) (%s, error) {
					var data %s
					for key, value := range row {
						switch key {`, funcName, typ.Name(), typ.Name())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		result += fmt.Sprintf(`
						case "%s":
							data.%s = value.(%s)`, field.Name, field.Name, field.Type)
	}
	result += `
						}
					}
					return data, nil
				}`
	return result
}

// DeserializationDemo demonstra a desserialização de dados.
func DeserializationDemo() {
	type Example struct {
		Name string `yaml:"name" json:"name"`
		Age  int    `yaml:"age" json:"age"`
	}

	example := Example{}
	fmt.Println(DeserializeJSONTo(&example))
	fmt.Println(DeserializeYAMLTo(&example))
	fmt.Println(DeserializeSQLTo(&example))
}
