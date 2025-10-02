package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const uploadPath = "/app/uploads"

// Variáveis de ambiente
var (
	appUser         string
	appPasswordHash string
	jwtSecret       []byte
)

// --- LÓGICA DE AUTENTICAÇÃO E JWT ---

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// authMiddleware protege as rotas que precisam de login
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Acesso não autorizado", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// loginHandler lida com a tentativa de login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Requisição inválida", http.StatusBadRequest)
		return
	}

	// Compara o usuário e o hash da senha
	if creds.Username != appUser || bcrypt.CompareHashAndPassword([]byte(appPasswordHash), []byte(creds.Password)) != nil {
		http.Error(w, "Usuário ou senha inválidos", http.StatusUnauthorized)
		return
	}

	// Gera o token JWT
	expirationTime := time.Now().Add(8 * time.Hour)
	claims := &Claims{
		Username: creds.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// --- INÍCIO DO PROGRAMA ---

func main() {
	// Carrega as variáveis de ambiente
	appUser = os.Getenv("APP_USER")
	appPasswordHash = os.Getenv("APP_PASSWORD_HASH")
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	log.Println("-------------------------------------------------")
	log.Printf("INICIANDO COM ==> Usuário: [%s], Hash: [%s]", appUser, appPasswordHash)
	log.Println("-------------------------------------------------")

	if appUser == "" || appPasswordHash == "" || len(jwtSecret) == 0 {
		log.Fatal("Variáveis de ambiente APP_USER, APP_PASSWORD_HASH, e JWT_SECRET devem ser definidas")
	}

	// --- ROTAS PÚBLICAS (não precisam de login) ---
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)                                                                       // A página HTML principal
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(uploadPath)))) // Download de arquivos
	http.HandleFunc("/login", loginHandler)                                                    // O endpoint de login

	// --- ROTAS PROTEGIDAS (precisam de login) ---
	http.HandleFunc("/list-files", authMiddleware(listFilesHandler))
	http.HandleFunc("/upload", authMiddleware(uploadHandler))
	http.HandleFunc("/delete/", authMiddleware(deleteHandler))
	http.HandleFunc("/delete-batch", authMiddleware(batchDeleteHandler))

	log.Println("Servidor iniciado em http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Não foi possível iniciar o servidor: %v", err)
	}
}

// (O restante do código, com os handlers 'listFilesHandler', 'uploadHandler', etc., continua aqui sem alterações)
// ...
// ... (cole aqui as funções 'listFilesHandler', 'batchDeleteHandler', 'uploadHandler' e 'deleteHandler' da versão anterior sem modificá-las)
type FileInfo struct {
	UniqueName   string `json:"uniqueName"`
	OriginalName string `json:"originalName"`
}

func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	entries, err := os.ReadDir(uploadPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]"))
			return
		}
		log.Printf("Erro ao ler o diretório de uploads: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}
	var files []FileInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			uniqueName := entry.Name()
			originalName := uniqueName
			parts := strings.SplitN(uniqueName, "-", 2)
			if len(parts) == 2 {
				originalName = parts[1]
			}
			files = append(files, FileInfo{
				UniqueName:   uniqueName,
				OriginalName: originalName,
			})
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].OriginalName < files[j].OriginalName })
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

type BatchDeleteRequest struct {
	Filenames []string `json:"filenames"`
}

func batchDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	var req BatchDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Corpo da requisição inválido", http.StatusBadRequest)
		return
	}
	results := struct {
		Success []string `json:"success"`
		Failed  []string `json:"failed"`
	}{}
	for _, fileName := range req.Filenames {
		safeFileName := filepath.Base(fileName)
		if safeFileName != fileName {
			results.Failed = append(results.Failed, fileName)
			continue
		}
		filePath := filepath.Join(uploadPath, safeFileName)
		err := os.Remove(filePath)
		if err != nil {
			log.Printf("Falha ao excluir %s em lote: %v", safeFileName, err)
			results.Failed = append(results.Failed, fileName)
		} else {
			log.Printf("Excluído %s em lote", safeFileName)
			results.Success = append(results.Success, fileName)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	r.ParseMultipartForm(1024 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Printf("Erro ao obter o arquivo: %v", err)
		http.Error(w, "Erro ao obter o arquivo", http.StatusBadRequest)
		return
	}
	defer file.Close()
	uniqueFileName := fmt.Sprintf("%d-%s", time.Now().Unix(), handler.Filename)
	filePath := filepath.Join(uploadPath, uniqueFileName)
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("Erro ao criar o arquivo de destino: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Erro ao salvar o arquivo: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(FileInfo{
		UniqueName:   uniqueFileName,
		OriginalName: handler.Filename,
	})
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	fileName := strings.TrimPrefix(r.URL.Path, "/delete/")
	if fileName == "" {
		http.Error(w, "Nome do arquivo não fornecido", http.StatusBadRequest)
		return
	}
	safeFileName := filepath.Base(fileName)
	if safeFileName != fileName {
		http.Error(w, "Nome de arquivo inválido", http.StatusBadRequest)
		return
	}
	filePath := filepath.Join(uploadPath, safeFileName)
	err := os.Remove(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Arquivo não encontrado", http.StatusNotFound)
			return
		}
		log.Printf("Erro ao excluir o arquivo %s: %v", safeFileName, err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}
	log.Printf("Arquivo excluído: %s", safeFileName)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Arquivo '%s' excluído com sucesso"}`, safeFileName)
}
