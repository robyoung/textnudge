from flask import Flask
from twilio.rest import TwilioRestClient
from twilio import twiml

app = Flask(__name__)
client = TwilioRestClient()


@app.route('/')
def hello():
    return "Hello world!"


@app.route('/receive', methods=['POST'])
def receive():
    r = twiml.Response()
    r.message("You are the bestest!")
    return str(r)
