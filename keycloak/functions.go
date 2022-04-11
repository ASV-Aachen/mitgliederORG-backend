package keycloak

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// keycloak
var hostname string = os.Getenv("keycloak_hostname")
var clientID string = os.Getenv("keycloak_clientID")
var clientSecret string = os.Getenv("keycloak_clientSecret")
var realm string = os.Getenv("keycloak_realm")

func Get_AdminToken() (string, error) {

	url := "http://" + hostname + "/sso/auth/realms/" + realm + "/protocol/openid-connect/token"

	payload := strings.NewReader("client_id=" + clientID + "&grant_type=client_credentials&client_secret=" + clientSecret)

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Default().Printf(err.Error())
		log.Default().Printf("client_id=" + clientID + "&grant_type=client_credentials&client_secret=" + clientSecret)
		return "", errors.New(err.Error())
	}

	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	answer := AdminToken{}
	json.Unmarshal(body, &answer)

	return answer.AccessToken, nil
}

func Get_UserID(token string) (string, error) {
	path := "http://" + hostname + "/sso/auth/realms/" + realm + "/protocol/openid-connect/userinfo"

	// payload := strings.NewReader("client_id=backend-check&grant_type=client_credentials")

	req, _ := http.NewRequest("GET", path, nil)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Default().Printf(err.Error())
		return "", errors.New(err.Error())
	}
	defer resp.Body.Close()

	if resp.Status != "200" {
		log.Default().Printf(resp.Status)
		log.Default().Printf(path)
		log.Default().Printf("[" + token + "]")

		return "", errors.New("unathorized")
	}

	body, _ := ioutil.ReadAll(resp.Body)

	answer := UserInfo{}
	json.Unmarshal(body, &answer)

	return answer.Sub, nil
}

func Get_UserGroups(token string, ID string) (GroupToken, error) {

	url := "http://" + hostname + "/sso/auth/admin/realms/" + realm + "/users/ " + ID + "/groups"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Default().Printf(err.Error())
		return nil, errors.New(err.Error())
	}

	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	answer := GroupToken{}
	json.Unmarshal(body, &answer)

	return answer, nil
}

func Check_IsUserPartOfGroup(gruppen [5]string, UserGruppen GroupToken) bool {
	for _, groupName := range gruppen{
		for _, token := range UserGruppen {
			if groupName == token.Name {
				return true
			}
		}
	}
	return false
}
