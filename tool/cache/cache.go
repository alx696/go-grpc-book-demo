package cache

var userTokenMap = make(map[string]string)
var tokenUserMap = make(map[string]string)

func init() {
	// TODO 仅供演示
	userTokenMap["测试"] = "d6449e41-e039-4458-8ab4-b47516aeacb1"
	tokenUserMap["d6449e41-e039-4458-8ab4-b47516aeacb1"] = "测试"
}

func SetUserToken(user_id, token string) {
	userTokenMap[user_id] = token
	tokenUserMap[token] = user_id
}

func GetUserToken(user_id string) string {
	return userTokenMap[user_id]
}

func GetTokenUser(token string) string {
	return tokenUserMap[token]
}

func DelUserToken(user_id string) {
	if token, exists := userTokenMap[user_id]; exists {
		delete(tokenUserMap, token)
	}

	delete(userTokenMap, user_id)
}
