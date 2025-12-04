class PaginationManager {
    constructor(options = {}) {
        this.defaultLimit = options.defaultLimit || 25;
        this.limitOptions = options.limitOptions || [10, 25, 50, 100];
        this.currentPage = 1;
        this.currentLimit = this.defaultLimit;
        this.currentSearch = '';
        this.onPageChange = options.onPageChange || (() => {});
        this.onLimitChange = options.onLimitChange || (() => {});
        this.onSearchChange = options.onSearchChange || (() => {});
    }

    init(containerId) {
        this.container = document.getElementById(containerId);
        if (!this.container) return;

        this.render();
    }

    render() {
        if (!this.container) return;

        const pagination = this.container.querySelector('.pagination-container');
        const limitSelector = this.container.querySelector('.limit-selector');
        const searchBox = this.container.querySelector('.search-box');

        // Render pagination controls
        if (pagination) {
            this.renderPaginationControls(pagination);
        }

        // Render limit selector
        if (limitSelector) {
            this.renderLimitSelector(limitSelector);
        }

        // Render search box
        if (searchBox) {
            this.renderSearchBox(searchBox);
        }
    }

    renderPaginationControls(container) {
        // Handle both snake_case (from API) and camelCase
        const currentPage = this.paginationData?.current_page || this.paginationData?.currentPage;
        const lastPage = this.paginationData?.last_page || this.paginationData?.lastPage;
        const hasMore = this.paginationData?.has_more !== undefined ? this.paginationData.has_more : this.paginationData?.hasMore;

        // Show default pagination even if no data yet
        if (!lastPage || lastPage <= 1) {
            container.innerHTML = `
                <div class="flex items-center justify-between">
                    <div class="text-sm text-gray-700">
                        Loading pagination...
                    </div>
                    <div class="flex items-center space-x-2">
                        <button class="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md opacity-50 cursor-not-allowed" disabled>
                            Previous
                        </button>
                        <button class="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md opacity-50 cursor-not-allowed" disabled>
                            Next
                        </button>
                    </div>
                </div>
            `;
            return;
        }

        const from = this.paginationData?.from || 0;
        const to = this.paginationData?.to || 0;
        const total = this.paginationData?.total || 0;

        let html = '<div class="flex items-center justify-between">';
        html += '<div class="text-sm text-gray-700">';
        html += `Showing <span class="font-medium">${from}</span> to <span class="font-medium">${to}</span> of <span class="font-medium">${total}</span> results`;
        html += '</div>';

        // Page navigation
        html += '<div class="flex items-center space-x-2">';

        // Previous button
        const prevDisabled = currentPage <= 1 ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-100';
        html += `<button onclick="paginationManager.goToPage(${currentPage - 1})" class="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md ${prevDisabled}" ${currentPage <= 1 ? 'disabled' : ''}>Previous</button>`;

        // Page numbers
        const pageNumbers = this.getPageNumbers(currentPage, lastPage);
        for (const page of pageNumbers) {
            if (page === '...') {
                html += '<span class="px-3 py-2 text-sm font-medium text-gray-500">...</span>';
            } else {
                const activeClass = page === currentPage ? 'bg-blue-500 text-white hover:bg-blue-600' : 'bg-white text-gray-700 hover:bg-gray-100 border border-gray-300';
                html += `<button onclick="paginationManager.goToPage(${page})" class="px-3 py-2 text-sm font-medium ${activeClass} rounded-md">${page}</button>`;
            }
        }

        // Next button
        const nextDisabled = !hasMore ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-100';
        html += `<button onclick="paginationManager.goToPage(${currentPage + 1})" class="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md ${nextDisabled}" ${!hasMore ? 'disabled' : ''}>Next</button>`;

        html += '</div>';
        html += '</div>';

        container.innerHTML = html;
    }

    renderLimitSelector(container) {
        let html = '<div class="flex items-center space-x-2">';
        html += '<label class="text-sm text-gray-700">Show:</label>';
        html += '<select onchange="paginationManager.changeLimit(this.value)" class="border border-gray-300 rounded-md px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">';

        for (const limit of this.limitOptions) {
            const selected = limit === this.currentLimit ? 'selected' : '';
            html += `<option value="${limit}" ${selected}>${limit}</option>`;
        }

        html += '</select>';
        html += '<span class="text-sm text-gray-700">per page</span>';
        html += '</div>';

        container.innerHTML = html;
    }

    renderSearchBox(container) {
        let html = '<div class="flex items-center space-x-2">';
        html += '<input type="text" placeholder="Search..." class="border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" value="' + this.currentSearch + '">';
        html += '<button onclick="paginationManager.performSearch()" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-md text-sm">Search</button>';
        if (this.currentSearch) {
            html += '<button onclick="paginationManager.clearSearch()" class="bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded-md text-sm">Clear</button>';
        }
        html += '</div>';

        container.innerHTML = html;

        // Attach enter key listener for search
        const input = container.querySelector('input[type="text"]');
        if (input) {
            input.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.performSearch();
                }
            });
        }
    }

    getPageNumbers(currentPage, lastPage) {
        const delta = 2;
        const range = [];
        const rangeWithDots = [];

        for (let i = Math.max(2, currentPage - delta); i <= Math.min(lastPage - 1, currentPage + delta); i++) {
            range.push(i);
        }

        if (currentPage - delta > 2) {
            rangeWithDots.push(1, '...');
        } else {
            rangeWithDots.push(1);
        }

        rangeWithDots.push(...range);

        if (currentPage + delta < lastPage - 1) {
            rangeWithDots.push('...', lastPage);
        } else {
            rangeWithDots.push(lastPage);
        }

        return rangeWithDots;
    }

    goToPage(page) {
        if (page < 1 || page === this.currentPage) return;

        // Check if page exceeds lastPage
        if (this.paginationData && this.paginationData.last_page && page > this.paginationData.last_page) {
            return;
        }

        this.currentPage = page;
        this.onPageChange(this.currentPage, this.currentLimit, this.currentSearch);
    }

    changeLimit(limit) {
        this.currentLimit = parseInt(limit);
        this.currentPage = 1; // Reset to first page when changing limit
        this.onLimitChange(this.currentPage, this.currentLimit, this.currentSearch);
        this.render(); // Re-render to update the select box
    }

    performSearch() {
        const searchInput = this.container?.querySelector('.search-box input[type="text"]');
        if (searchInput) {
            this.currentSearch = searchInput.value.trim();
            this.currentPage = 1; // Reset to first page when searching
            this.onSearchChange(this.currentPage, this.currentLimit, this.currentSearch);
            this.render(); // Re-render to show clear button
        }
    }

    clearSearch() {
        this.currentSearch = '';
        this.currentPage = 1;
        this.onSearchChange(this.currentPage, this.currentLimit, this.currentSearch);
        this.render(); // Re-render to hide clear button
    }

    updateData(paginationData) {
        console.log('Pagination data received:', paginationData);
        this.paginationData = paginationData;
        this.currentPage = paginationData.current_page || 1;
        this.currentLimit = paginationData.per_page || this.currentLimit;
        console.log('Current page:', this.currentPage, 'Last page:', paginationData.last_page, 'Has more:', paginationData.has_more);
        this.render();
    }

    reset() {
        this.currentPage = 1;
        this.currentLimit = this.defaultLimit;
        this.currentSearch = '';
        this.paginationData = null;
        this.render();
    }
}