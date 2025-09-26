package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	. "github.com/kubex-ecosystem/getl/etypes"
	. "github.com/kubex-ecosystem/getl/sql"
	. "github.com/kubex-ecosystem/getl/utils"
	gl "github.com/kubex-ecosystem/getl/internal/module/logger"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"
)

// VacuumCmd cria um comando Cobra para executar a limpeza de registros de uma tabela.
// Retorna um ponteiro para o comando Cobra configurado.
func VacuumCmd() *cobra.Command {
	var dbFilePath string

	cmd := &cobra.Command{
		Use:     "vacuum",
		Aliases: []string{"clean", "clear", "vacuumdb"},
		Short:   "Executa a limpeza de banco de dados SQLite",
		Long:    "Este comando executa a limpeza de compactação de registros, indexação e otimização de banco de dados SQLite.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ValidateArgs(dbFilePath); err != nil {
				gl.Log("error", fmt.Sprintf("falha ao validar argumentos: %v", err))
				return err
			}

			return VacuumDatabase(dbFilePath)
		},
	}

	cmd.Flags().StringVarP(&dbFilePath, "file", "f", "", "Caminho para o arquivo de banco de dados")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

// ExtractCmd cria um comando Cobra para extrair dados de uma fonte.
// Retorna um ponteiro para o comando Cobra configurado.
func ExtractCmd() *cobra.Command {
	var fileConfigPath, fileOutputPath, fileOutputFormat string

	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extrai dados de uma fonte",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Carregar a configuração da fonte
			sourceConfig, err := LoadConfigFile(fileConfigPath)
			if err != nil {
				return fmt.Errorf("falha ao carregar configuração da fonte: %w", err)
			}

			var data []Data
			var fieldsErr error

			// Extrai os dados com os tipos de coluna para contingência caso o tipo de coluna não seja informado
			data, _, fieldsErr = ExtractDataWithTypes(nil, sourceConfig)
			if fieldsErr != nil {
				return fmt.Errorf("falha ao extrair dados do destino: %w", fieldsErr)
			}

			if sourceConfig.OutputPath != "" && fileOutputPath == "" {
				fileOutputPath = sourceConfig.OutputPath
			}

			if fileOutputPath != "" {
				// Salvar os dados extraídos em um arquivo
				if saveDataErr := SaveData(fileOutputPath, data, fileOutputFormat); saveDataErr != nil {
					return fmt.Errorf("falha ao salvar os dados extraídos: %w", saveDataErr)
				}
				gl.Log("info", "Extração concluída com sucesso")
			} else {
				// Imprimir os dados extraídos no console
				gl.Log("info", "Extração concluída com sucesso")
				fmt.Printf("%s\n", data)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&fileConfigPath, "source", "s", "", "Caminho para o arquivo de configuração do source")
	cmd.Flags().StringVarP(&fileOutputPath, "output", "o", "", "Caminho para o arquivo de saída")
	cmd.Flags().StringVarP(&fileOutputFormat, "format", "F", "json", "Formato de saída dos dados")
	_ = cmd.MarkFlagRequired("source")

	return cmd
}

// loadCmd cria um comando Cobra para carregar os dados transformados em um destino.
// Retorna um ponteiro para o comando Cobra configurado.
func LoadCmd() *cobra.Command {
	var fileConfigPath string

	cmd := &cobra.Command{
		Use:   "load",
		Short: "Carrega os dados transformados em um destino",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Carregar a configuração do destino
			destinationConfig, err := LoadConfigFile(fileConfigPath)
			if err != nil {
				return fmt.Errorf("falha ao carregar configuração do destino: %w", err)
			}

			// Ler os dados transformados do arquivo JSON
			fileData, fileDataErr := os.ReadFile("extracted_data.json")
			if fileDataErr != nil {
				return fmt.Errorf("falha ao ler o arquivo de dados extraídos: %w", fileDataErr)
			}

			var data []Data
			if unmarshalErr := json.Unmarshal(fileData, &data); unmarshalErr != nil {
				return fmt.Errorf("falha ao processar dados JSON: %w", unmarshalErr)
			}

			// Carregar os dados no banco de destino
			if loadDataErr := LoadData(nil, destinationConfig); loadDataErr != nil {
				return fmt.Errorf("falha ao carregar os dados no destino: %w", loadDataErr)
			}

			gl.Log("info", "Carregamento concluído com sucesso")
			return nil
		},
	}

	cmd.Flags().StringVarP(&fileConfigPath, "file", "f", "", "Caminho para o arquivo de configuração do destino")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

// syncCmd cria um comando Cobra para executar as etapas extract, transform e load em sequência.
// Retorna um ponteiro para o comando Cobra configurado.
func SyncCmd() *cobra.Command {
	var fileConfigPath, fileOutputPath, outputFormat string
	var needCheck bool
	var checkMethod string

	sCmd := &cobra.Command{
		Use:     "sync",
		Aliases: []string{"s", "etl", "integrate"},
		Short:   "Executa as etapas extract, transform e load em sequência",
		Long:    "Este comando executa as etapas de extração, transformação e carregamento de dados em sequência.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if validateArgsErr := ValidateArgs(fileConfigPath); validateArgsErr != nil {
				gl.Log("error", fmt.Sprintf("falha ao validar argumentos: %v", validateArgsErr))
				return validateArgsErr
			}
			return ExecuteETL(fileConfigPath, fileOutputPath, outputFormat, needCheck, checkMethod)
		},
	}

	sCmd.Flags().StringVarP(&fileConfigPath, "file", "f", "", "Caminho para o arquivo de configuração das transformações")
	sCmd.Flags().StringVarP(&fileOutputPath, "output", "o", "", "Caminho para o arquivo de saída")
	sCmd.Flags().StringVarP(&outputFormat, "format", "F", "json", "Formato de saída dos dados")
	sCmd.Flags().BoolVarP(&needCheck, "check", "c", false, "Indica se é necessário realizar a verificação dos dados")
	sCmd.Flags().StringVarP(&checkMethod, "method", "m", "", "Método de verificação dos dados")

	_ = sCmd.MarkFlagRequired("file")

	return sCmd
}

// produceCmd cria um comando Cobra para produzir mensagens no Kafka.
// Retorna um ponteiro para o comando Cobra configurado.
func ProduceCmd() *cobra.Command {
	var kafkaURL, topic, message string

	cmd := &cobra.Command{
		Use:   "produce",
		Short: "Produz mensagens no Kafka",
		RunE: func(cmd *cobra.Command, args []string) error {
			writer := &kafka.Writer{
				Addr:  kafka.TCP(kafkaURL),
				Topic: topic,
			}
			defer func(writer *kafka.Writer) {
				_ = writer.Close()
			}(writer)

			err := writer.WriteMessages(context.Background(), kafka.Message{
				Value: []byte(message),
			})
			if err != nil {
				return fmt.Errorf("falha ao produzir mensagem: %w", err)
			}

			gl.Log("info", "Mensagem produzida com sucesso")
			return nil
		},
	}

	cmd.Flags().StringVarP(&kafkaURL, "kafka-url", "k", "localhost:9092", "URL do Kafka")
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "Tópico do Kafka")
	cmd.Flags().StringVarP(&message, "message", "m", "", "Mensagem a ser produzida")
	_ = cmd.MarkFlagRequired("topic")
	_ = cmd.MarkFlagRequired("message")

	return cmd
}

// consumeCmd cria um comando Cobra para consumir mensagens do Kafka.
// Retorna um ponteiro para o comando Cobra configurado.
func ConsumeCmd() *cobra.Command {
	var kafkaURL, topic, groupID string

	cmd := &cobra.Command{
		Use:   "consume",
		Short: "Consome mensagens do Kafka",
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := kafka.NewReader(kafka.ReaderConfig{
				Brokers: []string{kafkaURL},
				Topic:   topic,
				GroupID: groupID,
			})
			defer reader.Close()

			for {
				msg, err := reader.ReadMessage(context.Background())
				if err != nil {
					return fmt.Errorf("falha ao consumir mensagem: %w", err)
				}

				var data map[string]interface{}
				if err := json.Unmarshal(msg.Value, &data); err != nil {
					return fmt.Errorf("falha ao deserializar mensagem: %w", err)
				}

				gl.Log("info", fmt.Sprintf("Mensagem consumida: %v", data))
			}
		},
	}

	cmd.Flags().StringVarP(&kafkaURL, "kafka-url", "k", "localhost:9092", "URL do Kafka")
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "Tópico do Kafka")
	cmd.Flags().StringVarP(&groupID, "group-id", "g", "", "ID do grupo do Kafka")
	_ = cmd.MarkFlagRequired("topic")
	_ = cmd.MarkFlagRequired("group-id")

	return cmd
}

// dataTableCmd cria um comando Cobra para carregar os dados de uma tabela no banco de origem.
// Retorna um ponteiro para o comando Cobra configurado.
func DataTableCmd() *cobra.Command {
	var fileConfigPath, outputPath, outputFormat string
	var export bool

	cmd := &cobra.Command{
		Use:     "data-table",
		Aliases: []string{"dt", "table", "query"},
		Short:   "Carrega os dados de uma tabela no banco de origem",
		RunE: func(cmd *cobra.Command, args []string) error {
			if validateArgsErr := ValidateArgs(fileConfigPath); validateArgsErr != nil {
				gl.Log("error", fmt.Sprintf("falha ao validar argumentos: %v", validateArgsErr))
				return validateArgsErr
			}
			return ShowDataTableFromConfig(fileConfigPath, export, outputPath, outputFormat)
		},
	}

	cmd.Flags().StringVarP(&fileConfigPath, "file", "f", "", "Caminho para o arquivo de configuração do source")
	cmd.Flags().BoolVarP(&export, "export", "e", false, "Exportar os dados para um arquivo")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Caminho para o arquivo de saída")
	cmd.Flags().StringVarP(&outputFormat, "format", "F", "json", "Formato de saída dos dados")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

// validateArgs valida os argumentos passados para o comando.
// fileConfigPath: caminho para o arquivo de configuração das transformações.
// Retorna um erro se o caminho do arquivo de configuração estiver vazio.
func ValidateArgs(fileConfigPath string) error {
	if fileConfigPath == "" {
		return errors.New("o caminho para o arquivo de configuração das transformações é obrigatório")
	}
	return nil
}

// fieldsToSlice converte o mapa de campos para um slice de strings (nomes dos campos).
// fields: mapa de campos a ser convertido.
// Retorna um slice de strings contendo os nomes dos campos.
func FieldsToSlice(fields map[string][]Field) []string {
	var fieldNames []string
	for _, fieldGroup := range fields {
		for _, field := range fieldGroup {
			fieldNames = append(fieldNames, field["name"])
		}
	}
	return fieldNames
}
