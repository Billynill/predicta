package jira

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) GetAvatarURL(ctx context.Context, accountID string) (string, error) {
	user, err := c.GetUser(ctx, accountID)
	if err != nil {
		return "", err
	}
	if user.AvatarURL == "" {
		return "", fmt.Errorf("jira: no avatar for account %s", accountID)
	}
	return user.AvatarURL, nil
}

func (c *Client) GetUser(ctx context.Context, accountID string) (*AssignableUser, error) {
	q := url.Values{}
	q.Set("accountId", accountID)

	resp, err := c.doRequest(ctx, http.MethodGet, "/rest/api/3/user", q, nil)
	if err != nil {
		return nil, err
	}

	var raw struct {
		AccountID    string            `json:"accountId"`
		DisplayName  string            `json:"displayName"`
		EmailAddress string            `json:"emailAddress"`
		AvatarURLs   map[string]string `json:"avatarUrls"`
	}
	if err := c.decodeJSON(resp, &raw); err != nil {
		return nil, err
	}

	out := mapAssignable(jiraUser{
		AccountID:    raw.AccountID,
		DisplayName:  raw.DisplayName,
		EmailAddress: raw.EmailAddress,
		AvatarURLs:   raw.AvatarURLs,
	})
	return &out, nil
}

func pickAvatarURL(urls map[string]string) string {
	if urls == nil {
		return ""
	}
	for _, size := range []string{"48x48", "32x32", "24x24", "16x16"} {
		if u := urls[size]; u != "" {
			return u
		}
	}
	for _, u := range urls {
		if u != "" {
			return u
		}
	}
	return ""
}
