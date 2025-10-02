// --- LÓGICA DE TEMA E AUTENTICAÇÃO ---

const themeToggleButton = document.getElementById('theme-toggle');
const loginView = document.getElementById('login-view');
const appView = document.getElementById('app-view');
const loginForm = document.getElementById('login-form');
const loginError = document.getElementById('login-error');
const logoutButton = document.getElementById('logout-button');

// Função para fazer requisições autenticadas
async function fetchAuthenticated(url, options = {}) {
    const token = localStorage.getItem('token');
    if (!token) {
        showLoginView();
        throw new Error('Não autenticado');
    }

    const headers = {
        ...options.headers,
        'Authorization': `Bearer ${token}`
    };

    const response = await fetch(url, { ...options, headers });

    if (response.status === 401) { // Token expirado ou inválido
        logout();
        throw new Error('Sessão expirada');
    }

    return response;
}

const applyTheme = (theme) => {
    document.body.dataset.theme = theme;
    themeToggleButton.textContent = theme === 'dark' ? '☀️' : '🌙';
    localStorage.setItem('theme', theme);
};

themeToggleButton.addEventListener('click', () => {
    const newTheme = (document.body.dataset.theme || 'light') === 'light' ? 'dark' : 'light';
    applyTheme(newTheme);
});

const showLoginView = () => {
    loginView.style.display = 'block';
    appView.style.display = 'none';
};

const showAppView = () => {
    loginView.style.display = 'none';
    appView.style.display = 'block';
    loadInitialFiles();
};

const logout = () => {
    localStorage.removeItem('token');
    showLoginView();
};

loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    loginError.textContent = '';
    const username = e.target.username.value;
    const password = e.target.password.value;

    try {
        const response = await fetch('/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });
        if (!response.ok) {
            throw new Error('Usuário ou senha inválidos');
        }
        const data = await response.json();
        localStorage.setItem('token', data.token);
        showAppView();
    } catch (error) {
        loginError.textContent = error.message;
    }
});

logoutButton.addEventListener('click', logout);


// Inicialização da página
applyTheme(localStorage.getItem('theme') || 'light');
if (localStorage.getItem('token')) {
    showAppView();
} else {
    showLoginView();
}

// --- RESTANTE DO CÓDIGO DA APLICAÇÃO ---

const dropZone = document.getElementById('drop-zone');
const uploadButton = document.getElementById('upload-button');
const fileInput = document.getElementById('file-input');
const resultDiv = document.getElementById('result');
const bulkActionsDiv = document.getElementById('bulk-actions');
const selectAllCheckbox = document.getElementById('select-all-checkbox');
const bulkDeleteButton = document.getElementById('bulk-delete-button');

uploadButton.addEventListener('click', () => fileInput.click());
fileInput.addEventListener('change', () => handleFiles(fileInput.files));
['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => dropZone.addEventListener(eventName, preventDefaults));
dropZone.addEventListener('dragenter', () => dropZone.classList.add('dragover'));
dropZone.addEventListener('dragleave', () => dropZone.classList.remove('dragover'));
dropZone.addEventListener('drop', (e) => { dropZone.classList.remove('dragover'); handleFiles(e.dataTransfer.files); });
function preventDefaults(e) { e.preventDefault(); e.stopPropagation(); }
function handleFiles(files) { if (files.length > 0) [...files].forEach(uploadFile); }

async function loadInitialFiles() {
    try {
        const response = await fetchAuthenticated('/list-files');
        if (!response.ok) throw new Error('Falha ao buscar a lista de arquivos.');
        const files = await response.json();
        resultDiv.innerHTML = '';
        files.forEach(fileInfo => renderFileEntry(fileInfo));
        updateBulkActionsVisibility();
    } catch (error) { console.error('Erro ao carregar arquivos iniciais:', error); }
}

function uploadFile(file) {
    const fileId = `file-entry-${Date.now()}-${Math.random()}`;
    const placeholderHtml = `<div id="${fileId}" class="file-entry"><div class="file-info"><p>Enviando ${file.name}...</p><div class="progress-bar-container"><div class="progress-bar"></div></div></div></div>`;
    resultDiv.insertAdjacentHTML('beforeend', placeholderHtml);
    const fileEntryDiv = document.getElementById(fileId);
    const progressBar = fileEntryDiv.querySelector('.progress-bar');
    const formData = new FormData();
    formData.append('file', file);
    const xhr = new XMLHttpRequest();
    xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable) {
            const percentComplete = (event.loaded / event.total) * 100;
            progressBar.style.width = percentComplete + '%';
        }
    });
    xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
            const fileInfo = JSON.parse(xhr.responseText);
            renderFileEntry(fileInfo, fileEntryDiv);
        } else {
            fileEntryDiv.innerHTML = `<p style="color: red;">Falha ao enviar ${file.name} (Erro ${xhr.status})</p>`;
        }
    });
    xhr.addEventListener('error', () => { fileEntryDiv.innerHTML = `<p style="color: red;">Erro de rede ao enviar ${file.name}.</p>`; });
    xhr.open('POST', '/upload', true);
    const token = localStorage.getItem('token');
    xhr.setRequestHeader('Authorization', `Bearer ${token}`);
    xhr.send(formData);
}

function renderFileEntry(fileInfo, elementToUpdate = null) {
    const fileHtml = `<input type="checkbox" class="file-checkbox" data-filename="${fileInfo.uniqueName}"><div class="file-info"><p><a href="/files/${fileInfo.uniqueName}" target="_blank">${fileInfo.originalName}</a></p><small>${fileInfo.uniqueName}</small></div>`;
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

selectAllCheckbox.addEventListener('change', () => { document.querySelectorAll('.file-checkbox').forEach(checkbox => checkbox.checked = selectAllCheckbox.checked); updateBulkActionsState(); });
resultDiv.addEventListener('change', (e) => { if (e.target.classList.contains('file-checkbox')) { updateBulkActionsState(); } });
bulkDeleteButton.addEventListener('click', async () => {
    const checkedCheckboxes = document.querySelectorAll('.file-checkbox:checked');
    const filenamesToDelete = [...checkedCheckboxes].map(cb => cb.dataset.filename);
    if (filenamesToDelete.length === 0 || !confirm(`Tem certeza que deseja excluir ${filenamesToDelete.length} arquivo(s)?`)) return;

    try {
        const response = await fetchAuthenticated('/delete-batch', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ filenames: filenamesToDelete })
        });
        if (!response.ok) throw new Error('Erro na resposta do servidor.');
        const data = await response.json();
        data.success.forEach(filename => {
            const checkbox = document.querySelector(`.file-checkbox[data-filename="${filename}"]`);
            if (checkbox) checkbox.closest('.file-entry').remove();
        });
        if (data.failed && data.failed.length > 0) {
            alert(`Falha ao excluir os seguintes arquivos: ${data.failed.join(', ')}`);
        }
        updateBulkActionsVisibility();
        updateBulkActionsState();
    } catch (error) {
        console.error('Erro na exclusão em massa:', error);
        alert('Ocorreu um erro ao tentar excluir os arquivos.');
    }
});