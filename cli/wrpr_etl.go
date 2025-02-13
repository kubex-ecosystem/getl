package cli

import (
	"github.com/spf13/cobra"
)

// ETL representa a estrutura do módulo ETL.
type ETL struct{}

// Alias retorna o alias do módulo.
// Retorna uma string contendo o alias do módulo.
func (m *ETL) Alias() string {
	return "etl"
}

// ShortDescription retorna uma descrição curta do módulo.
// Retorna uma string contendo a descrição curta do módulo.
func (m *ETL) ShortDescription() string {
	return "ETL manager for data extraction, transformation, and loading."
}

// LongDescription retorna uma descrição longa do módulo.
// Retorna uma string contendo a descrição longa do módulo.
func (m *ETL) LongDescription() string {
	return "ETL manager for extracting, transforming, and loading data between different databases."
}

// Usage retorna a forma de uso do módulo.
// Retorna uma string contendo a forma de uso do módulo.
func (m *ETL) Usage() string {
	return "kbx etl [command] [args]"
}

// Examples retorna exemplos de uso do módulo.
// Retorna um slice de strings contendo exemplos de uso do módulo.
func (m *ETL) Examples() []string {
	return []string{"kbx etl extract [source]", "kbx etl transform [data]", "kbx etl load [destination]"}
}

// Active verifica se o módulo está ativo.
// Retorna um booleano indicando se o módulo está ativo.
func (m *ETL) Active() bool {
	return true
}

// Module retorna o nome do módulo.
// Retorna uma string contendo o nome do módulo.
func (m *ETL) Module() string {
	return "integration"
}

// Execute executa o comando especificado para o módulo.
// commandArgs: um slice de strings contendo os argumentos do comando.
// Retorna um erro, se houver.
func (m *ETL) Execute(args []string) error {
	cmdEtl := m.Command()
	if args != nil {
		parseFlagsErr := cmdEtl.ParseFlags(args)
		if parseFlagsErr != nil {
			return parseFlagsErr
		}
		return cmdEtl.Execute()
	} else {
		return cmdEtl.Execute()
	}
}

// concatenateExamples concatena os exemplos de uso do módulo.
// Retorna uma string contendo os exemplos concatenados.
func (m *ETL) concatenateExamples() string {
	examples := ""
	for _, example := range m.Examples() {
		examples += string(example) + "\n  "
	}
	return examples
}

// Command retorna o comando cobra para o módulo.
// Retorna um ponteiro para o comando cobra configurado.
func (m *ETL) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:     m.Module(),
		Aliases: []string{m.Alias(), "e", "etlz", "et"},
		Example: m.concatenateExamples(),
		Short:   m.ShortDescription(),
		Long:    m.LongDescription(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.Execute(args)
		},
	}

	cmd.AddCommand(SyncCmd())
	cmd.AddCommand(ExtractCmd())
	cmd.AddCommand(LoadCmd())
	cmd.AddCommand(ProduceCmd())
	cmd.AddCommand(ConsumeCmd())
	cmd.AddCommand(DataTableCmd())
	cmd.AddCommand(VacuumCmd())

	return cmd
}

// RegX registra e retorna uma nova instância de ETL.
// Retorna um ponteiro para uma nova instância de ETL.
func RegX() *ETL {
	return &ETL{}
}
