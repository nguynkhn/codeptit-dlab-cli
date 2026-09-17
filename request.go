package main

import "time"

type ChallengeResponse struct {
	Challenge     string `json:"challenge"`
	EncryptionKey []byte `json:"enc_key"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	SignKey      []byte `json:"sign_key"`
	ServerNow    int64  `json:"server_now"`
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

	c.signKey = challengeResp.EncryptionKey

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

	c.accessToken = loginResp.AccessToken
	c.refreshToken = loginResp.RefreshToken
	c.signKey = loginResp.SignKey
	c.skewMs = loginResp.ServerNow - time.Now().UnixMilli()

	return nil
}

type RefreshResponse struct {
	AccessToken string `json:"token"`
	SignKey     []byte `json:"sign_key"`
	ServerNow   int64  `json:"server_now"`
}

type SignKeyResponse struct {
	SignKey   []byte `json:"sign_key"`
	ServerNow int64  `json:"server_now"`
}

func (c *Client) Refresh() error {
	refreshReq, err := c.NewRequest[any]("POST", "/auth/refresh", nil)
	if err != nil {
		return err
	}

	refreshReq.Header.Set("Authorization", "Bearer "+c.refreshToken)
	refreshResp, err := c.Do[RefreshResponse](refreshReq)
	if err != nil {
		return err
	}

	c.accessToken = refreshResp.AccessToken
	c.signKey = refreshResp.SignKey
	c.skewMs = refreshResp.ServerNow - time.Now().UnixMilli()

	return nil
}
