package events

// TopicFormat is the format of strings used as Event Topics.
const TopicFormat = `^[-_[:alnum:]]+$`

// TopicReplacement is the format used to replace non-well-formed Topic strings.
const TopicReplacement = `[^_[:alnum:]]+`

// TopicEmpty is the replacement string for empty topics
const TopicEmpty = "-empty-"

// Topic is the type used for event labeling.
//
// Unlike vanilla strings, Topic instances should match the TopicFormat regexp,
// for debugging convenience.
type Topic string
