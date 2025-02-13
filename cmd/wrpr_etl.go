package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"strings"
)

type GETl struct{}

// Alias retorna o alias do módulo.
// Retorna uma string contendo o alias do módulo.
func (m *GETl) Alias() string {
	return "etl"
}

// ShortDescription retorna uma descrição curta do módulo.
// Retorna uma string contendo a descrição curta do módulo.
func (m *GETl) ShortDescription() string {
	return "GETl manager for data extraction, transformation, and loading."
}

// LongDescription retorna uma descrição longa do módulo.
// Retorna uma string contendo a descrição longa do módulo.
func (m *GETl) LongDescription() string {
	return "GETl manager for extracting, transforming, and loading data between different databases."
}

// Usage retorna a forma de uso do módulo.
// Retorna uma string contendo a forma de uso do módulo.
func (m *GETl) Usage() string {
	return "getl [command] [args]"
}

// Examples retorna exemplos de uso do módulo.
// Retorna um slice de strings contendo exemplos de uso do módulo.
func (m *GETl) Examples() []string {
	return []string{"kbx etl extract [source]", "kbx etl transform [data]", "kbx etl load [destination]"}
}

// Active verifica se o módulo está ativo.
// Retorna um booleano indicando se o módulo está ativo.
func (m *GETl) Active() bool {
	return true
}

// Module retorna o nome do módulo.
// Retorna uma string contendo o nome do módulo.
func (m *GETl) Module() string {
	return "integration"
}

// Execute executa o comando especificado para o módulo.
// commandArgs: um slice de strings contendo os argumentos do comando.
// Retorna um erro, se houver.
func (m *GETl) Execute(args []string) error {
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
func (m *GETl) concatenateExamples() string {
	examples := ""
	for _, example := range m.Examples() {
		examples += string(example) + "\n  "
	}
	return examples
}

// Command retorna o comando cobra para o módulo.
// Retorna um ponteiro para o comando cobra configurado.
func (m *GETl) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:         m.Module(),
		Aliases:     []string{m.Alias(), "e", "etl", "et"},
		Example:     m.concatenateExamples(),
		Annotations: m.getDescriptions(nil, true),
	}
	cmd.AddCommand(SyncCmd())
	cmd.AddCommand(ExtractCmd())
	cmd.AddCommand(LoadCmd())
	cmd.AddCommand(ProduceCmd())
	cmd.AddCommand(ConsumeCmd())
	cmd.AddCommand(DataTableCmd())
	cmd.AddCommand(VacuumCmd())
	setUsageDefinition(cmd)
	return cmd
}
func (m *GETl) getDescriptions(descriptionArg []string, hideBanner bool) map[string]string {
	var description, banner string
	if descriptionArg != nil {
		if strings.Contains(strings.Join(os.Args[0:], ""), "-h") {
			description = descriptionArg[0]
		} else {
			description = descriptionArg[1]
		}
	} else {
		if strings.Contains(strings.Join(os.Args[0:], ""), "-h") {
			description = m.LongDescription()
		} else {
			description = m.ShortDescription()
		}
	}
	//if !hideBanner {
	banner = ` ______      ______________ 
  / ____/     / ____/_  __/ / 
 / / ________/ __/   / / / /  
/ /_/ /_____/ /___  / / / /___
\____/     /_____/ /_/ /_____/
`
	//} else {
	//banner = ""
	//}
	return map[string]string{"banner": banner, "description": description}
}

// RegX registra e retorna uma nova instância de ETL.
// Retorna um ponteiro para uma nova instância de ETL.
func RegX() *GETl {
	return &GETl{}
}
