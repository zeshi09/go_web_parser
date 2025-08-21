package verification

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

// структура ответа VK API
type VKResponse struct {
	Response []struct {
		ID        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"response"`
	Error *struct {
		ErrorCode int    `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
	} `json:"error"`
}

func checkVKUser(username, token string) (bool, string, error) {
	apiURL := "https://api.vk.com/method/users.get"

	params := url.Values{}
	params.Set("user_ids", username)
	params.Set("access_token", token)
	params.Set("v", "5.199") // актуальная версия API

	resp, err := http.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var vkResp VKResponse
	if err := json.Unmarshal(body, &vkResp); err != nil {
		return false, "", err
	}

	if vkResp.Error != nil {
		return false, vkResp.Error.ErrorMsg, nil
	}

	if len(vkResp.Response) > 0 {
		user := vkResp.Response[0]
		return true, fmt.Sprintf("%s %s (id%d)", user.FirstName, user.LastName, user.ID), nil
	}

	return false, "unknown error", nil
}

func main() {
	// username = screen name или id
	username := "durov"
	// токен можно получить через VK Standalone App или личный access_token
	token := os.Getenv("VK_TOKEN")

	exists, info, err := checkVKUser(username, token)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	if exists {
		fmt.Println("Пользователь существует:", info)
	} else {
		fmt.Println("Пользователь не найден:", info)
	}
}

