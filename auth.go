package main

import "time"

type ChallengeResponse struct {
	Challenge     string `json:"challenge"`
	EncryptionKey string `json:"enc_key"`
}

type SignKeyResponse struct {
	SignKey   string `json:"sign_key"`
	ServerNow int64  `json:"server_now"`
}

type RefreshResponse struct {
	AccessToken string `json:"token"`
	SignKeyResponse
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
	c.signKey = []byte(resp.SignKey)
	c.skewMs = resp.ServerNow - time.Now().UnixMilli()
}

func (c *Client) updateToken(resp RefreshResponse) {
	c.updateSignKey(resp.SignKeyResponse)
	c.accessToken = resp.AccessToken
}

func (c *Client) Refresh() error {
	if c.accessToken == "" {
		refreshReq, err := c.NewRequest[any]("POST", "/auth/refresh", nil)
		if err != nil {
			return err
		}

		refreshReq.Header.Set("Authorization", "Bearer "+c.refreshToken)
		refreshResp, err := c.Do[RefreshResponse](refreshReq)
		if err != nil {
			return err
		}

		c.updateToken(refreshResp)

	} else if c.signKey == nil {
		signKeyReq, err := c.NewRequest[any]("GET", "/auth/sign-key", nil)
		if err != nil {
			return err
		}

		signKeyResp, err := c.Do[SignKeyResponse](signKeyReq)
		if err != nil {
			return err
		}

		c.updateSignKey(signKeyResp)
	}

	return nil
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

	c.signKey = []byte(challengeResp.EncryptionKey)

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

	c.refreshToken = loginResp.RefreshToken
	c.updateToken(loginResp.RefreshResponse)

	return nil
}
