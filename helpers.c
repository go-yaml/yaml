#include <yaml.h>


__typeof__(((yaml_event_t *)0)->data.scalar) * // Sadness.
event_scalar(yaml_event_t *event)
{
	return &event->data.scalar;
}
