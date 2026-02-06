# GQLGo

GraphQL сервис на Go с PostgreSQL через Docker Compose.

**Быстрый старт**
1. Создать `.env`. 

```env
USE_POSTGRES=true
DSN=postgres://admin:admin@db:5432/app?sslmode=disable
ADDR=0.0.0.0:8080
```
2. Запусти сервисы:

```bash
make up
```

3. Открой Playground:
   - `http://localhost:8080/` (встроенный)
   - `http://localhost:8080/query` (основной endpoint)

**Переменные окружения**

Пример `.env`:

```env
USE_POSTGRES=true
DSN=postgres://admin:admin@db:5432/app?sslmode=disable
ADDR=0.0.0.0:8080
```

**Полезные команды**

```bash
docker compose up -d
docker compose up --build
docker compose logs -f
docker compose down
docker compose down -v
```

**API (GraphQL)**

Схема:
- `Query`
  - `GetPosts(first: Int, after: String): PostConnection!`
  - `GetPost(id: ID!): Post`
  - `GetUsers(first: Int, after: String): UserConnection!`
  - `GetUser(id: ID!): User`
- `Mutation`
  - `createPost(input: CreatePostInput!): Post!`
  - `setCommentsEnabled(postId: ID!, enabled: Boolean!): Post!`
  - `addComment(input: AddCommentInput!): Comment!`
- `Subscription`
  - `commentAdded(postId: ID!): Comment!`

Пагинация:
- `first` — размер страницы
- `after` — курсор из `pageInfo.endCursor`
- `order` — `NEWEST` или `OLDEST` (для комментариев)

Примеры запросов:

```graphql
query ListPosts {
  GetPosts(first: 10) {
    totalCount
    pageInfo { hasNextPage endCursor }
    edges {
      cursor
      node { id title commentsEnabled author { id username } }
    }
  }
}
```

```graphql
query GetPostWithComments($id: ID!) {
  GetPost(id: $id) {
    id
    title
    body
    commentsEnabled
    comments(first: 20, order: NEWEST) {
      totalCount
      pageInfo { hasNextPage endCursor }
      edges {
        node {
          id
          body
          parentId
          depth
          childrenCount
          author { id username }
        }
      }
    }
  }
}
```

```graphql
mutation CreatePost {
  createPost(input: {
    authorId: "USER_ID"
    title: "Hello"
    body: "World"
    commentsEnabled: true
  }) {
    id
    title
  }
}
```

```graphql
mutation AddComment {
  addComment(input: {
    postId: "POST_ID"
    authorId: "USER_ID"
    parentId: "PARENT_COMMENT_ID"
    body: "Text"
  }) {
    id
    postId
    parentId
  }
}
```

```graphql
mutation ToggleComments {
  setCommentsEnabled(postId: "POST_ID", enabled: false) {
    id
    commentsEnabled
  }
}
```

Подписка на новые комментарии:

```graphql
subscription OnCommentAdded($postId: ID!) {
  commentAdded(postId: $postId) {
    id
    postId
    body
    parentId
    author { id username }
  }
}
```
