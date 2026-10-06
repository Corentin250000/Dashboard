package widgets

import (
	"errors"
	"strings"

	"dashboard-backend/internal/auth"
	"dashboard-backend/internal/models"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// minRefreshRateSeconds is the minimum polling interval for a widget instance.
const minRefreshRateSeconds = 60

// WidgetPosition is the zero-based top-left cell of an instance on the 6x6 desktop grid.
type WidgetPosition struct {
	X int `json:"x" example:"2"`
	Y int `json:"y" example:"1"`
}

// WidgetInstanceResponse is the API representation of a dashboard widget instance.
type WidgetInstanceResponse struct {
	ID                 string         `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Type               string         `json:"type" example:"sncf.departures"`
	Size               string         `json:"size" example:"big"`
	Config             map[string]any `json:"config"`
	RefreshRateSeconds int            `json:"refresh_rate_seconds" example:"120"`
	Position           WidgetPosition `json:"position"`
}

// WidgetInstanceListResponse wraps the dashboard instances of the current user.
type WidgetInstanceListResponse struct {
	Widgets []WidgetInstanceResponse `json:"widgets"`
}

// CreateWidgetRequest creates a dashboard widget instance.
type CreateWidgetRequest struct {
	Type               string          `json:"type" validate:"required" example:"sncf.departures"`
	Size               string          `json:"size" validate:"required" example:"big"`
	Config             map[string]any  `json:"config"`
	RefreshRateSeconds int             `json:"refresh_rate_seconds" validate:"required" example:"120"`
	Position           *WidgetPosition `json:"position" validate:"required"`
}

// UpdateWidgetRequest partially updates a dashboard widget instance.
// Only non-nil fields are applied.
type UpdateWidgetRequest struct {
	Size               *string         `json:"size" example:"wide"`
	Config             *map[string]any `json:"config"`
	RefreshRateSeconds *int            `json:"refresh_rate_seconds" example:"180"`
	Position           *WidgetPosition `json:"position"`
}

type WidgetBadRequestResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The request body must contain type, size, refresh_rate_seconds and position."`
}

type WidgetUnauthorizedResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Authentication required."`
}

type WidgetTypeNotFoundResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The requested widget type does not exist."`
}

type WidgetInstanceNotFoundResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The requested widget does not exist."`
}

type WidgetInvalidIDResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The widget ID is invalid."`
}

type WidgetConflictResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"This widget overlaps another widget."`
}

type WidgetUnprocessableResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"This size is not available for this widget."`
}

type WidgetServerErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Unable to create the widget."`
}

type WidgetUnavailableResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"The database is not configured."`
}

// SeedCatalog ensures the static widget catalog used by the instance routes exists.
// It mirrors backend/internal/about.Registry so type strings like "sncf.departures"
// resolve to database rows.
func SeedCatalog(database *gorm.DB) error {
	for _, serviceData := range []models.Service{
		{
			Name:        "spotify",
			DisplayName: "Spotify",
			Description: "Spotify account and playback data",
			AuthType:    "oauth2",
		},
		{
			Name:        "steam",
			DisplayName: "Steam",
			Description: "Steam account and game data",
			AuthType:    "oauth2",
		},
	} {
		var service models.Service
		if err := database.Where("name = ?", serviceData.Name).FirstOrCreate(&service, serviceData).Error; err != nil {
			return err
		}
	}

	service := models.Service{}
	if err := database.Where("name = ?", "sncf").FirstOrCreate(&service, models.Service{
		Name:        "sncf",
		DisplayName: "SNCF",
		Description: "French rail departures",
		AuthType:    "none",
	}).Error; err != nil {
		return err
	}

	widget := models.Widget{}
	if err := database.Where("service_id = ? AND name = ?", service.ID, "departures").FirstOrCreate(&widget, models.Widget{
		ServiceID:   service.ID,
		Name:        "departures",
		DisplayName: "Departures",
		Description: "Display train departures for a station",
	}).Error; err != nil {
		return err
	}

	param := models.WidgetParam{}
	if err := database.Where("widget_id = ? AND name = ?", widget.ID, "station").FirstOrCreate(&param, models.WidgetParam{
		WidgetID:     widget.ID,
		Name:         "station",
		Type:         "string",
		IsRequired:   true,
		DefaultValue: nil,
	}).Error; err != nil {
		return err
	}

	for _, sizeCode := range []string{"small", "wide", "tall", "big"} {
		allowed := models.WidgetAllowedSize{WidgetID: widget.ID, SizeCode: sizeCode}
		if err := database.Where("widget_id = ? AND size_code = ?", widget.ID, sizeCode).FirstOrCreate(&allowed).Error; err != nil {
			return err
		}
	}
	spotifyService := models.Service{}
	if err := database.Where("name = ?", "spotify").First(&spotifyService).Error; err != nil {
		return err
	}
	spotifyWidgets := []struct {
		name, displayName, description string
		params                         []models.WidgetParam
	}{
		{"currently_playing", "Currently playing", "Display the track currently playing on Spotify", []models.WidgetParam{
			{Name: "show_progress", Type: "integer", DefaultValue: stringPointer("1")},
		}},
		{"top_items", "Top items", "Display the user's top Spotify artists or tracks", []models.WidgetParam{
			{Name: "item_type", Type: "string", IsRequired: true},
			{Name: "time_range", Type: "string", IsRequired: true},
			{Name: "limit", Type: "integer", DefaultValue: stringPointer("5")},
		}},
	}
	for _, data := range spotifyWidgets {
		widget := models.Widget{}
		if err := database.Where("service_id = ? AND name = ?", spotifyService.ID, data.name).FirstOrCreate(&widget, models.Widget{
			ServiceID: spotifyService.ID, Name: data.name, DisplayName: data.displayName, Description: data.description,
		}).Error; err != nil {
			return err
		}
		if data.name == "currently_playing" {
			if err := database.Where("widget_id = ? AND name = ?", widget.ID, "layout").Delete(&models.WidgetParam{}).Error; err != nil {
				return err
			}
		}
		for _, paramData := range data.params {
			param := models.WidgetParam{}
			if err := database.Where("widget_id = ? AND name = ?", widget.ID, paramData.Name).FirstOrCreate(&param, models.WidgetParam{
				WidgetID: widget.ID, Name: paramData.Name, Type: paramData.Type, IsRequired: paramData.IsRequired, DefaultValue: paramData.DefaultValue,
			}).Error; err != nil {
				return err
			}
		}
		for _, sizeCode := range []string{"small", "wide", "tall", "big"} {
			allowed := models.WidgetAllowedSize{WidgetID: widget.ID, SizeCode: sizeCode}
			if err := database.Where("widget_id = ? AND size_code = ?", widget.ID, sizeCode).FirstOrCreate(&allowed).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func stringPointer(value string) *string { return &value }

// ListWidgets returns the widget instances of the authenticated user.
// @Summary List dashboard widgets
// @Description Return every widget instance placed on the dashboard of the current user, ordered by creation date. Requires the session_id cookie.
// @Tags widgets
// @Produce json
// @Success 200 {object} WidgetInstanceListResponse "The dashboard widgets of the current user."
// @Failure 401 {object} WidgetUnauthorizedResponse "Missing, expired or invalid session cookie."
// @Failure 500 {object} WidgetServerErrorResponse "Unable to load widgets."
// @Failure 503 {object} WidgetUnavailableResponse "The database is not configured."
// @Router /api/me/widgets [get]
func ListWidgets(c fiber.Ctx) error {
	if db == nil {
		return widgetError(c, fiber.StatusServiceUnavailable, "The database is not configured.")
	}
	user, ok := auth.CurrentUser(c)
	if !ok {
		return widgetError(c, fiber.StatusUnauthorized, "Authentication required.")
	}

	var instances []models.WidgetInstance
	if err := db.Preload("Widget.Service").Where("user_id = ?", user.ID).Order("created_at ASC").Find(&instances).Error; err != nil {
		return widgetError(c, fiber.StatusInternalServerError, "Unable to load widgets.")
	}

	response := WidgetInstanceListResponse{Widgets: make([]WidgetInstanceResponse, 0, len(instances))}
	for _, instance := range instances {
		response.Widgets = append(response.Widgets, toInstanceResponse(instance))
	}
	return c.JSON(response)
}

// CreateWidget adds a widget instance to the dashboard of the authenticated user.
// @Summary Create a dashboard widget
// @Description Create a widget instance. The type must use the "service.widget" format (e.g. "sncf.departures"), the size must be allowed for that widget, the refresh rate must be at least 60 seconds, and the position is the zero-based top-left cell on the 6x6 desktop grid. The new widget must fit inside the grid without overlapping an existing widget.
// @Tags widgets
// @Accept json
// @Produce json
// @Param widget body CreateWidgetRequest true "Widget to create (type, size, config, refresh_rate_seconds, position)"
// @Success 201 {object} WidgetInstanceResponse "The created widget instance."
// @Failure 400 {object} WidgetBadRequestResponse "Malformed JSON body or missing type, size, refresh_rate_seconds or position."
// @Failure 401 {object} WidgetUnauthorizedResponse "Missing, expired or invalid session cookie."
// @Failure 404 {object} WidgetTypeNotFoundResponse "Unknown widget type (unknown service, unknown widget, or malformed type string)."
// @Failure 409 {object} WidgetConflictResponse "The new widget overlaps another widget of the dashboard."
// @Failure 422 {object} WidgetUnprocessableResponse "Size not allowed for this widget, refresh rate below 60 seconds, invalid config parameter, or position outside the 6x6 grid."
// @Failure 500 {object} WidgetServerErrorResponse "Unable to create the widget."
// @Failure 503 {object} WidgetUnavailableResponse "The database is not configured."
// @Router /api/me/widgets [post]
func CreateWidget(c fiber.Ctx) error {
	if db == nil {
		return widgetError(c, fiber.StatusServiceUnavailable, "The database is not configured.")
	}
	user, ok := auth.CurrentUser(c)
	if !ok {
		return widgetError(c, fiber.StatusUnauthorized, "Authentication required.")
	}

	var req CreateWidgetRequest
	if err := c.Bind().Body(&req); err != nil {
		return widgetError(c, fiber.StatusBadRequest, "The request body must contain type, size, refresh_rate_seconds and position.")
	}
	req.Type = strings.TrimSpace(req.Type)
	req.Size = strings.TrimSpace(req.Size)
	if req.Type == "" || req.Size == "" || req.Position == nil {
		return widgetError(c, fiber.StatusBadRequest, "The request body must contain type, size, refresh_rate_seconds and position.")
	}
	if req.Config == nil {
		req.Config = map[string]any{}
	}

	widget, err := resolveWidgetType(req.Type)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return widgetError(c, fiber.StatusNotFound, "The requested widget type does not exist.")
		}
		return widgetError(c, fiber.StatusInternalServerError, "Unable to resolve the widget type.")
	}
	if err := validateSizeAllowed(widget.ID, req.Size); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return widgetError(c, fiber.StatusUnprocessableEntity, "This size is not available for this widget.")
		}
		return widgetError(c, fiber.StatusInternalServerError, "Unable to validate the widget size.")
	}
	if err := validateRefreshRate(req.RefreshRateSeconds); err != nil {
		return widgetError(c, fiber.StatusUnprocessableEntity, err.Error())
	}
	if err := validateConfig(widget.ID, req.Config); err != nil {
		return widgetError(c, fiber.StatusUnprocessableEntity, err.Error())
	}

	candidate := models.WidgetInstance{
		ID:                 uuid.New(),
		UserID:             user.ID,
		WidgetID:           widget.ID,
		Config:             req.Config,
		SizeCode:           req.Size,
		RefreshRateSeconds: req.RefreshRateSeconds,
		PositionX:          req.Position.X,
		PositionY:          req.Position.Y,
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		var existing []models.WidgetInstance
		if err := tx.Where("user_id = ?", user.ID).Find(&existing).Error; err != nil {
			return err
		}
		layout := append(existing, candidate)
		sizes, err := loadSizes(tx)
		if err != nil {
			return err
		}
		if err := models.ValidateDesktopLayout(layout, sizes); err != nil {
			return errLayout(err)
		}
		return tx.Create(&candidate).Error
	}); err != nil {
		if layoutErr, ok := asLayoutError(err); ok {
			return widgetLayoutConflict(c, layoutErr)
		}
		return widgetError(c, fiber.StatusInternalServerError, "Unable to create the widget.")
	}

	candidate.Widget = widget
	return c.Status(fiber.StatusCreated).JSON(toInstanceResponse(candidate))
}

// UpdateWidget partially updates a widget instance owned by the authenticated user.
// @Summary Update a dashboard widget
// @Description Update any subset of size, config, refresh rate or position. Omitted fields are left unchanged. The widget type cannot be changed. After applying the patch, the whole dashboard layout is revalidated: the widget must fit inside the 6x6 desktop grid without overlapping another widget.
// @Tags widgets
// @Accept json
// @Produce json
// @Param id path string true "Widget instance ID (UUID)"
// @Param widget body UpdateWidgetRequest true "Fields to update (only non-null fields are applied)"
// @Success 200 {object} WidgetInstanceResponse "The updated widget instance."
// @Failure 400 {object} WidgetBadRequestResponse "Malformed widget ID, malformed JSON body, or empty patch (no updatable field provided)."
// @Failure 401 {object} WidgetUnauthorizedResponse "Missing, expired or invalid session cookie."
// @Failure 404 {object} WidgetInstanceNotFoundResponse "No widget with this ID on the dashboard of the current user."
// @Failure 409 {object} WidgetConflictResponse "The updated widget would overlap another widget of the dashboard."
// @Failure 422 {object} WidgetUnprocessableResponse "Size not allowed for this widget, refresh rate below 60 seconds, invalid config parameter, or position outside the 6x6 grid."
// @Failure 500 {object} WidgetServerErrorResponse "Unable to update the widget."
// @Failure 503 {object} WidgetUnavailableResponse "The database is not configured."
// @Router /api/me/widgets/{id} [patch]
func UpdateWidget(c fiber.Ctx) error {
	if db == nil {
		return widgetError(c, fiber.StatusServiceUnavailable, "The database is not configured.")
	}
	user, ok := auth.CurrentUser(c)
	if !ok {
		return widgetError(c, fiber.StatusUnauthorized, "Authentication required.")
	}
	instanceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return widgetError(c, fiber.StatusBadRequest, "The widget ID is invalid.")
	}

	var instance models.WidgetInstance
	if err := db.Preload("Widget.Service").Where("id = ? AND user_id = ?", instanceID, user.ID).First(&instance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return widgetError(c, fiber.StatusNotFound, "The requested widget does not exist.")
		}
		return widgetError(c, fiber.StatusInternalServerError, "Unable to load the widget.")
	}

	var req UpdateWidgetRequest
	if err := c.Bind().Body(&req); err != nil {
		return widgetError(c, fiber.StatusBadRequest, "The request body is invalid.")
	}
	if req.Size == nil && req.Config == nil && req.RefreshRateSeconds == nil && req.Position == nil {
		return widgetError(c, fiber.StatusBadRequest, "No updatable field was provided.")
	}

	if req.Size != nil {
		trimmed := strings.TrimSpace(*req.Size)
		if trimmed == "" {
			return widgetError(c, fiber.StatusUnprocessableEntity, "The widget size is invalid.")
		}
		if err := validateSizeAllowed(instance.WidgetID, trimmed); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return widgetError(c, fiber.StatusUnprocessableEntity, "This size is not available for this widget.")
			}
			return widgetError(c, fiber.StatusInternalServerError, "Unable to validate the widget size.")
		}
		instance.SizeCode = trimmed
	}
	if req.RefreshRateSeconds != nil {
		if err := validateRefreshRate(*req.RefreshRateSeconds); err != nil {
			return widgetError(c, fiber.StatusUnprocessableEntity, err.Error())
		}
		instance.RefreshRateSeconds = *req.RefreshRateSeconds
	}
	if req.Config != nil {
		config := *req.Config
		if config == nil {
			config = map[string]any{}
		}
		if err := validateConfig(instance.WidgetID, config); err != nil {
			return widgetError(c, fiber.StatusUnprocessableEntity, err.Error())
		}
		instance.Config = config
	}
	if req.Position != nil {
		instance.PositionX = req.Position.X
		instance.PositionY = req.Position.Y
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		var existing []models.WidgetInstance
		if err := tx.Where("user_id = ?", user.ID).Find(&existing).Error; err != nil {
			return err
		}
		for index, other := range existing {
			if other.ID == instance.ID {
				existing[index] = instance
			}
		}
		sizes, err := loadSizes(tx)
		if err != nil {
			return err
		}
		if err := models.ValidateDesktopLayout(existing, sizes); err != nil {
			return errLayout(err)
		}
		return tx.Save(&instance).Error
	}); err != nil {
		if layoutErr, ok := asLayoutError(err); ok {
			return widgetLayoutConflict(c, layoutErr)
		}
		return widgetError(c, fiber.StatusInternalServerError, "Unable to update the widget.")
	}

	return c.JSON(toInstanceResponse(instance))
}

// DeleteWidget removes a widget instance owned by the authenticated user.
// @Summary Delete a dashboard widget
// @Description Delete a widget instance from the dashboard of the current user. Only the instance is removed; the widget catalog entry and the other widgets are untouched. Returns an empty body.
// @Tags widgets
// @Param id path string true "Widget instance ID (UUID)"
// @Success 204 "The widget was deleted (empty body)."
// @Failure 400 {object} WidgetInvalidIDResponse "The widget ID is not a valid UUID."
// @Failure 401 {object} WidgetUnauthorizedResponse "Missing, expired or invalid session cookie."
// @Failure 404 {object} WidgetInstanceNotFoundResponse "No widget with this ID on the dashboard of the current user."
// @Failure 500 {object} WidgetServerErrorResponse "Unable to delete the widget."
// @Failure 503 {object} WidgetUnavailableResponse "The database is not configured."
// @Router /api/me/widgets/{id} [delete]
func DeleteWidget(c fiber.Ctx) error {
	if db == nil {
		return widgetError(c, fiber.StatusServiceUnavailable, "The database is not configured.")
	}
	user, ok := auth.CurrentUser(c)
	if !ok {
		return widgetError(c, fiber.StatusUnauthorized, "Authentication required.")
	}
	instanceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return widgetError(c, fiber.StatusBadRequest, "The widget ID is invalid.")
	}

	result := db.Where("id = ? AND user_id = ?", instanceID, user.ID).Delete(&models.WidgetInstance{})
	if result.Error != nil {
		return widgetError(c, fiber.StatusInternalServerError, "Unable to delete the widget.")
	}
	if result.RowsAffected == 0 {
		return widgetError(c, fiber.StatusNotFound, "The requested widget does not exist.")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func toInstanceResponse(instance models.WidgetInstance) WidgetInstanceResponse {
	typeName := instance.Widget.Name
	if instance.Widget.Service.Name != "" {
		typeName = instance.Widget.Service.Name + "." + instance.Widget.Name
	}
	config := instance.Config
	if config == nil {
		config = map[string]any{}
	}
	return WidgetInstanceResponse{
		ID:                 instance.ID.String(),
		Type:               typeName,
		Size:               instance.SizeCode,
		Config:             config,
		RefreshRateSeconds: instance.RefreshRateSeconds,
		Position:           WidgetPosition{X: instance.PositionX, Y: instance.PositionY},
	}
}

func resolveWidgetType(typeName string) (models.Widget, error) {
	parts := strings.Split(typeName, ".")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return models.Widget{}, gorm.ErrRecordNotFound
	}
	var service models.Service
	if err := db.Where("name = ?", strings.TrimSpace(parts[0])).First(&service).Error; err != nil {
		return models.Widget{}, err
	}
	var widget models.Widget
	if err := db.Preload("Service").Where("service_id = ? AND name = ?", service.ID, strings.TrimSpace(parts[1])).First(&widget).Error; err != nil {
		return models.Widget{}, err
	}
	return widget, nil
}

func validateSizeAllowed(widgetID uuid.UUID, sizeCode string) error {
	var allowed models.WidgetAllowedSize
	return db.Where("widget_id = ? AND size_code = ?", widgetID, sizeCode).First(&allowed).Error
}

func validateRefreshRate(seconds int) error {
	if seconds < minRefreshRateSeconds {
		return errors.New("The refresh rate must be at least 60 seconds.")
	}
	return nil
}

func validateConfig(widgetID uuid.UUID, config map[string]any) error {
	var params []models.WidgetParam
	if err := db.Where("widget_id = ?", widgetID).Find(&params).Error; err != nil {
		return err
	}
	for _, param := range params {
		value, present := config[param.Name]
		if !present || value == nil {
			if param.IsRequired {
				return errors.New("Missing required parameter: " + param.Name + ".")
			}
			continue
		}
		switch param.Type {
		case "string":
			text, ok := value.(string)
			if !ok || (param.IsRequired && strings.TrimSpace(text) == "") {
				return errors.New("Invalid string parameter: " + param.Name + ".")
			}
		case "integer":
			if !isJSONInteger(value) {
				return errors.New("Invalid integer parameter: " + param.Name + ".")
			}
		default:
			return errors.New("Unsupported parameter type for: " + param.Name + ".")
		}
	}
	return nil
}

func isJSONInteger(value any) bool {
	switch number := value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	case float32:
		return float32(int(number)) == number
	case float64:
		return float64(int(number)) == number
	default:
		return false
	}
}

func loadSizes(tx *gorm.DB) (map[string]models.WidgetSize, error) {
	var sizes []models.WidgetSize
	if err := tx.Find(&sizes).Error; err != nil {
		return nil, err
	}
	out := make(map[string]models.WidgetSize, len(sizes))
	for _, size := range sizes {
		out[size.Code] = size
	}
	return out, nil
}

type layoutValidationError struct {
	err error
}

func (e *layoutValidationError) Error() string { return e.err.Error() }

func errLayout(err error) *layoutValidationError {
	return &layoutValidationError{err: err}
}

func asLayoutError(err error) (error, bool) {
	var layoutErr *layoutValidationError
	if errors.As(err, &layoutErr) {
		return layoutErr.err, true
	}
	return nil, false
}

func widgetLayoutConflict(c fiber.Ctx, err error) error {
	message := err.Error()
	if strings.Contains(message, "overlap") {
		return widgetError(c, fiber.StatusConflict, "This widget overlaps another widget.")
	}
	if strings.Contains(message, "outside") || strings.Contains(message, "negative") || strings.Contains(message, "unknown widget size") {
		return widgetError(c, fiber.StatusUnprocessableEntity, "This widget is outside the desktop grid.")
	}
	return widgetError(c, fiber.StatusUnprocessableEntity, "The widget layout is invalid.")
}

func widgetError(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"status": "error", "message": message})
}
