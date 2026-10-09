package widgets

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

const chessStatsURL = "https://api.chess.com/pub/player/%s/stats"

type chessStatsResponse struct {
	Rapid  chessTimeClassStats `json:"chess_rapid"`
	Blitz  chessTimeClassStats `json:"chess_blitz"`
	Bullet chessTimeClassStats `json:"chess_bullet"`
}

type chessTimeClassStats struct {
	Last   chessLastRating `json:"last"`
	Best   chessBestRating `json:"best"`
	Record chessRecord     `json:"record"`
}

type chessLastRating struct {
	Rating int   `json:"rating"`
	Date   int64 `json:"date"`
}

type chessBestRating struct {
	Rating int   `json:"rating"`
	Date   int64 `json:"date"`
}

type chessRecord struct {
	Win  int `json:"win"`
	Loss int `json:"loss"`
	Draw int `json:"draw"`
}

type ChessStatsResponse struct {
	Username   string              `json:"username" example:"ChessPlayer42"`
	TimeClass  string              `json:"time_class" example:"rapid"`
	Rating     int                 `json:"rating" example:"1350"`
	BestRating int                 `json:"best_rating" example:"1421"`
	UpdatedAt  int64               `json:"updated_at" example:"1791543807"`
	Record     ChessRecordResponse `json:"record"`
}

type ChessRecordResponse struct {
	Wins    int     `json:"wins" example:"915"`
	Losses  int     `json:"losses" example:"831"`
	Draws   int     `json:"draws" example:"94"`
	Games   int     `json:"games" example:"1840"`
	WinRate float64 `json:"win_rate" example:"49.73"`
}

type ChessStatsErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The Chess.com player could not be found."`
}

// GetChessStats returns the global rating and game record for a Chess.com
// player and one of the supported live time classes.
// @Summary Chess.com player statistics
// @Description Public route. Return the current and best Chess.com rating, plus the global wins, losses, draws and win rate for rapid, blitz or bullet.
// @Tags Chess widgets
// @Produce json
// @Param username query string true "Chess.com username" example(ChessPlayer42)
// @Param time_class query string true "Chess.com time class" Enums(rapid,blitz,bullet) example(rapid)
// @Success 200 {object} ChessStatsResponse "Global Chess.com player statistics."
// @Failure 400 {object} ChessStatsErrorResponse "A username is missing or time_class is not rapid, blitz or bullet."
// @Failure 404 {object} ChessStatsErrorResponse "The Chess.com player does not exist."
// @Failure 502 {object} ChessStatsErrorResponse "Chess.com is unavailable or returned an invalid response."
// @Router /api/chess/stats [get]
func GetChessStats(c fiber.Ctx) error {
	username := strings.TrimSpace(c.Query("username"))
	if username == "" {
		return chessStatsError(c, fiber.StatusBadRequest, "The username query parameter is required.")
	}

	timeClass := strings.ToLower(strings.TrimSpace(c.Query("time_class")))
	if timeClass != "rapid" && timeClass != "blitz" && timeClass != "bullet" {
		return chessStatsError(c, fiber.StatusBadRequest, "time_class must be rapid, blitz or bullet.")
	}

	requestURL := fmt.Sprintf(chessStatsURL, url.PathEscape(username))
	request, err := http.NewRequestWithContext(c.Context(), http.MethodGet, requestURL, nil)
	if err != nil {
		return chessStatsError(c, fiber.StatusBadGateway, "Unable to create the Chess.com request.")
	}
	request.Header.Set("User-Agent", "dashboard-chess-widget/1.0")

	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return chessStatsError(c, fiber.StatusBadGateway, "Chess.com is temporarily unavailable.")
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return chessStatsError(c, fiber.StatusNotFound, "The Chess.com player could not be found.")
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return chessStatsError(c, fiber.StatusBadGateway, "Chess.com returned an invalid response.")
	}

	var payload chessStatsResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return chessStatsError(c, fiber.StatusBadGateway, "Chess.com returned invalid statistics.")
	}

	stats := selectChessTimeClass(payload, timeClass)
	games := stats.Record.Win + stats.Record.Loss + stats.Record.Draw
	winRate := 0.0
	if games > 0 {
		winRate = math.Round(float64(stats.Record.Win)/float64(games)*10000) / 100
	}

	return c.JSON(ChessStatsResponse{
		Username:   username,
		TimeClass:  timeClass,
		Rating:     stats.Last.Rating,
		BestRating: stats.Best.Rating,
		UpdatedAt:  stats.Last.Date,
		Record: ChessRecordResponse{
			Wins:    stats.Record.Win,
			Losses:  stats.Record.Loss,
			Draws:   stats.Record.Draw,
			Games:   games,
			WinRate: winRate,
		},
	})
}

func selectChessTimeClass(payload chessStatsResponse, timeClass string) chessTimeClassStats {
	switch timeClass {
	case "blitz":
		return payload.Blitz
	case "bullet":
		return payload.Bullet
	default:
		return payload.Rapid
	}
}

func chessStatsError(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(ChessStatsErrorResponse{Status: "error", Message: message})
}
