package genx

import (
	"testing"
)

// TestDeserializeSQLTo testa a função DeserializeSQLTo.
// Verifica se a função gera a string correta para a desserialização de SQL para uma struct.
func TestDeserializeSQLTo(t *testing.T) {
	type Example struct {
		Name string `yaml:"name" json:"name"`
		Age  int    `yaml:"age" json:"age"`
	}

	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Example struct",
			args: args{v: &Example{}},
			want: `func DeserializeSQLToExample(row map[string]interface{}) (Example, error) {
			var data Example
			for key, value := range row {
				switch key {
				case "Name":
					data.Name = value.(string)
				case "Age":
					data.Age = value.(int)
				}
			}
			return data, nil
		}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeserializeSQLTo(tt.args.v); got != tt.want {
				t.Errorf("DeserializeSQLTo() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDeserializeJSONTo testa a função DeserializeJSONTo.
// Verifica se a função gera a string correta para a desserialização de JSON para uma struct.
func TestDeserializeJSONTo(t *testing.T) {
	type Example struct {
		Name string `yaml:"name" json:"name"`
		Age  int    `yaml:"age" json:"age"`
	}

	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Example struct",
			args: args{v: &Example{}},
			want: `func DeserializeJSONToExample(jsonData []byte) (Example, error) {
			var data Example
			if err := json.Unmarshal(jsonData, &data); err != nil {
				return data, err
			}
			return data, nil
		}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeserializeJSONTo(tt.args.v); got != tt.want {
				t.Errorf("DeserializeJSONTo() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDeserializeYAMLTo testa a função DeserializeYAMLTo.
// Verifica se a função gera a string correta para a desserialização de YAML para uma struct.
func TestDeserializeYAMLTo(t *testing.T) {
	type Example struct {
		Name string `yaml:"name" json:"name"`
		Age  int    `yaml:"age" json:"age"`
	}

	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Example struct",
			args: args{v: &Example{}},
			want: `func DeserializeYAMLToExample(yamlData []byte) (Example, error) {
			var data Example
			if err := yaml.Unmarshal(yamlData, &data); err != nil {
				return data, err
			}
			return data, nil
		}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeserializeYAMLTo(tt.args.v); got != tt.want {
				t.Errorf("DeserializeYAMLTo() = %v, want %v", got, tt.want)
			}
		})
	}
}
