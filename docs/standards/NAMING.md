# Convencoes de nomenclatura

- novos slugs de dataset devem usar kebab-case ASCII;
- nomes de tabelas devem usar snake_case;
- nomes de colunas devem usar snake_case;
- nomes tecnicos devem refletir a fonte oficial quando isso ajudar a auditoria;
- evite abreviacoes obscuras;
- nomes de tabelas de controle devem deixar claro o escopo do import.

## Exemplos

- pasta: `apis/novo-dataset`
- tabela: `novo_dataset`
- controle: `imports_novo_dataset`

Observacao: alguns modulos antigos podem carregar nomes historicos. Para novas contribuicoes, siga o padrao acima.
