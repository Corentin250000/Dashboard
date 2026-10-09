package widgets

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

const chessMonthlyGamesURL = "https://api.chess.com/pub/player/%s/games/%d/%02d"

type chessMonthlyGamesPayload struct {
	Games []chessGame `json:"games"`
}

type chessGame struct {
	URL        string           `json:"url"`
	PGN        string           `json:"pgn"`
	EndTime    int64            `json:"end_time"`
	Rated      bool             `json:"rated"`
	Accuracies *chessAccuracies `json:"accuracies"`
	TimeClass  string           `json:"time_class"`
	White      chessGamePlayer  `json:"white"`
	Black      chessGamePlayer  `json:"black"`
	ECOURL     string           `json:"eco"`
}

type chessAccuracies struct {
	White float64 `json:"white"`
	Black float64 `json:"black"`
}

type chessGamePlayer struct {
	Rating   int    `json:"rating"`
	Username string `json:"username"`
	Result   string `json:"result"`
}

type ChessMonthlyStatsResponse struct {
	Username        string                   `json:"username" example:"ChessPlayer42"`
	Year            int                      `json:"year" example:"2026"`
	Month           int                      `json:"month" example:"10"`
	TimeClass       string                   `json:"time_class,omitempty" example:"rapid"`
	Rating          *int                     `json:"rating,omitempty" example:"1350"`
	BestRating      *int                     `json:"best_rating,omitempty" example:"1421"`
	TotalGames      int                      `json:"total_games" example:"12"`
	Wins            int                      `json:"wins" example:"7"`
	Losses          int                      `json:"losses" example:"4"`
	Draws           int                      `json:"draws" example:"1"`
	AverageAccuracy *float64                 `json:"average_accuracy,omitempty" example:"78.42"`
	Openings        []ChessOpeningStats      `json:"openings"`
	Games           []ChessMonthlyGameDetail `json:"games"`
}

type ChessOpeningStats struct {
	Name  string `json:"name" example:"Caro-Kann Defense"`
	Games int    `json:"games" example:"3"`
}

type ChessMonthlyGameDetail struct {
	Color       string   `json:"color" example:"black"`
	Accuracy    *float64 `json:"accuracy,omitempty" example:"84.22"`
	Date        string   `json:"date" example:"2026-10-09T11:03:27Z"`
	Result      string   `json:"result" example:"win"`
	Termination string   `json:"termination,omitempty" example:"checkmate"`
	TimeClass   string   `json:"time_class" example:"rapid"`
	URL         string   `json:"url" example:"https://www.chess.com/game/live/123456789"`
}

// GetChessMonthlyStats returns a player's monthly Chess.com games and computed
// performance statistics. The route itself is public: the username identifies
// the public Chess.com profile being queried.
// @Summary Chess.com monthly statistics
// @Description Public route. Return a player's monthly global result counts, average accuracy, the three most frequent openings, and essential details for each game. The API returns every game present in Chess.com's monthly archive response and does not truncate the list.
// @Tags Chess widgets
// @Produce json
// @Param username query string true "Chess.com username" example(ChessPlayer42)
// @Param year query int true "Archive year" minimum(2000) maximum(2100) example(2026)
// @Param month query int true "Archive month" minimum(1) maximum(12) example(10)
// @Param time_class query string false "Optional time class filter" Enums(rapid,blitz,bullet) example(rapid)
// @Success 200 {object} ChessMonthlyStatsResponse "Monthly Chess.com statistics and games. When time_class is provided, counters, accuracy, openings and games are filtered by that cadence; rating and best_rating are also returned."
// @Failure 400 {object} ChessStatsErrorResponse "A parameter is missing or invalid."
// @Failure 404 {object} ChessStatsErrorResponse "The Chess.com player does not exist."
// @Failure 502 {object} ChessStatsErrorResponse "Chess.com is unavailable or returned an invalid response."
// @Router /api/chess/monthly-stats [get]
func GetChessMonthlyStats(c fiber.Ctx) error {
	username := strings.TrimSpace(c.Query("username"))
	if username == "" {
		return chessStatsError(c, fiber.StatusBadRequest, "The username query parameter is required.")
	}
	year, month, err := parseChessArchiveDate(c.Query("year"), c.Query("month"))
	if err != nil {
		return chessStatsError(c, fiber.StatusBadRequest, err.Error())
	}

	timeClass := strings.ToLower(strings.TrimSpace(c.Query("time_class")))
	if timeClass != "" && timeClass != "rapid" && timeClass != "blitz" && timeClass != "bullet" {
		return chessStatsError(c, fiber.StatusBadRequest, "time_class must be rapid, blitz or bullet.")
	}

	games, statusCode, err := fetchChessMonthlyGames(c.Context(), username, year, month)
	if err != nil {
		if statusCode == http.StatusNotFound {
			return chessStatsError(c, fiber.StatusNotFound, "The Chess.com player could not be found.")
		}
		return chessStatsError(c, fiber.StatusBadGateway, "Chess.com is temporarily unavailable.")
	}
	if timeClass != "" {
		filtered := make([]chessGame, 0, len(games))
		for _, game := range games {
			if game.TimeClass == timeClass {
				filtered = append(filtered, game)
			}
		}
		games = filtered
	}

	response := buildChessMonthlyStats(username, year, month, games)
	response.TimeClass = timeClass
	if timeClass != "" {
		statsPayload, statsStatus, statsErr := fetchChessStats(c.Context(), username)
		if statsErr != nil {
			if statsStatus == http.StatusNotFound {
				return chessStatsError(c, fiber.StatusNotFound, "The Chess.com player could not be found.")
			}
			return chessStatsError(c, fiber.StatusBadGateway, "Chess.com is temporarily unavailable.")
		}
		stats := selectChessTimeClass(statsPayload, timeClass)
		response.Rating = intPointer(stats.Last.Rating)
		response.BestRating = intPointer(stats.Best.Rating)
	}
	return c.JSON(response)
}

func fetchChessMonthlyGames(ctx context.Context, username string, year, month int) ([]chessGame, int, error) {
	requestURL := fmt.Sprintf(chessMonthlyGamesURL, url.PathEscape(username), year, month)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, 0, err
	}
	request.Header.Set("User-Agent", "dashboard-chess-widget/1.0")
	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(request)
	if err != nil {
		return nil, 0, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, response.StatusCode, fmt.Errorf("Chess.com returned HTTP %d", response.StatusCode)
	}
	var payload chessMonthlyGamesPayload
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, response.StatusCode, err
	}
	return payload.Games, response.StatusCode, nil
}

func parseChessArchiveDate(rawYear, rawMonth string) (int, int, error) {
	year, err := strconv.Atoi(rawYear)
	if err != nil || year < 2000 || year > 2100 {
		return 0, 0, fmt.Errorf("year must be an integer between 2000 and 2100")
	}
	month, err := strconv.Atoi(rawMonth)
	if err != nil || month < 1 || month > 12 {
		return 0, 0, fmt.Errorf("month must be an integer between 1 and 12")
	}
	return year, month, nil
}

func buildChessMonthlyStats(username string, year, month int, games []chessGame) ChessMonthlyStatsResponse {
	response := ChessMonthlyStatsResponse{
		Username: username,
		Year:     year,
		Month:    month,
		Openings: []ChessOpeningStats{},
		Games:    []ChessMonthlyGameDetail{},
	}
	var accuracySum float64
	var accuracyCount int
	openingStats := make(map[string]*ChessOpeningStats)

	for _, game := range games {
		color, player, opponent := chessPlayerSide(game, username)
		if player == nil || opponent == nil {
			continue
		}
		result := normalizeChessResult(player.Result)
		updateChessRecord(&response.Wins, &response.Losses, &response.Draws, result)
		response.TotalGames++

		if accuracy := chessPlayerAccuracy(game.Accuracies, color); accuracy != nil {
			accuracySum += *accuracy
			accuracyCount++
		}

		eco := chessPGNTag(game.PGN, "ECO")
		openingKey := eco + "|" + game.ECOURL
		opening := openingStats[openingKey]
		if opening == nil {
			opening = &ChessOpeningStats{Name: chessOpeningName(game.ECOURL)}
			openingStats[openingKey] = opening
		}
		opening.Games++

		accuracy := chessPlayerAccuracy(game.Accuracies, color)
		response.Games = append(response.Games, ChessMonthlyGameDetail{
			Color:       color,
			Accuracy:    accuracy,
			Date:        time.Unix(game.EndTime, 0).UTC().Format(time.RFC3339),
			Result:      result,
			Termination: chessTerminationCondition(chessPGNTag(game.PGN, "Termination")),
			TimeClass:   game.TimeClass,
			URL:         game.URL,
		})
	}
	if accuracyCount > 0 {
		value := math.Round(accuracySum/float64(accuracyCount)*100) / 100
		response.AverageAccuracy = &value
	}
	for _, opening := range openingStats {
		response.Openings = append(response.Openings, *opening)
	}
	sort.Slice(response.Openings, func(i, j int) bool {
		if response.Openings[i].Games == response.Openings[j].Games {
			return response.Openings[i].Name < response.Openings[j].Name
		}
		return response.Openings[i].Games > response.Openings[j].Games
	})
	if len(response.Openings) > 3 {
		response.Openings = response.Openings[:3]
	}
	sort.Slice(response.Games, func(i, j int) bool { return response.Games[i].Date > response.Games[j].Date })
	return response
}

func buildChessRecord(username string, games []chessGame, timeClass string) ChessMonthRecordResponse {
	var record ChessMonthRecordResponse
	for _, game := range games {
		if timeClass != "" && game.TimeClass != timeClass {
			continue
		}
		_, player, opponent := chessPlayerSide(game, username)
		if player == nil || opponent == nil {
			continue
		}
		updateChessRecord(&record.Wins, &record.Losses, &record.Draws, normalizeChessResult(player.Result))
		record.Games++
	}
	record.WinRate = chessWinRate(record.Wins, record.Games)
	return record
}

func intPointer(value int) *int { return &value }

func chessPlayerSide(game chessGame, username string) (string, *chessGamePlayer, *chessGamePlayer) {
	if strings.EqualFold(game.White.Username, username) {
		return "white", &game.White, &game.Black
	}
	return "black", &game.Black, &game.White
}

func normalizeChessResult(result string) string {
	switch result {
	case "win":
		return "win"
	case "checkmated", "resigned", "timeout":
		return "loss"
	default:
		return "draw"
	}
}

func updateChessRecord(wins, losses, draws *int, result string) {
	switch result {
	case "win":
		*wins++
	case "loss":
		*losses++
	default:
		*draws++
	}
}

func chessPlayerAccuracy(accuracies *chessAccuracies, color string) *float64 {
	if accuracies == nil {
		return nil
	}
	value := accuracies.Black
	if color == "white" {
		value = accuracies.White
	}
	return &value
}

func chessTerminationCondition(termination string) string {
	lower := strings.ToLower(termination)
	switch {
	case strings.Contains(lower, "checkmate"):
		return "checkmate"
	case strings.Contains(lower, "time"):
		return "time"
	case strings.Contains(lower, "resign"):
		return "resignation"
	case strings.Contains(lower, "stalemate"):
		return "stalemate"
	case strings.Contains(lower, "draw") || strings.Contains(lower, "agreement"):
		return "draw"
	default:
		return termination
	}
}

var chessPGNTagPattern = regexp.MustCompile(`(?m)^\[([^ ]+) "([^"]*)"\]`)

func chessPGNTag(pgn, name string) string {
	for _, match := range chessPGNTagPattern.FindAllStringSubmatch(pgn, -1) {
		if match[1] == name {
			return match[2]
		}
	}
	return ""
}

func chessOpeningName(openingURL string) string {
	if openingURL == "" {
		return "Unknown opening"
	}
	parts := strings.Split(strings.TrimSuffix(openingURL, "/"), "/")
	name := parts[len(parts)-1]
	name = strings.ReplaceAll(name, "-", " ")
	return name
}
