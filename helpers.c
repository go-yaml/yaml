#include "_cgo_export.h"
#include "helpers.h"

#define DEFINE_YUNION_FUNC(name) \
    __typeof__(((yaml_event_t *)0)->data.name) * \
    event_##name(yaml_event_t *event) { \
        return &event->data.name; \
    }

DEFINE_YUNION_FUNC(scalar)
DEFINE_YUNION_FUNC(alias)
DEFINE_YUNION_FUNC(mapping_start)
DEFINE_YUNION_FUNC(sequence_start)

void set_output_handler(yaml_emitter_t *e)
{
    yaml_emitter_set_output(e, (yaml_write_handler_t*)outputHandler, (void *)e);
}
