package praga

/*
 * This file exists to embed various build results from other tools to the Go server
 */

import "embed"

// "_" prefixes are ignored by default so need both of the below

//go:embed frontend/build
//go:embed frontend/build/_app
var EmbeddedFrontendBuild embed.FS

//go:embed email/verification.html
var VerificationEmail embed.FS
