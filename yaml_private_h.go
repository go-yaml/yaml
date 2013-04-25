package goyaml

const (
	// The size of the input raw buffer.
	input_raw_buffer_size = 16384

	// The size of the input buffer.
	// It should be possible to decode the whole raw buffer.
	input_buffer_size = input_raw_buffer_size * 3

	// The size of the output buffer.
	output_buffer_size = 16384

	// The size of the output raw buffer.
	// It should be possible to encode the whole output buffer.
	output_raw_buffer_size = (output_buffer_size*2 + 2)

	// The size of other stacks and queues.
	initial_stack_size  = 16
	initial_queue_size  = 16
	initial_string_size = 16
)

// Check if the character at the specified position is an alphabetical
// character, a digit, '_', or '-'.
func is_alpha(b []byte, i int) bool {
	return b[i] >= '0' && b[i] <= '9' || b[i] >= 'A' && b[i] <= 'Z' || b[i] >= 'a' && b[i] <= 'z' || b[i] == '_' || b[i] == '-'
}

// Check if the character at the specified position is a digit.
func is_digit(b []byte, i int) bool {
	return b[i] >= '0' && b[i] <= '9'
}

// Get the value of a digit.
func as_digit(b []byte, i int) int {
	return int(b[i]) - '0'
}

// Check if the character at the specified position is a hex-digit.
func is_hex(b []byte, i int) bool {
	return b[i] >= '0' && b[i] <= '9' || b[i] >= 'A' && b[i] <= 'F' || b[i] >= 'a' && b[i] <= 'f'
}

// Get the value of a hex-digit.
func as_hex(b []byte, i int) int {
	bi := b[i]
	if bi >= 'A' && bi <= 'F' {
		return int(bi) - 'A' + 10
	}
	if bi >= 'a' && bi <= 'f' {
		return int(bi) - 'a' + 10
	}
	return int(bi) - '0'
}

// Check if the character is ASCII.
func is_ascii(b []byte, i int) bool {
	return b[i] <= 0x7F
}

// Check if the character at the start of the buffer can be printed unescaped.
func is_printable(b []byte, i int) bool {
	return ((b[i] == 0x0A) || // . == #x0A
		(b[i] >= 0x20 && b[i] <= 0x7E) || // #x20 <= . <= #x7E
		(b[i] == 0xC2 && b[i+1] >= 0xA0) || // #0xA0 <= . <= #xD7FF
		(b[i] > 0xC2 && b[i] < 0xED) ||
		(b[i] == 0xED && b[i+1] < 0xA0) ||
		(b[i] == 0xEE) ||
		(b[i] == 0xEF && // #xE000 <= . <= #xFFFD
			!(b[i+1] == 0xBB && b[i+2] == 0xBF) && // && . != #xFEFF
			!(b[i+1] == 0xBF && (b[i+2] == 0xBE || b[i+2] == 0xBF))))
}

// Check if the character at the specified position is NUL.
func is_z(b []byte, i int) bool {
	return b[i] == 0x00
}

// Check if the beginning of the buffer is a BOM.
func is_bom(b []byte, i int) bool {
	return b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF
}

// Check if the character at the specified position is space.
func is_space(b []byte, i int) bool {
	return b[i] == ' '
}

// Check if the character at the specified position is tab.
func is_tab(b []byte, i int) bool {
	return b[i] == '\t'
}

// Check if the character at the specified position is blank (space or tab).
func is_blank(b []byte, i int) bool {
	return is_space(b, i) || is_tab(b, i)
}

// Check if the character at the specified position is a line break.
func is_break(b []byte, i int) bool {
	return (b[i] == '\r' || // CR (#xD)
		b[i] == '\n' || // LF (#xA)
		b[i] == 0xC2 && b[i+1] == 0x85 || // NEL (#x85)
		b[i] == 0xE2 && b[i+1] == 0x80 && b[i+2] == 0xA8 || // LS (#x2028)
		b[i] == 0xE2 && b[i+1] == 0x80 && b[i+2] == 0xA9) // PS (#x2029)
}

func is_crlf(b []byte, i int) bool {
	return b[i] == '\r' && b[i+1] == '\n'
}

// Check if the character is a line break or NUL.
func is_breakz(b []byte, i int) bool {
	return is_break(b, i) || b[i] == 0x00
}

// Check if the character is a line break, space, or NUL.
func is_spacez(b []byte, i int) bool {
	return is_space(b, i) || is_breakz(b, i)
}

// Check if the character is a line break, space, tab, or NUL.
func is_blankz(b []byte, i int) bool {
	return is_blank(b, i) || is_breakz(b, i)
}

// Determine the width of the character.
func width(b byte) int {
	switch {
	case b&0x80 == 0x00:
		return 1
	case b&0xE0 == 0xC0:
		return 2
	case b&0xF0 == 0xE0:
		return 3
	case b&0xF8 == 0xF0:
		return 4
	}
	return 0
}


///*
// * Event initializers.
// */
//
//#define EVENT_INIT(event,event_type,event_start_mark,event_end_mark)
//    (memset(&(event), 0, sizeof(yaml_event_t)),
//     (event).type = (event_type),
//     (event).start_mark = (event_start_mark),
//     (event).end_mark = (event_end_mark))
//

//#define DOCUMENT_END_EVENT_INIT(event,event_implicit,start_mark,end_mark)
//    (EVENT_INIT((event),YAML_DOCUMENT_END_EVENT,(start_mark),(end_mark)),
//     (event).data.document_end.implicit = (event_implicit))
//
//#define ALIAS_EVENT_INIT(event,event_anchor,start_mark,end_mark)
//    (EVENT_INIT((event),YAML_ALIAS_EVENT,(start_mark),(end_mark)),
//     (event).data.alias.anchor = (event_anchor))
//
//#define SCALAR_EVENT_INIT(event,event_anchor,event_tag,event_value,event_length,
//        event_plain_implicit, event_quoted_implicit,event_style,start_mark,end_mark)
//    (EVENT_INIT((event),YAML_SCALAR_EVENT,(start_mark),(end_mark)),
//     (event).data.scalar.anchor = (event_anchor),
//     (event).data.scalar.tag = (event_tag),
//     (event).data.scalar.value = (event_value),
//     (event).data.scalar.length = (event_length),
//     (event).data.scalar.plain_implicit = (event_plain_implicit),
//     (event).data.scalar.quoted_implicit = (event_quoted_implicit),
//     (event).data.scalar.style = (event_style))
//
//#define SEQUENCE_START_EVENT_INIT(event,event_anchor,event_tag,
//        event_implicit,event_style,start_mark,end_mark)
//    (EVENT_INIT((event),YAML_SEQUENCE_START_EVENT,(start_mark),(end_mark)),
//     (event).data.sequence_start.anchor = (event_anchor),
//     (event).data.sequence_start.tag = (event_tag),
//     (event).data.sequence_start.implicit = (event_implicit),
//     (event).data.sequence_start.style = (event_style))
//
//#define SEQUENCE_END_EVENT_INIT(event,start_mark,end_mark)
//    (EVENT_INIT((event),YAML_SEQUENCE_END_EVENT,(start_mark),(end_mark)))
//
//#define MAPPING_START_EVENT_INIT(event,event_anchor,event_tag,
//        event_implicit,event_style,start_mark,end_mark)
//    (EVENT_INIT((event),YAML_MAPPING_START_EVENT,(start_mark),(end_mark)),
//     (event).data.mapping_start.anchor = (event_anchor),
//     (event).data.mapping_start.tag = (event_tag),
//     (event).data.mapping_start.implicit = (event_implicit),
//     (event).data.mapping_start.style = (event_style))
//
//#define MAPPING_END_EVENT_INIT(event,start_mark,end_mark)
//    (EVENT_INIT((event),YAML_MAPPING_END_EVENT,(start_mark),(end_mark)))
//
///*
// * Document initializer.
// */
//
//#define DOCUMENT_INIT(document,document_nodes_start,document_nodes_end,
//        document_version_directive,document_tag_directives_start,
//        document_tag_directives_end,document_start_implicit,
//        document_end_implicit,document_start_mark,document_end_mark)
//    (memset(&(document), 0, sizeof(yaml_document_t)),
//     (document).nodes.start = (document_nodes_start),
//     (document).nodes.end = (document_nodes_end),
//     (document).nodes.top = (document_nodes_start),
//     (document).version_directive = (document_version_directive),
//     (document).tag_directives.start = (document_tag_directives_start),
//     (document).tag_directives.end = (document_tag_directives_end),
//     (document).start_implicit = (document_start_implicit),
//     (document).end_implicit = (document_end_implicit),
//     (document).start_mark = (document_start_mark),
//     (document).end_mark = (document_end_mark))
//
///*
// * Node initializers.
// */
//
//#define NODE_INIT(node,node_type,node_tag,node_start_mark,node_end_mark)
//    (memset(&(node), 0, sizeof(yaml_node_t)),
//     (node).type = (node_type),
//     (node).tag = (node_tag),
//     (node).start_mark = (node_start_mark),
//     (node).end_mark = (node_end_mark))
//
//#define SCALAR_NODE_INIT(node,node_tag,node_value,node_length,
//        node_style,start_mark,end_mark)
//    (NODE_INIT((node),YAML_SCALAR_NODE,(node_tag),(start_mark),(end_mark)),
//     (node).data.scalar.value = (node_value),
//     (node).data.scalar.length = (node_length),
//     (node).data.scalar.style = (node_style))
//
//#define SEQUENCE_NODE_INIT(node,node_tag,node_items_start,node_items_end,
//        node_style,start_mark,end_mark)
//    (NODE_INIT((node),YAML_SEQUENCE_NODE,(node_tag),(start_mark),(end_mark)),
//     (node).data.sequence.items.start = (node_items_start),
//     (node).data.sequence.items.end = (node_items_end),
//     (node).data.sequence.items.top = (node_items_start),
//     (node).data.sequence.style = (node_style))
//
//#define MAPPING_NODE_INIT(node,node_tag,node_pairs_start,node_pairs_end,
//        node_style,start_mark,end_mark)
//    (NODE_INIT((node),YAML_MAPPING_NODE,(node_tag),(start_mark),(end_mark)),
//     (node).data.mapping.pairs.start = (node_pairs_start),
//     (node).data.mapping.pairs.end = (node_pairs_end),
//     (node).data.mapping.pairs.top = (node_pairs_start),
//     (node).data.mapping.style = (node_style))
//
