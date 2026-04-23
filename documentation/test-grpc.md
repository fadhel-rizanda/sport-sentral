```go
grpcurl -plaintext -d '{
  "resource": "user",
  "action": "delete",
  "description": "Can delete users"
}' localhost:50051 rbac.v1.RBACService/CreatePermission
```

```go
grpcurl -plaintext -d '{
  "name": "admin",
  "description": "Administrator"
}' localhost:50051 rbac.v1.RBACService/CreateRole
```


```go
grpcurl -plaintext -d '{
    "email": "test@example.com",
    "username": "testuser2",
    "full_name": "Test User 2",
    "password": "password123"
}' localhost:50051 user.v1.UserService/CreateUser
```

```go
grpcurl -plaintext -d '{
  "email": "test@example.com",
  "password": "password123"
}' localhost:50052 auth.v1.AuthService/Login
```

```go
grpcurl -plaintext -d '{
  "token": "token_dari_email"
}' localhost:50051 user.v1.UserService/VerifyAccount
```

```go
grpcurl -plaintext -d '{
  "email": "test@example.com",
  "password": "password123"
}' localhost:50052 auth.v1.AuthService/Login
```

```go
grpcurl -plaintext -d '{
  "email": "test@example.com"
}' localhost:50051 user.v1.UserService/ForgotPassword
```

```go
grpcurl -plaintext -d '{
  "token": "token_dari_email",
  "password": "newpassword123"
}' localhost:50051 user.v1.UserService/ResetPassword
```

```go

```

