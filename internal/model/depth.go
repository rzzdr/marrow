package model

type Depth string

const (
	DepthSummary  Depth = "summary"  // one-liner per item
	DepthStandard Depth = "standard" // key fields, no reasoning text
	DepthFull     Depth = "full"     // everything
)

func ParseDepth(s string) Depth {
	switch s {
	case "standard":
		return DepthStandard
	case "full":
		return DepthFull
	default:
		return DepthSummary
	}
}
