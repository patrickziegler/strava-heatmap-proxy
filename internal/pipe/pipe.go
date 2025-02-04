package pipe

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type CookieClient interface {
	GetCloudFrontTokens() map[string]string
}

func ReplaceCloudFrontTokens(client CookieClient) {
	cookies := client.GetCloudFrontTokens()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Replace(line, "%CloudFront-Key-Pair-Id%", cookies["CloudFront-Key-Pair-Id"], -1)
		line = strings.Replace(line, "%CloudFront-Policy%", cookies["CloudFront-Policy"], -1)
		line = strings.Replace(line, "%CloudFront-Signature%", cookies["CloudFront-Signature"], -1)
		fmt.Println(line)
	}
}
