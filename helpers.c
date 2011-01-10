#include "_cgo_export.h"
#include "helpers.h"


__typeof__(((yaml_event_t *)0)->data.scalar) * // Sadness.
event_scalar(yaml_event_t *event)
{
    return &event->data.scalar;
}

void set_output_handler(yaml_emitter_t *e)
{
    yaml_emitter_set_output(e, outputHandler, (void *)e);
}
