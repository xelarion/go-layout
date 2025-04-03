# JWT认证集成指南

本指南提供了将go-layout JWT认证系统与前端应用程序集成的说明。

## 认证流程

认证系统实现了基于JWT的标准流程，使用短期访问令牌和刷新功能：

1. **用户登录**：使用凭据进行身份验证并接收访问令牌
2. **保护API调用**：在请求受保护的端点时包含令牌
3. **令牌过期**：在令牌过期之前/时通过刷新处理令牌过期
4. **令牌刷新**：刷新令牌以获取新的访问令牌，无需重新登录

## API端点

| 端点 | 方法 | 描述 |
|----------|--------|-------------|
| `/api/v1/login` | POST | 进行身份验证并接收访问令牌 |
| `/api/v1/refresh_token` | GET | 刷新已过期或即将过期的令牌 |

## 响应结构

所有认证端点都返回具有以下结构的响应：

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

## 前端实现

### 1. 登录

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
      throw new Error(result.message || '登录失败');
    }

    // 存储认证数据
    const authData = result.data;
    localStorage.setItem('token', authData.token);
    localStorage.setItem('tokenExpiry', authData.expire);
    localStorage.setItem('tokenType', authData.token_type);

    // 设置令牌刷新
    setupTokenRefresh(authData.expires_in);

    return authData;
  } catch (error) {
    console.error('登录错误:', error);
    throw error;
  }
}
```

### 2. 向请求添加认证头

```javascript
// Axios示例
axios.interceptors.request.use(config => {
  const token = localStorage.getItem('token');
  const tokenType = localStorage.getItem('tokenType') || 'Bearer';

  if (token) {
    config.headers.Authorization = `${tokenType} ${token}`;
  }
  return config;
});

// Fetch API示例
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

### 3. 令牌刷新机制

```javascript
// 在令牌过期前设置刷新计时器
function setupTokenRefresh(expiresInSeconds) {
  // 在过期前1分钟刷新
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
      // 如果刷新失败，重定向到登录页
      redirectToLogin();
      return;
    }

    // 更新存储的认证数据
    const authData = result.data;
    localStorage.setItem('token', authData.token);
    localStorage.setItem('tokenExpiry', authData.expire);

    // 设置下一次刷新
    setupTokenRefresh(authData.expires_in);
  } catch (error) {
    console.error('令牌刷新错误:', error);
    redirectToLogin();
  }
}
```

### 4. 处理未授权错误

```javascript
// Axios示例
axios.interceptors.response.use(
  response => response,
  async error => {
    const originalRequest = error.config;

    // 防止无限重试循环
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        // 尝试刷新令牌
        await refreshToken();

        // 使用新令牌重试原始请求
        const token = localStorage.getItem('token');
        const tokenType = localStorage.getItem('tokenType') || 'Bearer';
        originalRequest.headers.Authorization = `${tokenType} ${token}`;

        return axios(originalRequest);
      } catch (refreshError) {
        // 如果刷新失败，重定向到登录页
        redirectToLogin();
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);
```

### 5. 令牌验证辅助函数

```javascript
// 检查令牌是否仍然有效
function isTokenValid() {
  const expiryStr = localStorage.getItem('tokenExpiry');

  if (!expiryStr) {
    return false;
  }

  const expiryTime = new Date(expiryStr).getTime();
  const currentTime = Date.now();

  // 如果未过期，则认为令牌有效
  return expiryTime > currentTime;
}

// 检查令牌是否即将过期（2分钟内）
function isTokenExpiringSoon() {
  const expiryStr = localStorage.getItem('tokenExpiry');

  if (!expiryStr) {
    return true;
  }

  const expiryTime = new Date(expiryStr).getTime();
  const currentTime = Date.now();

  // 检查令牌是否在2分钟内过期
  return expiryTime - currentTime < 2 * 60 * 1000;
}
```

## 安全注意事项

1. **安全存储令牌**：对于单页应用使用localStorage，但对于需要更高安全性的应用，请考虑使用HttpOnly cookies等更安全的替代方案。

2. **在服务器上验证令牌**：始终在服务器端验证令牌，不要仅信任客户端验证。

3. **使用短期令牌**：我们的系统默认使用30分钟的令牌，这有助于限制令牌被盗的影响。

4. **仅使用HTTPS**：在生产环境中始终使用HTTPS以防止令牌被拦截。

5. **退出时清除令牌**：实现适当的退出功能，清除存储的令牌。

```javascript
function logout() {
  localStorage.removeItem('token');
  localStorage.removeItem('tokenExpiry');
  localStorage.removeItem('tokenType');
  redirectToLogin();
}
```

## 其他资源

- [JWT.io](https://jwt.io/) - 用于调试的JWT解码和验证工具
- [OWASP认证备忘单](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
