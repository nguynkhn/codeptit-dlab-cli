package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const BASE_URL = "https://code.ptit.edu.vn/api"

type Client struct {
	httpClient *http.Client

	accessToken  string
	refreshToken string

	signKey []byte
	skewMs  int64
}

func NewClient() Client {
	return Client{httpClient: &http.Client{}}
}

func (c *Client) NewRequest[V any](method, endpoint string, data *V) (*http.Request, error) {
	var buffer bytes.Buffer
	if data != nil {
		dataBytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		if err := EncryptEnvelope(&buffer, c.signKey, dataBytes); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, BASE_URL+endpoint, &buffer)
	if err != nil {
		return nil, err
	}

	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	if buffer.Len() > 0 {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-PTIT-Enc-Req", "1")
	}

	if c.signKey != nil && c.accessToken != "" {
		signature, err := CreateSignature(c.signKey, req.Method, req.URL.Path, buffer.Bytes(), c.skewMs)
		if err != nil {
			return nil, err
		}

		req.Header.Set("X-PTIT-Sign", signature.Sign)
		req.Header.Set("X-PTIT-Sign-Nonce", signature.Nonce)
		req.Header.Set("X-PTIT-Sign-Ts", fmt.Sprint(signature.Timestamp))
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

	if resp.Header.Get("X-PTIT-Enc-Res") == "1" {
		dataBytes, err := DecryptEnvelope(resp.Body, c.signKey)
		if err != nil {
			return data, err
		}

		if err = json.Unmarshal(dataBytes, &data); err != nil {
			return data, err
		}
	} else if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return data, err
	}

	return data, err
}
