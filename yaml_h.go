package goyaml

import (
	"io"
)

// The version directive data.
type yaml_version_directive_t struct {
	major int // The major version number.
	minor int // The minor version number.
}

// The tag directive data.
type yaml_tag_directive_t struct {
	handle []byte // The tag handle.
	prefix []byte // The tag prefix.
}

type yaml_encoding_t int

// The stream encoding.
const (
	// Let the parser choose the encoding.
	yaml_ANY_ENCODING yaml_encoding_t = iota

	yaml_UTF8_ENCODING    // The default UTF-8 encoding.
	yaml_UTF16LE_ENCODING // The UTF-16-LE encoding with BOM.
	yaml_UTF16BE_ENCODING // The UTF-16-BE encoding with BOM.
)

type yaml_break_t int

// Line break types.
const (
	// Let the parser choose the break type.
	yaml_ANY_BREAK yaml_break_t = iota

	yaml_CR_BREAK   // Use CR for line breaks (Mac style).
	yaml_LN_BREAK   // Use LN for line breaks (Unix style).
	yaml_CRLN_BREAK // Use CR LN for line breaks (DOS style).
)

type yaml_error_type_t int

// Many bad things could happen with the parser and emitter.
const (
	// No error is produced.
	yaml_NO_ERROR yaml_error_type_t = iota

	yaml_MEMORY_ERROR   // Cannot allocate or reallocate a block of memory.
	yaml_READER_ERROR   // Cannot read or decode the input stream.
	yaml_SCANNER_ERROR  // Cannot scan the input stream.
	yaml_PARSER_ERROR   // Cannot parse the input stream.
	yaml_COMPOSER_ERROR // Cannot compose a YAML document.
	yaml_WRITER_ERROR   // Cannot write to the output stream.
	yaml_EMITTER_ERROR  // Cannot emit a YAML stream.
)

// The pointer position.
type yaml_mark_t struct {
	index  int // The position index.
	line   int // The position line.
	column int // The position column.
}

// Node Styles

type yaml_scalar_style_t int

// Scalar styles.
const (
	// Let the emitter choose the style.
	yaml_ANY_SCALAR_STYLE yaml_scalar_style_t = iota

	yaml_PLAIN_SCALAR_STYLE         // The plain scalar style.
	yaml_SINGLE_QUOTED_SCALAR_STYLE // The single-quoted scalar style.
	yaml_DOUBLE_QUOTED_SCALAR_STYLE // The double-quoted scalar style.
	yaml_LITERAL_SCALAR_STYLE       // The literal scalar style.
	yaml_FOLDED_SCALAR_STYLE        // The folded scalar style.
)

type yaml_sequence_style_t int

// Sequence styles.
const (
	// Let the emitter choose the style.
	yaml_ANY_SEQUENCE_STYLE yaml_sequence_style_t = iota

	yaml_BLOCK_SEQUENCE_STYLE // The block sequence style.
	yaml_FLOW_SEQUENCE_STYLE  // The flow sequence style.
)

type yaml_mapping_style_t int

// Mapping styles.
const (
	// Let the emitter choose the style.
	YAML_ANY_MAPPING_STYLE yaml_mapping_style_t = iota

	yaml_BLOCK_MAPPING_STYLE // The block mapping style.
	yaml_FLOW_MAPPING_STYLE  // The flow mapping style.
)

// Tokens

type yaml_token_type_t int

// Token types.
const (
	// An empty token.
	yaml_NO_TOKEN yaml_token_type_t = iota

	yaml_STREAM_START_TOKEN // A STREAM-START token.
	yaml_STREAM_END_TOKEN   // A STREAM-END token.

	yaml_VERSION_DIRECTIVE_TOKEN // A VERSION-DIRECTIVE token.
	yaml_TAG_DIRECTIVE_TOKEN     // A TAG-DIRECTIVE token.
	yaml_DOCUMENT_START_TOKEN    // A DOCUMENT-START token.
	yaml_DOCUMENT_END_TOKEN      // A DOCUMENT-END token.

	yaml_BLOCK_SEQUENCE_START_TOKEN // A BLOCK-SEQUENCE-START token.
	yaml_BLOCK_MAPPING_START_TOKEN  // A BLOCK-SEQUENCE-END token.
	yaml_BLOCK_END_TOKEN            // A BLOCK-END token.

	yaml_FLOW_SEQUENCE_START_TOKEN // A FLOW-SEQUENCE-START token.
	yaml_FLOW_SEQUENCE_END_TOKEN   // A FLOW-SEQUENCE-END token.
	yaml_FLOW_MAPPING_START_TOKEN  // A FLOW-MAPPING-START token.
	yaml_FLOW_MAPPING_END_TOKEN    // A FLOW-MAPPING-END token.

	yaml_BLOCK_ENTRY_TOKEN // A BLOCK-ENTRY token.
	yaml_FLOW_ENTRY_TOKEN  // A FLOW-ENTRY token.
	yaml_KEY_TOKEN         // A KEY token.
	yaml_VALUE_TOKEN       // A VALUE token.

	yaml_ALIAS_TOKEN  // An ALIAS token.
	yaml_ANCHOR_TOKEN // An ANCHOR token.
	yaml_TAG_TOKEN    // A TAG token.
	yaml_SCALAR_TOKEN // A SCALAR token.
)

func (tt yaml_token_type_t) String() string {
	switch tt {
	case yaml_NO_TOKEN:
		return "yaml_NO_TOKEN"
	case yaml_STREAM_START_TOKEN:
		return "yaml_STREAM_START_TOKEN"
	case yaml_STREAM_END_TOKEN:
		return "yaml_STREAM_END_TOKEN"
	case yaml_VERSION_DIRECTIVE_TOKEN:
		return "yaml_VERSION_DIRECTIVE_TOKEN"
	case yaml_TAG_DIRECTIVE_TOKEN:
		return "yaml_TAG_DIRECTIVE_TOKEN"
	case yaml_DOCUMENT_START_TOKEN:
		return "yaml_DOCUMENT_START_TOKEN"
	case yaml_DOCUMENT_END_TOKEN:
		return "yaml_DOCUMENT_END_TOKEN"
	case yaml_BLOCK_SEQUENCE_START_TOKEN:
		return "yaml_BLOCK_SEQUENCE_START_TOKEN"
	case yaml_BLOCK_MAPPING_START_TOKEN:
		return "yaml_BLOCK_MAPPING_START_TOKEN"
	case yaml_BLOCK_END_TOKEN:
		return "yaml_BLOCK_END_TOKEN"
	case yaml_FLOW_SEQUENCE_START_TOKEN:
		return "yaml_FLOW_SEQUENCE_START_TOKEN"
	case yaml_FLOW_SEQUENCE_END_TOKEN:
		return "yaml_FLOW_SEQUENCE_END_TOKEN"
	case yaml_FLOW_MAPPING_START_TOKEN:
		return "yaml_FLOW_MAPPING_START_TOKEN"
	case yaml_FLOW_MAPPING_END_TOKEN:
		return "yaml_FLOW_MAPPING_END_TOKEN"
	case yaml_BLOCK_ENTRY_TOKEN:
		return "yaml_BLOCK_ENTRY_TOKEN"
	case yaml_FLOW_ENTRY_TOKEN:
		return "yaml_FLOW_ENTRY_TOKEN"
	case yaml_KEY_TOKEN:
		return "yaml_KEY_TOKEN"
	case yaml_VALUE_TOKEN:
		return "yaml_VALUE_TOKEN"
	case yaml_ALIAS_TOKEN:
		return "yaml_ALIAS_TOKEN"
	case yaml_ANCHOR_TOKEN:
		return "yaml_ANCHOR_TOKEN"
	case yaml_TAG_TOKEN:
		return "yaml_TAG_TOKEN"
	case yaml_SCALAR_TOKEN:
		return "yaml_SCALAR_TOKEN"
	}
	return "<unknown token>"
}

// The token structure.
type yaml_token_t struct {

	// The token type.
	typ yaml_token_type_t

	// The token data.

	// The stream start (for yaml_STREAM_START_TOKEN).
	stream_start struct {
		encoding yaml_encoding_t // The stream encoding.
	}

	// The alias (for yaml_ALIAS_TOKEN).
	alias struct {
		value []byte // The alias value.
	}

	// The anchor (for yaml_ANCHOR_TOKEN).
	anchor struct {
		value []byte // The anchor value.
	}

	// The tag (for yaml_TAG_TOKEN).
	tag struct {
		handle []byte // The tag handle.
		suffix []byte // The tag suffix.
	}

	// The scalar value (for yaml_SCALAR_TOKEN).
	scalar struct {
		value []byte              // The scalar value.
		style yaml_scalar_style_t // The scalar style.
	}

	// The version directive (for yaml_VERSION_DIRECTIVE_TOKEN).
	version_directive struct {
		major int // The major version number.
		minor int // The minor version number.
	}

	// The tag directive (for yaml_TAG_DIRECTIVE_TOKEN).
	tag_directive struct {
		handle []byte // The tag handle.
		prefix []byte // The tag prefix.
	}

	// The beginning of the token.
	start_mark yaml_mark_t
	// The end of the token.
	end_mark yaml_mark_t
}

// Events

type yaml_event_type_t int

// Event types.
const (
	// An empty event.
	yaml_NO_EVENT yaml_event_type_t = iota

	yaml_STREAM_START_EVENT   // A STREAM-START event.
	yaml_STREAM_END_EVENT     // A STREAM-END event.
	yaml_DOCUMENT_START_EVENT // A DOCUMENT-START event.
	yaml_DOCUMENT_END_EVENT   // A DOCUMENT-END event.
	yaml_ALIAS_EVENT          // An ALIAS event.
	yaml_SCALAR_EVENT         // A SCALAR event.
	yaml_SEQUENCE_START_EVENT // A SEQUENCE-START event.
	yaml_SEQUENCE_END_EVENT   // A SEQUENCE-END event.
	yaml_MAPPING_START_EVENT  // A MAPPING-START event.
	yaml_MAPPING_END_EVENT    // A MAPPING-END event.
)

// The event structure.
type yaml_event_t struct {

	// The event type.
	typ yaml_event_type_t

	// The event data.

	// The stream parameters (for yaml_STREAM_START_EVENT).
	stream_start struct {
		encoding yaml_encoding_t // The document encoding.
	}

	// The document parameters (for yaml_DOCUMENT_START_EVENT).
	document_start struct {
		version_directive *yaml_version_directive_t // The version directive.

		// The list of tag directives.
		tag_directives []yaml_tag_directive_t
		implicit       bool // Is the document indicator implicit?
	}

	// The document end parameters (for yaml_DOCUMENT_END_EVENT).
	document_end struct {
		implicit bool // Is the document end indicator implicit?
	}

	// The alias parameters (for yaml_ALIAS_EVENT).
	alias struct {
		anchor []byte // The anchor.
	}

	// The scalar parameters (for yaml_SCALAR_EVENT).
	scalar struct {
		anchor          []byte              // The anchor.
		tag             []byte              // The tag.
		value           []byte              // The scalar value.
		length          int                 // The length of the scalar value.
		plain_implicit  bool                // Is the tag optional for the plain style?
		quoted_implicit bool                // Is the tag optional for any non-plain style?
		style           yaml_scalar_style_t // The scalar style.
	}

	// The sequence parameters (for yaml_SEQUENCE_START_EVENT).
	sequence_start struct {
		anchor   []byte                // The anchor.
		tag      []byte                // The tag.
		implicit bool                  // Is the tag optional?
		style    yaml_sequence_style_t // The sequence style.
	}

	// The mapping parameters (for yaml_MAPPING_START_EVENT).
	mapping_start struct {
		anchor   []byte               // The anchor.
		tag      []byte               // The tag.
		implicit bool                 // Is the tag optional?
		style    yaml_mapping_style_t // The mapping style.
	}

	start_mark yaml_mark_t // The beginning of the event.
	end_mark   yaml_mark_t // The end of the event.

}

///**
// * Create the STREAM-START event.
// *
// * @param[out]      event       An empty event object.
// * @param[in]       encoding    The stream encoding.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_stream_start_event_initialize(yaml_event_t *event,
//        yaml_encoding_t encoding);
//
///**
// * Create the STREAM-END event.
// *
// * @param[out]      event       An empty event object.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_stream_end_event_initialize(yaml_event_t *event);
//
///**
// * Create the DOCUMENT-START event.
// *
// * The @a implicit argument is considered as a stylistic parameter and may be
// * ignored by the emitter.
// *
// * @param[out]      event                   An empty event object.
// * @param[in]       version_directive       The %YAML directive value or
// *                                          @c NULL.
// * @param[in]       tag_directives_start    The beginning of the %TAG
// *                                          directives list.
// * @param[in]       tag_directives_end      The end of the %TAG directives
// *                                          list.
// * @param[in]       implicit                If the document start indicator is
// *                                          implicit.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_document_start_event_initialize(yaml_event_t *event,
//        yaml_version_directive_t *version_directive,
//        yaml_tag_directive_t *tag_directives_start,
//        yaml_tag_directive_t *tag_directives_end,
//        int implicit);
//
///**
// * Create the DOCUMENT-END event.
// *
// * The @a implicit argument is considered as a stylistic parameter and may be
// * ignored by the emitter.
// *
// * @param[out]      event       An empty event object.
// * @param[in]       implicit    If the document end indicator is implicit.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_document_end_event_initialize(yaml_event_t *event, int implicit);
//
///**
// * Create an ALIAS event.
// *
// * @param[out]      event       An empty event object.
// * @param[in]       anchor      The anchor value.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_alias_event_initialize(yaml_event_t *event, yaml_char_t *anchor);
//
///**
// * Create a SCALAR event.
// *
// * The @a style argument may be ignored by the emitter.
// *
// * Either the @a tag attribute or one of the @a plain_implicit and
// * @a quoted_implicit flags must be set.
// *
// * @param[out]      event           An empty event object.
// * @param[in]       anchor          The scalar anchor or @c NULL.
// * @param[in]       tag             The scalar tag or @c NULL.
// * @param[in]       value           The scalar value.
// * @param[in]       length          The length of the scalar value.
// * @param[in]       plain_implicit  If the tag may be omitted for the plain
// *                                  style.
// * @param[in]       quoted_implicit If the tag may be omitted for any
// *                                  non-plain style.
// * @param[in]       style           The scalar style.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_scalar_event_initialize(yaml_event_t *event,
//        yaml_char_t *anchor, yaml_char_t *tag,
//        yaml_char_t *value, int length,
//        int plain_implicit, int quoted_implicit,
//        yaml_scalar_style_t style);
//
///**
// * Create a SEQUENCE-START event.
// *
// * The @a style argument may be ignored by the emitter.
// *
// * Either the @a tag attribute or the @a implicit flag must be set.
// *
// * @param[out]      event       An empty event object.
// * @param[in]       anchor      The sequence anchor or @c NULL.
// * @param[in]       tag         The sequence tag or @c NULL.
// * @param[in]       implicit    If the tag may be omitted.
// * @param[in]       style       The sequence style.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_sequence_start_event_initialize(yaml_event_t *event,
//        yaml_char_t *anchor, yaml_char_t *tag, int implicit,
//        yaml_sequence_style_t style);
//
///**
// * Create a SEQUENCE-END event.
// *
// * @param[out]      event       An empty event object.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_sequence_end_event_initialize(yaml_event_t *event);
//
///**
// * Create a MAPPING-START event.
// *
// * The @a style argument may be ignored by the emitter.
// *
// * Either the @a tag attribute or the @a implicit flag must be set.
// *
// * @param[out]      event       An empty event object.
// * @param[in]       anchor      The mapping anchor or @c NULL.
// * @param[in]       tag         The mapping tag or @c NULL.
// * @param[in]       implicit    If the tag may be omitted.
// * @param[in]       style       The mapping style.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_mapping_start_event_initialize(yaml_event_t *event,
//        yaml_char_t *anchor, yaml_char_t *tag, int implicit,
//        yaml_mapping_style_t style);
//
///**
// * Create a MAPPING-END event.
// *
// * @param[out]      event       An empty event object.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_mapping_end_event_initialize(yaml_event_t *event);
//
///**
// * Free any memory allocated for an event object.
// *
// * @param[in,out]   event   An event object.
// */
//
//YAML_DECLARE(void)
//yaml_event_delete(yaml_event_t *event);
//
//// @}
//

// Nodes

const (
	yaml_NULL_TAG      = "tag:yaml.org,2002:null"      // The tag !!null with the only possible value: null.
	yaml_BOOL_TAG      = "tag:yaml.org,2002:bool"      // The tag !!bool with the values: true and false.
	yaml_STR_TAG       = "tag:yaml.org,2002:str"       // The tag !!str for string values.
	yaml_INT_TAG       = "tag:yaml.org,2002:int"       // The tag !!int for integer values.
	yaml_FLOAT_TAG     = "tag:yaml.org,2002:float"     // The tag !!float for float values.
	yaml_TIMESTAMP_TAG = "tag:yaml.org,2002:timestamp" // The tag !!timestamp for date and time values.

	yaml_SEQ_TAG = "tag:yaml.org,2002:seq" // The tag !!seq is used to denote sequences.
	yaml_MAP_TAG = "tag:yaml.org,2002:map" // The tag !!map is used to denote mapping.

	yaml_DEFAULT_SCALAR_TAG   = yaml_STR_TAG // The default scalar tag is !!str.
	yaml_DEFAULT_SEQUENCE_TAG = yaml_SEQ_TAG // The default sequence tag is !!seq.
	yaml_DEFAULT_MAPPING_TAG  = yaml_MAP_TAG // The default mapping tag is !!map.
)

type yaml_node_type_t int

// Node types.
const (
	// An empty node.
	yaml_NO_NODE yaml_node_type_t = iota

	yaml_SCALAR_NODE   // A scalar node.
	yaml_SEQUENCE_NODE // A sequence node.
	yaml_MAPPING_NODE  // A mapping node.
)

// An element of a sequence node.
type yaml_node_item_t int

// An element of a mapping node.
type yaml_node_pair_t struct {
	key   int // The key of the element.
	value int // The value of the element.
}

// The node structure.
type yaml_node_t struct {
	typ yaml_node_type_t // The node type.
	tag []byte           // The node tag.

	// The node data.

	// The scalar parameters (for yaml_SCALAR_NODE).
	scalar struct {
		value  []byte              // The scalar value.
		length int                 // The length of the scalar value.
		style  yaml_scalar_style_t // The scalar style.
	}

	// The sequence parameters (for YAML_SEQUENCE_NODE).
	sequence struct {
		items_data []yaml_node_item_t    // The stack of sequence items.
		style      yaml_sequence_style_t // The sequence style.
	}

	// The mapping parameters (for yaml_MAPPING_NODE).
	mapping struct {
		pairs_data  []yaml_node_pair_t   // The stack of mapping pairs (key, value).
		pairs_start *yaml_node_pair_t    // The beginning of the stack.
		pairs_end   *yaml_node_pair_t    // The end of the stack.
		pairs_top   *yaml_node_pair_t    // The top of the stack.
		style       yaml_mapping_style_t // The mapping style.
	}

	start_mark yaml_mark_t // The beginning of the node.
	end_mark   yaml_mark_t // The end of the node.

}

// The document structure.
type yaml_document_t struct {

	// The document nodes.
	nodes []yaml_node_t

	// The version directive.
	version_directive *yaml_version_directive_t

	// The list of tag directives.
	tag_directives_data  []yaml_tag_directive_t
	tag_directives_start int // The beginning of the tag directives list.
	tag_directives_end   int // The end of the tag directives list.

	start_implicit int // Is the document start indicator implicit?
	end_implicit   int // Is the document end indicator implicit?

	start_mark yaml_mark_t // The beginning of the document.
	end_mark   yaml_mark_t // The end of the document.

}

// The prototype of a read handler.
//
// The read handler is called when the parser needs to read more bytes from the
// source. The handler should write not more than size bytes to the buffer.
// The number of written bytes should be set to the size_read variable.
//
// [in,out]   data        A pointer to an application data specified by
//                        yaml_parser_set_input().
// [out]      buffer      The buffer to write the data from the source.
// [in]       size        The size of the buffer.
// [out]      size_read   The actual number of bytes read from the source.
//
// On success, the handler should return 1.  If the handler failed,
// the returned value should be 0. On EOF, the handler should set the
// size_read to 0 and return 1.
type yaml_read_handler_t func(parser *yaml_parser_t, buffer []byte) (n int, err error)

// This structure holds information about a potential simple key.
type yaml_simple_key_t struct {
	possible     bool        // Is a simple key possible?
	required     bool        // Is a simple key required?
	token_number int         // The number of the token.
	mark         yaml_mark_t // The position mark.
}

// The states of the parser.
type yaml_parser_state_t int

const (
	yaml_PARSE_STREAM_START_STATE yaml_parser_state_t = iota

	yaml_PARSE_IMPLICIT_DOCUMENT_START_STATE           // Expect the beginning of an implicit document.
	yaml_PARSE_DOCUMENT_START_STATE                    // Expect DOCUMENT-START.
	yaml_PARSE_DOCUMENT_CONTENT_STATE                  // Expect the content of a document.
	yaml_PARSE_DOCUMENT_END_STATE                      // Expect DOCUMENT-END.
	yaml_PARSE_BLOCK_NODE_STATE                        // Expect a block node.
	yaml_PARSE_BLOCK_NODE_OR_INDENTLESS_SEQUENCE_STATE // Expect a block node or indentless sequence.
	yaml_PARSE_FLOW_NODE_STATE                         // Expect a flow node.
	yaml_PARSE_BLOCK_SEQUENCE_FIRST_ENTRY_STATE        // Expect the first entry of a block sequence.
	yaml_PARSE_BLOCK_SEQUENCE_ENTRY_STATE              // Expect an entry of a block sequence.
	yaml_PARSE_INDENTLESS_SEQUENCE_ENTRY_STATE         // Expect an entry of an indentless sequence.
	yaml_PARSE_BLOCK_MAPPING_FIRST_KEY_STATE           // Expect the first key of a block mapping.
	yaml_PARSE_BLOCK_MAPPING_KEY_STATE                 // Expect a block mapping key.
	yaml_PARSE_BLOCK_MAPPING_VALUE_STATE               // Expect a block mapping value.
	yaml_PARSE_FLOW_SEQUENCE_FIRST_ENTRY_STATE         // Expect the first entry of a flow sequence.
	yaml_PARSE_FLOW_SEQUENCE_ENTRY_STATE               // Expect an entry of a flow sequence.
	yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_KEY_STATE   // Expect a key of an ordered mapping.
	yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_VALUE_STATE // Expect a value of an ordered mapping.
	yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_END_STATE   // Expect the and of an ordered mapping entry.
	yaml_PARSE_FLOW_MAPPING_FIRST_KEY_STATE            // Expect the first key of a flow mapping.
	yaml_PARSE_FLOW_MAPPING_KEY_STATE                  // Expect a key of a flow mapping.
	yaml_PARSE_FLOW_MAPPING_VALUE_STATE                // Expect a value of a flow mapping.
	yaml_PARSE_FLOW_MAPPING_EMPTY_VALUE_STATE          // Expect an empty value of a flow mapping.
	yaml_PARSE_END_STATE                               // Expect nothing.
)

func (ps yaml_parser_state_t) String() string {
	switch ps {
	case yaml_PARSE_STREAM_START_STATE:
		return "yaml_PARSE_STREAM_START_STATE"
	case yaml_PARSE_IMPLICIT_DOCUMENT_START_STATE:
		return "yaml_PARSE_IMPLICIT_DOCUMENT_START_STATE"
	case yaml_PARSE_DOCUMENT_START_STATE:
		return "yaml_PARSE_DOCUMENT_START_STATE"
	case yaml_PARSE_DOCUMENT_CONTENT_STATE:
		return "yaml_PARSE_DOCUMENT_CONTENT_STATE"
	case yaml_PARSE_DOCUMENT_END_STATE:
		return "yaml_PARSE_DOCUMENT_END_STATE"
	case yaml_PARSE_BLOCK_NODE_STATE:
		return "yaml_PARSE_BLOCK_NODE_STATE"
	case yaml_PARSE_BLOCK_NODE_OR_INDENTLESS_SEQUENCE_STATE:
		return "yaml_PARSE_BLOCK_NODE_OR_INDENTLESS_SEQUENCE_STATE"
	case yaml_PARSE_FLOW_NODE_STATE:
		return "yaml_PARSE_FLOW_NODE_STATE"
	case yaml_PARSE_BLOCK_SEQUENCE_FIRST_ENTRY_STATE:
		return "yaml_PARSE_BLOCK_SEQUENCE_FIRST_ENTRY_STATE"
	case yaml_PARSE_BLOCK_SEQUENCE_ENTRY_STATE:
		return "yaml_PARSE_BLOCK_SEQUENCE_ENTRY_STATE"
	case yaml_PARSE_INDENTLESS_SEQUENCE_ENTRY_STATE:
		return "yaml_PARSE_INDENTLESS_SEQUENCE_ENTRY_STATE"
	case yaml_PARSE_BLOCK_MAPPING_FIRST_KEY_STATE:
		return "yaml_PARSE_BLOCK_MAPPING_FIRST_KEY_STATE"
	case yaml_PARSE_BLOCK_MAPPING_KEY_STATE:
		return "yaml_PARSE_BLOCK_MAPPING_KEY_STATE"
	case yaml_PARSE_BLOCK_MAPPING_VALUE_STATE:
		return "yaml_PARSE_BLOCK_MAPPING_VALUE_STATE"
	case yaml_PARSE_FLOW_SEQUENCE_FIRST_ENTRY_STATE:
		return "yaml_PARSE_FLOW_SEQUENCE_FIRST_ENTRY_STATE"
	case yaml_PARSE_FLOW_SEQUENCE_ENTRY_STATE:
		return "yaml_PARSE_FLOW_SEQUENCE_ENTRY_STATE"
	case yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_KEY_STATE:
		return "yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_KEY_STATE"
	case yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_VALUE_STATE:
		return "yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_VALUE_STATE"
	case yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_END_STATE:
		return "yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_END_STATE"
	case yaml_PARSE_FLOW_MAPPING_FIRST_KEY_STATE:
		return "yaml_PARSE_FLOW_MAPPING_FIRST_KEY_STATE"
	case yaml_PARSE_FLOW_MAPPING_KEY_STATE:
		return "yaml_PARSE_FLOW_MAPPING_KEY_STATE"
	case yaml_PARSE_FLOW_MAPPING_VALUE_STATE:
		return "yaml_PARSE_FLOW_MAPPING_VALUE_STATE"
	case yaml_PARSE_FLOW_MAPPING_EMPTY_VALUE_STATE:
		return "yaml_PARSE_FLOW_MAPPING_EMPTY_VALUE_STATE"
	case yaml_PARSE_END_STATE:
		return "yaml_PARSE_END_STATE"
	}
	return "<unknown parser state>"
}

// This structure holds aliases data.
type yaml_alias_data_t struct {
	anchor []byte      // The anchor.
	index  int         // The node id.
	mark   yaml_mark_t // The anchor mark.
}

// The parser structure.
//
// All members are internal. Manage the structure using the
// yaml_parser_ family of functions.
type yaml_parser_t struct {

	// Error handling

	error yaml_error_type_t // Error type.

	problem string // Error description.

	// The byte about which the problem occured.
	problem_offset int
	problem_value  int
	problem_mark   yaml_mark_t

	// The error context.
	context      string
	context_mark yaml_mark_t

	// Reader stuff

	read_handler yaml_read_handler_t // Read handler.

	input_file io.Reader // File input data.
	input      []byte    // String input data.
	input_pos  int

	eof bool // EOF flag

	buffer     []byte // The working buffer.
	buffer_pos int    // The current position of the buffer.

	unread int // The number of unread characters in the buffer.

	raw_buffer     []byte // The raw buffer.
	raw_buffer_pos int    // The current position of the buffer.

	encoding yaml_encoding_t // The input encoding.

	offset int         // The offset of the current position (in bytes).
	mark   yaml_mark_t // The mark of the current position.

	// Scanner stuff

	stream_start_produced bool // Have we started to scan the input stream?
	stream_end_produced   bool // Have we reached the end of the input stream?

	flow_level int // The number of unclosed '[' and '{' indicators.

	tokens          []yaml_token_t // The tokens queue.
	tokens_head     int            // The head of the tokens queue.
	tokens_parsed   int            // The number of tokens fetched from the queue.
	token_available bool           // Does the tokens queue contain a token ready for dequeueing.

	indent  int   // The current indentation level.
	indents []int // The indentation levels stack.

	simple_key_allowed bool                // May a simple key occur at the current position?
	simple_keys        []yaml_simple_key_t // The stack of simple keys.

	// Parser stuff

	state          yaml_parser_state_t    // The current parser state.
	states         []yaml_parser_state_t  // The parser states stack.
	marks          []yaml_mark_t          // The stack of marks.
	tag_directives []yaml_tag_directive_t // The list of TAG directives.

	// Dumper stuff

	aliases []yaml_alias_data_t // The alias data.

	document *yaml_document_t // The currently parsed document.
}

///**
// * @defgroup emitter Emitter Definitions
// * @{
// */
//
///**
// * The prototype of a write handler.
// *
// * The write handler is called when the emitter needs to flush the accumulated
// * characters to the output.  The handler should write @a size bytes of the
// * @a buffer to the output.
// *
// * @param[in,out]   data        A pointer to an application data specified by
// *                              yaml_emitter_set_output().
// * @param[in]       buffer      The buffer with bytes to be written.
// * @param[in]       size        The size of the buffer.
// *
// * @returns On success, the handler should return @c 1.  If the handler failed,
// * the returned value should be @c 0.
// */
//
//typedef int yaml_write_handler_t(void *data, unsigned char *buffer, int size);
//
//// The emitter states.
//typedef enum yaml_emitter_state_e {
//    // Expect STREAM-START.
//    YAML_EMIT_STREAM_START_STATE,
//    // Expect the first DOCUMENT-START or STREAM-END.
//    YAML_EMIT_FIRST_DOCUMENT_START_STATE,
//    // Expect DOCUMENT-START or STREAM-END.
//    YAML_EMIT_DOCUMENT_START_STATE,
//    // Expect the content of a document.
//    YAML_EMIT_DOCUMENT_CONTENT_STATE,
//    // Expect DOCUMENT-END.
//    YAML_EMIT_DOCUMENT_END_STATE,
//    // Expect the first item of a flow sequence.
//    YAML_EMIT_FLOW_SEQUENCE_FIRST_ITEM_STATE,
//    // Expect an item of a flow sequence.
//    YAML_EMIT_FLOW_SEQUENCE_ITEM_STATE,
//    // Expect the first key of a flow mapping.
//    YAML_EMIT_FLOW_MAPPING_FIRST_KEY_STATE,
//    // Expect a key of a flow mapping.
//    YAML_EMIT_FLOW_MAPPING_KEY_STATE,
//    // Expect a value for a simple key of a flow mapping.
//    YAML_EMIT_FLOW_MAPPING_SIMPLE_VALUE_STATE,
//    // Expect a value of a flow mapping.
//    YAML_EMIT_FLOW_MAPPING_VALUE_STATE,
//    // Expect the first item of a block sequence.
//    YAML_EMIT_BLOCK_SEQUENCE_FIRST_ITEM_STATE,
//    // Expect an item of a block sequence.
//    YAML_EMIT_BLOCK_SEQUENCE_ITEM_STATE,
//    // Expect the first key of a block mapping.
//    YAML_EMIT_BLOCK_MAPPING_FIRST_KEY_STATE,
//    // Expect the key of a block mapping.
//    YAML_EMIT_BLOCK_MAPPING_KEY_STATE,
//    // Expect a value for a simple key of a block mapping.
//    YAML_EMIT_BLOCK_MAPPING_SIMPLE_VALUE_STATE,
//    // Expect a value of a block mapping.
//    YAML_EMIT_BLOCK_MAPPING_VALUE_STATE,
//    // Expect nothing.
//    YAML_EMIT_END_STATE
//} yaml_emitter_state_t;
//
///**
// * The emitter structure.
// *
// * All members are internal.  Manage the structure using the @c yaml_emitter_
// * family of functions.
// */
//
//typedef struct yaml_emitter_s {
//
//    /**
//     * @name Error handling
//     * @{
//     */
//
//    // Error type.
//    yaml_error_type_t error;
//    // Error description.
//    char *problem;
//
//    /**
//     * @}
//     */
//
//    /**
//     * @name Writer stuff
//     * @{
//     */
//
//    // Write handler.
//    yaml_write_handler_t *write_handler;
//
//    // A pointer for passing to the white handler.
//    void *write_handler_data;
//
//    // Standard (string or file) output data.
//    union {
//        // String output data.
//        struct {
//            // The buffer pointer.
//            unsigned char *buffer;
//            // The buffer size.
//            int size;
//            // The number of written bytes.
//            int *size_written;
//        } string;
//
//        // File output data.
//        FILE *file;
//    } output;
//
//    // The working buffer.
//    struct {
//        // The beginning of the buffer.
//        yaml_char_t *start;
//        // The end of the buffer.
//        yaml_char_t *end;
//        // The current position of the buffer.
//        yaml_char_t *pointer;
//        // The last filled position of the buffer.
//        yaml_char_t *last;
//    } buffer;
//
//    // The raw buffer.
//    struct {
//        // The beginning of the buffer.
//        unsigned char *start;
//        // The end of the buffer.
//        unsigned char *end;
//        // The current position of the buffer.
//        unsigned char *pointer;
//        // The last filled position of the buffer.
//        unsigned char *last;
//    } raw_buffer;
//
//    // The stream encoding.
//    yaml_encoding_t encoding;
//
//    /**
//     * @}
//     */
//
//    /**
//     * @name Emitter stuff
//     * @{
//     */
//
//    // If the output is in the canonical style?
//    int canonical;
//    // The number of indentation spaces.
//    int best_indent;
//    // The preferred width of the output lines.
//    int best_width;
//    // Allow unescaped non-ASCII characters?
//    int unicode;
//    // The preferred line break.
//    yaml_break_t line_break;
//
//    // The stack of states.
//    struct {
//        // The beginning of the stack.
//        yaml_emitter_state_t *start;
//        // The end of the stack.
//        yaml_emitter_state_t *end;
//        // The top of the stack.
//        yaml_emitter_state_t *top;
//    } states;
//
//    // The current emitter state.
//    yaml_emitter_state_t state;
//
//    // The event queue.
//    struct {
//        // The beginning of the event queue.
//        yaml_event_t *start;
//        // The end of the event queue.
//        yaml_event_t *end;
//        // The head of the event queue.
//        yaml_event_t *head;
//        // The tail of the event queue.
//        yaml_event_t *tail;
//    } events;
//
//    // The stack of indentation levels.
//    struct {
//        // The beginning of the stack.
//        int *start;
//        // The end of the stack.
//        int *end;
//        // The top of the stack.
//        int *top;
//    } indents;
//
//    // The list of tag directives.
//    struct {
//        // The beginning of the list.
//        yaml_tag_directive_t *start;
//        // The end of the list.
//        yaml_tag_directive_t *end;
//        // The top of the list.
//        yaml_tag_directive_t *top;
//    } tag_directives;
//
//    // The current indentation level.
//    int indent;
//
//    // The current flow level.
//    int flow_level;
//
//    // Is it the document root context?
//    int root_context;
//    // Is it a sequence context?
//    int sequence_context;
//    // Is it a mapping context?
//    int mapping_context;
//    // Is it a simple mapping key context?
//    int simple_key_context;
//
//    // The current line.
//    int line;
//    // The current column.
//    int column;
//    // If the last character was a whitespace?
//    int whitespace;
//    // If the last character was an indentation character (' ', '-', '?', ':')?
//    int indention;
//    // If an explicit document end is required?
//    int open_ended;
//
//    // Anchor analysis.
//    struct {
//        // The anchor value.
//        yaml_char_t *anchor;
//        // The anchor length.
//        int anchor_length;
//        // Is it an alias?
//        int alias;
//    } anchor_data;
//
//    // Tag analysis.
//    struct {
//        // The tag handle.
//        yaml_char_t *handle;
//        // The tag handle length.
//        int handle_length;
//        // The tag suffix.
//        yaml_char_t *suffix;
//        // The tag suffix length.
//        int suffix_length;
//    } tag_data;
//
//    // Scalar analysis.
//    struct {
//        // The scalar value.
//        yaml_char_t *value;
//        // The scalar length.
//        int length;
//        // Does the scalar contain line breaks?
//        int multiline;
//        // Can the scalar be expessed in the flow plain style?
//        int flow_plain_allowed;
//        // Can the scalar be expressed in the block plain style?
//        int block_plain_allowed;
//        // Can the scalar be expressed in the single quoted style?
//        int single_quoted_allowed;
//        // Can the scalar be expressed in the literal or folded styles?
//        int block_allowed;
//        // The output style.
//        yaml_scalar_style_t style;
//    } scalar_data;
//
//    /**
//     * @}
//     */
//
//    /**
//     * @name Dumper stuff
//     * @{
//     */
//
//    // If the stream was already opened?
//    int opened;
//    // If the stream was already closed?
//    int closed;
//
//    // The information associated with the document nodes.
//    struct {
//        // The number of references.
//        int references;
//        // The anchor id.
//        int anchor;
//        // If the node has been emitted?
//        int serialized;
//    } *anchors;
//
//    // The last assigned anchor id.
//    int last_anchor_id;
//
//    // The currently emitted document.
//    yaml_document_t *document;
//
//    /**
//     * @}
//     */
//
//} yaml_emitter_t;
//
///**
// * Initialize an emitter.
// *
// * This function creates a new emitter object.  An application is responsible
// * for destroying the object using the yaml_emitter_delete() function.
// *
// * @param[out]      emitter     An empty parser object.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_emitter_initialize(yaml_emitter_t *emitter);
//
///**
// * Destroy an emitter.
// *
// * @param[in,out]   emitter     An emitter object.
// */
//
//YAML_DECLARE(void)
//yaml_emitter_delete(yaml_emitter_t *emitter);
//
///**
// * Set a string output.
// *
// * The emitter will write the output characters to the @a output buffer of the
// * size @a size.  The emitter will set @a size_written to the number of written
// * bytes.  If the buffer is smaller than required, the emitter produces the
// * YAML_WRITE_ERROR error.
// *
// * @param[in,out]   emitter         An emitter object.
// * @param[in]       output          An output buffer.
// * @param[in]       size            The buffer size.
// * @param[in]       size_written    The pointer to save the number of written
// *                                  bytes.
// */
//
//YAML_DECLARE(void)
//yaml_emitter_set_output_string(yaml_emitter_t *emitter,
//        unsigned char *output, int size, size_t *size_written);
//
///**
// * Set a file output.
// *
// * @a file should be a file object open for writing.  The application is
// * responsible for closing the @a file.
// *
// * @param[in,out]   emitter     An emitter object.
// * @param[in]       file        An open file.
// */
//
//YAML_DECLARE(void)
//yaml_emitter_set_output_file(yaml_emitter_t *emitter, FILE *file);
//
///**
// * Set a generic output handler.
// *
// * @param[in,out]   emitter     An emitter object.
// * @param[in]       handler     A write handler.
// * @param[in]       data        Any application data for passing to the write
// *                              handler.
// */
//
//YAML_DECLARE(void)
//yaml_emitter_set_output(yaml_emitter_t *emitter,
//        yaml_write_handler_t *handler, void *data);
//
///**
// * Set the output encoding.
// *
// * @param[in,out]   emitter     An emitter object.
// * @param[in]       encoding    The output encoding.
// */
//
//YAML_DECLARE(void)
//yaml_emitter_set_encoding(yaml_emitter_t *emitter, yaml_encoding_t encoding);
//
///**
// * Set if the output should be in the "canonical" format as in the YAML
// * specification.
// *
// * @param[in,out]   emitter     An emitter object.
// * @param[in]       canonical   If the output is canonical.
// */
//
//YAML_DECLARE(void)
//yaml_emitter_set_canonical(yaml_emitter_t *emitter, int canonical);
//
///**
// * Set the intendation increment.
// *
// * @param[in,out]   emitter     An emitter object.
// * @param[in]       indent      The indentation increment (1 < . < 10).
// */
//
//YAML_DECLARE(void)
//yaml_emitter_set_indent(yaml_emitter_t *emitter, int indent);
//
///**
// * Set the preferred line width. @c -1 means unlimited.
// *
// * @param[in,out]   emitter     An emitter object.
// * @param[in]       width       The preferred line width.
// */
//
//YAML_DECLARE(void)
//yaml_emitter_set_width(yaml_emitter_t *emitter, int width);
//
///**
// * Set if unescaped non-ASCII characters are allowed.
// *
// * @param[in,out]   emitter     An emitter object.
// * @param[in]       unicode     If unescaped Unicode characters are allowed.
// */
//
//YAML_DECLARE(void)
//yaml_emitter_set_unicode(yaml_emitter_t *emitter, int unicode);
//
///**
// * Set the preferred line break.
// *
// * @param[in,out]   emitter     An emitter object.
// * @param[in]       line_break  The preferred line break.
// */
//
//YAML_DECLARE(void)
//yaml_emitter_set_break(yaml_emitter_t *emitter, yaml_break_t line_break);
//
///**
// * Emit an event.
// *
// * The event object may be generated using the yaml_parser_parse() function.
// * The emitter takes the responsibility for the event object and destroys its
// * content after it is emitted. The event object is destroyed even if the
// * function fails.
// *
// * @param[in,out]   emitter     An emitter object.
// * @param[in,out]   event       An event object.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_emitter_emit(yaml_emitter_t *emitter, yaml_event_t *event);
//
///**
// * Start a YAML stream.
// *
// * This function should be used before yaml_emitter_dump() is called.
// *
// * @param[in,out]   emitter     An emitter object.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_emitter_open(yaml_emitter_t *emitter);
//
///**
// * Finish a YAML stream.
// *
// * This function should be used after yaml_emitter_dump() is called.
// *
// * @param[in,out]   emitter     An emitter object.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_emitter_close(yaml_emitter_t *emitter);
//
///**
// * Emit a YAML document.
// *
// * The documen object may be generated using the yaml_parser_load() function
// * or the yaml_document_initialize() function.  The emitter takes the
// * responsibility for the document object and destoys its content after
// * it is emitted. The document object is destroyedeven if the function fails.
// *
// * @param[in,out]   emitter     An emitter object.
// * @param[in,out]   document    A document object.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_emitter_dump(yaml_emitter_t *emitter, yaml_document_t *document);
//
///**
// * Flush the accumulated characters to the output.
// *
// * @param[in,out]   emitter     An emitter object.
// *
// * @returns @c 1 if the function succeeded, @c 0 on error.
// */
//
//YAML_DECLARE(int)
//yaml_emitter_flush(yaml_emitter_t *emitter);
//
//// @}
//
//#ifdef __cplusplus
//}
//#endif
//
//#endif /* #ifndef YAML_H */
//
