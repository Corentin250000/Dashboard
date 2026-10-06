package widgets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var ErrFriendListPrivate = errors.New("steam friend list is private")

type PlayerSummary struct {
	SteamID                  string  `json:"steamid"`
	PersonaName              string  `json:"personaname"`
	CommunityVisibilityState int     `json:"communityvisibilitystate"`
	ProfileState             *int    `json:"profilestate"`
	PersonaState             int     `json:"personastate"`
	GameID                   *string `json:"gameid"`
	GameExtraInfo            *string `json:"gameextrainfo"`
	LastLogoff               *int64  `json:"lastlogoff"`
}

type SteamClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewSteamClient(apiKey string, httpClient *http.Client) *SteamClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &SteamClient{apiKey: apiKey, httpClient: httpClient}
}

type friendListResponse struct {
	FriendsList *struct {
		Friends *[]struct {
			SteamID string `json:"steamid"`
		} `json:"friends"`
	} `json:"friendslist"`
}

func (c *SteamClient) GetFriendList(ctx context.Context, steamID string) ([]string, error) {
	query := url.Values{"key": {c.apiKey}, "steamid": {steamID}, "relationship": {"friend"}, "format": {"json"}}
	var response friendListResponse
	if err := c.getJSON(ctx, "https://api.steampowered.com/ISteamUser/GetFriendList/v1/", query, &response); err != nil {
		return nil, err
	}
	if response.FriendsList == nil || response.FriendsList.Friends == nil {
		return nil, ErrFriendListPrivate
	}
	ids := make([]string, 0, len(*response.FriendsList.Friends))
	for _, friend := range *response.FriendsList.Friends {
		if friend.SteamID != "" {
			ids = append(ids, friend.SteamID)
		}
	}
	return ids, nil
}

type playerResponse struct {
	Response struct {
		Players []PlayerSummary `json:"players"`
	} `json:"response"`
}

func (c *SteamClient) GetPlayerSummaries(ctx context.Context, steamIDs []string) ([]PlayerSummary, error) {
	if len(steamIDs) == 0 {
		return []PlayerSummary{}, nil
	}
	chunks := make([][]string, 0, (len(steamIDs)+99)/100)
	for start := 0; start < len(steamIDs); start += 100 {
		end := start + 100
		if end > len(steamIDs) {
			end = len(steamIDs)
		}
		chunks = append(chunks, steamIDs[start:end])
	}
	results := make([][]PlayerSummary, len(chunks))
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup
	var firstErr error
	var errMu sync.Mutex
	for i, chunk := range chunks {
		wg.Add(1)
		go func(index int, ids []string) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()
			query := url.Values{"key": {c.apiKey}, "steamids": {strings.Join(ids, ",")}, "format": {"json"}}
			var response playerResponse
			if err := c.getJSON(ctx, "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/", query, &response); err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				errMu.Unlock()
				return
			}
			results[index] = response.Response.Players
		}(i, chunk)
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	aggregated := make([]PlayerSummary, 0, len(steamIDs))
	for _, result := range results {
		aggregated = append(aggregated, result...)
	}
	return aggregated, nil
}

func (c *SteamClient) getJSON(ctx context.Context, endpoint string, query url.Values, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return fmt.Errorf("create Steam request: %w", err)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("request Steam API: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return ErrFriendListPrivate
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("Steam API returned HTTP %d", response.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(target); err != nil {
		return fmt.Errorf("decode Steam response: %w", err)
	}
	return nil
}
