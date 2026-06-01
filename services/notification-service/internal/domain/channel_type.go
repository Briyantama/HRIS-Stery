package domain

import "fmt"

type ChannelType int

const (
	ChannelTypeUnknown ChannelType = iota
	ChannelTypeEmail
	ChannelTypeInApp
)

func (ct ChannelType) String() string {
	switch ct {
	case ChannelTypeEmail:
		return "EMAIL"
	case ChannelTypeInApp:
		return "IN_APP"
	default:
		return "UNKNOWN"
	}
}

func ChannelTypeFromString(s string) (ChannelType, error) {
	switch s {
	case "EMAIL":
		return ChannelTypeEmail, nil
	case "IN_APP":
		return ChannelTypeInApp, nil
	default:
		return ChannelTypeUnknown, fmt.Errorf("invalid channel type: %s", s)
	}
}

func (ct ChannelType) IsValid() bool {
	return ct == ChannelTypeEmail || ct == ChannelTypeInApp
}
