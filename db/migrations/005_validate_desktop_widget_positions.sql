-- Desktop dashboard coordinates are zero-based on a 6x6 grid.
ALTER TABLE widget_instances
    DROP CONSTRAINT IF EXISTS widget_instances_position_x_check,
    DROP CONSTRAINT IF EXISTS widget_instances_position_y_check;

ALTER TABLE widget_instances
    ADD CONSTRAINT widget_instances_position_x_check
        CHECK (position_x >= 0 AND position_x < 6),
    ADD CONSTRAINT widget_instances_position_y_check
        CHECK (position_y >= 0 AND position_y < 6);
