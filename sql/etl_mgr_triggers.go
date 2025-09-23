package sql

import (
	"bytes"
	"database/sql"
	"fmt"
	"text/template"

	. "github.com/kubex-ecosystem/getl/etypes"
)

// createTrigger cria um trigger no banco de dados especificado.
// db: a conexão com o banco de dados.
// dbType: o tipo de banco de dados (e.g., sqlite, postgres, mysql).
// trigger: a estrutura do trigger a ser criado.
// Retorna um erro, se houver.
func createTrigger(db *sql.DB, dbType string, trigger Trigger) error {
	tmpl, ok := triggerTemplates[dbType]
	if !ok {
		return fmt.Errorf("template de trigger não encontrado para o banco de dados: %s", dbType)
	}

	t, err := template.New("trigger").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("falha ao analisar template de trigger: %w", err)
	}

	var query bytes.Buffer
	if err := t.Execute(&query, trigger); err != nil {
		return fmt.Errorf("falha ao executar template de trigger: %w", err)
	}

	_, err = db.Exec(query.String())
	if err != nil {
		return fmt.Errorf("falha ao criar trigger: %w", err)
	}

	return nil
}

// triggerTemplates contém os templates de triggers para diferentes tipos de bancos de dados.
var triggerTemplates = map[string]string{
	"sqlite": `
        CREATE TRIGGER IF NOT EXISTS {{.Name}}
        {{.Event}} ON {{.Table}}
        FOR EACH ROW
        BEGIN
            {{.Statement}}
        END;
    `,
	"postgres": `
        CREATE OR REPLACE FUNCTION {{.Name}}_func() RETURNS TRIGGER AS $$
        BEGIN
            {{.Statement}};
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;

        CREATE TRIGGER {{.Name}}
        {{.Event}} ON {{.Table}}
        FOR EACH ROW
        EXECUTE FUNCTION {{.Name}}_func();
    `,
	"mysql": `
        CREATE TRIGGER {{.Name}}
        {{.Event}} ON {{.Table}}
        FOR EACH ROW
        BEGIN
            {{.Statement}};
        END;
    `,
	// Adicione outros templates conforme necessário
}
