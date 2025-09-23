// Package module provides internal types and functions for the GoBE application.
package module

import (
	"os"
	"strings"

	clii "github.com/kubex-ecosystem/getl/cmd/cli"
	"github.com/kubex-ecosystem/getl/internal/module/version"
	"github.com/spf13/cobra"
)

type GETl struct {
	hideBannerV bool
}

// Alias retorna o alias do módulo.
// Retorna uma string contendo o alias do módulo.
func (m *GETl) Alias() string {
	return "etl"
}

// ShortDescription retorna uma descrição curta do módulo.
// Retorna uma string contendo a descrição curta do módulo.
func (m *GETl) ShortDescription() string {
	return "GETl: Efficient manager for data extraction, transformation, and loading."
}

// LongDescription retorna uma descrição longa do módulo.
// Retorna uma string contendo a descrição longa do módulo.
func (m *GETl) LongDescription() string {
	return "GETl is a comprehensive manager designed to streamline the processes of extracting, transforming, and loading data between various databases. It offers a robust and flexible solution for handling complex data workflows, ensuring seamless integration and efficient data management."
}

// Usage retorna a forma de uso do módulo.
// Retorna uma string contendo a forma de uso do módulo.
func (m *GETl) Usage() string {
	return "getl [command] [args]"
}

// Examples retorna exemplos de uso do módulo.
// Retorna um slice de strings contendo exemplos de uso do módulo.
func (m *GETl) Examples() []string {
	return []string{"getl extract [source]", "getl transform [data]", "getl load [destination]"}
}

// Active verifica se o módulo está ativo.
// Retorna um booleano indicando se o módulo está ativo.
func (m *GETl) Active() bool {
	return true
}

// Module retorna o nome do módulo.
// Retorna uma string contendo o nome do módulo.
func (m *GETl) Module() string {
	return "getl"
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
		Use:     m.Module(),
		Aliases: []string{m.Alias(), "e", "etl", "et"},
		Example: m.concatenateExamples(),
		Annotations: clii.GetDescriptions(
			[]string{
				"This is a efficient sync manager for almost any database, any environment and any data source. \nYou can vizualize before, after, when you want and how you want.\nYou will extract, transform and load data from almost any source to almost any destination.\nSweet yourself with many flavors... Enjoy!",
				"Sync manager for almost any database, any environment and any data source.",
			}, m.hideBannerV,
		),
	}
	cmd.AddCommand(clii.SyncCmd())
	cmd.AddCommand(clii.ExtractCmd())
	cmd.AddCommand(clii.LoadCmd())
	cmd.AddCommand(clii.ProduceCmd())
	cmd.AddCommand(clii.ConsumeCmd())
	cmd.AddCommand(clii.DataTableCmd())
	cmd.AddCommand(clii.VacuumCmd())
	cmd.AddCommand(version.CliCommand())

	// Set usage definitions for the command and its subcommands
	setUsageDefinition(cmd)
	for _, c := range cmd.Commands() {
		setUsageDefinition(c)
		if !strings.Contains(strings.Join(os.Args, " "), c.Use) {
			if c.Short == "" {
				c.Short = c.Annotations["description"]
			}
		}
	}

	return cmd
}
