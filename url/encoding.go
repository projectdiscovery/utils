package urlutil

import (
	"github.com/projectdiscovery/utils/env"
)

// SpaceEncoding determines how spaces are encoded in URLs via external environment variable:
// - When empty (""), spaces are encoded as "+"
// - When set to "percent", spaces are encoded as "%20"
var SpaceEncoding = env.GetEnvOrDefault("SPACE_ENCODING", "")
