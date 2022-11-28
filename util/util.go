package util

import "strconv"

func ParseUrl(path string) (int, int, error) {
	i, s := 0, 0
	userIdStr := ""
	for j, b := range path {
		if b == '/' {
			i++
			if i == 3 {
				userIdStr = path[s+1 : j]
			}
			s = j
		}
	}
	toCollectEnergyId, _ := strconv.Atoi(path[s+1:])
	userId, _ := strconv.Atoi(userIdStr)
	return userId, toCollectEnergyId, nil
}
