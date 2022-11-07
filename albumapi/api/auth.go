package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/AndrewBoyarsky/albumapi/users"
	u "github.com/AndrewBoyarsky/albumapi/users"
	"github.com/AndrewBoyarsky/common/config"
	ci "github.com/AndrewBoyarsky/common/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {

	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	repo := u.NewUserRepo()
	byUserName := repo.GetByUserName(input.Username)
	if byUserName != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "Already registered with username: " + byUserName.UserName})
		return
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 8)
	user := u.User{
		UserName: input.Username,
		Password: string(hashedPassword),
	}

	id := repo.Save(user)
	c.JSON(http.StatusOK, gin.H{"message": "User saved with id=" + id})

}

func Login(c *gin.Context) {

	var input RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	repo := u.NewUserRepo()

	user := repo.GetByUserName(input.Username)
	if user == nil {
		logrus.New().Infof("Attempt with nonexistent user to login is rejected: %s", input.Username)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Bad credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		logrus.New().Infof("User %s supplied a wrong password", input.Username)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Bad credentials"})
		return
	}
	uuid, _ := uuid.NewRandom()
	duration, _ := time.ParseDuration("24h")
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.RegisteredClaims{
		ID:        uuid.String(),
		Issuer:    "AwesomeProject",
		Subject:   input.Username,
		Audience:  jwt.ClaimStrings{"USER"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	})
	signed, errr := token.SignedString(ci.Config.JwtSecretKey)

	if errr != nil {
		logrus.Fatalf("Unable to sign token with edd25519 key")
	}
	c.JSON(http.StatusOK, gin.H{"message": "User logged in successfullt ", "accessToken": signed})
}

func AuthMiddleware(c *gin.Context) {
	headerContent := c.GetHeader("Authorization")
	if headerContent == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, Error{"No JWT token provided for auth in 'Authorization' header"})
		return
	}
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(strings.TrimPrefix(headerContent, "Bearer "), &claims, func(t *jwt.Token) (interface{}, error) {
		return config.Config.JwtSecretKey.Public(), nil
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, Error{"Bad token"})
		return
	}
	err_valid := token.Claims.Valid()
	if err_valid != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, Error{"Token expired"})
		return
	}
	repo := users.NewUserRepo()
	c.Set("user", repo.GetByUserName(claims.Subject))
}
