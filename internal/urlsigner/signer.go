package urlsigner

import (
	"fmt"
	"log"
	"strings"
	"time"

	goalone "github.com/bwmarrin/go-alone"
)

// create type signer to hold secret object
type Signer struct {
	Secret []byte
}

// create function to generate new token to url
func (s *Signer) CreateNewToken(url string) string {
	// create variable to hold url with token
	var urlToken string

	// create noalone object
	crypt := goalone.New(s.Secret, goalone.Timestamp)

	// check for url, is it contain params or not
	containParams := strings.Contains(url, "?")

	if containParams {
		//if url contains params
		urlToken = fmt.Sprintf("%s&hash=", url)
	} else {
		// if url is not contains params
		urlToken = fmt.Sprintf("%s?hash=", url)
	}

	// assign token to url
	tokenAssign := crypt.Sign([]byte(urlToken))

	// convert token to string
	getToken := string(tokenAssign)

	// return value
	return getToken
}

// create function to check if token valid or not
func (s *Signer) ValidateToken(token string) (bool, error) {
	// create noalone object
	crypt := goalone.New(s.Secret, goalone.Timestamp)

	var result = false

	// check token signature
	_, err := crypt.Unsign([]byte(token))

	// check for an error
	if err != nil {
		log.Println("token is not valid : ", err)
		return result, err
	} else {
		log.Println("token is valid...")
		result = true
	}

	return result, nil
}

// create function to check if duration valid
func (s *Signer) ValidateDuration(token string, times int) bool {
	// create noalone object
	crypt := goalone.New(s.Secret, goalone.Timestamp)

	// get time stamp from crypt
	ts := crypt.Parse([]byte(token))

	// get time stamp from token
	// akan mereturn waktu yang sudah berjalan sejak timestamp diambil hingga saat ini
	// misalnya token pertama kali dijalankan jam 16.30 maka saat sekaran 16.40 maka
	// variable timeStamp dibawah akan bernilai 10 menit
	timeStamp := time.Since(ts.Timestamp)

	// get time duration as minute
	// akan mereturn waktu secara eksplisit seperti 10 menit
	// karean akan dilakukan pengecekan apakah waktu yang sudah berlalau lebih dari berapa menit
	// yang sudah di set saat token pertama kali diassign
	timeDur := time.Duration(times) * time.Minute

	// compare
	compare := timeStamp < timeDur

	return compare
}
