// Global auth helper for sending Authorization header
class AuthHelper {
    constructor() {
        this.token = localStorage.getItem('access_token');
        this.user = JSON.parse(localStorage.getItem('user') || '{}');
        this.init();
    }

    init() {
        // Check if token exists on page load
        if (!this.token) {
            // If we're on a protected page, redirect to login
            const protectedPages = ['/', '/accounts', '/additional-analyses', '/koreksi-rules', '/obyek-rules', '/uploads'];
            const currentPath = window.location.pathname;

            if (protectedPages.some(path => currentPath.startsWith(path))) {
                window.location.href = '/login';
                return;
            }
        }

        // Add Authorization header to all fetch requests
        this.interceptFetch();

        // Add token to all forms
        this.addTokenToForms();

        // Display user info
        this.displayUserInfo();
    }

    interceptFetch() {
        const originalFetch = window.fetch;
        window.fetch = async (...args) => {
            let [url, options = {}] = args;

            // Only add Authorization header for API requests
            if (typeof url === 'string' && url.includes('/api/v1/') && this.token) {
                options.headers = {
                    ...options.headers,
                    'Authorization': `Bearer ${this.token}`
                };
            }

            return originalFetch.apply(window, args);
        };
    }

    addTokenToForms() {
        // Add hidden token input to all forms
        const forms = document.querySelectorAll('form');
        forms.forEach(form => {
            // Check if form already has token input
            if (!form.querySelector('input[name="auth_token"]')) {
                const tokenInput = document.createElement('input');
                tokenInput.type = 'hidden';
                tokenInput.name = 'auth_token';
                tokenInput.value = this.token || '';
                form.appendChild(tokenInput);
            }
        });
    }

    displayUserInfo() {
        const userDisplay = document.getElementById('userDisplay');
        const userDisplayMobile = document.getElementById('userDisplayMobile');

        if (userDisplay && this.user.username) {
            userDisplay.textContent = this.user.username;
        }
        if (userDisplayMobile && this.user.username) {
            userDisplayMobile.textContent = this.user.username;
        }
    }

    login(token, user) {
        this.token = token;
        this.user = user;
        localStorage.setItem('access_token', token);
        localStorage.setItem('user', JSON.stringify(user));
        this.displayUserInfo();
        this.addTokenToForms();
    }

    logout() {
        this.token = null;
        this.user = {};
        localStorage.removeItem('access_token');
        localStorage.removeItem('user');
        window.location.href = '/login';
    }

    isAuthenticated() {
        return !!this.token;
    }

    getUser() {
        return this.user;
    }
}

// Create global auth helper instance
window.authHelper = new AuthHelper();

// Make logout globally available
window.logout = function() {
    window.authHelper.logout();
};

// Helper function to make authenticated API calls
window.authenticatedFetch = async (url, options = {}) => {
    const token = localStorage.getItem('access_token');
    if (!token) {
        throw new Error('No authentication token found');
    }

    const authOptions = {
        ...options,
        headers: {
            ...options.headers,
            'Authorization': `Bearer ${token}`
        }
    };

    return fetch(url, authOptions);
};