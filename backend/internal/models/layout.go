package models

import (
	"fmt"
)

const (
	DesktopGridColumns = 6
	DesktopGridRows    = 6
)

// ValidateDesktopLayout validates the complete desktop dashboard layout.
// Sizes are expressed in grid cells and must already be restricted to the
// widget by widget_allowed_sizes before this function is called.
func ValidateDesktopLayout(instances []WidgetInstance, sizes map[string]WidgetSize) error {
	for index, instance := range instances {
		size, ok := sizes[instance.SizeCode]
		if !ok {
			return fmt.Errorf("unknown widget size %q for instance %s", instance.SizeCode, instance.ID)
		}

		if instance.PositionX < 0 || instance.PositionY < 0 {
			return fmt.Errorf("widget instance %s has a negative position", instance.ID)
		}
		if instance.PositionX+size.Width > DesktopGridColumns ||
			instance.PositionY+size.Height > DesktopGridRows {
			return fmt.Errorf("widget instance %s is outside the desktop grid", instance.ID)
		}

		for _, other := range instances[index+1:] {
			otherSize, ok := sizes[other.SizeCode]
			if !ok {
				return fmt.Errorf("unknown widget size %q for instance %s", other.SizeCode, other.ID)
			}
			if layoutsOverlap(instance, size, other, otherSize) {
				return fmt.Errorf("widget instances %s and %s overlap", instance.ID, other.ID)
			}
		}
	}
	return nil
}

func layoutsOverlap(first WidgetInstance, firstSize WidgetSize, second WidgetInstance, secondSize WidgetSize) bool {
	return first.PositionX < second.PositionX+secondSize.Width &&
		first.PositionX+firstSize.Width > second.PositionX &&
		first.PositionY < second.PositionY+secondSize.Height &&
		first.PositionY+firstSize.Height > second.PositionY
}
