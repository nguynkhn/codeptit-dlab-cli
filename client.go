package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	httpClient *http.Client
	session    Session
	store      Store
}

func NewClient(store Store) (*Client, error) {
	c := Client{httpClient: &http.Client{}, store: store}
	if err := c.store.Load(&c.session); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Client) LoggedIn() bool {
	return c.session.RefreshToken != nil
}

func (c *Client) NewRequest[V any](method, endpoint string, data *V) (*http.Request, error) {
	var body bytes.Buffer
	if data != nil && c.session.SignKey != nil {
		dataBytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		if err := EncryptEnvelope(&body, c.session.SignKey, dataBytes); err != nil {
			return nil, err
		}
	}

	const BASE_URL = "https://code.ptit.edu.vn/api"
	req, err := http.NewRequest(method, BASE_URL+endpoint, &body)
	if err != nil {
		return nil, err
	}

	if body.Len() > 0 {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-PTIT-Enc-Req", "1")
	}

	if c.session.AccessToken != nil {
		req.Header.Set("Authorization", "Bearer "+*c.session.AccessToken)

		if c.session.SignKey != nil {
			signature, err := CreateSignature(
				c.session.SignKey,
				req.Method,
				req.URL.Path,
				body.Bytes(),
				c.session.SkewMs,
			)

			if err != nil {
				return nil, err
			}

			req.Header.Set("X-PTIT-Sign", signature.Sign)
			req.Header.Set("X-PTIT-Sign-Nonce", signature.Nonce)
			req.Header.Set("X-PTIT-Sign-Ts", fmt.Sprint(signature.Timestamp))
		}
	}

	return req, nil
}

func (c *Client) Do[V any](req *http.Request) (V, error) {
	var data V

	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return data, fmt.Errorf("Non-OK HTTP status received: %d", resp.StatusCode)
	}

	if resp.Header.Get("X-PTIT-Enc-Res") == "1" {
		dataBytes, err := DecryptEnvelope(resp.Body, c.session.SignKey)
		if err != nil {
			return data, err
		}

		if err = json.Unmarshal(dataBytes, &data); err != nil {
			return data, err
		}
	} else if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return data, err
	}

	return data, nil
}
