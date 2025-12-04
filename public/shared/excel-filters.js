class ExcelFilterManager {
    constructor(options = {}) {
        this.container = options.container || null;
        this.columns = options.columns || [];
        this.onFilterChange = options.onFilterChange || (() => {});
        this.filters = {};
        this.activeFilters = {};
        this.showAllEntries = options.showAllEntries || false;
        this.maxEntries = options.maxEntries || 10000;
    }

    init() {
        this.render();
        this.attachEventListeners();
    }

    render() {
        if (!this.container) return;

        const filterHtml = `
            <div class="excel-filter-container bg-white rounded-lg border border-gray-200 p-4 mb-4">
                <!-- Show Entries Control -->
                <div class="flex items-center justify-between mb-4">
                    <div class="flex items-center space-x-3">
                        <label class="text-sm font-medium text-gray-700">Show entries:</label>
                        <select id="showEntriesSelect" class="border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                            <option value="10">10</option>
                            <option value="25" selected>25</option>
                            <option value="50">50</option>
                            <option value="100">100</option>
                            <option value="250">250</option>
                            <option value="500">500</option>
                            <option value="1000">1,000</option>
                            <option value="2500">2,500</option>
                            <option value="5000">5,000</option>
                            <option value="10000">10,000</option>
                        </select>
                        <span class="text-sm text-gray-500">entries</span>
                    </div>

                    <div class="flex items-center space-x-3">
                        <button id="clearAllFiltersBtn" class="text-sm text-blue-600 hover:text-blue-700 font-medium">
                            <i class="fas fa-filter-circle-xmark mr-1"></i>
                            Clear All Filters
                        </button>
                        <button id="toggleFiltersBtn" class="text-sm text-gray-600 hover:text-gray-700 font-medium">
                            <i class="fas fa-filter mr-1"></i>
                            <span id="filterToggleText">Show Filters</span>
                        </button>
                    </div>
                </div>

                <!-- Column Filters -->
                <div id="columnFilters" class="hidden">
                    <div class="border-t pt-4">
                        <h4 class="text-sm font-semibold text-gray-700 mb-3">Column Filters</h4>
                        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                            ${this.renderColumnFilters()}
                        </div>

                        <div class="mt-4 flex justify-end">
                            <button id="applyFiltersBtn" class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors">
                                <i class="fas fa-check mr-1"></i>
                                Apply Filters
                            </button>
                        </div>
                    </div>
                </div>

                <!-- Active Filters Display -->
                <div id="activeFiltersDisplay" class="hidden mt-3">
                    <div class="flex flex-wrap gap-2">
                        <span class="text-sm text-gray-600">Active filters:</span>
                        <div id="activeFiltersList" class="flex flex-wrap gap-2"></div>
                    </div>
                </div>
            </div>
        `;

        this.container.innerHTML = filterHtml;
    }

    renderColumnFilters() {
        return this.columns.map(column => this.renderSingleFilter(column)).join('');
    }

    renderSingleFilter(column) {
        const filterId = `filter_${column.key}`;
        const activeFilter = this.activeFilters[column.key] || '';

        switch (column.type) {
            case 'text':
                return `
                    <div class="filter-group">
                        <label class="block text-xs font-medium text-gray-700 mb-1">${column.label}</label>
                        <div class="relative">
                            <input type="text"
                                   id="${filterId}"
                                   class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm pr-8 focus:outline-none focus:ring-2 focus:ring-blue-500"
                                   placeholder="Search ${column.label.toLowerCase()}..."
                                   value="${activeFilter}">
                            <button class="absolute right-2 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 clear-filter-btn" data-column="${column.key}">
                                <i class="fas fa-times text-xs"></i>
                            </button>
                        </div>
                    </div>
                `;

            case 'select':
                const options = column.options || [];
                return `
                    <div class="filter-group">
                        <label class="block text-xs font-medium text-gray-700 mb-1">${column.label}</label>
                        <select id="${filterId}" class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
                            <option value="">All ${column.label}</option>
                            ${options.map(option => `
                                <option value="${option.value}" ${activeFilter === option.value ? 'selected' : ''}>
                                    ${option.label}
                                </option>
                            `).join('')}
                        </select>
                    </div>
                `;

            case 'number':
                return `
                    <div class="filter-group">
                        <label class="block text-xs font-medium text-gray-700 mb-1">${column.label}</label>
                        <div class="flex space-x-2">
                            <input type="number"
                                   id="${filterId}_min"
                                   class="w-1/2 border border-gray-300 rounded-md px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                   placeholder="Min"
                                   value="${this.extractMinValue(activeFilter)}">
                            <input type="number"
                                   id="${filterId}_max"
                                   class="w-1/2 border border-gray-300 rounded-md px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                   placeholder="Max"
                                   value="${this.extractMaxValue(activeFilter)}">
                        </div>
                    </div>
                `;

            default:
                return '';
        }
    }

    extractMinValue(filterValue) {
        if (!filterValue || !filterValue.includes('-')) return '';
        return filterValue.split('-')[0];
    }

    extractMaxValue(filterValue) {
        if (!filterValue || !filterValue.includes('-')) return '';
        return filterValue.split('-')[1];
    }

    attachEventListeners() {
        // Show entries control
        const showEntriesSelect = document.getElementById('showEntriesSelect');
        if (showEntriesSelect) {
            showEntriesSelect.addEventListener('change', (e) => {
                this.onFilterChange();
            });
        }

        // Toggle filters
        const toggleFiltersBtn = document.getElementById('toggleFiltersBtn');
        if (toggleFiltersBtn) {
            toggleFiltersBtn.addEventListener('click', () => {
                this.toggleFilterVisibility();
            });
        }

        // Apply filters
        const applyFiltersBtn = document.getElementById('applyFiltersBtn');
        if (applyFiltersBtn) {
            applyFiltersBtn.addEventListener('click', () => {
                this.applyFilters();
            });
        }

        // Clear all filters
        const clearAllFiltersBtn = document.getElementById('clearAllFiltersBtn');
        if (clearAllFiltersBtn) {
            clearAllFiltersBtn.addEventListener('click', () => {
                this.clearAllFilters();
            });
        }

        // Individual filter clear buttons
        document.addEventListener('click', (e) => {
            if (e.target.closest('.clear-filter-btn')) {
                const column = e.target.closest('.clear-filter-btn').dataset.column;
                this.clearFilter(column);
            }
        });

        // Enter key support for text inputs
        document.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && e.target.id && e.target.id.startsWith('filter_')) {
                this.applyFilters();
            }
        });
    }

    toggleFilterVisibility() {
        const columnFilters = document.getElementById('columnFilters');
        const toggleText = document.getElementById('filterToggleText');

        if (columnFilters.classList.contains('hidden')) {
            columnFilters.classList.remove('hidden');
            toggleText.textContent = 'Hide Filters';
        } else {
            columnFilters.classList.add('hidden');
            toggleText.textContent = 'Show Filters';
        }
    }

    applyFilters() {
        // Collect filter values
        this.filters = {};

        this.columns.forEach(column => {
            const filterId = `filter_${column.key}`;
            let filterValue = '';

            switch (column.type) {
                case 'text':
                case 'select':
                    const element = document.getElementById(filterId);
                    if (element) {
                        filterValue = element.value.trim();
                    }
                    break;

                case 'number':
                    const minElement = document.getElementById(`${filterId}_min`);
                    const maxElement = document.getElementById(`${filterId}_max`);

                    const minValue = minElement ? minElement.value.trim() : '';
                    const maxValue = maxElement ? maxElement.value.trim() : '';

                    if (minValue && maxValue) {
                        filterValue = `${minValue}-${maxValue}`;
                    } else if (minValue) {
                        filterValue = minValue;
                    } else if (maxValue) {
                        filterValue = `0-${maxValue}`;
                    }
                    break;
            }

            if (filterValue) {
                this.filters[column.key] = filterValue;
                this.activeFilters[column.key] = filterValue;
            } else {
                delete this.activeFilters[column.key];
            }
        });

        // Update active filters display
        this.updateActiveFiltersDisplay();

        // Trigger filter change callback
        this.onFilterChange();
    }

    clearFilter(column) {
        const filterId = `filter_${column}`;

        // Clear text and select inputs
        const textInput = document.getElementById(filterId);
        if (textInput) {
            textInput.value = '';
        }

        // Clear number range inputs
        const minInput = document.getElementById(`${filterId}_min`);
        const maxInput = document.getElementById(`${filterId}_max`);
        if (minInput) minInput.value = '';
        if (maxInput) maxInput.value = '';

        // Remove from active filters
        delete this.activeFilters[column];
        delete this.filters[column];

        // Update display and apply
        this.updateActiveFiltersDisplay();
        this.applyFilters();
    }

    clearAllFilters() {
        // Clear all filter inputs
        this.columns.forEach(column => {
            this.clearFilter(column.key);
        });

        // Reset filters object
        this.filters = {};
        this.activeFilters = {};

        // Update display
        this.updateActiveFiltersDisplay();
        this.applyFilters();
    }

    updateActiveFiltersDisplay() {
        const activeFiltersDisplay = document.getElementById('activeFiltersDisplay');
        const activeFiltersList = document.getElementById('activeFiltersList');

        if (Object.keys(this.activeFilters).length === 0) {
            activeFiltersDisplay.classList.add('hidden');
            return;
        }

        activeFiltersDisplay.classList.remove('hidden');

        const filterTags = Object.entries(this.activeFilters).map(([key, value]) => {
            const column = this.columns.find(col => col.key === key);
            const label = column ? column.label : key;

            return `
                <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                    ${label}: ${value}
                    <button class="ml-1 text-blue-600 hover:text-blue-800" onclick="excelFilterManager.clearFilter('${key}')">
                        <i class="fas fa-times text-xs"></i>
                    </button>
                </span>
            `;
        }).join('');

        activeFiltersList.innerHTML = filterTags;
    }

    // Public methods
    getShowEntries() {
        const showEntriesSelect = document.getElementById('showEntriesSelect');
        return showEntriesSelect ? parseInt(showEntriesSelect.value) : 25;
    }

    getFilters() {
        return { ...this.filters };
    }

    getAllParameters() {
        return {
            limit: this.getShowEntries(),
            filters: this.getFilters()
        };
    }

    setInitialValues(values) {
        if (values.limit) {
            const showEntriesSelect = document.getElementById('showEntriesSelect');
            if (showEntriesSelect) {
                showEntriesSelect.value = values.limit;
            }
        }

        if (values.filters) {
            this.activeFilters = { ...values.filters };
            this.render(); // Re-render with initial values
        }
    }

    // Utility methods
    highlightFilteredResults() {
        // Add visual indication for filtered results
        const filteredElements = document.querySelectorAll('[data-filter-highlight]');
        filteredElements.forEach(el => {
            el.style.backgroundColor = '#fef3c7';
            setTimeout(() => {
                el.style.backgroundColor = '';
            }, 2000);
        });
    }

    exportFilters() {
        return {
            filters: this.getFilters(),
            showEntries: this.getShowEntries(),
            timestamp: new Date().toISOString()
        };
    }

    importFilters(filterData) {
        if (filterData.filters) {
            this.activeFilters = { ...filterData.filters };
        }

        if (filterData.showEntries) {
            const showEntriesSelect = document.getElementById('showEntriesSelect');
            if (showEntriesSelect) {
                showEntriesSelect.value = filterData.showEntries;
            }
        }

        this.render();
        this.applyFilters();
    }
}