package widgets

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"dashboard-backend/internal/auth"
	"dashboard-backend/internal/models"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var steamClient *SteamClient

const refreshInterval = 2 * time.Minute

func SetClient(client *SteamClient) { steamClient = client }

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
	if db == nil || steamClient == nil {
		return errors.New("Steam friends service is not configured")
	}
	friendIDs, err := steamClient.GetFriendList(ctx, ownerSteamID)
	if err != nil {
		return err
	}
	summaries, err := steamClient.GetPlayerSummaries(ctx, friendIDs)
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

type SteamFriendsErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"La liste d'amis Steam est privée."`
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
// @Failure 400 {object} SteamFriendsErrorResponse "Malformed JSON, empty pseudo, pseudo longer than 64 Unicode characters, or no Steam account connected."
// @Failure 401 {object} SteamFriendsErrorResponse "The session_id cookie is missing, expired or invalid."
// @Failure 403 {object} SteamFriendsErrorResponse "The Steam friend list is private."
// @Failure 502 {object} SteamFriendsErrorResponse "Steam is unreachable, timed out, returned an unexpected status, or returned an invalid response."
// @Failure 503 {object} SteamFriendsErrorResponse "The Steam friends service is not configured."
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
			return steamError(c, fiber.StatusBadRequest, "A Steam account is not connected.")
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
		if err := refreshFound(c.Context(), ownerID, ids); err != nil {
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
		return steamError(c, fiber.StatusForbidden, "La liste d'amis Steam est privée.")
	}
	return steamError(c, fiber.StatusBadGateway, "Steam est momentanément indisponible.")
}
func steamError(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"status": "error", "message": message})
}
