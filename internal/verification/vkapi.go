package verification

import (
	"fmt"
	"log"

	"github.com/SevereCloud/vksdk/v2/api"
)

func CheckVKApi(url string) bool {
	token := "<TOKEN>" // рекомендуется использовать os.Getenv("TOKEN")
	vk := api.NewVK(token)

	// Получаем информацию о группе
	group, err := vk.GroupsGetByID(api.Params{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(group)

	return true
}
