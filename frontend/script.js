// Lógica de Tema (sem alterações)
const themeToggleButton = document.getElementById('theme-toggle');
const applyTheme = (theme) => {
    document.body.dataset.theme = theme;
    themeToggleButton.textContent = theme === 'dark' ? '☀️' : '🌙';
    localStorage.setItem('theme', theme);
};
themeToggleButton.addEventListener('click', () => {
    const newTheme = (document.body.dataset.theme || 'light') === 'light' ? 'dark' : 'light';
    applyTheme(newTheme);
});
applyTheme(localStorage.getItem('theme') || 'light');


// Elementos da UI
const dropZone = document.getElementById('drop-zone');
const uploadButton = document.getElementById('upload-button');
const fileInput = document.getElementById('file-input');
const resultDiv = document.getElementById('result');
const bulkActionsDiv = document.getElementById('bulk-actions');
const selectAllCheckbox = document.getElementById('select-all-checkbox');
const bulkDeleteButton = document.getElementById('bulk-delete-button');

document.addEventListener('DOMContentLoaded', loadInitialFiles);

// Lógica de Upload (sem alterações, exceto a chamada para uploadFile)
uploadButton.addEventListener('click', () => fileInput.click());
fileInput.addEventListener('change', () => handleFiles(fileInput.files));
['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => dropZone.addEventListener(eventName, preventDefaults));
dropZone.addEventListener('dragenter', () => dropZone.classList.add('dragover'));
dropZone.addEventListener('dragleave', () => dropZone.classList.remove('dragover'));
dropZone.addEventListener('drop', (e) => {
    dropZone.classList.remove('dragover');
    handleFiles(e.dataTransfer.files);
});

function preventDefaults(e) { e.preventDefault(); e.stopPropagation(); }
function handleFiles(files) { if (files.length > 0) [...files].forEach(uploadFile); }


/**
 * --- FUNÇÃO ATUALIZADA COM XMLHttpRequest E BARRA DE PROGRESSO ---
 * Faz o upload de UM arquivo e atualiza a UI.
 * @param {File} file 
 */
function uploadFile(file) {
    const fileId = `file-entry-${Date.now()}-${Math.random()}`;
    // Novo placeholder com a estrutura da barra de progresso
    const placeholderHtml = `
        <div id="${fileId}" class="file-entry">
            <div class="file-info">
                <p>Enviando ${file.name}...</p>
                <div class="progress-bar-container">
                    <div class="progress-bar"></div>
                </div>
            </div>
        </div>
    `;
    resultDiv.insertAdjacentHTML('beforeend', placeholderHtml);
    const fileEntryDiv = document.getElementById(fileId);
    const progressBar = fileEntryDiv.querySelector('.progress-bar');

    const formData = new FormData();
    formData.append('file', file);

    const xhr = new XMLHttpRequest();

    // Evento para atualizar a barra de progresso
    xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable) {
            const percentComplete = (event.loaded / event.total) * 100;
            progressBar.style.width = percentComplete + '%';
        }
    });

    // Evento para quando o upload termina (com sucesso)
    xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
            const fileInfo = JSON.parse(xhr.responseText);
            renderFileEntry(fileInfo, fileEntryDiv);
        } else {
            // Se o servidor retornar um erro
            fileEntryDiv.innerHTML = `<p style="color: red;">Falha ao enviar ${file.name} (Erro ${xhr.status})</p>`;
        }
    });

    // Evento para erros de rede
    xhr.addEventListener('error', () => {
        fileEntryDiv.innerHTML = `<p style="color: red;">Erro de rede ao enviar ${file.name}.</p>`;
    });

    xhr.open('POST', '/upload', true);
    xhr.send(formData);
}


/**
 * --- FUNÇÃO ATUALIZADA SEM A MENSAGEM DE ERRO ---
 * Busca a lista de arquivos do servidor e os renderiza na tela.
 */
async function loadInitialFiles() {
    try {
        const response = await fetch('/list-files');
        if (!response.ok) throw new Error('Falha ao buscar a lista de arquivos.');
        const files = await response.json();
        resultDiv.innerHTML = '';
        files.forEach(fileInfo => renderFileEntry(fileInfo));
        updateBulkActionsVisibility();
    } catch (error) {
        // A MENSAGEM FOI REMOVIDA DAQUI
        console.error('Erro ao carregar arquivos iniciais:', error);
    }
}

// O restante do código (renderFileEntry e Ações em Massa) permanece o mesmo
function renderFileEntry(fileInfo, elementToUpdate = null) {
    const fileHtml = `
        <input type="checkbox" class="file-checkbox" data-filename="${fileInfo.uniqueName}">
        <div class="file-info">
            <p><a href="/files/${fileInfo.uniqueName}" target="_blank">${fileInfo.originalName}</a></p>
            <small>${fileInfo.uniqueName}</small>
        </div>
    `;
    if (elementToUpdate) {
        elementToUpdate.innerHTML = fileHtml;
    } else {
        const fileEntryDiv = document.createElement('div');
        fileEntryDiv.className = 'file-entry';
        fileEntryDiv.innerHTML = fileHtml;
        resultDiv.appendChild(fileEntryDiv);
    }
    updateBulkActionsVisibility();
}

function updateBulkActionsVisibility() {
    const fileEntries = document.querySelectorAll('.file-entry');
    bulkActionsDiv.style.display = fileEntries.length > 0 ? 'flex' : 'none';
    if (fileEntries.length === 0) {
        selectAllCheckbox.checked = false;
        selectAllCheckbox.indeterminate = false;
        updateBulkActionsState();
    }
}

function updateBulkActionsState() {
    const allCheckboxes = document.querySelectorAll('.file-checkbox');
    const checkedCheckboxes = document.querySelectorAll('.file-checkbox:checked');
    bulkDeleteButton.disabled = checkedCheckboxes.length === 0;
    if (allCheckboxes.length > 0 && checkedCheckboxes.length === allCheckboxes.length) {
        selectAllCheckbox.checked = true;
        selectAllCheckbox.indeterminate = false;
    } else if (checkedCheckboxes.length > 0) {
        selectAllCheckbox.checked = false;
        selectAllCheckbox.indeterminate = true;
    } else {
        selectAllCheckbox.checked = false;
        selectAllCheckbox.indeterminate = false;
    }
}

selectAllCheckbox.addEventListener('change', () => {
    document.querySelectorAll('.file-checkbox').forEach(checkbox => checkbox.checked = selectAllCheckbox.checked);
    updateBulkActionsState();
});

resultDiv.addEventListener('change', (e) => {
    if (e.target.classList.contains('file-checkbox')) {
        updateBulkActionsState();
    }
});

bulkDeleteButton.addEventListener('click', () => {
    const checkedCheckboxes = document.querySelectorAll('.file-checkbox:checked');
    const filenamesToDelete = [...checkedCheckboxes].map(cb => cb.dataset.filename);
    if (filenamesToDelete.length === 0) return;
    if (!confirm(`Tem certeza que deseja excluir ${filenamesToDelete.length} arquivo(s)?`)) return;
    fetch('/delete-batch', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ filenames: filenamesToDelete })
    })
    .then(response => {
        if (!response.ok) throw new Error('Erro na resposta do servidor.');
        return response.json();
    })
    .then(data => {
        data.success.forEach(filename => {
            const checkbox = document.querySelector(`.file-checkbox[data-filename="${filename}"]`);
            if (checkbox) checkbox.closest('.file-entry').remove();
        });
        if (data.failed && data.failed.length > 0) {
            alert(`Falha ao excluir os seguintes arquivos: ${data.failed.join(', ')}`);
        }
        updateBulkActionsVisibility();
        updateBulkActionsState();
    })
    .catch(error => {
        console.error('Erro na exclusão em massa:', error);
        alert('Ocorreu um erro ao tentar excluir os arquivos.');
    });
});