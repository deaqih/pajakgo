class CursorPaginationManager {
    constructor(options = {}) {
        this.defaultLimit = options.defaultLimit || 25;
        this.limitOptions = options.limitOptions || [10, 25, 50, 100];
        this.defaultMode = options.defaultMode || 'cursor'; // 'cursor' or 'offset'

        // Current state
        this.currentMode = this.defaultMode;
        this.currentPage = 1;
        this.currentLimit = this.defaultLimit;
        this.currentSearch = '';
        this.currentOrderBy = 'created_at';
        this.currentOrderDir = 'desc';
        this.currentCursor = null;
        this.prevCursors = []; // History for backward navigation

        // Callbacks
        this.onPageChange = options.onPageChange || (() => {});
        this.onLimitChange = options.onLimitChange || (() => {});
        this.onSearchChange = options.onSearchChange || (() => {});
        this.onModeChange = options.onModeChange || (() => {});
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
        const modeSelector = this.container.querySelector('.mode-selector');

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

        // Render mode selector
        if (modeSelector) {
            this.renderModeSelector(modeSelector);
        }
    }

    renderPaginationControls(container) {
        // Handle both snake_case (from API) and camelCase
        const currentPage = this.paginationData?.current_page || this.paginationData?.currentPage || 1;
        const lastPage = this.paginationData?.last_page || this.paginationData?.lastPage || 1;
        const hasMore = this.paginationData?.has_more !== undefined ? this.paginationData.has_more : this.paginationData?.hasMore;
        const mode = this.paginationData?.mode || this.currentMode;
        const nextCursor = this.paginationData?.next_cursor || this.paginationData?.nextCursor;
        const prevCursor = this.paginationData?.prev_cursor || this.paginationData?.prevCursor;

        const from = this.paginationData?.from || 0;
        const to = this.paginationData?.to || 0;
        const total = this.paginationData?.total || 0;

        let html = '<div class="flex items-center justify-between">';
        html += '<div class="text-sm text-gray-700">';

        // Show mode indicator
        const modeBadge = mode === 'cursor' ?
            '<span class="ml-2 px-2 py-1 text-xs bg-green-100 text-green-800 rounded-full">Cursor</span>' :
            '<span class="ml-2 px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded-full">Offset</span>';

        html += `Showing <span class="font-medium">${from}</span> to <span class="font-medium">${to}</span> of <span class="font-medium">${total}</span> results${modeBadge}`;

        // Show max records warning
        if (total >= 10000) {
            html += '<span class="ml-2 text-xs text-amber-600 font-medium">(max 10,000 records)</span>';
        }

        html += '</div>';

        // Page navigation
        html += '<div class="flex items-center space-x-2">';

        if (mode === 'cursor') {
            // Cursor-based navigation
            const prevDisabled = !prevCursor || this.currentPage <= 1 ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-100';
            const nextDisabled = !nextCursor || !hasMore ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-100';

            html += `<button onclick="cursorPaginationManager.goPrevious()" class="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md ${prevDisabled}" ${!prevCursor ? 'disabled' : ''}>Previous</button>`;

            // Current page indicator
            html += `<span class="px-3 py-2 text-sm font-medium text-gray-700 bg-gray-100 border border-gray-300 rounded-md">Page ${currentPage}</span>`;

            html += `<button onclick="cursorPaginationManager.goNext()" class="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md ${nextDisabled}" ${!hasMore ? 'disabled' : ''}>Next</button>`;

        } else {
            // Traditional offset-based pagination
            const prevDisabled = currentPage <= 1 ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-100';
            const nextDisabled = !hasMore ? 'opacity-50 cursor-not-allowed' : 'hover:bg-gray-100';

            html += `<button onclick="cursorPaginationManager.goToPage(${currentPage - 1})" class="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md ${prevDisabled}" ${currentPage <= 1 ? 'disabled' : ''}>Previous</button>`;

            // Page numbers (show fewer for performance)
            const pageNumbers = this.getPageNumbers(currentPage, lastPage);
            for (const page of pageNumbers) {
                if (page === '...') {
                    html += '<span class="px-3 py-2 text-sm font-medium text-gray-500">...</span>';
                } else {
                    const activeClass = page === currentPage ? 'bg-blue-500 text-white hover:bg-blue-600' : 'bg-white text-gray-700 hover:bg-gray-100 border border-gray-300';
                    html += `<button onclick="cursorPaginationManager.goToPage(${page})" class="px-3 py-2 text-sm font-medium ${activeClass} rounded-md">${page}</button>`;
                }
            }

            html += `<button onclick="cursorPaginationManager.goToPage(${currentPage + 1})" class="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md ${nextDisabled}" ${!hasMore ? 'disabled' : ''}>Next</button>`;
        }

        html += '</div>';
        html += '</div>';

        container.innerHTML = html;
    }

    renderLimitSelector(container) {
        let html = '<div class="flex items-center space-x-2">';
        html += '<label class="text-sm text-gray-700">Show:</label>';
        html += '<select onchange="cursorPaginationManager.changeLimit(this.value)" class="border border-gray-300 rounded-md px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">';

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
        html += '<input type="text" placeholder="Search..." id="searchInput" class="border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" value="' + this.currentSearch + '">';
        html += '<button onclick="cursorPaginationManager.performSearch()" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-md text-sm">Search</button>';
        if (this.currentSearch) {
            html += '<button onclick="cursorPaginationManager.clearSearch()" class="bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded-md text-sm">Clear</button>';
        }
        html += '</div>';

        container.innerHTML = html;

        // Attach enter key listener for search
        const input = container.querySelector('#searchInput');
        if (input) {
            input.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.performSearch();
                }
            });
        }
    }

    renderModeSelector(container) {
        let html = '<div class="flex items-center space-x-2">';
        html += '<label class="text-sm text-gray-700">Mode:</label>';
        html += '<select onchange="cursorPaginationManager.changeMode(this.value)" class="border border-gray-300 rounded-md px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">';

        const cursorSelected = this.currentMode === 'cursor' ? 'selected' : '';
        const offsetSelected = this.currentMode === 'offset' ? 'selected' : '';

        html += `<option value="cursor" ${cursorSelected}>Cursor (Fast)</option>`;
        html += `<option value="offset" ${offsetSelected}>Offset (Traditional)</option>`;

        html += '</select>';
        html += '</div>';

        container.innerHTML = html;
    }

    getPageNumbers(currentPage, lastPage) {
        const delta = 1; // Show fewer pages for cursor mode
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

    // Cursor-based navigation
    goNext() {
        if (this.currentMode !== 'cursor') {
            this.goToPage(this.currentPage + 1);
            return;
        }

        if (!this.paginationData?.next_cursor && !this.paginationData?.nextCursor) {
            return;
        }

        // Store current cursor for backward navigation
        if (this.currentCursor) {
            this.prevCursors.push(this.currentCursor);
        }

        this.currentCursor = this.paginationData?.next_cursor || this.paginationData?.nextCursor;
        this.currentPage++;
        this.triggerChange();
    }

    goPrevious() {
        if (this.currentMode !== 'cursor') {
            this.goToPage(this.currentPage - 1);
            return;
        }

        if (!this.paginationData?.prev_cursor && !this.paginationData?.prevCursor && this.prevCursors.length === 0) {
            return;
        }

        // Use previous cursor if available, otherwise use API prev cursor
        if (this.prevCursors.length > 0) {
            this.currentCursor = this.prevCursors.pop();
        } else {
            this.currentCursor = this.paginationData?.prev_cursor || this.paginationData?.prevCursor;
        }

        this.currentPage--;
        this.triggerChange();
    }

    // Traditional page-based navigation (for offset mode)
    goToPage(page) {
        if (page < 1 || page === this.currentPage) return;

        // Check if page exceeds lastPage
        if (this.paginationData && this.paginationData.last_page && page > this.paginationData.last_page) {
            return;
        }

        this.currentPage = page;
        this.currentCursor = null; // Reset cursor for offset mode
        this.triggerChange();
    }

    changeLimit(limit) {
        this.currentLimit = parseInt(limit);
        this.reset();
        this.triggerChange();
        this.render();
    }

    changeMode(mode) {
        if (mode === this.currentMode) return;

        this.currentMode = mode;
        this.reset();
        this.onModeChange(mode);
        this.triggerChange();
        this.render();
    }

    performSearch() {
        const searchInput = this.container?.querySelector('#searchInput');
        if (searchInput) {
            this.currentSearch = searchInput.value.trim();
            this.reset();
            this.triggerChange();
            this.render();
        }
    }

    clearSearch() {
        this.currentSearch = '';
        this.reset();
        this.triggerChange();
        this.render();
    }

    triggerChange() {
        // Build parameters based on current mode
        const params = {
            page: this.currentPage,
            limit: this.currentLimit,
            search: this.currentSearch,
            mode: this.currentMode,
            order_by: this.currentOrderBy,
            order_dir: this.currentOrderDir
        };

        if (this.currentMode === 'cursor' && this.currentCursor) {
            params.cursor = this.currentCursor;
        }

        this.onPageChange(params);
    }

    updateData(paginationData) {
        // Reduced logging for performance
        this.paginationData = paginationData;

        // Update current cursor from API response
        if (paginationData.next_cursor || paginationData.nextCursor) {
            this.currentCursor = paginationData.next_cursor || paginationData.nextCursor;
        }

        this.currentPage = paginationData.current_page || paginationData.currentPage || 1;
        this.currentLimit = paginationData.per_page || paginationData.perPage || this.currentLimit;

        this.render();
    }

    reset() {
        this.currentPage = 1;
        this.currentCursor = null;
        this.prevCursors = [];
        this.paginationData = null;
    }

    // Get current API parameters
    getApiParams() {
        const params = {
            limit: this.currentLimit,
            mode: this.currentMode,
            order_by: this.currentOrderBy,
            order_dir: this.currentOrderDir
        };

        if (this.currentSearch) {
            params.search = this.currentSearch;
        }

        if (this.currentMode === 'cursor') {
            if (this.currentCursor) {
                params.cursor = this.currentCursor;
            }
            // Don't include page in cursor mode
        } else {
            params.page = this.currentPage;
        }

        return params;
    }

    // Utility method to build query string
    buildQueryString() {
        const params = this.getApiParams();
        return Object.keys(params)
            .filter(key => params[key] !== undefined && params[key] !== '')
            .map(key => `${encodeURIComponent(key)}=${encodeURIComponent(params[key])}`)
            .join('&');
    }
}