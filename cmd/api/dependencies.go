//go:build tools
// +build tools

package main

import (
_ "github.com/aws/aws-sdk-go-v2/config"
_ "github.com/aws/aws-sdk-go-v2/credentials"
_ "github.com/aws/aws-sdk-go-v2/service/s3"
_ "github.com/gin-gonic/gin"
_ "github.com/golang-jwt/jwt/v5"
_ "github.com/golang-migrate/migrate/v4"
_ "github.com/golang-migrate/migrate/v4/database/postgres"
_ "github.com/golang-migrate/migrate/v4/source/file"
_ "github.com/google/uuid"
_ "github.com/jackc/pgx/v5"
_ "github.com/microcosm-cc/bluemonday"
_ "github.com/stretchr/objx"
_ "github.com/stretchr/testify"
_ "golang.org/x/crypto/bcrypt"
_ "gopkg.in/gomail.v2"
)
