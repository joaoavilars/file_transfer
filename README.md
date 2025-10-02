File Transfer Intranet

Uma aplicação web simples, segura e eficiente para transferência de arquivos, projetada para ser executada em uma rede interna (intranet). A aplicação é containerizada com Docker para fácil implantação e gerenciamento.

O backend é construído em Go (Golang) e o frontend em HTML, CSS e JavaScript puros, sem a necessidade de frameworks.

✨ Funcionalidades
🔒 Autenticação Segura: Acesso à área de gerenciamento protegido por usuário e senha, com autenticação baseada em JWT (JSON Web Tokens).

🚀 Upload Simples e Múltiplo: Envie arquivos através de um botão de seleção ou arrastando e soltando (drag-and-drop) na interface.

🔗 Geração de Links Únicos: Cada arquivo enviado gera um link único para download direto, que pode ser compartilhado. O download não requer autenticação.

📋 Listagem Persistente: A lista de arquivos enviados é carregada automaticamente ao acessar a página, refletindo o estado atual do servidor.

☑️ Gerenciamento em Massa: Exclua múltiplos arquivos de uma só vez utilizando checkboxes e um botão "Excluir Selecionados".

📊 Barra de Progresso: Acompanhe o andamento do upload de arquivos grandes com uma barra de progresso visual para cada item.

🌓 Tema Dark/Light: Alterne entre os temas claro e escuro para melhor conforto visual. A sua preferência é salva no navegador.

🐳 Containerizado com Docker: A aplicação é totalmente gerenciada pelo Docker e Docker Compose, simplificando a instalação e a execução.

🛠️ Tecnologias Utilizadas
Backend: Go (Golang)

Frontend: HTML5, CSS3, JavaScript (Vanilla JS)

Containerização: Docker, Docker Compose

Segurança:

Bcrypt para hash de senhas.

JWT (JSON Web Tokens) para gerenciamento de sessões.

🚀 Como Executar o Projeto
Siga os passos abaixo para configurar e executar a aplicação no seu ambiente.

Pré-requisitos
Docker

Docker Compose

1. Clone este repositório
git clone https://github.com/joaoavilars/file_transfer.git
cd file_transfer

2. Gere um Hash para sua Senha
Nunca armazene senhas em texto puro. Use o comando Docker abaixo para gerar um hash seguro para a senha que você deseja usar.

Substitua suaSenhaSuperForte123 pela senha de sua escolha:

docker run --rm golang:1.25-alpine sh -c "go run -e 'package main; import (\"fmt\"; \"os\"; \"golang.org/x/crypto/bcrypt\"); func main() { hash, _ := bcrypt.GenerateFromPassword([]byte(os.Args[1]), bcrypt.DefaultCost); fmt.Println(string(hash)) }' 'suaSenhaSuperForte123'"

Copie o resultado gerado (algo como $2a$10$...).

3. Configure o docker-compose.yml
Abra o arquivo docker-compose.yml e edite a seção environment com suas credenciais.

services:
  file-transfer-intranet:
    # ... (outras configurações)
    environment:
      # Defina o nome de usuário para o login
      - APP_USER=seu_usuario

      # Cole o hash gerado, substituindo cada '$' por '$$'
      - APP_PASSWORD_HASH=$$2a$$10$$... (seu hash com os '$' duplicados)

      # Crie um segredo longo e aleatório para os tokens JWT
      - JWT_SECRET=seu-segredo-super-secreto-e-longo

⚠️ Importante: O Docker Compose usa o caractere $ para variáveis. Para que ele interprete o seu hash corretamente, você deve substituir cada $ do hash por $$.

Exemplo:

Hash gerado: $2a$10$AbCdEf...

Valor no .yml: $$2a$$10$$AbCdEf...

(usuario inicial: admin - senha inicial: senha123)

4. Inicie a Aplicação
Com tudo configurado, execute o seguinte comando na raiz do projeto:

docker-compose up --build

O comando irá construir a imagem Docker e iniciar o container. Após a conclusão, a aplicação estará acessível no seu navegador em:

http://localhost:8080

📂 Estrutura do Projeto
.
├── backend/          # Contém o código-fonte do servidor Go
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── frontend/         # Contém os arquivos da interface do usuário
│   ├── index.html
│   └── script.js
├── uploads/          # Diretório onde os arquivos enviados são armazenados
├── docker-compose.yml # Orquestra a construção e execução do container
├── Dockerfile        # Define como construir a imagem Docker da aplicação
└── README.md         # Este arquivo
