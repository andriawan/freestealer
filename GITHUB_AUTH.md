# GitHub OAuth Setup Guide

## Getting GitHub OAuth Credentials

1. Go to https://github.com/settings/developers
2. Click "OAuth Apps" → "New OAuth App"
3. Fill in the application details:
   - **Application name**: Free Tier API
   - **Homepage URL**: http://localhost:5050
   - **Authorization callback URL**: http://localhost:5050/auth/github/callback
4. Click "Register application"
5. Copy the **Client ID**
6. Click "Generate a new client secret" and copy the **Client Secret**

## Configure Environment Variables

Update your `.env` file with the credentials:

```env
GITHUB_CLIENT_ID=your_actual_client_id
GITHUB_CLIENT_SECRET=your_actual_client_secret
GITHUB_CALLBACK_URL=http://localhost:5050/auth/github/callback
SESSION_SECRET=generate-a-random-secret-key-here
```

## Authentication Flow

### Login Process

1. **Start OAuth**: Navigate to `/auth/github`
   - User is redirected to GitHub for authorization

2. **GitHub Callback**: After user approves, GitHub redirects to `/auth/github/callback`
   - System receives user data from GitHub
   - Creates or updates user in database
   - Creates session for the user
   - Returns user info

3. **Get Current User**: `/auth/me`
   - Returns currently logged-in user info
   - Requires active session

4. **Logout**: `/auth/logout`
   - Destroys user session

### API Endpoints

#### Start GitHub Login
```
GET /auth/github
```
Redirects to GitHub OAuth authorization page.

#### GitHub Callback (automatic)
```
GET /auth/github/callback?code=xxx&state=xxx
```
Response:
```json
{
  "message": "Authentication successful",
  "user": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com",
    "avatar_url": "https://avatars.githubusercontent.com/..."
  }
}
```

#### Get Current User
```
GET /auth/me
```
Response:
```json
{
  "id": 1,
  "username": "johndoe",
  "email": "john@example.com",
  "github_id": "12345678",
  "github_login": "johndoe",
  "avatar_url": "https://avatars.githubusercontent.com/...",
  "created_at": "2026-01-21T10:00:00Z"
}
```

#### Logout
```
GET /auth/logout
```
Response:
```json
{
  "message": "Logged out successfully"
}
```

## Protected Routes

Use the `auth.RequireAuth` middleware to protect routes:

```go
http.HandleFunc("/protected", auth.RequireAuth(yourHandler))
```

This ensures only authenticated users can access the endpoint.

## Testing the Flow

1. Start the server: `go run main.go` or `air`
2. Open browser: `http://localhost:5050/auth/github`
3. Authorize the app on GitHub
4. You'll be redirected back with user info
5. Check session: `http://localhost:5050/auth/me`

## Session Management

- Sessions are stored in cookies
- Session cookie name: `auth-session`
- Sessions persist across server restarts (in-memory storage)
- For production, use a persistent session store (Redis, database, etc.)

## Security Notes

- **Never commit** your `.env` file with real credentials
- Use strong `SESSION_SECRET` in production (32+ random characters)
- For production, enable HTTPS and set secure cookie flags
- Consider adding CSRF protection for production
- Rate limit authentication endpoints to prevent abuse

## User Data Stored

From GitHub OAuth, we store:
- GitHub ID (unique identifier)
- GitHub username
- Email address
- Avatar URL
- Access token (encrypted, not exposed in API)
- Refresh token (encrypted, not exposed in API)

## Integration with Existing Features

Once logged in, the user's ID is available in session:
- Creating tiers automatically links to logged-in user
- Voting requires authentication
- Comments require authentication

Example of getting user from session in handlers:
```go
session, _ := store.Get(r, "auth-session")
userID, ok := session.Values["user_id"].(uint)
```
