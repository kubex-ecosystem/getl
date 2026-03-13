// Package cli fornece comandos e funcionalidades para a interface de linha de comando.
package cli

import (
	"math/rand"
	"os"
	"strings"
)

var banners = []string{
	`
 /      \|        \        \  \
|  ▓▓▓▓▓▓\ ▓▓▓▓▓▓▓▓\▓▓▓▓▓▓▓▓ ▓▓
| ▓▓ __\▓▓ ▓▓__      | ▓▓  | ▓▓
| ▓▓|    \ ▓▓  \     | ▓▓  | ▓▓
| ▓▓ \▓▓▓▓ ▓▓▓▓▓     | ▓▓  | ▓▓
| ▓▓__| ▓▓ ▓▓_____   | ▓▓  | ▓▓_____
 \▓▓    ▓▓ ▓▓     \  | ▓▓  | ▓▓     \
  \▓▓▓▓▓▓ \▓▓▓▓▓▓▓▓   \▓▓   \▓▓▓▓▓▓▓▓
`,
}

func GetDescriptions(descriptionArg []string, hideBanner bool) map[string]string {
	var description, banner string

	if descriptionArg != nil {
		if strings.Contains(strings.Join(os.Args[0:], ""), "-h") {
			description = descriptionArg[0]
		} else {
			description = descriptionArg[1]
		}
	} else {
		description = ""
	}

	bannerRandLen := len(banners)
	bannerRandIndex := rand.Intn(bannerRandLen)
	banner = banners[bannerRandIndex]

	if hideBanner {
		banner = ""
	}

	return map[string]string{"banner": banner, "description": description}
}
