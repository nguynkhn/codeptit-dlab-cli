package main

import "time"

type RefreshResponse struct {
	AccessToken string `json:"token"`
	SignKey     string `json:"sign_key"`
	ServerNow   int64  `json:"server_now"`
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

func (c *Client) updateToken(resp RefreshResponse) error {
	c.session.SignKey = []byte(resp.SignKey)
	c.session.SkewMs = resp.ServerNow - time.Now().UnixMilli()
	c.session.AccessToken = resp.AccessToken
	return c.store.Save(&c.session)
}

func (c *Client) Refresh() error {
	refreshReq, err := c.NewRequest[struct{}]("POST", "/auth/refresh", nil)
	if err != nil {
		return err
	}

	refreshReq.Header.Set("Authorization", "Bearer "+c.session.RefreshToken)
	refreshResp, err := c.Do[RefreshResponse](refreshReq)
	if err != nil {
		return err
	}

	return c.updateToken(refreshResp)
}

func (c *Client) Login(username, passsword string) error {
	challengeReq, err := c.NewRequest[struct{}]("POST", "/auth/pre-auth-challenge", nil)
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

	c.session.RefreshToken = loginResp.RefreshToken
	return c.updateToken(loginResp.RefreshResponse)
}
