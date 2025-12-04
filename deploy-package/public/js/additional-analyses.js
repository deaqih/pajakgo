// Global variables
let currentPage = 1;
let totalItems = 0;
let totalPages = 1;
let itemsPerPage = 25;
let currentEditingId = null;
let deleteItemId = null;
let accounts = [];
let analysisTypes = [];

// Initialize page
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
    loadAnalyses();
    loadAccounts();
    loadAnalysisTypes();
    setupEventListeners();
});

// Initialize app
function initializeApp() {
    // Check authentication
    const token = localStorage.getItem('access_token');
    if (!token) {
        window.location.href = '/login';
        return;
    }

    // Display user info
    displayUserInfo();
}

// Setup event listeners
function setupEventListeners() {
    // Search input
    document.getElementById('searchInput').addEventListener('input', debounce(handleSearch, 300));

    // Filter dropdowns
    document.getElementById('accountFilter').addEventListener('change', handleFilter);
    document.getElementById('typeFilter').addEventListener('change', handleFilter);
    document.getElementById('statusFilter').addEventListener('change', handleFilter);

    // Limit selector
    document.getElementById('limitSelect').addEventListener('change', handleLimitChange);

    // Form submissions
    document.getElementById('analysisForm').addEventListener('submit', handleAnalysisSubmit);
    document.getElementById('importForm').addEventListener('submit', handleImportSubmit);

    // File input change
    document.getElementById('importFile').addEventListener('change', handleFileSelect);
}

// Display user information
function displayUserInfo() {
    const userStr = localStorage.getItem('user');
    if (userStr) {
        const user = JSON.parse(userStr);
        const userDisplay = document.getElementById('userDisplay');
        if (userDisplay) {
            userDisplay.textContent = user.username || user.name || 'User';
        }
    }
}

// Load additional analyses
async function loadAnalyses() {
    try {
        const params = new URLSearchParams({
            page: currentPage,
            limit: itemsPerPage,
            search: document.getElementById('searchInput').value,
            account_code: document.getElementById('accountFilter').value,
            analysis_type: document.getElementById('typeFilter').value,
            status: document.getElementById('statusFilter').value,
            sort_by: 'created_at',
            sort_order: 'desc'
        });

        const response = await fetch(`/api/v1/additional-analyses?${params}`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
                'Content-Type': 'application/json'
            }
        });

        const result = await response.json();

        if (result.success) {
            renderTable(result.data.data);
            updatePagination(result.data.pagination);
            updateStatistics(result.data.data);
        } else {
            showError(result.message || 'Failed to load additional analyses');
        }
    } catch (error) {
        console.error('Error loading analyses:', error);
        showError('Error loading additional analyses');
    }
}

// Render table data
function renderTable(data) {
    const tableBody = document.getElementById('tableBody');

    if (data.length === 0) {
        tableBody.innerHTML = `
            <tr>
                <td colspan="6" class="px-6 py-12 text-center">
                    <div class="flex flex-col items-center">
                        <i class="fas fa-chart-line text-4xl text-gray-300 mb-4"></i>
                        <p class="text-gray-500 font-medium">No additional analyses found</p>
                        <p class="text-gray-400 text-sm mt-1">Get started by adding your first additional analysis</p>
                    </div>
                </td>
            </tr>
        `;
        return;
    }

    tableBody.innerHTML = data.map(item => `
        <tr class="hover:bg-gray-50 transition-colors">
            <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm font-medium text-gray-900">${item.account_name || 'N/A'}</div>
                <div class="text-xs text-gray-500">${item.account_code || 'N/A'}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm text-gray-900">${item.analysis_type}</div>
            </td>
            <td class="px-6 py-4">
                <div class="text-sm text-gray-900 font-medium">${item.analysis_title}</div>
            </td>
            <td class="px-6 py-4">
                <div class="text-sm text-gray-600 max-w-xs truncate" title="${item.notes || ''}">
                    ${item.notes || '-'}
                </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-center">
                <span class="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full status-${item.status}">
                    ${item.status}
                </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-center text-sm font-medium">
                <button onclick="editAnalysis(${item.id})" class="text-indigo-600 hover:text-indigo-900 mr-3 transition-colors">
                    <i class="fas fa-edit"></i>
                </button>
                <button onclick="deleteAnalysis(${item.id})" class="text-red-600 hover:text-red-900 transition-colors">
                    <i class="fas fa-trash"></i>
                </button>
            </td>
        </tr>
    `).join('');
}

// Update pagination
function updatePagination(pagination) {
    currentPage = pagination.currentPage;
    totalPages = pagination.totalPages;
    totalItems = pagination.totalItems;

    // Update showing info
    const from = (currentPage - 1) * itemsPerPage + 1;
    const to = Math.min(currentPage * itemsPerPage, totalItems);

    document.getElementById('showingFrom').textContent = from;
    document.getElementById('showingTo').textContent = to;
    document.getElementById('totalRecords').textContent = totalItems;
    document.getElementById('currentPageInfo').textContent = currentPage;
    document.getElementById('totalPagesInfo').textContent = totalPages;

    // Update pagination buttons
    updatePaginationButtons();
}

// Update pagination buttons
function updatePaginationButtons() {
    const firstPageBtn = document.getElementById('firstPageBtn');
    const prevPageBtn = document.getElementById('prevPageBtn');
    const nextPageBtn = document.getElementById('nextPageBtn');
    const lastPageBtn = document.getElementById('lastPageBtn');

    firstPageBtn.disabled = currentPage === 1;
    prevPageBtn.disabled = currentPage === 1;
    nextPageBtn.disabled = currentPage === totalPages;
    lastPageBtn.disabled = currentPage === totalPages;

    // Update page numbers
    const pageNumbers = document.getElementById('pageNumbers');
    pageNumbers.innerHTML = generatePageNumbers(currentPage, totalPages);
}

// Generate page numbers
function generatePageNumbers(current, total) {
    const maxVisible = 5;
    let pages = [];

    if (total <= maxVisible) {
        for (let i = 1; i <= total; i++) {
            pages.push(i);
        }
    } else {
        if (current <= 3) {
            pages = [1, 2, 3, 4, '...', total];
        } else if (current >= total - 2) {
            pages = [1, '...', total - 3, total - 2, total - 1, total];
        } else {
            pages = [1, '...', current - 1, current, current + 1, '...', total];
        }
    }

    return pages.map(page => {
        if (page === '...') {
            return '<span class="px-3 py-2 text-gray-500">...</span>';
        } else {
            const isActive = page === currentPage;
            return `
                <button onclick="goToPage(${page})"
                    class="px-3 py-2 text-sm font-medium ${isActive
                        ? 'bg-primary-600 text-white border-primary-600'
                        : 'text-gray-700 bg-white border-gray-300 hover:bg-gray-50'}
                    border rounded-lg transition-colors">
                    ${page}
                </button>
            `;
        }
    }).join('');
}

// Update statistics
function updateStatistics(data) {
    const totalAnalysesCount = document.getElementById('totalAnalysesCount');
    const activeAnalysesCount = document.getElementById('activeAnalysesCount');
    const analysisTypesCount = document.getElementById('analysisTypesCount');

    if (totalAnalysesCount) totalAnalysesCount.textContent = data.length;
    if (activeAnalysesCount) activeAnalysesCount.textContent = data.filter(item => item.status === 'active').length;
    if (analysisTypesCount) analysisTypesCount.textContent = analysisTypes.length;
}

// Load accounts for dropdowns
async function loadAccounts() {
    try {
        const response = await fetch('/api/v1/accounts?limit=1000', {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
                'Content-Type': 'application/json'
            }
        });

        const result = await response.json();

        if (result.success) {
            accounts = result.data.accounts || result.data || [];
            populateAccountDropdowns();
        }
    } catch (error) {
        console.error('Error loading accounts:', error);
    }
}

// Populate account dropdowns
function populateAccountDropdowns() {
    const accountSelect = document.getElementById('accountCode');
    const accountFilter = document.getElementById('accountFilter');

    if (accountSelect) {
        accountSelect.innerHTML = '<option value="">Select Account</option>' +
            accounts.map(account => `
                <option value="${account.account_code}">${account.account_code} - ${account.account_name}</option>
            `).join('');
    }

    if (accountFilter) {
        accountFilter.innerHTML = '<option value="">All Accounts</option>' +
            accounts.map(account => `
                <option value="${account.account_code}">${account.account_code} - ${account.account_name}</option>
            `).join('');
    }
}

// Load analysis types
async function loadAnalysisTypes() {
    try {
        const response = await fetch('/api/v1/additional-analyses/types', {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
                'Content-Type': 'application/json'
            }
        });

        const result = await response.json();

        if (result.success) {
            analysisTypes = result.data;
            populateAnalysisTypeFilter();
        }
    } catch (error) {
        console.error('Error loading analysis types:', error);
    }
}

// Populate analysis type filter
function populateAnalysisTypeFilter() {
    const typeFilter = document.getElementById('typeFilter');

    if (typeFilter) {
        typeFilter.innerHTML = '<option value="">All Types</option>' +
            analysisTypes.map(type => `
                <option value="${type}">${type}</option>
            `).join('');
    }
}

// Show add modal
function showAddModal() {
    currentEditingId = null;
    document.getElementById('modalTitle').textContent = 'Add Additional Analysis';
    document.getElementById('analysisForm').reset();
    document.getElementById('analysisModal').classList.remove('hidden');
}

// Show edit modal
async function editAnalysis(id) {
    currentEditingId = id;
    document.getElementById('modalTitle').textContent = 'Edit Additional Analysis';

    try {
        const response = await fetch(`/api/v1/additional-analyses/${id}`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
                'Content-Type': 'application/json'
            }
        });

        const result = await response.json();

        if (result.success) {
            const data = result.data;
            document.getElementById('accountCode').value = data.account_code || '';
            document.getElementById('analysisType').value = data.analysis_type || '';
            document.getElementById('analysisTitle').value = data.analysis_title || '';
            document.getElementById('status').value = data.status || 'active';
            document.getElementById('notes').value = data.notes || '';

            document.getElementById('analysisModal').classList.remove('hidden');
        } else {
            showError(result.message || 'Failed to load additional analysis');
        }
    } catch (error) {
        console.error('Error loading analysis:', error);
        showError('Error loading additional analysis');
    }
}

// Hide modal
function hideModal() {
    document.getElementById('analysisModal').classList.add('hidden');
    document.getElementById('analysisForm').reset();
    currentEditingId = null;
}

// Handle analysis form submission
async function handleAnalysisSubmit(e) {
    e.preventDefault();

    const formData = {
        account_code: document.getElementById('accountCode').value,
        analysis_type: document.getElementById('analysisType').value,
        analysis_title: document.getElementById('analysisTitle').value,
        status: document.getElementById('status').value,
        notes: document.getElementById('notes').value
    };

    try {
        const url = currentEditingId
            ? `/api/v1/additional-analyses/${currentEditingId}`
            : '/api/v1/additional-analyses';

        const method = currentEditingId ? 'PUT' : 'POST';

        const response = await fetch(url, {
            method: method,
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(formData)
        });

        const result = await response.json();

        if (result.success) {
            hideModal();
            showSuccess(currentEditingId ? 'Additional analysis updated successfully' : 'Additional analysis created successfully');
            loadAnalyses();
        } else {
            showError(result.message || 'Failed to save additional analysis');
        }
    } catch (error) {
        console.error('Error saving analysis:', error);
        showError('Error saving additional analysis');
    }
}

// Show delete modal
function deleteAnalysis(id) {
    deleteItemId = id;
    document.getElementById('deleteModal').classList.remove('hidden');
}

// Hide delete modal
function hideDeleteModal() {
    document.getElementById('deleteModal').classList.add('hidden');
    deleteItemId = null;
}

// Confirm delete
async function confirmDelete() {
    if (!deleteItemId) return;

    try {
        const response = await fetch(`/api/v1/additional-analyses/${deleteItemId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
                'Content-Type': 'application/json'
            }
        });

        const result = await response.json();

        if (result.success) {
            hideDeleteModal();
            showSuccess('Additional analysis deleted successfully');
            loadAnalyses();
        } else {
            showError(result.message || 'Failed to delete additional analysis');
        }
    } catch (error) {
        console.error('Error deleting analysis:', error);
        showError('Error deleting additional analysis');
    }
}

// Show import modal
function showImportModal() {
    document.getElementById('importModal').classList.remove('hidden');
}

// Hide import modal
function hideImportModal() {
    document.getElementById('importModal').classList.add('hidden');
    document.getElementById('importForm').reset();
    document.getElementById('fileName').classList.add('hidden');
}

// Handle file selection
function handleFileSelect(e) {
    const file = e.target.files[0];
    const fileNameDiv = document.getElementById('fileName');

    if (file) {
        fileNameDiv.textContent = `Selected: ${file.name}`;
        fileNameDiv.classList.remove('hidden');
    } else {
        fileNameDiv.classList.add('hidden');
    }
}

// Handle import form submission
async function handleImportSubmit(e) {
    e.preventDefault();

    const fileInput = document.getElementById('importFile');
    const file = fileInput.files[0];

    if (!file) {
        showError('Please select a file to import');
        return;
    }

    const formData = new FormData();
    formData.append('file', file);

    try {
        const response = await fetch('/api/v1/additional-analyses/import', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: formData
        });

        const result = await response.json();

        if (result.success) {
            hideImportModal();
            showSuccess(`Import completed: ${result.data.success} records imported, ${result.data.failed} failed`);
            loadAnalyses();
        } else {
            showError(result.message || 'Failed to import additional analyses');
        }
    } catch (error) {
        console.error('Error importing analyses:', error);
        showError('Error importing additional analyses');
    }
}

// Export analyses
async function exportAnalyses() {
    try {
        const exportData = {
            account_code: document.getElementById('accountFilter').value || undefined,
            analysis_type: document.getElementById('typeFilter').value || undefined,
            status: document.getElementById('statusFilter').value || undefined,
            search: document.getElementById('searchInput').value || undefined,
            format: 'excel'
        };

        const response = await fetch('/api/v1/additional-analyses/export', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(exportData)
        });

        if (response.ok) {
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `additional_analyses_${new Date().toISOString().split('T')[0]}.xlsx`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
            showSuccess('Export completed successfully');
        } else {
            showError('Failed to export additional analyses');
        }
    } catch (error) {
        console.error('Error exporting analyses:', error);
        showError('Error exporting additional analyses');
    }
}

// Download template
async function downloadTemplate() {
    try {
        const response = await fetch('/api/v1/additional-analyses/template', {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('access_token')}`
            }
        });

        if (response.ok) {
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'additional_analyses_template.csv';
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
            showSuccess('Template downloaded successfully');
        } else {
            showError('Failed to download template');
        }
    } catch (error) {
        console.error('Error downloading template:', error);
        showError('Error downloading template');
    }
}

// Handle search
function handleSearch() {
    currentPage = 1;
    loadAnalyses();
}

// Handle filter change
function handleFilter() {
    currentPage = 1;
    loadAnalyses();
}

// Handle limit change
function handleLimitChange(e) {
    itemsPerPage = parseInt(e.target.value);
    currentPage = 1;
    loadAnalyses();
}

// Pagination functions
function goToPage(page) {
    if (page >= 1 && page <= totalPages) {
        currentPage = page;
        loadAnalyses();
    }
}

function goToPreviousPage() {
    if (currentPage > 1) {
        goToPage(currentPage - 1);
    }
}

function goToNextPage() {
    if (currentPage < totalPages) {
        goToPage(currentPage + 1);
    }
}

function goToLastPage() {
    goToPage(totalPages);
}

// Utility functions
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

function showSuccess(message) {
    // Use your preferred notification system
    alert('Success: ' + message);
}

function showError(message) {
    // Use your preferred notification system
    alert('Error: ' + message);
}

function showInfo(message) {
    // Use your preferred notification system
    alert('Info: ' + message);
}

function toggleMobileMenu() {
    const mobileMenu = document.getElementById('mobile-menu');
    mobileMenu.classList.toggle('hidden');
}

function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    window.location.href = '/login';
}