// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gitea

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

// BasicAuthEncode generate base64 of basic auth head
func BasicAuthEncode(user, pass string) string {
	return base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
}

// AccessToken represents an API access token.
// swagger:response AccessToken
type AccessToken struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Token          string `json:"token"`
	HashedToken    string `json:"hashed_token"`
	TokenLastEight string `json:"token_last_eight"`
}

// AccessTokenList represents a list of API access token.
// swagger:response AccessTokenList
type AccessTokenList []*AccessToken

// ListAccessTokens lista all the access tokens of user
func (c *Client) ListAccessTokens(user, pass string) ([]*AccessToken, error) {
	tokens := make([]*AccessToken, 0, 10)
	return tokens, c.getParsedResponse("GET", fmt.Sprintf("/users/%s/tokens", user),
		http.Header{"Authorization": []string{"Basic " + BasicAuthEncode(user, pass)}}, nil, &tokens)
}

// CreateAccessTokenOption options when create access token
// swagger:parameters userCreateToken
type CreateAccessTokenOption struct {
	Name string `json:"name" binding:"Required"`
}

// CreateAccessToken create one access token with options
func (c *Client) CreateAccessToken(user, pass string, opt CreateAccessTokenOption) (*AccessToken, error) {
	t := new(AccessToken)
	return t, c.getParsedResponse("POST", fmt.Sprintf("/users/%s/tokens", user),
		http.Header{
			"content-type":  []string{"application/json"},
			"Authorization": []string{"Basic " + BasicAuthEncode(user, pass)}},
		opt, t)
}

// DeleteAccessToken delete token with key id
func (c *Client) DeleteAccessToken(user string, keyID int64) error {
	_, err := c.getResponse("DELETE", fmt.Sprintf("/user/%s/tokens/%d", user, keyID), nil, nil)
	return err
}
