package widgets

import (
	"context"
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
	Username    string                   `json:"username" example:"ChessPlayer42"`
	TimeClass   string                   `json:"time_class" example:"rapid"`
	Rating      int                      `json:"rating" example:"1350"`
	BestRating  int                      `json:"best_rating" example:"1421"`
	UpdatedAt   int64                    `json:"updated_at" example:"1791543807"`
	Record      ChessRecordResponse      `json:"record"`
	MonthRecord ChessMonthRecordResponse `json:"month_record"`
}

type ChessRecordResponse struct {
	Wins    int     `json:"wins" example:"915"`
	Losses  int     `json:"losses" example:"831"`
	Draws   int     `json:"draws" example:"94"`
	Games   int     `json:"games" example:"1840"`
	WinRate float64 `json:"win_rate" example:"49.73"`
}

type ChessMonthRecordResponse struct {
	Wins    int     `json:"wins" example:"7"`
	Losses  int     `json:"losses" example:"4"`
	Draws   int     `json:"draws" example:"1"`
	Games   int     `json:"games" example:"12"`
	WinRate float64 `json:"win_rate" example:"58.33"`
}

type ChessStatsErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The Chess.com player could not be found."`
}

// GetChessStats returns the global rating and game record for a Chess.com
// player and one of the supported live time classes.
// @Summary Chess.com player statistics
// @Description Public route. Return the current and best Chess.com rating, global results, and the current calendar month's results in UTC for rapid, blitz or bullet.
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

	payload, statusCode, err := fetchChessStats(c.Context(), username)
	if err != nil {
		if statusCode == http.StatusNotFound {
			return chessStatsError(c, fiber.StatusNotFound, "The Chess.com player could not be found.")
		}
		return chessStatsError(c, fiber.StatusBadGateway, "Chess.com is temporarily unavailable.")
	}

	stats := selectChessTimeClass(payload, timeClass)
	games := stats.Record.Win + stats.Record.Loss + stats.Record.Draw
	winRate := chessWinRate(stats.Record.Win, games)

	now := time.Now().UTC()
	monthGames, monthStatus, monthErr := fetchChessMonthlyGames(c.Context(), username, now.Year(), int(now.Month()))
	if monthErr != nil {
		if monthStatus == http.StatusNotFound {
			return chessStatsError(c, fiber.StatusNotFound, "The Chess.com player could not be found.")
		}
		return chessStatsError(c, fiber.StatusBadGateway, "Chess.com is temporarily unavailable.")
	}
	monthRecord := buildChessRecord(username, monthGames, timeClass)

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
		MonthRecord: monthRecord,
	})
}

func fetchChessStats(ctx context.Context, username string) (chessStatsResponse, int, error) {
	requestURL := fmt.Sprintf(chessStatsURL, url.PathEscape(username))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return chessStatsResponse{}, 0, err
	}
	request.Header.Set("User-Agent", "dashboard-chess-widget/1.0")

	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	if err != nil {
		return chessStatsResponse{}, 0, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return chessStatsResponse{}, response.StatusCode, fmt.Errorf("Chess.com returned HTTP %d", response.StatusCode)
	}

	var payload chessStatsResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return chessStatsResponse{}, response.StatusCode, err
	}
	return payload, response.StatusCode, nil
}

func chessWinRate(wins, games int) float64 {
	if games == 0 {
		return 0
	}
	return math.Round(float64(wins)/float64(games)*10000) / 100
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
