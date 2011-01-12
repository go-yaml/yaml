#ifndef helpers_h
#define helpers_h

#define CGO_LDFLAGS "-lm -lpthread"
#define CGO_CFLAGS "-I. -DHAVE_CONFIG_H=1"

#include <yaml.h>

#define DECLARE_YUNION_FUNC(name) \
    __typeof__(((yaml_event_t *)0)->data.name) *\
    event_##name(yaml_event_t *event);

DECLARE_YUNION_FUNC(scalar)
DECLARE_YUNION_FUNC(alias)
DECLARE_YUNION_FUNC(mapping_start)
DECLARE_YUNION_FUNC(sequence_start)

void set_output_handler(yaml_emitter_t *e);

#endif
