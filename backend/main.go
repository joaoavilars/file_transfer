package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort" // Novo import para ordenar os arquivos
	"strings"
	"time"
)

const uploadPath = "/app/uploads"

// Struct para a requisição de exclusão em massa
type BatchDeleteRequest struct {
	Filenames []string `json:"filenames"`
}

// --- NOVO ---
// Struct para enviar informações dos arquivos para o frontend
type FileInfo struct {
	UniqueName   string `json:"uniqueName"`
	OriginalName string `json:"originalName"`
	// Podemos adicionar mais campos no futuro, como tamanho (Size) ou data (ModTime)
}

// --- FIM NOVO ---

func main() {
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)
	http.HandleFunc("/upload", uploadHandler)
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(uploadPath))))
	http.HandleFunc("/delete/", deleteHandler)
	http.HandleFunc("/delete-batch", batchDeleteHandler)

	// --- NOVO ENDPOINT ---
	http.HandleFunc("/list-files", listFilesHandler)
	// --- FIM NOVO ---

	log.Println("Servidor iniciado em http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Não foi possível iniciar o servidor: %v", err)
	}
}

// --- NOVA FUNÇÃO ---
// listFilesHandler lê o diretório de uploads e retorna uma lista de arquivos.
func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	entries, err := os.ReadDir(uploadPath)
	if err != nil {
		// Se a pasta ainda não existir, retorna uma lista vazia em vez de um erro.
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]")) // Retorna um array JSON vazio
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
			// Extrai o nome original do arquivo (remove o timestamp prefixado)
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

	// Ordena a lista de arquivos em ordem alfabética (opcional, mas bom para consistência)
	sort.Slice(files, func(i, j int) bool {
		return files[i].OriginalName < files[j].OriginalName
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// --- FIM NOVA FUNÇÃO ---

// (As outras funções: batchDeleteHandler, uploadHandler e deleteHandler permanecem exatamente as mesmas da versão anterior)
// ...
// ... (cole aqui as funções 'batchDeleteHandler', 'uploadHandler' e 'deleteHandler' da versão anterior sem modificá-las)
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

	// Linha do 'link' removida
	w.Header().Set("Content-Type", "application/json")
	// Retorna mais informações para o frontend com a sintaxe correta
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
