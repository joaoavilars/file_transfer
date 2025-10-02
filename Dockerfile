# --- Estágio de Build ---
# Usamos uma imagem Go para compilar nossa aplicação
FROM golang:1.25-alpine AS builder

# Define o diretório de trabalho dentro do container
WORKDIR /app

# Copia os arquivos do backend para o container
COPY backend/ ./

# Compila o código Go. 
# CGO_ENABLED=0 e GOOS=linux garantem que o binário seja pequeno e compatível com a imagem final.
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-file-server .


# --- Estágio Final ---
# Usamos uma imagem mínima para a execução, tornando o container mais leve e seguro
FROM alpine:latest

# Define o diretório de trabalho
WORKDIR /app

# Cria o diretório para uploads
RUN mkdir -p /app/uploads

# Copia os arquivos do frontend do nosso contexto local para dentro da imagem
COPY frontend/ ./frontend/

# Copia o binário compilado do estágio de build para a imagem final
COPY --from=builder /go-file-server .

# Expõe a porta 8080 para que possamos acessá-la de fora do container
EXPOSE 8080

# O comando que será executado quando o container iniciar
CMD ["./go-file-server"]