package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

type SignKeyResponse struct {
	SignKey   string `json:"sign_key"`
	ServerNow int64  `json:"server_now"`
}

type RefreshResponse struct {
	AccessToken string `json:"token"`
	SignKeyResponse
}

type ChallengeResponse struct {
	Challenge     string `json:"challenge"`
	EncryptionKey string `json:"enc_key"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	RefreshToken string `json:"refresh_token"`
	RefreshResponse
}

func (c *Client) updateSignKey(resp SignKeyResponse) {
	c.session.SignKey = []byte(resp.SignKey)
	c.session.SkewMs = resp.ServerNow - time.Now().UnixMilli()
}

func (c *Client) updateToken(resp RefreshResponse) error {
	c.session.AccessToken = &resp.AccessToken
	c.updateSignKey(resp.SignKeyResponse)
	return c.store.Save(&c.session)
}

func (c *Client) SignKey() error {
	signKeyReq, err := c.NewRequest[any]("GET", "/auth/sign-key", nil)
	if err != nil {
		return err
	}

	signKeyResp, err := c.Do[SignKeyResponse](signKeyReq)
	if err != nil {
		return err
	}

	c.updateSignKey(signKeyResp)
	return nil
}

func (c *Client) Refresh() error {
	refreshReq, err := c.NewRequest[any]("POST", "/auth/refresh", nil)
	if err != nil {
		return err
	}

	refreshReq.Header.Set("Authorization", "Bearer "+*c.session.RefreshToken)
	refreshResp, err := c.Do[RefreshResponse](refreshReq)
	if err != nil {
		return err
	}

	return c.updateToken(refreshResp)
}

func (c *Client) Login(username, passsword string) error {
	challengeReq, err := c.NewRequest[any]("POST", "/auth/pre-auth-challenge", nil)
	if err != nil {
		return err
	}

	challengeResp, err := c.Do[ChallengeResponse](challengeReq)
	if err != nil {
		return err
	}

	c.session.SignKey = []byte(challengeResp.EncryptionKey)

	loginReq, err := c.NewRequest("POST", "/auth/login", &LoginRequest{
		Username: username, Password: passsword,
	})
	if err != nil {
		return err
	}

	loginReq.Header.Set("X-PTIT-Pow", challengeResp.Challenge+":0")
	loginResp, err := c.Do[LoginResponse](loginReq)
	if err != nil {
		return err
	}

	c.session.RefreshToken = &loginResp.RefreshToken
	return c.updateToken(loginResp.RefreshResponse)
}

func LoginCommand(c *Client) {
	reader := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter your username: ")
	reader.Scan()
	username := reader.Text()

	fmt.Print("Enter your password: ")
	reader.Scan()
	password := reader.Text()

	if err := c.Login(username, password); err != nil {
		fmt.Println("Login failed:", err)
		return
	}

	fmt.Println("Logged in successfully")
}
