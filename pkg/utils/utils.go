package utils

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/accuknox/kubearmor/pkg/models"
	"github.com/kubearmor/KubeArmor/KubeArmor/types"
)

func ConvertToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		// Handle conversion error gracefully, e.g., log error and return default value
		fmt.Printf("Error converting %s to int: %v\n", s, err)
		return 0 // Default value (or handle as appropriate)
	}
	return i
}

func Random(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func Extractdata(body string) map[string]string {

	pairs := strings.Split(body, " ")

	// Initialize a map to store extracted values
	dataMap := make(map[string]string)

	// Loop through each key-value pair
	for _, pair := range pairs {
		// Split each pair by '=' to separate key and value
		parts := strings.Split(pair, "=")
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			dataMap[key] = value
		}
	}
	return dataMap
}

func ResolveIp(remoteIP string) string {
	return ""
}

func IsCorrectLog(log types.Log, qm models.QueryModel) bool {
	if log.Operation != qm.Operation {
		return false
	}

	if log.NamespaceName != "All" && log.NamespaceName != qm.NamespaceQuery {
		return false
	}

	if log.Labels != "All" && log.Labels != qm.LabelQuery {
		return false
	}

	return true
}
