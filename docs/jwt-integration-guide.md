# JWT Authentication Integration Guide

This guide provides instructions for integrating the go-layout JWT authentication system with frontend applications.

## Authentication Flow

The authentication system implements a standard JWT-based flow with short-lived access tokens and refresh capability:

1. **User Login**: Authenticate with credentials and receive an access token
2. **Protected API Calls**: Include the token in requests to protected endpoints
3. **Token Expiration**: Handle token expiration by refreshing before/when it expires
4. **Token Refresh**: Refresh the token to get a new access token without requiring re-login

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/login` | POST | Authenticate and receive access token |
| `/api/v1/refresh_token` | GET | Refresh an expired or about-to-expire token |

## Response Structure

All authentication endpoints return responses with this structure:

```json
{
  "code": 0,
  "message": "",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expire": "2023-05-01T15:04:05Z",
    "expires_in": 1800,
    "token_type": "Bearer"
  }
}
```

## Frontend Implementation

### 1. Login

```javascript
async function login(email, password) {
  try {
    const response = await fetch('/api/v1/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ email, password })
    });

    const result = await response.json();

    if (result.code !== 0) {
      throw new Error(result.message || 'Login failed');
    }

    // Store auth data
    const authData = result.data;
    localStorage.setItem('token', authData.token);
    localStorage.setItem('tokenExpiry', authData.expire);
    localStorage.setItem('tokenType', authData.token_type);

    // Setup token refresh
    setupTokenRefresh(authData.expires_in);

    return authData;
  } catch (error) {
    console.error('Login error:', error);
    throw error;
  }
}
```

### 2. Adding Auth Headers to Requests

```javascript
// Axios example
axios.interceptors.request.use(config => {
  const token = localStorage.getItem('token');
  const tokenType = localStorage.getItem('tokenType') || 'Bearer';

  if (token) {
    config.headers.Authorization = `${tokenType} ${token}`;
  }
  return config;
});

// Fetch API example
async function fetchWithAuth(url, options = {}) {
  const token = localStorage.getItem('token');
  const tokenType = localStorage.getItem('tokenType') || 'Bearer';

  const headers = new Headers(options.headers || {});
  if (token) {
    headers.append('Authorization', `${tokenType} ${token}`);
  }

  return fetch(url, {
    ...options,
    headers
  });
}
```

### 3. Token Refresh Mechanism

```javascript
// Setup refresh timer before token expires
function setupTokenRefresh(expiresInSeconds) {
  // Refresh 1 minute before expiration
  const refreshTime = (expiresInSeconds - 60) * 1000;
  setTimeout(refreshToken, refreshTime);
}

async function refreshToken() {
  try {
    const response = await fetch('/api/v1/refresh_token', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });

    const result = await response.json();

    if (result.code !== 0) {
      // Redirect to login if refresh fails
      redirectToLogin();
      return;
    }

    // Update stored auth data
    const authData = result.data;
    localStorage.setItem('token', authData.token);
    localStorage.setItem('tokenExpiry', authData.expire);

    // Setup next refresh
    setupTokenRefresh(authData.expires_in);
  } catch (error) {
    console.error('Token refresh error:', error);
    redirectToLogin();
  }
}
```

### 4. Handling Unauthorized Errors

```javascript
// Axios example
axios.interceptors.response.use(
  response => response,
  async error => {
    const originalRequest = error.config;

    // Prevent infinite retry loops
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        // Try to refresh the token
        await refreshToken();

        // Retry the original request with new token
        const token = localStorage.getItem('token');
        const tokenType = localStorage.getItem('tokenType') || 'Bearer';
        originalRequest.headers.Authorization = `${tokenType} ${token}`;

        return axios(originalRequest);
      } catch (refreshError) {
        // If refresh fails, redirect to login
        redirectToLogin();
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);
```

### 5. Token Validation Helper

```javascript
// Check if token is still valid
function isTokenValid() {
  const expiryStr = localStorage.getItem('tokenExpiry');

  if (!expiryStr) {
    return false;
  }

  const expiryTime = new Date(expiryStr).getTime();
  const currentTime = Date.now();

  // Consider token valid if not expired
  return expiryTime > currentTime;
}

// Check if token will expire soon (within 2 minutes)
function isTokenExpiringSoon() {
  const expiryStr = localStorage.getItem('tokenExpiry');

  if (!expiryStr) {
    return true;
  }

  const expiryTime = new Date(expiryStr).getTime();
  const currentTime = Date.now();

  // Check if token expires within 2 minutes
  return expiryTime - currentTime < 2 * 60 * 1000;
}
```

## Security Considerations

1. **Store tokens securely**: Use localStorage for SPAs, but consider more secure alternatives like HttpOnly cookies for applications requiring higher security.

2. **Validate tokens on the server**: Always validate tokens on the server side, never trust client-side validation alone.

3. **Keep tokens short-lived**: Our system uses 30-minute tokens by default, which helps limit the impact of token theft.

4. **HTTPS only**: Always use HTTPS in production to prevent token interception.

5. **Clear tokens on logout**: Implement a proper logout function that clears stored tokens.

```javascript
function logout() {
  localStorage.removeItem('token');
  localStorage.removeItem('tokenExpiry');
  localStorage.removeItem('tokenType');
  redirectToLogin();
}
```

## Additional Resources

- [JWT.io](https://jwt.io/) - Decode and verify JWTs for debugging
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
