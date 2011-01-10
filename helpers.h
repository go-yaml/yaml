#ifndef helpers_h
#define helpers_h

#include <yaml.h>

__typeof__(((yaml_event_t *)0)->data.scalar) *event_scalar(yaml_event_t *event);

void set_output_handler(yaml_emitter_t *e);

#endif
