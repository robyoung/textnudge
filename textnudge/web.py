import os
from flask import Flask, request
from twilio.rest import TwilioRestClient
from twilio import twiml

app = Flask(__name__)
app.config['PERSON_ONE'] = os.getenv('PERSON_ONE')
app.config['PERSON_TWO'] = os.getenv('PERSON_TWO')
client = TwilioRestClient()


@app.route('/')
def hello():
    return "Hello world!"


@app.route('/receive', methods=['POST'])
def receive():
    message = client.messages.create(
        to=get_forward_number(request.form['From']),
        from_=request.form['To'],
        body=request.form['Body'])
    print("{}".format(message))

    r = twiml.Response()
    return str(r)


def get_forward_number(from_):
    if from_ == app.config['PERSON_ONE']:
        return app.config['PERSON_TWO']
    elif from_ == app.config['PERSON_TWO']:
        return app.config['PERSON_ONE']
    else:
        raise ValueError("Invalid from address {}".format(from_))

if __name__ == '__main__':
    app.debug = True
    app.run()
