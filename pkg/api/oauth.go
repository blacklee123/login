package api

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type OAuth struct {
	logger          *zap.Logger
	private_key     *rsa.PrivateKey
	public_key      *rsa.PublicKey
	clientid        string
	secrets         string
	authorize_url   string
	accesstoken_url string
	user_url        string
	expire          int
}

type AccessToken struct {
	Access_token string `json:"access_token"`
	Token_type   string `json:"token_type"`
	Scope        string `json:"scope"`
}

type GitUser struct {
	Login               string `json:"login"`
	Id                  int    `json:"id"`
	Node_id             string `json:"node_id"`
	Avatar_url          string `json:"avatar_url"`
	Gravatar_id         string `json:"gravatar_id"`
	Url                 string `json:"url"`
	Html_url            string `json:"html_url"`
	Followers_url       string `json:"followers_url"`
	Following_url       string `json:"following_url"`
	Gists_url           string `json:"gists_url"`
	Subscriptions_url   string `json:"subscriptions_url"`
	Organizations_url   string `json:"organizations_url"`
	Repos_url           string `json:"repos_url"`
	Events_url          string `json:"events_url"`
	Received_events_url string `json:"received_events_url"`
	Type                string `json:"type"`
	Site_admin          bool   `json:"site_admin"`
	Name                string `json:"name"`
	Company             string `json:"company"`
	Blog                string `json:"blog"`
	Location            string `json:"location"`
	Email               string `json:"email"`
	Hireable            string `json:"hireable"`
	Bio                 string `json:"bio"`
	Twitter_username    string `json:"twitter_username"`
	Public_repos        int    `json:"public_repos"`
	Public_gists        int    `json:"public_gists"`
	Followers           int    `json:"followers"`
	Following           int    `json:"following"`
	Created_at          string `json:"created_at"`
	Updated_at          string `json:"updated_at"`
}

func NewOAuth(logger *zap.Logger) *OAuth {
	private_key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(viper.GetString("rsa.private_key")))
	if err != nil {
		logger.Panic("parse error")
	}
	public_key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(viper.GetString("rsa.public_key")))
	if err != nil {
		logger.Panic("parse error")
	}
	oauth := &OAuth{
		logger:          logger,
		private_key:     private_key,
		public_key:      public_key,
		clientid:        viper.GetString("LOGIN_CLIENTID"),
		secrets:         viper.GetString("LOGIN_SECRETS"),
		authorize_url:   viper.GetString("github.url.authorize"),
		accesstoken_url: viper.GetString("github.url.accesstoken"),
		user_url:        viper.GetString("github.url.user"),
		expire:          viper.GetInt("token.auth_exp_d"),
	}
	return oauth
}

func (o *OAuth) authorize(c *gin.Context) string {
	next := c.DefaultQuery("next", "/")
	nextUrl := url.URL{
		Scheme:   c.Request.URL.Scheme,
		Host:     c.Request.Host,
		Path:     "/web/login/callback",
		RawQuery: fmt.Sprintf("next=%s", next),
	}
	o.logger.Info("authorize_url", zap.String("authorize_url", nextUrl.String()))
	return fmt.Sprintf(o.authorize_url, o.clientid, url.QueryEscape(nextUrl.String()))
}

func (o *OAuth) Code2Token(code string) string {
	return o.user2Token(o.code2User(code))
}

func (o *OAuth) user2Token(user *GitUser) string {
	claims := make(jwt.MapClaims)
	claims["fullname"] = user.Name
	claims["avatar"] = strings.TrimSpace(user.Avatar_url)
	claims["email"] = user.Email
	claims["userid"] = strconv.Itoa(user.Id)
	claims["salt"] = uuid.NewString()
	claims["exp"] = time.Now().Add(time.Hour * 24 * time.Duration(o.expire)).UnixNano()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ss, _ := token.SignedString(o.private_key)
	return ss
}

func (o *OAuth) code2User(code string) *GitUser {
	client := resty.New()
	access_token := &AccessToken{}
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetResult(&access_token).
		Post(fmt.Sprintf(o.accesstoken_url, o.clientid, o.secrets, code))
	if err != nil || resp.StatusCode() != http.StatusOK {
		fmt.Printf("request failed status = [%v], error = %v", resp.StatusCode(), err)
	}
	gitUser := &GitUser{}
	resp, err = client.R().
		SetHeader("Accept", "application/json").
		SetAuthToken(access_token.Access_token).
		SetAuthScheme("Bearer").
		SetResult(&gitUser).
		Get(o.user_url)
	if err != nil || resp.StatusCode() != http.StatusOK {
		fmt.Printf("request failed status = [%v], error = %v", resp.StatusCode(), err)
	}
	return gitUser
}

func (o *OAuth) SetJWTCookie(c *gin.Context, token, domain string) {
	expiry := time.Now().Add(time.Hour * 24 * 10).UnixNano()
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return o.public_key, nil
	})
	if err != nil {
		o.logger.Panic("token parse error")
	}
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		c.SetCookie("jwt", token, int(expiry), "/", domain, false, true)
		c.SetCookie("fullname", claims["fullname"].(string), int(expiry), "/", domain, true, false)
		c.SetCookie("email", claims["email"].(string), int(expiry), "/", domain, false, false)
		c.SetCookie("avatar", claims["avatar"].(string), int(expiry), "/", domain, false, false)
		c.SetCookie("userid", claims["userid"].(string), int(expiry), "/", domain, false, false)
	} else {
		o.logger.Panic("token parse error")
	}

}

func (o *OAuth) DelJWTCookie(c *gin.Context, domain string) {
	c.SetCookie("jwt", "", 0, "/", domain, true, true)
	c.SetCookie("fullname", "", 0, "/", domain, true, false)
	c.SetCookie("email", "", 0, "/", domain, true, false)
	c.SetCookie("avatar", "", 0, "/", domain, true, false)
	c.SetCookie("userid", "", 0, "/", domain, true, false)
}
