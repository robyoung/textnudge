package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"bitbucket.org/ckvist/twilio/twiml"
	"bitbucket.org/ckvist/twilio/twirest"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/soveran/redisurl"
)

type Config struct {
	PORT               string
	PERSON_ONE         string
	PERSON_TWO         string
	REDISCLOUD_URL     string
	TWILIO_ACCOUNT_SID string
	TWILIO_AUTH_TOKEN  string

	twilio_client *twirest.TwilioClient
	redis_client  redis.Conn
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world")
}

func getForm(r *http.Request, name string) (string, error) {
	if len(r.PostForm[name]) > 0 {
		return r.PostForm[name][0], nil
	} else if len(r.Form[name]) > 0 {
		return r.Form[name][0], nil
	} else {
		return "", fmt.Errorf("%s not found in request (URL: %v, POST: %v)", name, r.Form, r.PostForm)
	}
}

func ReceiveHandler(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Print(err)
			http.Error(w, "Invalid request", 400)
		}
		from_number, err := getForm(r, "From")
		if err != nil {
			log.Print(err)
			http.Error(w, "No from number", 400)
			return
		}
		to_number, err := getForwardNumber(config, from_number)
		if err != nil {
			log.Print(err)
			http.Error(w, "Invalid from number", 400)
			return
		}
		twilio_number, err := getForm(r, "To")
		if err != nil {
			log.Print(err)
			http.Error(w, "No to number", 400)
			return
		}
		body, err := getForm(r, "Body")
		if err != nil {
			log.Print(err)
			http.Error(w, "No message body", 400)
			return
		}

		err = sendMessage(config, twilio_number, to_number, body)
		if err != nil {
			log.Print(err)
			http.Error(w, "Failed to send message", 500)
			return
		}

		_, err = config.redis_client.Do("LPUSH", fmt.Sprintf("textnudge.unreplied.%s", to_number), time.Now().Format(time.RFC3339))
		if err != nil {
			log.Print(err)
			http.Error(w, "Failed to record from number", 500)
			return
		}
		length, err := redis.Int(config.redis_client.Do("LLEN", fmt.Sprintf("textnudge.unreplied.%s", to_number)))
		if err != nil {
			log.Print(err)
		}
		log.Printf("Length is %s", length)

		_, err = config.redis_client.Do("DEL", fmt.Sprintf("textnudge.unreplied.%s", from_number))
		if err != nil {
			log.Print(err)
			http.Error(w, "Failed to reset to number", 500)
			return
		}

		time.AfterFunc(15*time.Second, nudge(config, twilio_number, to_number))

		resp := twiml.NewResponse()
		resp.Send(w)
	}
}

func sendMessage(config Config, from string, to string, message string) error {
	log.Printf("Sending message from %s to %s", from, to)
	msg := twirest.SendMessage{
		To:   to,
		From: from,
		Text: message,
	}
	_, err := config.twilio_client.Request(msg)
	return err
}

func getForwardNumber(config Config, from string) (string, error) {
	if from == config.PERSON_ONE {
		return config.PERSON_TWO, nil
	} else if from == config.PERSON_TWO {
		return config.PERSON_ONE, nil
	} else {
		return "", fmt.Errorf("Unknown number %s", from)
	}
}

func nudge(config Config, twilio_number string, to_number string) func() {
	log.Printf("Starting nudge to %s", to_number)
	return func() {
		log.Printf("Firing nudge to %s", to_number)
		length, err := redis.Int(config.redis_client.Do("LLEN", fmt.Sprintf("textnudge.unreplied.%s", to_number)))
		if err != nil {
			log.Print(err)
			return
		}
		if length > 0 {
			message := fmt.Sprintf("You have %d unreplied message.")
			err := sendMessage(config, twilio_number, to_number, message)
			if err != nil {
				log.Print(err)
			}
			time.AfterFunc(5*time.Minute, nudge(config, to_number, twilio_number))
		}
	}
}

func main() {
	config := Config{
		PORT:               os.Getenv("PORT"),
		PERSON_ONE:         os.Getenv("PERSON_ONE"),
		PERSON_TWO:         os.Getenv("PERSON_TWO"),
		REDISCLOUD_URL:     os.Getenv("REDISCLOUD_URL"),
		TWILIO_ACCOUNT_SID: os.Getenv("TWILIO_ACCOUNT_SID"),
		TWILIO_AUTH_TOKEN:  os.Getenv("TWILIO_AUTH_TOKEN"),
	}
	var err error

	config.twilio_client = twirest.NewClient(config.TWILIO_ACCOUNT_SID, config.TWILIO_AUTH_TOKEN)
	_, err = config.twilio_client.Request(twirest.Accounts{})
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to connect to twilio: %s", err))
	}
	config.redis_client, err = redisurl.ConnectToURL(config.REDISCLOUD_URL)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to connect to redis: %s", err))
	}

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/receive", ReceiveHandler(config))

	log.Printf("Listening on :%s", config.PORT)
	http.ListenAndServe(":"+config.PORT, r)
}
