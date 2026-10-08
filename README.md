# Leitor de Certificados Digitais

Aplicativo desktop para Windows que lê certificados digitais **ICP-Brasil** (e-CPF e e-CNPJ) e mostra as informações principais: titular, CPF/CNPJ, número de série, vencimento e autoridade certificadora.

Feito em Go com [Wails](https://wails.io) e React.

## Funcionalidades

- **Certificados instalados no Windows:** lista os certificados do repositório pessoal do usuário. Os A3 em token ou cartão também aparecem quando o driver do dispositivo está instalado.
  - Busca por nome, CPF, CNPJ ou número de série
  - Filtro por situação: válidos, vencendo (até 30 dias) e vencidos
- **Arquivos A1 (`.pfx` / `.p12`):** abre o arquivo com a senha. Dá para arrastar o arquivo para a janela ou escolher no computador.
- **Exportar como PFX:** gera um novo `.pfx` de um certificado instalado, com uma nova senha, incluindo a chave privada e a cadeia de certificação.
- **Remover certificados:** remove um ou vários certificados do repositório do Windows, sempre com confirmação.
- Botão de copiar ao lado de cada campo e tema claro/escuro.

### Informações exibidas

| Campo | Origem |
|---|---|
| Titular, CPF ou CNPJ | Campos ICP-Brasil do certificado (OIDs `2.16.76.1.3.*`); se não existirem, o nome no formato `NOME:DOCUMENTO` |
| Responsável pelo e-CNPJ e CPF do responsável | Campos ICP-Brasil `2.16.76.1.3.2` e `2.16.76.1.3.4` |
| Data de nascimento | Campos ICP-Brasil de pessoa física |
| Tipo (A1, A3...) | Política de certificação ICP-Brasil (`2.16.76.1.2.*`) |
| Número de série | Hexadecimal e decimal |
| Validade | Início, fim, dias restantes e situação |
| Outros | Autoridade certificadora, e-mail, algoritmos, thumbprints SHA-1 e SHA-256, se possui chave privada |

## Privacidade

- Tudo roda **localmente**. O aplicativo não acessa a internet e não envia dados para lugar nenhum.
- Senhas não são gravadas.
- Ao abrir um `.pfx`, nada é instalado no Windows. A chave privada só é carregada em memória.

## Download

Baixe o `Certificado.exe` na página de [Releases](../../releases).

O executável não tem assinatura digital, então o Windows SmartScreen pode mostrar um aviso na primeira execução. Para continuar, clique em **Mais informações → Executar assim mesmo**. Se preferir, compile você mesmo a partir do código (instruções abaixo).

Requisitos: Windows 10 ou 11 com o [WebView2](https://developer.microsoft.com/microsoft-edge/webview2/), que já vem instalado nas versões atuais.

## Limitações

- **Certificados A3 não podem ser exportados:** a chave privada fica no token ou cartão e não sai de lá. O mesmo vale para A1 instalados sem a opção "marcar esta chave como exportável".
- **Remover funciona como no gerenciador de certificados do Windows:** sai o certificado, mas a chave privada continua no sistema. Um A3 pode voltar a aparecer quando o token for reconectado.
- A listagem, a exportação e a remoção funcionam só no Windows.

## Compilando

Pré-requisitos:

- [Go](https://go.dev/dl/) 1.26 ou superior
- [Node.js](https://nodejs.org/) 20 ou superior
- Wails CLI v2: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

```bash
git clone <url-do-repositorio>
cd Certificado

wails dev     # modo desenvolvimento, com recarga automática
wails build   # gera build/bin/Certificado.exe
```

## Testes

```bash
go test ./...
```

Os testes usam certificados sintéticos, gerados na hora. Alguns testes que só rodam no Windows acessam o repositório de certificados do usuário que executa:

- lista os certificados instalados;
- exporta um certificado **só em memória** para conferir a exportação, sem gravar arquivo;
- instala e em seguida remove um certificado de teste descartável.

Nenhum desses testes altera os certificados existentes.

## Licença

[MIT](LICENSE)
