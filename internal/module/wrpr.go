package module

import (
	"os"
	"strings"
)

func RegX() *GETl {
	var hideBannerV = os.Getenv("GOBE_HIDE_BANNER")
	if hideBannerV == "" {
		hideBannerV = "true"
	}

	return &GETl{
		hideBannerV: strings.ToLower(hideBannerV) == "true",
	}
}
