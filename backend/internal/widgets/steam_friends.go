package widgets

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"dashboard-backend/internal/auth"
	"dashboard-backend/internal/models"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var steamClient *SteamClient

// These groups prevent several widgets for the same user from refreshing the
// same Steam data simultaneously. They do not serialize normal searches: only
// calls sharing the same owner key are coalesced while the refresh is running.
var steamFriendsSyncGroup singleflight.Group
var steamFriendsRefreshGroup singleflight.Group
var steamFriendsSummaryGroup singleflight.Group

type steamSummaryCacheEntry struct {
	summaries []PlayerSummary
	expiresAt time.Time
}

var steamSummaryCache = struct {
	sync.RWMutex
	entries map[string]steamSummaryCacheEntry
}{entries: make(map[string]steamSummaryCacheEntry)}

const steamSummaryCacheDuration = time.Minute

const refreshInterval = 2 * time.Minute

func SetClient(client *SteamClient) {
	steamClient = client
	steamSummaryCache.Lock()
	steamSummaryCache.entries = make(map[string]steamSummaryCacheEntry)
	steamSummaryCache.Unlock()
}

// NormalizeName keeps the displayed persona untouched while making searches stable.
func NormalizeName(value string) string {
	value = strings.Map(func(r rune) rune {
		if (r >= '\u202a' && r <= '\u202e') || (r >= '\u2066' && r <= '\u2069') || r == '\u200e' || r == '\u200f' {
			return -1
		}
		return r
	}, value)
	return strings.TrimSpace(strings.ToLower(value))
}

func normalizeName(value string) string { return NormalizeName(value) }

func SyncFriends(ctx context.Context, ownerSteamID string) error {
	_, err, _ := steamFriendsSyncGroup.Do(ownerSteamID, func() (interface{}, error) {
		return nil, syncFriends(ctx, ownerSteamID)
	})
	return err
}

func syncFriends(ctx context.Context, ownerSteamID string) error {
	if db == nil || steamClient == nil {
		return errors.New("Steam friends service is not configured")
	}
	friendIDs, err := steamClient.GetFriendList(ctx, ownerSteamID)
	if err != nil {
		return err
	}
	summaries, err := cachedPlayerSummaries(ctx, ownerSteamID, friendIDs)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	rows := make([]models.SteamFriend, 0, len(summaries))
	for _, summary := range summaries {
		rows = append(rows, steamFriendFromSummary(ownerSteamID, summary, now))
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(rows) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "owner_steam_id"}, {Name: "friend_steam_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"persona_name", "search_name", "profile_visibility", "profile_configured", "persona_state", "current_game_id", "current_game_name", "last_logoff_at", "synced_at"}),
			}).Create(&rows).Error; err != nil {
				return err
			}
		}
		if len(friendIDs) == 0 {
			return tx.Where("owner_steam_id = ?", ownerSteamID).Delete(&models.SteamFriend{}).Error
		}
		return tx.Where("owner_steam_id = ? AND friend_steam_id NOT IN ?", ownerSteamID, friendIDs).Delete(&models.SteamFriend{}).Error
	})
}

func steamFriendFromSummary(owner string, summary PlayerSummary, syncedAt time.Time) models.SteamFriend {
	configured := summary.ProfileState != nil && *summary.ProfileState == 1
	var lastLogoff *time.Time
	if summary.LastLogoff != nil {
		value := time.Unix(*summary.LastLogoff, 0).UTC()
		lastLogoff = &value
	}
	return models.SteamFriend{OwnerSteamID: owner, FriendSteamID: summary.SteamID, PersonaName: summary.PersonaName, SearchName: NormalizeName(summary.PersonaName), ProfileVisibility: summary.CommunityVisibilityState, ProfileConfigured: configured, PersonaState: summary.PersonaState, CurrentGameID: summary.GameID, CurrentGameName: summary.GameExtraInfo, LastLogoffAt: lastLogoff, SyncedAt: syncedAt}
}

type SteamFriendStatusRequest struct {
	Pseudo string `json:"pseudo" example:"chrystoval"`
}

type SteamOnlineFriendsResponse struct {
	TotalOnline int                 `json:"totalOnline" example:"14"`
	Returned    int                 `json:"returned" example:"5"`
	Friends     []SteamOnlineFriend `json:"friends"`
}

type SteamOnlineFriend struct {
	SteamID     string            `json:"steamId" example:"76561198978982665"`
	Pseudo      string            `json:"pseudo" example:"chrystoval♛"`
	Status      string            `json:"status" example:"En ligne"`
	CurrentGame *SteamCurrentGame `json:"currentGame"`
}
type SteamFriendStatusResponse struct {
	Results []SteamFriendStatus `json:"results"`
}
type SteamFriendStatus struct {
	SteamID           string            `json:"steamId" example:"76561198978982665"`
	Pseudo            string            `json:"pseudo" example:"chrystoval\u2067\u2067♛"`
	ProfileVisibility string            `json:"profileVisibility" example:"public"`
	ProfileConfigured bool              `json:"profileConfigured" example:"true"`
	Status            string            `json:"status" example:"En ligne"`
	CurrentGame       *SteamCurrentGame `json:"currentGame"`
	LastLogoff        *time.Time        `json:"lastLogoff" example:"2026-10-05T19:51:31Z"`
}
type SteamCurrentGame struct {
	ID   string  `json:"id" example:"730"`
	Name *string `json:"name,omitempty" example:"Counter-Strike 2"`
}

type SteamFriendsBadRequestResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The pseudo must contain between 1 and 64 characters."`
}

type SteamFriendsUnauthorizedResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Authentication required."`
}

type SteamFriendsPrivateResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The Steam friend list is private."`
}

type SteamFriendsBadGatewayResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Steam is temporarily unavailable."`
}

type SteamFriendsServerResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Unable to search Steam friends."`
}

type SteamFriendsUnavailableResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The Steam friends service is not configured."`
}

// Status searches the authenticated user's synchronized Steam friends by pseudo
// and returns their translated online status and currently visible game.
// @Summary Search a Steam friend's status
// @Description Search the connected user's Steam friends with a normalized pseudo. The local friend list is synchronized after Steam connection and refreshed when its data is older than two minutes. Directional Unicode control characters are removed only for searching; the original Steam pseudo is returned unchanged.
// @Tags Steam
// @Accept json
// @Produce json
// @Param request body SteamFriendStatusRequest true "Pseudo to search (1 to 64 Unicode characters)" example({"pseudo":"chrystoval"})
// @Success 200 {object} SteamFriendStatusResponse "Matching friends, limited to 10 results. Exact normalized matches are returned first. Example: {\"results\":[{\"steamId\":\"76561198978982665\",\"pseudo\":\"chrystoval\\u2067\\u2067♛\",\"profileVisibility\":\"public\",\"profileConfigured\":true,\"status\":\"En ligne\",\"currentGame\":{\"id\":\"730\",\"name\":\"Counter-Strike 2\"},\"lastLogoff\":\"2026-10-05T19:51:31Z\"}]}"
// @Failure 400 {object} SteamFriendsBadRequestResponse "Malformed JSON, empty pseudo, or pseudo longer than 64 Unicode characters."
// @Failure 401 {object} SteamFriendsUnauthorizedResponse "The session_id cookie is missing, expired or invalid."
// @Failure 403 {object} SteamFriendsPrivateResponse "The Steam account is not connected or the Steam friend list is private."
// @Failure 500 {object} SteamFriendsServerResponse "The local Steam friends data could not be searched or refreshed."
// @Failure 502 {object} SteamFriendsBadGatewayResponse "Steam is unreachable, timed out, returned an unexpected status, or returned an invalid response."
// @Failure 503 {object} SteamFriendsUnavailableResponse "The Steam friends service is not configured."
// @Router /api/steam/friends/status [post]
func Status(c fiber.Ctx) error {
	if db == nil || steamClient == nil {
		return steamError(c, fiber.StatusServiceUnavailable, "Steam friends service is not configured.")
	}
	user, ok := auth.CurrentUser(c)
	if !ok {
		return steamError(c, fiber.StatusUnauthorized, "Authentication required.")
	}
	var request SteamFriendStatusRequest
	if err := c.Bind().Body(&request); err != nil {
		return steamError(c, fiber.StatusBadRequest, "The request body must contain a valid pseudo.")
	}
	searchName := NormalizeName(request.Pseudo)
	if searchName == "" || utf8.RuneCountInString(searchName) > 64 {
		return steamError(c, fiber.StatusBadRequest, "The pseudo must contain between 1 and 64 characters.")
	}
	ownerID, err := steamIDForUser(user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return steamError(c, fiber.StatusForbidden, "A Steam account is not connected.")
		}
		return steamError(c, fiber.StatusInternalServerError, "Unable to load the Steam connection.")
	}
	pattern := "%" + escapeLike(searchName) + "%"
	var friends []models.SteamFriend
	if err := db.WithContext(c.Context()).Where("owner_steam_id = ? AND search_name ILIKE ? ESCAPE '\\'", ownerID, pattern).Order("persona_name ASC").Limit(10).Find(&friends).Error; err != nil {
		return steamError(c, fiber.StatusInternalServerError, "Unable to search Steam friends.")
	}
	if len(friends) == 0 {
		if err := SyncFriends(c.Context(), ownerID); err != nil {
			return steamSyncError(c, err)
		}
		if err := db.WithContext(c.Context()).Where("owner_steam_id = ? AND search_name ILIKE ? ESCAPE '\\'", ownerID, pattern).Order("persona_name ASC").Limit(10).Find(&friends).Error; err != nil {
			return steamError(c, fiber.StatusInternalServerError, "Unable to search Steam friends.")
		}
	}
	stale := false
	for _, friend := range friends {
		if time.Since(friend.SyncedAt) > refreshInterval {
			stale = true
			break
		}
	}
	if stale {
		ids := make([]string, 0, len(friends))
		for _, friend := range friends {
			ids = append(ids, friend.FriendSteamID)
		}
		if err := refreshFoundOnce(c.Context(), ownerID, ids); err != nil {
			return steamSyncError(c, err)
		}
		if err := db.WithContext(c.Context()).Where("owner_steam_id = ? AND search_name ILIKE ? ESCAPE '\\'", ownerID, pattern).Order("persona_name ASC").Limit(10).Find(&friends).Error; err != nil {
			return steamError(c, fiber.StatusInternalServerError, "Unable to search Steam friends.")
		}
	}
	// Exact matches are ordered first without changing the user's original name.
	exact := NormalizeName(request.Pseudo)
	for i := range friends {
		if friends[i].SearchName == exact {
			friends = append([]models.SteamFriend{friends[i]}, append(friends[:i], friends[i+1:]...)...)
			break
		}
	}
	response := SteamFriendStatusResponse{Results: make([]SteamFriendStatus, 0, len(friends))}
	for _, friend := range friends {
		response.Results = append(response.Results, toStatus(friend))
	}
	return c.JSON(response)
}

// Online returns the authenticated user's friends currently visible as online.
// Steam reports private or "appears offline" profiles as offline, so they cannot
// be detected by this endpoint.
// @Summary List Steam friends online
// @Description Return fresh online statuses for the authenticated user's Steam friends. A friend is online when Steam reports personastate from 1 to 6; private profiles and profiles appearing offline report personastate 0 and cannot be detected. Results are cached for 60 seconds per Steam account.
// @Tags Steam
// @Produce json
// @Param limit query int false "Maximum number of friends to return" minimum(1) maximum(10) default(5) example(5)
// @Param inGameOnly query bool false "Only return friends currently playing a game" default(false) example(false)
// @Success 200 {object} SteamOnlineFriendsResponse "Example: {\"totalOnline\":14,\"returned\":5,\"friends\":[{\"steamId\":\"76561198978982665\",\"pseudo\":\"chrystoval♛\",\"status\":\"En ligne\",\"currentGame\":{\"id\":\"730\",\"name\":\"Counter-Strike 2\"}}]}"
// @Failure 400 {object} SteamFriendsBadRequestResponse "Invalid limit (must be an integer from 1 to 10) or invalid inGameOnly value."
// @Failure 401 {object} SteamFriendsUnauthorizedResponse "The session_id cookie is missing, expired or invalid."
// @Failure 403 {object} SteamFriendsPrivateResponse "The Steam account is not connected or the Steam friend list is private."
// @Failure 500 {object} SteamFriendsServerResponse "The local Steam friends data could not be loaded or updated."
// @Failure 502 {object} SteamFriendsBadGatewayResponse "Steam is unreachable, timed out, returned an unexpected status, or returned an invalid response."
// @Failure 503 {object} SteamFriendsUnavailableResponse "The Steam friends service is not configured."
// @Router /api/steam/friends/online [get]
func Online(c fiber.Ctx) error {
	if db == nil || steamClient == nil {
		return steamError(c, fiber.StatusServiceUnavailable, "Steam friends service is not configured.")
	}
	user, ok := auth.CurrentUser(c)
	if !ok {
		return steamError(c, fiber.StatusUnauthorized, "Authentication required.")
	}
	limit, err := parseOnlineLimit(c.Query("limit"))
	if err != nil {
		return steamError(c, fiber.StatusBadRequest, err.Error())
	}
	inGameOnly, err := parseOnlineBool(c.Query("inGameOnly"))
	if err != nil {
		return steamError(c, fiber.StatusBadRequest, err.Error())
	}
	ownerID, err := steamIDForUser(user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return steamError(c, fiber.StatusForbidden, "A Steam account is not connected.")
		}
		return steamError(c, fiber.StatusInternalServerError, "Unable to load the Steam connection.")
	}
	var friends []models.SteamFriend
	if err := db.WithContext(c.Context()).Where("owner_steam_id = ?", ownerID).Find(&friends).Error; err != nil {
		return steamError(c, fiber.StatusInternalServerError, "Unable to load Steam friends.")
	}
	if len(friends) == 0 {
		if err := SyncFriends(c.Context(), ownerID); err != nil {
			return steamSyncError(c, err)
		}
		if err := db.WithContext(c.Context()).Where("owner_steam_id = ?", ownerID).Find(&friends).Error; err != nil {
			return steamError(c, fiber.StatusInternalServerError, "Unable to load Steam friends.")
		}
	}
	ids := make([]string, 0, len(friends))
	for _, friend := range friends {
		ids = append(ids, friend.FriendSteamID)
	}
	summaries, err := cachedPlayerSummaries(c.Context(), ownerID, ids)
	if err != nil {
		return steamSyncError(c, err)
	}
	if err := persistSummaryUpdates(c.Context(), ownerID, summaries); err != nil {
		return steamError(c, fiber.StatusInternalServerError, "Unable to update Steam friends.")
	}
	byID := make(map[string]models.SteamFriend, len(friends))
	for _, friend := range friends {
		byID[friend.FriendSteamID] = friend
	}
	for _, summary := range summaries {
		byID[summary.SteamID] = steamFriendFromSummary(ownerID, summary, time.Now().UTC())
	}
	filtered := make([]models.SteamFriend, 0, len(summaries))
	for _, summary := range summaries {
		if summary.PersonaState < 1 || summary.PersonaState > 6 || (inGameOnly && summary.GameID == nil) {
			continue
		}
		filtered = append(filtered, byID[summary.SteamID])
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		iInGame, jInGame := filtered[i].CurrentGameID != nil, filtered[j].CurrentGameID != nil
		if iInGame != jInGame {
			return iInGame
		}
		return NormalizeName(filtered[i].PersonaName) < NormalizeName(filtered[j].PersonaName)
	})
	response := SteamOnlineFriendsResponse{TotalOnline: len(filtered), Friends: make([]SteamOnlineFriend, 0, limit)}
	if limit > len(filtered) {
		limit = len(filtered)
	}
	for _, friend := range filtered[:limit] {
		response.Friends = append(response.Friends, toOnlineFriend(friend))
	}
	response.Returned = len(response.Friends)
	return c.JSON(response)
}

func parseOnlineLimit(value string) (int, error) {
	if value == "" {
		return 5, nil
	}
	limit, err := strconv.Atoi(value)
	if err != nil || limit < 1 || limit > 10 {
		return 0, errors.New("limit must be an integer between 1 and 10")
	}
	return limit, nil
}

func parseOnlineBool(value string) (bool, error) {
	if value == "" {
		return false, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, errors.New("inGameOnly must be a boolean")
	}
	return parsed, nil
}

func toOnlineFriend(friend models.SteamFriend) SteamOnlineFriend {
	result := SteamOnlineFriend{SteamID: friend.FriendSteamID, Pseudo: friend.PersonaName, Status: translatedStatus(friend)}
	if friend.CurrentGameID != nil {
		result.CurrentGame = &SteamCurrentGame{ID: *friend.CurrentGameID, Name: friend.CurrentGameName}
	}
	return result
}

func cachedPlayerSummaries(ctx context.Context, owner string, ids []string) ([]PlayerSummary, error) {
	now := time.Now()
	steamSummaryCache.RLock()
	entry, found := steamSummaryCache.entries[owner]
	steamSummaryCache.RUnlock()
	if found && now.Before(entry.expiresAt) {
		return entry.summaries, nil
	}
	value, err, _ := steamFriendsSummaryGroup.Do(owner, func() (interface{}, error) {
		steamSummaryCache.RLock()
		entry, found := steamSummaryCache.entries[owner]
		steamSummaryCache.RUnlock()
		if found && time.Now().Before(entry.expiresAt) {
			return entry.summaries, nil
		}
		summaries, err := steamClient.GetPlayerSummaries(ctx, ids)
		if err != nil {
			return nil, err
		}
		steamSummaryCache.Lock()
		steamSummaryCache.entries[owner] = steamSummaryCacheEntry{summaries: summaries, expiresAt: time.Now().Add(steamSummaryCacheDuration)}
		steamSummaryCache.Unlock()
		return summaries, nil
	})
	if err != nil {
		return nil, err
	}
	return value.([]PlayerSummary), nil
}

func persistSummaryUpdates(ctx context.Context, owner string, summaries []PlayerSummary) error {
	now := time.Now().UTC()
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, summary := range summaries {
			row := steamFriendFromSummary(owner, summary, now)
			if err := tx.Model(&models.SteamFriend{}).
				Where("owner_steam_id = ? AND friend_steam_id = ?", owner, row.FriendSteamID).
				Select("persona_name", "search_name", "profile_visibility", "profile_configured", "persona_state", "current_game_id", "current_game_name", "last_logoff_at", "synced_at").
				Updates(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// refreshFoundOnce coalesces concurrent refreshes for the same Steam account.
// The first request supplies the IDs; subsequent requests wait for that same
// refresh instead of issuing another GetPlayerSummaries call.
func refreshFoundOnce(ctx context.Context, owner string, ids []string) error {
	_, err, _ := steamFriendsRefreshGroup.Do(owner, func() (interface{}, error) {
		return nil, refreshFound(ctx, owner, ids)
	})
	return err
}

func refreshFound(ctx context.Context, owner string, ids []string) error {
	summaries, err := steamClient.GetPlayerSummaries(ctx, ids)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, summary := range summaries {
			row := steamFriendFromSummary(owner, summary, now)
			if err := tx.Model(&models.SteamFriend{}).
				Where("owner_steam_id = ? AND friend_steam_id = ?", owner, row.FriendSteamID).
				Select("persona_name", "search_name", "profile_visibility", "profile_configured", "persona_state", "current_game_id", "current_game_name", "last_logoff_at", "synced_at").
				Updates(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func steamIDForUser(userID interface{}) (string, error) {
	var service models.UserService
	return serviceExternalSteam(userID, &service)
}
func serviceExternalSteam(userID interface{}, service *models.UserService) (string, error) {
	if err := db.Where("user_id = ? AND service_id = (SELECT id FROM services WHERE name = 'steam')", userID).First(service).Error; err != nil {
		return "", err
	}
	if service.ExternalUsername == nil || *service.ExternalUsername == "" {
		return "", gorm.ErrRecordNotFound
	}
	return *service.ExternalUsername, nil
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "%", `\%`)
	return strings.ReplaceAll(value, "_", `\_`)
}

func toStatus(friend models.SteamFriend) SteamFriendStatus {
	result := SteamFriendStatus{SteamID: friend.FriendSteamID, Pseudo: friend.PersonaName, ProfileVisibility: "private", ProfileConfigured: friend.ProfileConfigured, Status: translatedStatus(friend), LastLogoff: friend.LastLogoffAt}
	if friend.ProfileVisibility == 3 {
		result.ProfileVisibility = "public"
	}
	if friend.CurrentGameID != nil {
		result.CurrentGame = &SteamCurrentGame{ID: *friend.CurrentGameID, Name: friend.CurrentGameName}
	}
	return result
}
func translatedStatus(friend models.SteamFriend) string {
	if friend.ProfileVisibility != 3 && friend.PersonaState == 0 {
		return "Statut masqué"
	}
	if status, ok := map[int]string{0: "Hors ligne", 1: "En ligne", 2: "Occupé", 3: "Absent", 4: "Sieste", 5: "Souhaite échanger", 6: "Souhaite jouer"}[friend.PersonaState]; ok {
		return status
	}
	return "Inconnu"
}
func steamSyncError(c fiber.Ctx, err error) error {
	if errors.Is(err, ErrFriendListPrivate) {
		return steamError(c, fiber.StatusForbidden, "The Steam friend list is private.")
	}
	return steamError(c, fiber.StatusBadGateway, "Steam is temporarily unavailable.")
}
func steamError(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"status": "error", "message": message})
}
