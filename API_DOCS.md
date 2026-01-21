# Free Tier API

A system to track and share information about free tier offerings from hosting platforms like Railway, Koyeb, Vercel, etc. Users can add tiers, vote on them, and leave comments.

## Authentication

**GitHub OAuth Integration** - Users can login with their GitHub account.

See [GITHUB_AUTH.md](GITHUB_AUTH.md) for detailed setup instructions.

### Auth Endpoints

- `GET /auth/github` - Start GitHub OAuth login
- `GET /auth/github/callback` - OAuth callback (automatic)
- `GET /auth/me` - Get current authenticated user
- `GET /auth/logout` - Logout current user

## Database Schema

**Efficient SQLite design with:**
- Indexed fields for fast queries
- Denormalized vote/comment counts for performance
- Soft deletes with GORM
- Composite unique indexes to prevent duplicate votes
- Custom indexes for common query patterns

### Models

#### User
- `id`, `username` (unique), `email` (unique)
- Tracks all tiers, votes, and comments created by the user

#### Tier
- Platform details (Railway, Koyeb, etc.)
- Resource limits (CPU, memory, storage, bandwidth, hours)
- Privacy setting (`is_public`)
- Denormalized counts: `upvote_count`, `downvote_count`, `comment_count`
- Indexed for queries by platform and votes

#### Vote
- User can upvote (+1) or downvote (-1) a tier
- One vote per user per tier (composite unique index)
- Toggle support: clicking same vote removes it
- Automatically updates tier vote counts

#### Comment
- Max 100 characters
- Links user to tier
- Automatically updates tier comment count

## API Endpoints

### Health Check
```
GET /health
```

### Users

**Create User**
```
POST /users
Content-Type: application/json

{
  "username": "john_doe",
  "email": "john@example.com"
}
```

**Get All Users**
```
GET /users
```

### Tiers

**Create Tier**
```
POST /tiers
Content-Type: application/json

{
  "user_id": 1,
  "platform": "Railway",
  "name": "Railway Free Tier",
  "description": "Great for hobby projects",
  "is_public": true,
  "cpu_limit": "0.5 vCPU",
  "memory_limit": "512MB",
  "storage_limit": "1GB",
  "bandwidth_limit": "100GB",
  "monthly_hours": "500 hours",
  "url": "https://railway.app/pricing"
}
```

**Get Tiers (with filters)**
```
GET /tiers?platform=Railway&sort=recent&page=1
GET /tiers?user_id=1  (shows user's private + public tiers)
GET /tiers            (shows only public tiers, sorted by upvotes)

Query params:
- platform: filter by platform name
- user_id: show specific user's tiers (including private)
- sort: "recent" or default (by upvotes)
- page: pagination (20 items per page)
```

**Get Single Tier**
```
GET /tiers/{id}
```

**Update Tier**
```
PUT /tiers/{id}
Content-Type: application/json

{
  "name": "Updated Tier Name",
  "is_public": false
}
```

**Delete Tier**
```
DELETE /tiers/{id}
```

### Votes

**Vote on Tier**
```
POST /votes
Content-Type: application/json

{
  "user_id": 1,
  "tier_id": 5,
  "vote_type": 1    // 1 for upvote, -1 for downvote
}

Behavior:
- First vote: Creates vote
- Same vote again: Removes vote (toggle off)
- Different vote: Changes vote type
- Automatically updates tier vote counts in transaction
```

### Comments

**Create Comment**
```
POST /comments
Content-Type: application/json

{
  "user_id": 1,
  "tier_id": 5,
  "content": "This tier is perfect for small projects!"
}

Max 100 characters
```

**Get Comments for Tier**
```
GET /comments?tier_id=5
```

**Delete Comment**
```
DELETE /comments/{id}
```

## Environment Variables

Create a `.env` file:
```env
PORT=5050
DB_PATH=./freetier.db
```

## Running the Application

**Development (with hot reload):**
```powershell
$env:PATH += ";$env:USERPROFILE\go\bin"
air
```

**Production:**
```powershell
go run main.go
```

**Debug Mode (F5 in VS Code):**
- Set breakpoints in code
- Press F5 to start debugging
- Environment variables loaded from `.env`

## Database Features

### Performance Optimizations
1. **Denormalized Counts**: Vote and comment counts stored on tier for fast reads
2. **Composite Indexes**: `(user_id, tier_id)` for unique vote constraint
3. **Custom Indexes**: `(is_public, upvote_count DESC)` for homepage queries
4. **Platform Index**: Fast filtering by platform
5. **Soft Deletes**: Data preserved but hidden from queries

### Data Integrity
- Transactions for vote operations (count + vote record)
- Unique constraints on username/email
- One vote per user per tier enforced at DB level
- Foreign key relationships maintained by GORM

## Example Usage Flow

1. **Create users**: POST to `/users`
2. **Add tiers**: POST to `/tiers` with platform info
3. **Browse public tiers**: GET `/tiers` (sorted by votes)
4. **Vote**: POST to `/votes` to upvote/downvote
5. **Comment**: POST to `/comments` with max 100 chars
6. **Filter**: GET `/tiers?platform=Railway&sort=recent`

## Schema Efficiency Notes

- **No N+1 queries**: Using GORM Preload for relations
- **Indexed queries**: All list endpoints use indexed fields
- **Pagination**: 20 items per page to limit response size
- **Vote counts cached**: No need to count votes on each request
- **Soft deletes**: Maintains data integrity and allows recovery
